package aio

import (
	"io"
	"sync"
)

// onceError is an object that will only store an error once.
type onceError struct {
	sync.Mutex // guards following
	err        error
}

func (a *onceError) Store(err error) {
	a.Lock()
	defer a.Unlock()
	if a.err != nil {
		return
	}
	a.err = err
}
func (a *onceError) Load() error {
	a.Lock()
	defer a.Unlock()
	return a.err
}

// A pipe is the shared pipe structure underlying [PipeReader] and [PipeWriter].
type pipe struct {
	wrMu sync.Mutex // Serializes WriteSamples operations
	wrCh chan []float64
	rdCh chan int

	once sync.Once // Protects closing done
	done chan struct{}
	rerr onceError
	werr onceError
}

func (p *pipe) readSamples(b []float64) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.readCloseError()
	default:
	}

	select {
	case bw := <-p.wrCh:
		nr := copy(b, bw)
		p.rdCh <- nr
		return nr, nil
	case <-p.done:
		return 0, p.readCloseError()
	}
}

func (p *pipe) closeRead(err error) error {
	if err == nil {
		err = io.ErrClosedPipe
	}
	p.rerr.Store(err)
	p.once.Do(func() { close(p.done) })
	return nil
}

func (p *pipe) writeSamples(b []float64) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.writeCloseError()
	default:
		p.wrMu.Lock()
		defer p.wrMu.Unlock()
	}

	for once := true; once || len(b) > 0; once = false {
		select {
		case p.wrCh <- b:
			nw := <-p.rdCh
			b = b[nw:]
			n += nw
		case <-p.done:
			return n, p.writeCloseError()
		}
	}
	return n, nil
}

func (p *pipe) closeWrite(err error) error {
	if err == nil {
		err = io.EOF
	}
	p.werr.Store(err)
	p.once.Do(func() { close(p.done) })
	return nil
}

// readCloseError is considered internal to the pipe type.
func (p *pipe) readCloseError() error {
	rerr := p.rerr.Load()
	if werr := p.werr.Load(); rerr == nil && werr != nil {
		return werr
	}
	return io.ErrClosedPipe
}

// writeCloseError is considered internal to the pipe type.
func (p *pipe) writeCloseError() error {
	werr := p.werr.Load()
	if rerr := p.rerr.Load(); werr == nil && rerr != nil {
		return rerr
	}
	return io.ErrClosedPipe
}

// A PipeReader is the read half of a pipe.
type PipeReader struct{ pipe }

// Read implements the standard Read interface:
// it reads data from the pipe, blocking until a writer
// arrives or the write end is closed.
// If the write end is closed with an error, that error is
// returned as err; otherwise err is EOF.
func (r *PipeReader) ReadSamples(data []float64) (n int, err error) {
	return r.pipe.readSamples(data)
}

// Close closes the reader; subsequent writes to the
// write half of the pipe will return the error [io.ErrClosedPipe].
func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

// CloseWithError closes the reader; subsequent writes
// to the write half of the pipe will return the error err.
//
// CloseWithError never overwrites the previous error if it exists
// and always returns nil.
func (r *PipeReader) CloseWithError(err error) error {
	return r.pipe.closeRead(err)
}

// A PipeWriter is the write half of a pipe.
type PipeWriter struct{ r PipeReader }

// Write implements the standard Write interface:
// it writes data to the pipe, blocking until one or more readers
// have consumed all the data or the read end is closed.
// If the read end is closed with an error, that err is
// returned as err; otherwise err is [io.ErrClosedPipe].
func (w *PipeWriter) WriteSamples(data []float64) (n int, err error) {
	return w.r.pipe.writeSamples(data)
}

// Close closes the writer; subsequent reads from the
// read half of the pipe will return no samples and [io.EOF].
func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

// CloseWithError closes the writer; subsequent reads from the
// read half of the pipe will return no samples and the error err,
// or [io.EOF] if err is nil.
//
// CloseWithError never overwrites the previous error if it exists
// and always returns nil.
func (w *PipeWriter) CloseWithError(err error) error {
	return w.r.pipe.closeWrite(err)
}

// Pipe creates a synchronous in-memory pipe.
// It can be used to connect code expecting an [io.Reader]
// with code expecting an [io.Writer].
//
// Reads and writes on the pipe are matched one to one
// except when multiple reads are needed to consume a single write.
// That is, each write to the [PipeWriter] blocks until it has satisfied
// one or more reads from the [PipeReader] that fully consume
// the written data.
// The data is copied directly from the write to the corresponding
// read (or reads); there is no internal buffering.
//
// It is safe to call ReadSamples and WriteSamples in parallel with each other or with Close.
// Parallel calls to ReadSampels and parallel calls to WriteSamples are also safe:
// the individual calls will be gated sequentially.
func Pipe() (*PipeReader, *PipeWriter) {
	pw := &PipeWriter{r: PipeReader{pipe: pipe{
		wrCh: make(chan []float64),
		rdCh: make(chan int),
		done: make(chan struct{}),
	}}}
	return &pw.r, pw
}
