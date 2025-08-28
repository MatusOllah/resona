package playback

import (
	"sync"

	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/audio"
)

// Player represents an audio player.
type Player struct {
	mux *audio.Mixer
	src *pausable
}

// NewPlayer creates a new [Player].
func (ctx *Context) NewPlayer(src aio.SampleReader) *Player {
	return &Player{
		mux: ctx.mux,
		src: &pausable{r: src},
	}
}

// Play starts the playback.
func (p *Player) Play() {
	p.mux.Add(p.src)
}

// PlayWithDone starts the playback and returns a channel that closes when the player has finished playing and drained.
func (p *Player) PlayWithDone() chan struct{} {
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	wrapped := aio.CallbackReader(p.src, func() {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	p.mux.Add(wrapped)

	return done
}

// PlayAndWait starts the playback and blocks until the player has finished playing and drained.
func (p *Player) PlayAndWait() {
	<-p.PlayWithDone()
}

// Pause pauses the playback.
func (p *Player) Pause() {
	p.src.paused = true
}

// Unpause resumes the playback.
func (p *Player) Unpause() {
	p.src.paused = false
}

// pausable is a wrapper around a regular aio.SampleReader that adds pausing/unpausing capabilities.
// It outputs silence when paused.
type pausable struct {
	r      aio.SampleReader
	paused bool
}

func (r *pausable) ReadSamples(p []float64) (int, error) {
	if r.paused {
		clear(p)
		return len(p), nil
	}
	return r.r.ReadSamples(p)
}
