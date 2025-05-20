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

var (
	WaveID riff.FourCC = riff.FourCC{'W', 'A', 'V', 'E'}
	FmtID  riff.FourCC = riff.FourCC{'f', 'm', 't', ' '}
	ListID riff.FourCC = riff.FourCC{'L', 'I', 'S', 'T'}
	DataID riff.FourCC = riff.FourCC{'d', 'a', 't', 'a'}
)

const (
	formatInt   uint16 = 1
	formatFloat uint16 = 3
)

const magic string = "RIFF????WAVE"

type Decoder struct {
	riffR *riff.Reader

	audioFormat   uint16
	numChannels   uint16
	sampleRate    uint32
	bytesPerSec   uint32
	bytesPerBlock uint16
	bitsPerSample uint16

	listChunk *riff.Chunk

	dataChunk *riff.Chunk
	dataRead  int
}

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

	chunk, err := d.riffR.NextChunk()
	if err != nil {
		return nil, err
	}
	switch {
	case bytes.Equal(chunk.ID[:], ListID[:]):
		d.listChunk = chunk
		// TODO: parse metadata???
		d.dataChunk, err = d.riffR.NextChunk()
		if err != nil {
			return nil, err
		}
	case bytes.Equal(chunk.ID[:], DataID[:]):
		d.dataChunk = chunk
	}
	if d.dataChunk == nil {
		return nil, fmt.Errorf("missing data chunk")
	}

	return d, err
}

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

	return nil
}

func (d *Decoder) Format() afmt.Format {
	return afmt.Format{
		SampleRate:  afmt.SampleRate(freq.Frequency(d.sampleRate) * freq.Hertz),
		NumChannels: int(d.numChannels),
	}
}

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

func (d *Decoder) ReadSamples(p []float64) (n int, err error) {
	if d.dataRead >= d.dataChunk.Len {
		return 0, io.EOF
	}

	numChannels := int(d.numChannels)
	sampleSize := int(d.bitsPerSample / 8)
	frameSize := sampleSize * numChannels

	numFrames := len(p) / numChannels
	numBytes := min(numFrames*frameSize, d.dataChunk.Len-d.dataRead)

	buf := make([]byte, numBytes)
	readBytes, err := d.dataChunk.Reader.Read(buf)
	if err != nil && err != io.EOF {
		return 0, err
	}

	readFrames := readBytes / frameSize

	for frame := range readFrames {
		for ch := range numChannels {
			offset := frame*frameSize + ch*sampleSize

			switch d.bitsPerSample {
			case 8:
				// Unsigned 8-bit
				val := buf[offset]
				p[frame*numChannels+ch] = (float64(val) - 128.0) / 128.0

			case 16:
				val := int16(binary.LittleEndian.Uint16(buf[offset:]))
				p[frame*numChannels+ch] = float64(val) / 32768.0

			case 24:
				b := buf[offset : offset+3]
				val := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16
				if val&(1<<23) != 0 {
					val |= ^0xFFFFFF
				}
				p[frame*numChannels+ch] = float64(val) / 8388608.0

			case 32:
				if d.audioFormat == formatInt {
					val := int32(binary.LittleEndian.Uint32(buf[offset:]))
					p[frame*numChannels+ch] = float64(val) / 2147483648.0
				} else if d.audioFormat == formatFloat {
					bits := binary.LittleEndian.Uint32(buf[offset:])
					p[frame*numChannels+ch] = float64(math.Float32frombits(bits))
				}

			default:
				return 0, fmt.Errorf("unsupported bit depth: %d", d.bitsPerSample)
			}
		}
	}

	d.dataRead += readBytes
	return readFrames, nil
}

func (d *Decoder) Len() int {
	return d.dataChunk.Len
}

func init() {
	codec.RegisterFormat("wav", magic, NewDecoder)
}
