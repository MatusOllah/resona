package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/codec"
	"github.com/MatusOllah/resona/codec/wav/internal/riff"
	"github.com/MatusOllah/resona/freq"
)

// Chunk IDs for the WAVE file format.
var (
	WaveID riff.FourCC = riff.FourCC{'W', 'A', 'V', 'E'}
	FmtID  riff.FourCC = riff.FourCC{'f', 'm', 't', ' '}
	ListID riff.FourCC = riff.FourCC{'L', 'I', 'S', 'T'}
	DataID riff.FourCC = riff.FourCC{'d', 'a', 't', 'a'}
)

// WAVE formats.
const (
	formatInt   uint16 = 1 // PCM Integer
	formatFloat uint16 = 3 // IEEE Float
)

const magic string = "RIFF????WAVE"

// Decoder represents the decoder for the WAVE file format.
// It implements codec.Decoder.
type Decoder struct {
	riffR *riff.Reader

	audioFormat   uint16
	numChannels   uint16
	sampleRate    uint32
	bytesPerSec   uint32
	bytesPerBlock uint16
	bitsPerSample uint16

	dataChunk *riff.Chunk
	dataRead  int

	pcmBuf []byte
}

// NewDecoder creates a new [Decoder] and decodes the headers.
func NewDecoder(r io.Reader) (_ codec.Decoder, err error) {
	d := &Decoder{}

	var id riff.FourCC
	id, d.riffR, err = riff.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RIFF stream: %w", err)
	}

	if !bytes.Equal(id[:], WaveID[:]) {
		return nil, fmt.Errorf("invalid WAVE header")
	}

	if err := d.parseFmt(); err != nil {
		return nil, fmt.Errorf("failed to parse fmt chunk: %w", err)
	}

	for {
		chunk, err := d.riffR.NextChunk()
		if err != nil {
			return nil, err
		}

		switch {
		case bytes.Equal(chunk.ID[:], DataID[:]):
			d.dataChunk = chunk
			return d, nil // success
		default:
			// Skip unknown chunk
			_, _ = io.Copy(io.Discard, chunk.Reader)
		}
	}
}

// parseFmt reads and parses the "fmt " chunk.
func (d *Decoder) parseFmt() error {
	chunk, err := d.riffR.NextChunk()
	if err != nil {
		return err
	}

	if !bytes.Equal(chunk.ID[:], FmtID[:]) {
		return fmt.Errorf("invalid fmt chunk")
	}

	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.audioFormat); err != nil {
		return fmt.Errorf("failed to read audio format: %w", err)
	}
	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.numChannels); err != nil {
		return fmt.Errorf("failed to read number of channels: %w", err)
	}
	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.sampleRate); err != nil {
		return fmt.Errorf("failed to read sample rate: %w", err)
	}
	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.bytesPerSec); err != nil {
		return fmt.Errorf("failed to read bytes pre second: %w", err)
	}
	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.bytesPerBlock); err != nil {
		return fmt.Errorf("failed to read bytes per block: %w", err)
	}
	if err := binary.Read(chunk.Reader, binary.LittleEndian, &d.bitsPerSample); err != nil {
		return fmt.Errorf("failed to read bits per sample: %w", err)
	}

	_, _ = io.Copy(io.Discard, chunk.Reader)

	return nil
}

// Format returns the audio stream format.
func (d *Decoder) Format() afmt.Format {
	return afmt.Format{
		SampleRate:  freq.Frequency(d.sampleRate) * freq.Hertz,
		NumChannels: int(d.numChannels),
	}
}

// SampleFormat returns the sample format.
func (d *Decoder) SampleFormat() afmt.SampleFormat {
	var enc afmt.SampleEncoding
	switch d.audioFormat {
	case formatInt:
		enc = afmt.SampleEncodingInt
	case formatFloat:
		enc = afmt.SampleEncodingFloat
	}
	if d.bitsPerSample == 8 {
		enc = afmt.SampleEncodingUint // 8-bit is always unsigned
	}

	return afmt.SampleFormat{
		BitDepth: int(d.bitsPerSample),
		Encoding: enc,
		Endian:   binary.LittleEndian,
	}
}

// ReadSamples reads float64 samples from the data chunk into p.
// It returns the number of samples read and/or an error.
func (d *Decoder) ReadSamples(p []float64) (n int, err error) {
	if d.dataRead >= d.dataChunk.Len {
		return 0, io.EOF
	}

	numChannels := int(d.numChannels)
	sampleSize := int(d.bitsPerSample / 8)
	frameSize := sampleSize * numChannels

	numFrames := len(p) / numChannels // Number of frames we can store in p
	numBytes := min(numFrames*frameSize, d.dataChunk.Len-d.dataRead)

	if len(d.pcmBuf) != numBytes {
		d.pcmBuf = make([]byte, numBytes)
	}
	readBytes, err := d.dataChunk.Reader.Read(d.pcmBuf)
	if err != nil {
		return 0, err
	}

	readFrames := readBytes / frameSize
	actualSamples := readFrames * numChannels

	for frame := range readFrames {
		for ch := range numChannels {
			offset := frame*frameSize + ch*sampleSize
			idx := frame*numChannels + ch

			if offset+sampleSize > len(d.pcmBuf) {
				return idx, io.ErrUnexpectedEOF
			}

			switch d.bitsPerSample {
			case 8:
				s := d.pcmBuf[offset]
				p[idx] = (float64(s) - 128.0) / 128.0

			case 16:
				s := int16(binary.LittleEndian.Uint16(d.pcmBuf[offset:]))
				p[idx] = float64(s) / 32767.0

			case 24:
				if offset+3 > len(d.pcmBuf) {
					return idx, io.ErrUnexpectedEOF
				}
				b := d.pcmBuf[offset : offset+3]
				s := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16
				if s&(1<<23) != 0 {
					s |= ^0xFFFFFF
				}
				p[idx] = float64(s) / 8388608.0

			case 32:
				switch d.audioFormat {
				case formatInt:
					s := int32(binary.LittleEndian.Uint32(d.pcmBuf[offset:]))
					p[idx] = float64(s) / 2147483647.0
				case formatFloat:
					bits := binary.LittleEndian.Uint32(d.pcmBuf[offset:])
					p[idx] = float64(math.Float32frombits(bits))
				}

			default:
				return idx, fmt.Errorf("unsupported bit depth: %d", d.bitsPerSample)
			}
		}
	}

	d.dataRead += readBytes
	return actualSamples, nil
}

// Len returns the total number of frames.
func (d *Decoder) Len() int {
	frameSize := int(d.numChannels) * int(d.bitsPerSample/8)
	if frameSize == 0 {
		return 0
	}
	return d.dataChunk.Len / frameSize
}

// Seek seeks to the specified frame.
// Seek offset is measured in frames, where one frame contains one sample per channel.
// It returns the new offset relative to the start and/or an error.
// It will return an error if the source is not an [io.Seeker].
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	frameSize := int64(d.numChannels) * int64(d.bitsPerSample/8)
	totalFrames := int64(d.dataChunk.Len) / frameSize

	var targetFrame int64
	switch whence {
	case io.SeekStart:
		targetFrame = offset
	case io.SeekCurrent:
		targetFrame = int64(d.dataRead)/frameSize + offset
	case io.SeekEnd:
		targetFrame = totalFrames + offset
	default:
		return 0, fmt.Errorf("wav: invalid seek whence")
	}

	if targetFrame < 0 || targetFrame > totalFrames {
		return 0, fmt.Errorf("wav: seek out of bounds")
	}

	byteOffset := targetFrame * frameSize

	_, err := d.dataChunk.Reader.Seek(byteOffset, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("wav: failed to seek: %w", err)
	}

	d.dataRead = int(byteOffset)
	return targetFrame, nil
}

func init() {
	codec.RegisterFormat("wav", magic, NewDecoder)
}
