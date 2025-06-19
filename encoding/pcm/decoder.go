package pcm

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
)

var ErrInvalidSampleEncoding error = errors.New("pcm: invalid sample encoding")
var ErrInvalidBitDepth error = errors.New("pcm: invalid bit depth")

type decoder struct {
	r            io.Reader
	sampleFormat afmt.SampleFormat
	pcmBuf       []byte
}

func NewDecoder(r io.Reader, sampleFormat afmt.SampleFormat) aio.SampleReader {
	if sampleFormat.Endian == nil {
		sampleFormat.Endian = binary.NativeEndian
	}

	return &decoder{
		r:            r,
		sampleFormat: sampleFormat,
	}
}

func (d *decoder) ReadSamples(p []float64) (int, error) {
	sampleSize := d.sampleFormat.BytesPerSample()
	numBytes := sampleSize * len(p)

	if cap(d.pcmBuf) < numBytes {
		d.pcmBuf = make([]byte, numBytes)
	} else {
		d.pcmBuf = d.pcmBuf[:numBytes]
	}

	n, err := d.r.Read(d.pcmBuf)
	if err != nil {
		return 0, err
	}

	for i := range n / sampleSize {
		offset := i * sampleSize
		if offset+sampleSize > len(d.pcmBuf) {
			return i, io.ErrUnexpectedEOF
		}

		switch d.sampleFormat.Encoding {
		case afmt.SampleEncodingInt:
			switch d.sampleFormat.BitDepth {
			case 8:
				v := d.pcmBuf[offset]
				p[i] = float64(v) / (1<<7 - 1)
			case 16:
				v := int16(d.sampleFormat.Endian.Uint16(d.pcmBuf[offset:]))
				p[i] = float64(v) / (1<<15 - 1)
			case 24:
				b := d.pcmBuf[offset : offset+3]
				v := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16
				if v&(1<<23) != 0 {
					v |= ^0xFFFFFF
				}
				p[i] = float64(v) / (1<<23 - 1)
			case 32:
				v := int32(d.sampleFormat.Endian.Uint32(d.pcmBuf[offset:]))
				p[i] = float64(v) / (1<<31 - 1)
			case 64:
				v := int64(d.sampleFormat.Endian.Uint64(d.pcmBuf[offset:]))
				p[i] = float64(v) / (1<<63 - 1)
			default:
				return 0, ErrInvalidBitDepth
			}
		case afmt.SampleEncodingUint:
			switch d.sampleFormat.BitDepth {
			case 8:
				v := d.pcmBuf[offset]
				p[i] = (float64(v) - 128) / 127
			case 16:
				v := d.sampleFormat.Endian.Uint16(d.pcmBuf[offset:])
				p[i] = float64(v) / (1<<16 - 1)
			case 24:
				if offset+3 > len(d.pcmBuf) {
					return i, io.ErrUnexpectedEOF
				}
				b := d.pcmBuf[offset : offset+3]
				v := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
				p[i] = float64(v) / (1<<24 - 1)
			case 32:
				v := d.sampleFormat.Endian.Uint32(d.pcmBuf[offset:])
				p[i] = float64(v) / (1<<32 - 1)
			case 64:
				v := d.sampleFormat.Endian.Uint64(d.pcmBuf[offset:])
				p[i] = float64(v) / (1<<64 - 1)
			default:
				return 0, ErrInvalidBitDepth
			}
		case afmt.SampleEncodingFloat:
			switch d.sampleFormat.BitDepth {
			case 32:
				bits := d.sampleFormat.Endian.Uint32(d.pcmBuf[offset:])
				p[i] = float64(math.Float32frombits(bits))
			case 64:
				bits := d.sampleFormat.Endian.Uint64(d.pcmBuf[offset:])
				p[i] = math.Float64frombits(bits)
			default:
				return 0, ErrInvalidBitDepth
			}
		default:
			return 0, ErrInvalidSampleEncoding
		}
	}

	return n / sampleSize, nil
}
