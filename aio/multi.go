package aio

import "io"

type eofReader struct{}

func (eofReader) ReadSamples([]float64) (int, error) {
	return 0, io.EOF
}

type multiReader struct {
	readers []SampleReader
}

func (mr *multiReader) ReadSamples(p []float64) (n int, err error) {
	for len(mr.readers) > 0 {
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReader); ok {
				mr.readers = r.readers
				continue
			}
		}
		n, err = mr.readers[0].ReadSamples(p)
		if err == io.EOF {
			mr.readers[0] = eofReader{} // permit earlier GC
			mr.readers = mr.readers[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

func (mr *multiReader) WriteSamplesTo(w SampleWriter) (sum int64, err error) {
	return mr.writeSamplesToWithBuffer(w, make([]float64, 1024*32))
}

func (mr *multiReader) writeSamplesToWithBuffer(w SampleWriter, buf []float64) (sum int64, err error) {
	for i, r := range mr.readers {
		var n int64
		if subMr, ok := r.(*multiReader); ok { // reuse buffer with nested multiReaders
			n, err = subMr.writeSamplesToWithBuffer(w, buf)
		} else {
			n, err = copyBuffer(w, r, buf)
		}
		sum += n
		if err != nil {
			mr.readers = mr.readers[i:] // permit resume / retry after error
			return sum, err
		}
		mr.readers[i] = nil // permit early GC
	}
	mr.readers = nil
	return sum, nil
}

var _ SampleWriterTo = (*multiReader)(nil)

// MultiReader returns a [SampleReader] that's the logical concatenation of
// the provided input readers. They're read sequentially. Once all
// inputs have returned EOF, ReadSamples will return EOF. If any of the readers
// return a non-nil, non-EOF error, ReadSamples will return that error.
func MultiReader(readers ...SampleReader) SampleReader {
	r := make([]SampleReader, len(readers))
	copy(r, readers)
	return &multiReader{r}
}

type multiWriter struct {
	writers []SampleWriter
}

func (t *multiWriter) WriteSamples(p []float64) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.WriteSamples(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
//
// Each write is written to each listed writer, one at a time.
// If a listed writer returns an error, that overall write operation
// stops and returns the error; it does not continue down the list.
func MultiWriter(writers ...SampleWriter) SampleWriter {
	allWriters := make([]SampleWriter, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &multiWriter{allWriters}
}
