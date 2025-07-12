package g711

import (
	"io"

	"github.com/MatusOllah/resona/aio"
)

// EncodeUlaw encodes a slice of float64 samples into μ-law.
func EncodeUlaw(s []float64) []byte {
	b := make([]byte, len(s))
	for i := range s {
		i16 := int16(s[i] * (1<<15 - 1))
		if i16 >= 0 {
			b[i] = ulawEnc[i16>>4]
		} else {
			b[i] = 0x7F & ulawEnc[-i16>>4]
		}
	}
	return b
}

type ulawEncoder struct {
	w io.Writer
}

// NewUlawEncoder returns an aio.SampleWriter that encodes and writes μ-law samples to the provided [io.Writer].
func NewUlawEncoder(w io.Writer) aio.SampleWriter {
	return &ulawEncoder{w: w}
}

func (e *ulawEncoder) WriteSamples(p []float64) (int, error) {
	ulaw := EncodeUlaw(p)

	n, err := e.w.Write(ulaw)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// DecodeUlaw decodes μ-law encoded samples.
func DecodeUlaw(b []byte) []float64 {
	s := make([]float64, len(b))
	for i := range b {
		i16 := ulawDec[b[i]]
		s[i] = float64(i16) / (1<<15 - 1)
	}
	return s
}

type ulawDecoder struct {
	r   io.Reader
	buf []byte
}

// NewUlawDecoder returns an aio.SampleReader that reads and decodes μ-law encoded samples from the provided [io.Reader].
func NewUlawDecoder(r io.Reader) aio.SampleReader {
	return &ulawDecoder{r: r}
}

func (d *ulawDecoder) ReadSamples(p []float64) (int, error) {
	numSamples := len(p)
	if cap(d.buf) < numSamples {
		d.buf = make([]byte, numSamples)
	} else {
		d.buf = d.buf[:numSamples]
	}

	n, err := d.r.Read(d.buf)
	if err != nil {
		return n, err
	}

	f64buf := DecodeUlaw(d.buf[:n])
	copy(p, f64buf)

	return n, nil
}
