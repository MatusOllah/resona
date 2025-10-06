package pcm

import (
	"encoding/binary"
	"io"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
)

// NewEbitenS16Encoder creates an [io.Reader] that encodes samples from
// the given aio.SampleReader to 16-bit signed little-endian linear PCM format
// for use with Ebiten's audio APIs.
func NewEbitenS16Encoder(src aio.SampleReader) io.Reader {
	pr, pw := io.Pipe()
	enc := NewEncoder(pw, afmt.SampleFormat{
		BitDepth: 16,
		Encoding: afmt.SampleEncodingInt,
		Endian:   binary.LittleEndian,
	})

	go func() {
		defer pw.Close()
		buf := make([]float32, 1024)
		for {
			n, err := src.ReadSamples(buf)
			if n > 0 {
				if _, err := enc.WriteSamples(buf[:n]); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr
}

// NewEbitenF32Encoder creates an [io.Reader] that encodes samples from
// the given aio.SampleReader to 32-bit float little-endian linear PCM format
// for use with Ebiten's F32 audio APIs.
func NewEbitenF32Encoder(src aio.SampleReader) io.Reader {
	pr, pw := io.Pipe()
	enc := NewEncoder(pw, afmt.SampleFormat{
		BitDepth: 32,
		Encoding: afmt.SampleEncodingFloat,
		Endian:   binary.LittleEndian,
	})

	go func() {
		defer pw.Close()
		buf := make([]float32, 1024)
		for {
			n, err := src.ReadSamples(buf)
			if n > 0 {
				if _, err := enc.WriteSamples(buf[:n]); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr
}
