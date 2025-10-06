// Portions of this code are derived from the Go standard library's io package.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package aio defines basic interfaces for audio I/O primitives.
// It provides shared abstractions over various audio backends,
// enabling modular and interoperable audio processing components.
//
// These interfaces typically wrap low-level implementations.
// Unless explicitly stated, implementations should not be assumed
// to be safe for concurrent use.
//
// # Interleaved Audio Format
//
// All sample slices in Resona and this package represent interleaved, multi-channel audio data.
// In an interleaved layout, samples for each channel are stored in sequence for each frame.
// For example, in stereo (2-channel) audio, the slice:
//
//	[]float32{L0, R0, L1, R1, L2, R2, ...}
//
// contains successive frames, where each frame consists of one sample per channel (left, then right).
// This layout is common in many audio APIs and is efficient for streaming and hardware buffers.
// However, this may be less convenient for certain processing tasks.
// If channel-separated (planar) access is required, callers may convert interleaved slices using the audio package.
//
// The number of channels is not specified by the aio interfaces themselves;
// it is an implicit contract between the caller and the implementation.
// Implementations should clearly document the expected or provided number of channels and sample rate.
//
// # Sample Format
//
// All samples are represented as 32-bit floating-point numbers (float32).
// The value range is typically normalized between -1.0 and +1.0, where:
//
//   - 0.0 represents silence
//   - -1.0 to +1.0 represents full-scale audio signal
//   - Values outside this range may be clipped or distorted depending on the backend
//
// # Seeking
//
// All seekable streams in Resona and this package implement the [io.Seeker] interface.
// Seek offset is measured in frames, where one frame contains one sample per channel.
package aio

import (
	"errors"
	"io"
	"sync"
)

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

// SampleReader is the interface for types that are readable.
//
// ReadSamples reads up to len(p) 32-bit floating-point interleaved audio samples into p.
// It returns the number of samples read (0 <= n <= len(p)) and any error encountered.
// If some samples are available but fewer than len(p), ReadSamples may return
// early with the available data rather than blocking.
//
// If an error or end-of-stream occurs after successfully reading n > 0 samples,
// ReadSamples returns the samples along with a non-nil error.
// The next call should then return 0 and [io.EOF].
//
// Callers should always handle the returned samples before inspecting err.
// This ensures correct handling of partial reads and end-of-stream behavior.
//
// If len(p) == 0, ReadSamples must return n == 0.
// It may return a non-nil error if an error is known (e.g., EOF).
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
//
// Implementations must not retain nor modify p.
type SampleReader interface {
	ReadSamples(p []float32) (n int, err error)
}

// Source is an alias for [SampleReader].
type Source = SampleReader

// SampleWriter is the interface for types that are writable.
//
// WriteSamples writes up to len(p) 32-bit floating-point interleaved audio samples from p.
// It returns the number of samples written (0 <= n <= len(p)) and any error encountered.
// Implementations should return a non-nil error if they write fewer than len(p).
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
//
// Implementations must not retain nor modify p.
type SampleWriter interface {
	WriteSamples(p []float32) (n int, err error)
}

// Sink is an alias for [SampleWriter].
type Sink = SampleWriter

// SampleReadWriter combines [SampleReader] and [SampleWriter].
type SampleReadWriter interface {
	SampleReader
	SampleWriter
}

// SampleReadCloser combines [SampleReader] and [io.Closer].
type SampleReadCloser interface {
	SampleReader
	io.Closer
}

// SampleWriteCloser combines [SampleWriter] and [io.Closer].
type SampleWriteCloser interface {
	SampleWriter
	io.Closer
}

// SampleReadWriteCloser combines [SampleReader], [SampleWriter] and [io.Closer].
type SampleReadWriteCloser interface {
	SampleReader
	SampleWriter
	io.Closer
}

// SampleReadSeeker combines [SampleReader] and [io.Seeker].
type SampleReadSeeker interface {
	SampleReader
	io.Seeker
}

// SampleReadSeekCloser combines [SampleReader], [io.Seeker] and [io.Closer].
type SampleReadSeekCloser interface {
	SampleReader
	io.Seeker
	io.Closer
}

// SampleWriteSeeker combines [SampleWriter] and [io.Seeker].
type SampleWriteSeeker interface {
	SampleWriter
	io.Seeker
}

// SampleReadWriteSeeker combines [SampleReader], [SampleWriter] and [io.Seeker].
type SampleReadWriteSeeker interface {
	SampleReader
	SampleWriter
	io.Seeker
}

// SampleReaderFrom is the interface for types that can read audio samples from a [SampleReader].
//
// ReadFrom reads data from r until EOF or error.
// The return value n is the number of samples read.
// Any error except EOF encountered during the read is also returned.
//
// The [Copy] function uses [SampleReaderFrom] if available.
type SampleReaderFrom interface {
	ReadSamplesFrom(r SampleReader) (n int64, err error)
}

// SampleWriterTo is the interface for types that can write audio samples to a [SampleWriter].
//
// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of samples
// written. Any error encountered during the write is also returned.
//
// The [Copy] function uses [SampleWriterTo] if available.
type SampleWriterTo interface {
	WriteSamplesTo(w SampleWriter) (n int64, err error)
}

// SampleReaderAt is the interface for types that are readable at an offset.
//
// ReadSamplesAt reads up to len(p) 32-bit floating-point interleaved audio samples into p
// starting at offset off in the underlying input source.
//
// When ReadSamplesAt returns n < len(p), it returns a non-nil error
// explaining why more samples were not returned. In this respect,
// ReadSamplesAt is stricter than ReadSamples.
//
// Callers should always handle the returned samples before inspecting err.
// This ensures correct handling of partial reads and end-of-stream behavior.
//
// If len(p) == 0, ReadSamples must return n == 0.
// It may return a non-nil error if an error is known (e.g., EOF).
//
// If ReadSamplesAt is reading from an input source with a seek offset,
// it should not affect nor be affected by the underlying seek offset.
//
// Clients of ReadSamplesAt can execute parallel calls on the same input source.
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
//
// Implementations must not retain nor modify p.
type SampleReaderAt interface {
	ReadSamplesAt(p []float32, off int64) (n int, err error)
}

// SampleWriterAt is the interface for types that are writable at an offset.
//
// WriteSamplesAt writes up to len(p) 32-bit floating-point interleaved audio samples into p
// at offset off in the underlying input source.
// It returns the number of samples written (0 <= n <= len(p)) and any error encountered.
// Implementations should return a non-nil error if they write fewer than len(p).
//
// If WriteSamplesAt is writing to an input source with a seek offset,
// it should not affect nor be affected by the underlying seek offset.
//
// Clients of WriteSamplesAt can execute parallel calls on the same input source.
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
//
// Implementations must not retain nor modify p.
type SampleWriterAt interface {
	WriteSamplesAt(p []float32, off int64) (n int, err error)
}

// ReadAtLeast reads from r into buf until it has read at least min samples.
// It returns the number of samples copied and an error if fewer samples were read.
// The error is EOF only if no samples were read.
// If an EOF happens after reading fewer than min samples,
// ReadAtLeast returns [io.ErrUnexpectedEOF].
// If min is greater than the length of buf, ReadAtLeast returns [io.ErrShortBuffer].
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min samples, the error is dropped.
func ReadAtLeast(r SampleReader, buf []float32, min int) (n int, err error) {
	if len(buf) < min {
		return 0, io.ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.ReadSamples(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

// ReadFull reads exactly len(buf) samples from r into buf.
// It returns the number of samples copied and an error if fewer samples were read.
// The error is EOF only if no samples were read.
// If an EOF happens after reading some but not all the samples,
// ReadFull returns [io.ErrUnexpectedEOF].
// On return, n == len(buf) if and only if err == nil.
// If r returns an error having read at least len(buf) samples, the error is dropped.
func ReadFull(r SampleReader, buf []float32) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}

// CopyN copies n samples (or until an error) from src to dst.
// It returns the number of samples copied and the earliest
// error encountered while copying.
// On return, written == n if and only if err == nil.
//
// If dst implements [SampleReaderFrom], the copy is implemented using it.
func CopyN(dst SampleWriter, src SampleReader, n int64) (written int64, err error) {
	written, err = Copy(dst, LimitReader(src, n))
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}
	return
}

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of samples
// copied and the first error encountered while copying, if any.
//
// A successful Copy returns err == nil, not err == EOF.
// Because Copy is defined to read from src until EOF, it does
// not treat an EOF from Read as an error to be reported.
//
// If src implements [SampleWriterTo],
// the copy is implemented by calling src.WriteTo(dst).
// Otherwise, if dst implements [SampleReaderFrom],
// the copy is implemented by calling dst.ReadFrom(src).
func Copy(dst SampleWriter, src SampleReader) (written int64, err error) {
	return copyBuffer(dst, src, nil)
}

// CopyBuffer is identical to Copy except that it stages through the
// provided buffer (if one is required) rather than allocating a
// temporary one. If buf is nil, one is allocated; otherwise if it has
// zero length, CopyBuffer panics.
//
// If either src implements [SampleWriterTo] or dst implements [SampleReaderFrom],
// buf will not be used to perform the copy.
func CopyBuffer(dst SampleWriter, src SampleReader, buf []float32) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBuffer(dst SampleWriter, src SampleReader, buf []float32) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(SampleWriterTo); ok {
		return wt.WriteSamplesTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rf, ok := dst.(SampleReaderFrom); ok {
		return rf.ReadSamplesFrom(src)
	}
	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]float32, size)
	}
	for {
		nr, er := src.ReadSamples(buf)
		if nr > 0 {
			nw, ew := dst.WriteSamples(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// LimitReader returns a [SampleReader] that reads from r
// but stops with EOF after n samples.
// The underlying implementation is a *LimitedReader.
func LimitReader(r SampleReader, n int64) SampleReader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N samples. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReader struct {
	R SampleReader // underlying reader
	N int64        // max samples remaining
}

func (l *LimitedReader) ReadSamples(p []float32) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.ReadSamples(p)
	l.N -= int64(n)
	return
}

// NewSectionReader returns a [SectionReader] that reads from r
// starting at offset off and stops with EOF after n samples.
func NewSectionReader(r SampleReaderAt, off int64, n int64) *SectionReader {
	var remaining int64
	const maxint64 = 1<<63 - 1
	if off <= maxint64-n {
		remaining = n + off
	} else {
		// Overflow, with no way to return error.
		// Assume we can read up to an offset of 1<<63 - 1.
		remaining = maxint64
	}
	return &SectionReader{r, off, off, remaining, n}
}

// SectionReader implements Read, Seek, and ReadAt on a section
// of an underlying [SampleReaderAt].
type SectionReader struct {
	r     SampleReaderAt // constant after creation
	base  int64          // constant after creation
	off   int64
	limit int64 // constant after creation
	n     int64 // constant after creation
}

func (s *SectionReader) ReadSamples(p []float32) (n int, err error) {
	if s.off >= s.limit {
		return 0, io.EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadSamplesAt(p, s.off)
	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
		offset += s.base
	case io.SeekCurrent:
		offset += s.off
	case io.SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

func (s *SectionReader) ReadSamplesAt(p []float32, off int64) (n int, err error) {
	if off < 0 || off >= s.Size() {
		return 0, io.EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.ReadSamplesAt(p, off)
		if err == nil {
			err = io.EOF
		}
		return n, err
	}
	return s.r.ReadSamplesAt(p, off)
}

// Size returns the size of the section in samples.
func (s *SectionReader) Size() int64 { return s.limit - s.base }

// Outer returns the underlying [SampleReaderAt] and offsets for the section.
//
// The returned values are the same that were passed to [NewSectionReader]
// when the [SectionReader] was created.
func (s *SectionReader) Outer() (r SampleReaderAt, off int64, n int64) {
	return s.r, s.base, s.n
}

// An OffsetWriter maps writes at offset base to offset base+off in the underlying writer.
type OffsetWriter struct {
	w    SampleWriterAt
	base int64 // the original offset
	off  int64 // the current offset
}

// NewOffsetWriter returns an [OffsetWriter] that writes to w
// starting at offset off.
func NewOffsetWriter(w SampleWriterAt, off int64) *OffsetWriter {
	return &OffsetWriter{w, off, off}
}

func (o *OffsetWriter) WriteSamples(p []float32) (n int, err error) {
	n, err = o.w.WriteSamplesAt(p, o.off)
	o.off += int64(n)
	return
}

func (o *OffsetWriter) WriteSamplesAt(p []float32, off int64) (n int, err error) {
	if off < 0 {
		return 0, errOffset
	}

	off += o.base
	return o.w.WriteSamplesAt(p, off)
}

func (o *OffsetWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
		offset += o.base
	case io.SeekCurrent:
		offset += o.off
	}
	if offset < o.base {
		return 0, errOffset
	}
	o.off = offset
	return offset - o.base, nil
}

// TeeReader returns a [SampleReader] that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReader(r SampleReader, w SampleWriter) SampleReader {
	return &teeReader{r, w}
}

type teeReader struct {
	r SampleReader
	w SampleWriter
}

func (t *teeReader) ReadSamples(p []float32) (n int, err error) {
	n, err = t.r.ReadSamples(p)
	if n > 0 {
		if n, err := t.w.WriteSamples(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

// Discard is a [SampleWriter] on which all WriteSamples calls succeed
// without doing anything.
var Discard SampleWriter = discard{}

type discard struct{}

// discard implements [SampleReaderFrom] as an optimization so [Copy] to
// [io.Discard] can avoid doing unnecessary work.
var _ SampleReaderFrom = discard{}

func (discard) WriteSamples(p []float32) (int, error) {
	return len(p), nil
}

var blackHolePool = sync.Pool{
	New: func() any {
		b := make([]float32, 8192)
		return &b
	},
}

func (discard) ReadSamplesFrom(r SampleReader) (n int64, err error) {
	bufp := blackHolePool.Get().(*[]float32)
	readSize := 0
	for {
		readSize, err = r.ReadSamples(*bufp)
		n += int64(readSize)
		if err != nil {
			blackHolePool.Put(bufp)
			if err == io.EOF {
				return n, nil
			}
			return
		}
	}
}

// NopCloser returns a [SampleReadCloser] with a no-op Close method wrapping
// the provided [SampleReader] r.
// If r implements [SampleWriterTo], the returned [SampleReadCloser] will implement [SampleWriterTo]
// by forwarding calls to r.
func NopCloser(r SampleReader) SampleReadCloser {
	if _, ok := r.(SampleWriterTo); ok {
		return nopCloserWriterTo{r}
	}
	return nopCloser{r}
}

type nopCloser struct {
	SampleReader
}

func (nopCloser) Close() error { return nil }

type nopCloserWriterTo struct {
	SampleReader
}

func (nopCloserWriterTo) Close() error { return nil }

func (c nopCloserWriterTo) WriteSamplesTo(w SampleWriter) (n int64, err error) {
	return c.SampleReader.(SampleWriterTo).WriteSamplesTo(w)
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAll(r SampleReader) ([]float32, error) {
	b := make([]float32, 0, 512)
	for {
		n, err := r.ReadSamples(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}

		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
	}
}

type callbackReader struct {
	r     SampleReader
	done  func()
	drain bool
}

// CallbackReader returns a [SampleReader] that calls done when r drains.
func CallbackReader(r SampleReader, done func()) SampleReader {
	return &callbackReader{r: r, done: done}
}

func (cr *callbackReader) ReadSamples(p []float32) (int, error) {
	if cr.drain {
		return 0, io.EOF
	}
	n, err := cr.r.ReadSamples(p)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if err == io.EOF || n == 0 {
		cr.drain = true
		if cr.done != nil {
			cr.done()
		}
	}
	return n, err
}

// PausableReader is a [SampleReader] that can be paused and resumed.
// When paused, ReadSamples outputs silence and doesn't read anything from the underlying [SampleReader].
type PausableReader struct {
	r      SampleReader
	mu     sync.RWMutex
	paused bool
}

// NewPausableReader creates a new [PausableReader].
func NewPausableReader(r SampleReader) *PausableReader {
	return &PausableReader{r: r}
}

// Pause pauses the reader. While paused, it outputs silence.
func (r *PausableReader) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused = true
}

// Resume resumes the reader.
func (r *PausableReader) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused = false
}

// IsPaused reports whether the reader is currently paused.
func (r *PausableReader) IsPaused() bool {
	r.mu.RLock()
	paused := r.paused
	r.mu.RUnlock()
	return paused
}

func (r *PausableReader) ReadSamples(p []float32) (int, error) {
	r.mu.RLock()
	paused := r.paused
	r.mu.RUnlock()
	if paused {
		clear(p)
		return len(p), nil
	}
	return r.r.ReadSamples(p)
}
