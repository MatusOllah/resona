package aio

import "io"

type seq struct {
	readers []SampleReader
	i       int
}

func (s *seq) ReadSamples(p []float64) (n int, err error) {
	for s.i < len(s.readers) && len(p) > 0 {
		rn, rerr := s.readers[s.i].ReadSamples(p)
		p = p[rn:]
		n += rn

		if rerr == io.EOF && rn == 0 {
			s.i++
		} else if rerr != nil {
			return n, rerr
		}
	}
	if s.i >= len(s.readers) {
		return n, io.EOF
	}
	return n, nil
}

// Seq returns a [SampleReader] that outputs each reader sequentially, without pauses.
func Seq(readers ...SampleReader) SampleReader {
	return &seq{readers: readers}
}
