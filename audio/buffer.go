// Portions of this code are inspired by the Go standard library's bytes.Buffer.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package audio

import "io"

// Buffer is a simple variable-sized audio buffer of float32 samples with
// [Buffer.Read], [Buffer.Write], and other helper methods.
//
// The zero value for [Buffer] is an empty buffer ready to use.
type Buffer struct {
	buf []float32
	off int
}

// NewBuffer creates and initializes a new Buffer using buf as its initial contents.
// The new [Buffer] takes ownership of buf, and the caller should not use buf after this call.
//
// In most cases, new([Buffer]) (or just declaring a [Buffer] variable) is sufficient
// to create a new [Buffer].
func NewBuffer(buf []float32) *Buffer {
	return &Buffer{buf: buf}
}

// NewBufferSize creates and initializes a new Buffer with the given capacity.
func NewBufferSize(size int) *Buffer {
	return &Buffer{buf: make([]float32, 0, size), off: 0}
}

// float32s returns a slice of the unread portion of the buffer.
// The slice is only valid until the next buffer modification (e.g. reading, writing, truncating).
// The slice aliases the buffer content at least until the next buffer modification,
// so changes to the slice will affect the buffer content and vice versa.
func (b *Buffer) float32s() []float32 {
	return b.buf[b.off:]
}

// Len returns the number of unread samples in the buffer;
// that is, [Buffer.Len] = len([Buffer.float32s]).
func (b *Buffer) Len() int {
	return len(b.buf) - b.off
}

// Cap returns the capacity of the buffer's underlying slice.
func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// Available returns how many samples are unused in the buffer.
func (b *Buffer) Available() int {
	return cap(b.buf) - len(b.buf)
}

// Truncate discards all but the first n unread samples from the buffer.
func (b *Buffer) Truncate(n int) {
	if n == 0 {
		b.Reset() // reset buffer
		return
	}
	if n < 0 || n > b.Len() {
		panic("audio Buffer.Truncate: truncation out of bounds")
	}
	b.buf = b.buf[:b.off+n]
}

// Reset resets and wipes the buffer content to be empty.
// The capacity is unchanged.
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

const minBufferSize = 64

func (b *Buffer) grow(n int) int {
	m := b.Len()
	if m == 0 && b.off != 0 {
		b.Reset() // reset buffer
	}
	// try to grow by reslicing
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l
	}
	if b.buf == nil || n <= minBufferSize {
		b.buf = make([]float32, n, minBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// slide things down
		copy(b.buf, b.buf[b.off:])
	} else if c > int(^uint(0)>>1)-c-n {
		panic("audio Buffer: too large")
	} else {
		// grow slice
		buf := make([]float32, m+n, 2*c+n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// Grow grows the buffer's capacity by n samples.
// If n is negative, Grow panics.
func (b *Buffer) Grow(n int) {
	if n < 0 {
		panic("audio Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
// The return value n is the length of p; err is always nil.
func (b *Buffer) Write(p []float32) (n int, err error) {
	m := b.grow(len(p))
	return copy(b.buf[m:], p), nil
}

// Read reads up to len(p) samples from the buffer into p.
func (b *Buffer) Read(p []float32) (n int, err error) {
	if len(b.buf) <= b.off { // empty
		b.Reset() // reset to recover space
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}
