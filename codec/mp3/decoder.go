package mp3

import (
	"encoding/binary"
	"io"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/codec"
	"github.com/MatusOllah/resona/freq"
	"github.com/hajimehoshi/go-mp3"
)

const frameSize = 4 // int16 le stereo = 4 bytes

// Decoder represents the decoder for the MP3 file format.
// It implements codec.Decoder.
type Decoder struct {
	d      *mp3.Decoder
	pcmBuf []byte
}

// NewDecoder creates a new [Decoder] and decodes the headers.
func NewDecoder(r io.Reader) (_ codec.Decoder, err error) {
	var d *mp3.Decoder
	if rs, ok := r.(io.ReadSeeker); ok {
		d, err = mp3.NewDecoder(rs)
		if err != nil {
			return nil, err
		}
	} else {
		d, err = mp3.NewDecoder(r)
		if err != nil {
			return nil, err
		}
	}

	return &Decoder{d: d}, nil
}

// Format returns the audio stream format.
func (d *Decoder) Format() afmt.Format {
	return afmt.Format{
		SampleRate:  freq.Frequency(d.d.SampleRate()) * freq.Hertz,
		NumChannels: 2,
	}
}

// SampleFormat returns the sample format.
func (d *Decoder) SampleFormat() afmt.SampleFormat {
	return afmt.SampleFormat{
		BitDepth: 16,
		Encoding: afmt.SampleEncodingInt,
		Endian:   binary.LittleEndian,
	}
}

// ReadSamples reads float64 samples into p.
// It returns the number of samples read and/or an error.
func (d *Decoder) ReadSamples(p []float64) (int, error) {
	numFrames := len(p) / 2
	numBytes := numFrames * frameSize

	if cap(d.pcmBuf) < numBytes {
		d.pcmBuf = make([]byte, numBytes)
	}
	buf := d.pcmBuf[:numBytes]

	n, err := d.d.Read(buf)
	if n == 0 {
		return 0, err
	}

	readFrames := n / frameSize
	for frame := range readFrames {
		for ch := range 2 {
			offset := frame*frameSize + ch*2
			s := int16(binary.LittleEndian.Uint16(buf[offset:]))
			p[frame*2+ch] = float64(s) / 32767.0
		}
	}
	return readFrames * 2, err
}

// Len returns the total number of frames.
func (d *Decoder) Len() int {
	return int(d.d.Length()) / frameSize
}

// Seek seeks to the specified frame.
// Seek offset is measured in frames, where one frame contains one sample per channel.
// It returns the new offset relative to the start and/or an error.
// It will return an error if the source is not an [io.Seeker].
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	pos, err := d.d.Seek(offset*frameSize, whence)
	if err != nil {
		return 0, err
	}

	return pos / frameSize, nil
}

func init() {
	// Without ID3v2
	codec.RegisterFormat("mp3", string([]byte{0xFF, 0xFB}), NewDecoder)
	codec.RegisterFormat("mp3", string([]byte{0xFF, 0xF3}), NewDecoder)
	codec.RegisterFormat("mp3", string([]byte{0xFF, 0xF2}), NewDecoder)

	// With ID3v2
	codec.RegisterFormat("mp3", string([]byte{0x49, 0x44, 0x33}), NewDecoder)
}
