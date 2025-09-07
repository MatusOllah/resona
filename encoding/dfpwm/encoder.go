package dfpwm

// Original C implementation: https://github.com/ChenThread/dfpwm/blob/master/1a/aucmp.c
/*
DFPWM1a (Dynamic Filter Pulse Width Modulation) codec - C Implementation
by Ben "GreaseMonkey" Russell, 2012, 2016
Public Domain

Compression Component
*/

/*
type encoder struct {
	w   io.Writer
	buf []byte

	q  int
	s  int
	lt int
}

func NewEncoder(w io.Writer) aio.SampleWriter {
	return &encoder{
		w:  w,
		q:  0,
		s:  0,
		lt: -128,
	}
}

func (e *encoder) WriteSamples(p []float64) (int, error) {
	return 0, nil
}

func Encode(s []float64) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	_, err := enc.WriteSamples(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodedLen(x int) int {
	return x / 8
}
*/
