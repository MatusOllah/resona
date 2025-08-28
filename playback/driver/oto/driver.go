// Package oto provides a cross-platform [Oto]-based playback driver.
//
// [Oto]: https://github.com/ebitengine/oto
package oto

import (
	"encoding/binary"
	"math"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/playback"
	"github.com/ebitengine/oto/v3"
)

// Driver represents the driver.
type Driver struct {
	ctx    *oto.Context
	player *oto.Player
}

// Init initializes the driver based on the format and source.
// It blocks until the driver is ready.
func (d *Driver) Init(format afmt.Format, src aio.SampleReader) error {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   int(format.SampleRate.Hertz()),
		ChannelCount: format.NumChannels,
		Format:       oto.FormatFloat32LE,
	})
	if err != nil {
		return err
	}
	<-ready

	d.ctx = ctx

	d.player = ctx.NewPlayer(&pcmReader{src: src})
	d.player.Play()
	return nil
}

// Close closes audio playback.
// However, the underlying driver keeps existing until the process dies,
// as closing it is not supported (see [Oto issue #149]).
//
// In most cases, there is no need to call Close even when the program doesn't play
// audio anymore, because the driver closes when the process dies.
//
// [Oto issue #149]: https://github.com/ebitengine/oto/issues/149
func (d *Driver) Close() error {
	if err := d.player.Close(); err != nil {
		return err
	}
	return d.ctx.Suspend()
}

// pcmReader is an [io.Reader] that wraps aio.SampleReader and encodes audio to float32 little endian PCM.
type pcmReader struct {
	src aio.SampleReader
}

func (r *pcmReader) Read(p []byte) (int, error) {
	const sampleSize = 4 // float32 size = 4 bytes

	buf := make([]float64, len(p)/sampleSize)

	n, err := r.src.ReadSamples(buf)
	if err != nil {
		return 0, err
	}
	for i := range buf[:n] {
		binary.LittleEndian.PutUint32(p[i*sampleSize:], math.Float32bits(float32(buf[i])))
	}
	return n * sampleSize, nil
}

func init() {
	playback.Register("oto", &Driver{}) // register driver
}
