package riff

import "io"

var (
	RiffID FourCC = FourCC{'R', 'I', 'F', 'F'}
	ListID FourCC = FourCC{'L', 'I', 'S', 'T'}
)

// FourCC represents a four character code.
type FourCC [4]byte

// String returns a string representation of f.
func (f FourCC) String() string {
	return string(f[:])
}

// Chunk represents a RIFF chunk.
type Chunk struct {
	ID     FourCC
	Len    int
	Reader io.ReadSeeker
}
