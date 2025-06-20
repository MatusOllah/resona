package pcm

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/audio"
)

type encoder struct {
	r            aio.SampleReader
	sampleFormat afmt.SampleFormat
	f64buf       []float64
}

func NewEncoder(r aio.SampleReader, sampleFormat afmt.SampleFormat) io.Reader {
	if sampleFormat.Endian == nil {
		sampleFormat.Endian = binary.NativeEndian
	}

	return &encoder{
		r:            r,
		sampleFormat: sampleFormat,
	}
}

func (e *encoder) Read(p []byte) (int, error) {
	if e.sampleFormat.BitDepth <= 0 {
		return 0, ErrInvalidBitDepth
	}
	if e.sampleFormat.Encoding <= 0 {
		return 0, ErrInvalidSampleEncoding
	}

	sampleSize := e.sampleFormat.BytesPerSample()
	numSamples := len(p) / sampleSize

	if cap(e.f64buf) < numSamples {
		e.f64buf = make([]float64, numSamples)
	} else {
		e.f64buf = e.f64buf[:numSamples]
	}

	n, err := e.r.ReadSamples(e.f64buf)
	if err != nil {
		return 0, err
	}

	for i := range n {
		e.f64buf[i] = math.Max(-1, math.Min(1, e.f64buf[i]))

		offset := i * sampleSize
		switch e.sampleFormat.Encoding {
		case afmt.SampleEncodingInt:
			switch e.sampleFormat.BitDepth {
			case 8:
				v := int8(e.f64buf[i] * 127)
				p[offset] = byte(v)
			case 16:
				v := int16(e.f64buf[i] * (1<<15 - 1))
				e.sampleFormat.Endian.PutUint16(p[offset:], uint16(v))
			case 24:
				v := int32(e.f64buf[i] * (1<<23 - 1))
				putUint24(p[offset:], uint32(v), e.sampleFormat.Endian)
			case 32:
				v := int32(e.f64buf[i] * (1<<31 - 1))
				e.sampleFormat.Endian.PutUint32(p[offset:], uint32(v))
				/*
					case 64:
						v := int64(e.f64buf[i] * (1<<63 - 1))
						e.sampleFormat.Endian.PutUint64(p[offset:], uint64(v))
				*/
			default:
				return 0, ErrInvalidBitDepth
			}
		case afmt.SampleEncodingUint:
			switch e.sampleFormat.BitDepth {
			case 8:
				v := byte((e.f64buf[i] + 1.0) * 0.5 * ((1 << 8) - 1))
				p[offset] = v
				/*
					case 16:
						v := uint16((e.f64buf[i] + 1.0) * 0.5 * ((1 << 16) - 1))
						e.sampleFormat.Endian.PutUint16(p[offset:], v)
					case 24:
						v := uint32((e.f64buf[i] + 1.0) * 0.5 * ((1 << 24) - 1))
						putUint24(p[offset:], v, e.sampleFormat.Endian)
					case 32:
						v := uint32((e.f64buf[i] + 1.0) * 0.5 * float64(math.MaxUint32))
						e.sampleFormat.Endian.PutUint32(p[offset:], v)
					case 64:
						v := uint64((e.f64buf[i] + 1.0) * 0.5 * float64(math.MaxUint64))
						e.sampleFormat.Endian.PutUint64(p[offset:], v)
				*/
			default:
				return 0, ErrInvalidBitDepth
			}
		case afmt.SampleEncodingFloat:
			switch e.sampleFormat.BitDepth {
			case 32:
				e.sampleFormat.Endian.PutUint32(p[offset:], math.Float32bits(float32(e.f64buf[i])))
			case 64:
				e.sampleFormat.Endian.PutUint64(p[offset:], math.Float64bits(e.f64buf[i]))
			default:
				return 0, ErrInvalidBitDepth
			}
		default:
			return 0, ErrInvalidSampleEncoding
		}
	}

	return n * sampleSize, nil
}

func putUint24(p []byte, v uint32, endian binary.ByteOrder) {
	if len(p) < 3 {
		return
	}
	switch endian {
	case binary.BigEndian:
		p[0] = byte(v >> 16)
		p[1] = byte(v >> 8)
		p[2] = byte(v)
	case binary.LittleEndian:
		p[0] = byte(v)
		p[1] = byte(v >> 8)
		p[2] = byte(v >> 16)
	default:
		panic("unsupported byte order")
	}
}

// Encode encodes a slice of float64 samples into a PCM byte slice.
func Encode(s []float64, sampleFormat afmt.SampleFormat) ([]byte, error) {
	enc := NewEncoder(audio.NewReader(s), sampleFormat)
	b := make([]byte, len(s)*sampleFormat.BytesPerSample())
	n, err := enc.Read(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
