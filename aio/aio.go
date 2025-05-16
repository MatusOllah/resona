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
//	[]float64{L0, R0, L1, R1, L2, R2, ...}
//
// contains successive frames, where each frame consists of one sample per channel (left, then right).
// This layout is common in many audio APIs and is efficient for streaming and hardware buffers.
// However, this may be less convenient for certain processing tasks.
// If channel-separated (planar) access is required, callers may convert interleaved slices
// using [Interleave] and [Deinterleave] accordingly.
//
// The number of channels is not specified by the aio interfaces themselves;
// it is an implicit contract between the caller and the implementation.
// Implementations should clearly document the expected or provided number of channels and sample rate.
//
// # Sample Format
//
// All samples are represented as 64-bit floating-point numbers (float64).
// The value range is typically normalized between -1.0 and +1.0, where:
//
//   - 0.0 represents silence
//   - -1.0 to +1.0 represents full-scale audio signal
//   - Values outside this range may be clipped or distorted depending on the backend
package aio

import "io"

// SampleReader is the interface for types that can read audio samples.
//
// ReadSamples reads up to len(p) 64-bit floating-point interleaved audio samples into p.
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
	ReadSamples(p []float64) (n int, err error)
}

// SampleWriter is the interface for types that can write audio samples.
//
// WriteSamples writes up to len(p) 64-bit floating-point interleaved audio samples from p.
// It returns the number of samples written (0 <= n <= len(p)) and any error encountered.
// Implementations should return a non-nil error if they write fewer than len(p).
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
//
// Implementations must not retain nor modify p.
type SampleWriter interface {
	WriteSamples(p []float64) (n int, err error)
}

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

// SampleReadWriteCloser combines [SampleReader], [SampleWriter] and [io.Seeker].
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

// SampleReaderAt is the interface for types that can read audio samples at an offset.
//
// ReadSamplesAt reads up to len(p) 64-bit floating-point interleaved audio samples into p
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
	ReadSamplesAt(p []float64, off int64) (n int, err error)
}

// SampleWriterAt is the interface for types that can write audio samples at an offset.
//
// WriteSamplesAt writes up to len(p) 64-bit floating-point interleaved audio samples into p
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
type SampleWriteAt interface {
	WriteSamplesAt(p []float64, off int64) (n int, err error)
}

// SingleSampleReader is the interface for types that can read single audio samples.
//
// ReadSample reads and returns the next 64-bit floating-point interleaved audio sample
// from the input or any error encountered. If ReadSample returns an error, no input
// sample was consumed, and the returned sample is undefined.
//
// The sample rate of the data is unspecified by this interface.
// Implementations should document their sample rate expectations.
type SingleSampleReader interface {
	ReadSample() (float64, error)
}

// SingleSampleScanner combines [SingleSampleReader] and the UnreadSample method.
//
// UnreadSample causes the next call to ReadSample to return the last audio sample read.
// If the last operation was not a successful call to ReadSample, UnreadSample may
// return an error, unread the last sample read (or the sample prior to the
// last-unread one), or (in implementations that support the [io.Seeker] interface)
// seek to one sample before the current offset.
type SingleSampleScanner interface {
	SingleSampleReader
	UnreadSample() error
}

// SingleSampleWriter is the interface for types that can write single audio samples.
type SingleSampleWriter interface {
	WriteSample(s float64) error
}
