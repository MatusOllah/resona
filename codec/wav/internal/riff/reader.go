package riff

import (
	"bytes"
	"errors"
	"io"
	"math"
)

var (
	ErrMissingPaddingByte     = errors.New("riff: missing padding byte")
	ErrMissingRIFFChunkHeader = errors.New("riff: missing RIFF chunk header")
	ErrListSubchunkTooLong    = errors.New("riff: list subchunk too long")
	ErrShortChunkData         = errors.New("riff: short chunk data")
	ErrShortChunkHeader       = errors.New("riff: short chunk header")
	ErrStaleReader            = errors.New("riff: stale reader")
	ErrSeekingUnsupported     = errors.New("riff: resource does not support seeking")
	ErrInvalidSeekWhence      = errors.New("riff: invalid seek whence")
	ErrSeekOutOfRange         = errors.New("riff: seek out of range")
)

func u32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

const chunkHeaderSize = 8

type Reader struct {
	r   io.Reader
	err error

	totalLen uint32
	chunkLen uint32

	chunkReader *chunkReader
	buf         [chunkHeaderSize]byte
	padded      bool
}

// NewReader creates a new [Reader] for a RIFF stream.
func NewReader(r io.Reader) (formType FourCC, data *Reader, err error) {
	var buf [chunkHeaderSize]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = ErrMissingRIFFChunkHeader
		}
		return FourCC{}, nil, err
	}
	if !bytes.Equal(buf[:4], RiffID[:]) {
		return FourCC{}, nil, ErrMissingRIFFChunkHeader
	}
	return NewListReader(u32(buf[4:]), r)
}

// NewListReader creates a new [Reader] for a LIST chunk.
func NewListReader(chunkLen uint32, chunkData io.Reader) (listType FourCC, data *Reader, err error) {
	if chunkLen < 4 {
		return FourCC{}, nil, ErrShortChunkData
	}
	z := &Reader{r: chunkData}
	if _, err := io.ReadFull(chunkData, z.buf[:4]); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = ErrShortChunkData
		}
		return FourCC{}, nil, err
	}
	z.totalLen = chunkLen - 4
	return FourCC{z.buf[0], z.buf[1], z.buf[2], z.buf[3]}, z, nil
}

// NextChunk decodes and returns the next chunk.
func (z *Reader) NextChunk() (c *Chunk, err error) {
	if z.err != nil {
		return nil, z.err
	}

	// Drain the rest of the previous chunk.
	if z.chunkLen != 0 {
		want := z.chunkLen
		var got int64
		got, z.err = io.Copy(io.Discard, z.chunkReader)
		if z.err == nil && uint32(got) != want {
			z.err = ErrShortChunkData
		}
		if z.err != nil {
			return nil, z.err
		}
	}
	z.chunkReader = nil
	if z.padded {
		if z.totalLen == 0 {
			z.err = ErrListSubchunkTooLong
			return nil, z.err
		}
		z.totalLen--
		_, z.err = io.ReadFull(z.r, z.buf[:1])
		if z.err != nil {
			if z.err == io.EOF {
				z.err = ErrMissingPaddingByte
			}
			return nil, z.err
		}
	}

	// We are done if we have no more data.
	if z.totalLen == 0 {
		z.err = io.EOF
		return nil, z.err
	}

	// Read the next chunk header.
	if z.totalLen < chunkHeaderSize {
		z.err = ErrShortChunkHeader
		return nil, z.err
	}
	z.totalLen -= chunkHeaderSize
	if _, z.err = io.ReadFull(z.r, z.buf[:chunkHeaderSize]); z.err != nil {
		if z.err == io.EOF || z.err == io.ErrUnexpectedEOF {
			z.err = ErrShortChunkHeader
		}
		return nil, z.err
	}
	chunkID := FourCC{z.buf[0], z.buf[1], z.buf[2], z.buf[3]}
	z.chunkLen = u32(z.buf[4:])
	if z.chunkLen > z.totalLen {
		z.err = ErrListSubchunkTooLong
		return nil, z.err
	}
	z.padded = z.chunkLen&1 == 1

	var startPos int64 = 0
	if s, ok := z.r.(io.Seeker); ok {
		startPos, err = s.Seek(0, io.SeekCurrent)
		if err != nil {
			z.err = err
			return nil, z.err
		}
	}
	z.chunkReader = &chunkReader{
		z:         z,
		start:     startPos,
		totalSize: int64(z.chunkLen),
	}

	return &Chunk{
		ID:     chunkID,
		Len:    int(z.chunkLen),
		Reader: z.chunkReader,
	}, nil
}

type chunkReader struct {
	z         *Reader
	start     int64
	offset    int64
	totalSize int64
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c != c.z.chunkReader {
		return 0, ErrStaleReader
	}
	z := c.z
	if z.err != nil {
		if z.err == io.EOF {
			return 0, ErrStaleReader
		}
		return 0, z.err
	}

	n := int(z.chunkLen)
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 {
		n = math.MaxInt32
	}
	if n > len(p) {
		n = len(p)
	}
	n, err := z.r.Read(p[:n])
	z.totalLen -= uint32(n)
	z.chunkLen -= uint32(n)
	c.offset += int64(n)
	if err != io.EOF {
		z.err = err
	}
	return n, err
}

func (c *chunkReader) Seek(offset int64, whence int) (int64, error) {
	s, ok := c.z.r.(io.Seeker)
	if !ok {
		return 0, ErrSeekingUnsupported
	}
	if c != c.z.chunkReader {
		return 0, ErrStaleReader
	}

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = c.offset + offset
	case io.SeekEnd:
		abs = c.totalSize + offset
	default:
		return 0, ErrInvalidSeekWhence
	}

	if abs < 0 || abs > c.totalSize {
		return 0, ErrSeekOutOfRange
	}

	_, err := s.Seek(c.start+abs, io.SeekStart)
	if err != nil {
		return 0, err
	}

	c.offset = abs
	c.z.chunkLen = uint32(c.totalSize - abs)
	return abs, nil
}
