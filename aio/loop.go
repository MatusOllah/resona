package aio

import (
	"io"
	"math"
)

// LoopInfinite loops indefinitely.
const LoopInfinite int = -1

// A LoopReader reads the [SampleReadSeeker] in a loop.
type LoopReader struct {
	rs      SampleReadSeeker
	remains int
	start   int
	end     int
}

// NewLoopReader creates a new [LoopReader] that loops n times.
// If [LoopInfinite] is provided, it loops indefinitely.
func NewLoopReader(rs SampleReadSeeker, n int) *LoopReader {
	return &LoopReader{
		rs:      rs,
		remains: n,
		start:   0,
		end:     math.MaxInt,
	}
}

// SetNumLoops sets the number of loops.
// If [LoopInfinite] is provided, it loops indefinitely.
func (l *LoopReader) SetNumLoops(n int) {
	l.remains = n
}

// SetStart sets the start position of the loop in samples.
func (l *LoopReader) SetStart(start int) {
	if start < 0 {
		panic("aio: start position out of bounds")
	}

	l.start = start
}

// SetEnd sets the start position of the loop in samples.
func (l *LoopReader) SetEnd(end int) {
	l.end = end
}

func (l *LoopReader) ReadSamples(p []float64) (int, error) {
	if l.remains == 0 {
		return 0, io.EOF
	}

	total := 0

	for len(p) > 0 {
		toRead := len(p)

		pos, err := l.rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return total, err
		}

		samplesUntilEnd := l.end - int(pos)
		if samplesUntilEnd <= 0 {
			// Loop boundary hit
			if l.remains > 0 {
				l.remains--
			}
			if l.remains == 0 {
				return total, io.EOF
			}
			if _, err := l.rs.Seek(int64(l.start), io.SeekStart); err != nil {
				return total, err
			}
			continue
		}

		toRead = min(samplesUntilEnd, toRead)
		n, err := l.rs.ReadSamples(p[:toRead])
		total += n
		p = p[n:]

		if err != nil && err != io.EOF {
			return total, err
		}
	}

	return total, nil
}
