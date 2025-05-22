package playback

import (
	"fmt"
	"sync"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/audio"
	"github.com/ebitengine/oto/v3"
)

var (
	format          afmt.Format
	bitDepthInBytes int
	bytesPerSample  int
	otoFormat       oto.Format
)

var (
	mu     sync.Mutex
	mixer  audio.Mixer
	otoCtx *oto.Context
	player *oto.Player
)

// Init initializes audio playback. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(audioFormat afmt.Format, bufferSize int) error {
	if otoCtx != nil {
		return fmt.Errorf("playback cannot be initialized more than once")
	}

	mixer = audio.Mixer{}
	format = audioFormat
	bitDepthInBytes = 4 // float32 = 4 bytes
	bytesPerSample = bitDepthInBytes * format.NumChannels
	otoFormat = oto.FormatFloat32LE

	// split buffer size between driver and player like Beep does
	driverBufSize := bufferSize / 2
	playerBufSize := bufferSize / 2

	var err error
	var readyChan chan struct{}
	otoCtx, readyChan, err = oto.NewContext(&oto.NewContextOptions{
		SampleRate:   int(audioFormat.SampleRate.Hertz()),
		ChannelCount: audioFormat.NumChannels,
		Format:       otoFormat,
		BufferSize:   afmt.NumSamplesToDuration(audioFormat.SampleRate, driverBufSize),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize driver: %w", err)
	}
	<-readyChan // wait for driver to init on channel

	player = otoCtx.NewPlayer(&pcmReader{r: &mixer})
	player.SetBufferSize(playerBufSize * bytesPerSample)
	go player.Play()

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
func Close() error {
	if player == nil {
		return nil
	}
	Clear()
	if err := player.Close(); err != nil {
		return err
	}
	player = nil
	if err := otoCtx.Suspend(); err != nil {
		return err
	}
	return nil
}

// Lock locks the playback mutex. While locked, the driver won't pull any new data from the playing sources.
// Lock if you want to modify any currently playing SampleReaders to avoid race conditions.
//
// Always lock for as little time as possible, to avoid playback glitches.
func Lock() {
	mu.Lock()
}

// TryLock tries to lock the playback mutex and reports whether it succeeded.
// While locked, the driver won't pull any new data from the playing sources.
// Lock if you want to modify any currently playing SampleReaders to avoid race conditions.
//
// Always lock for as little time as possible, to avoid playback glitches.
func TryLock() bool {
	return mu.TryLock()
}

// Unlock unlocks the playback mutex.
// Call this after locking and modifying any currently playing SampleReaders.
func Unlock() {
	mu.Unlock()
}

// Clear removes all currently playing sources from the internal mixer.
// Previously buffered samples may still be played.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	mixer.Clear()
}

// Suspend suspends the entire audio playback.
func Suspend() error {
	return otoCtx.Suspend()
}

// Resume resumes the entire audio playback, after being suspended by [Suspend].
func Resume() error {
	return otoCtx.Resume()
}

// Play starts playing all provided sources.
func Play(readers ...aio.SampleReader) {
	mu.Lock()
	defer mu.Unlock()
	mixer.Add(readers...)
}

// PlayWithDone starts playing all provided sources and returns a channel that closes when all sources have finished playing and drained.
func PlayWithDone(readers ...aio.SampleReader) chan struct{} {
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(len(readers))

	wrapped := make([]aio.SampleReader, len(readers))
	for i := range readers {
		wrapped = append(wrapped, aio.CallbackReader(readers[i], func() {
			wg.Done()
		}))
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	mu.Lock()
	defer mu.Unlock()
	mixer.Add(wrapped...)

	return done
}

// PlayAndWait plays all provided sources and waits until all sources have finished playing and drained.
func PlayAndWait(readers ...aio.SampleReader) {
	done := PlayWithDone(readers...)
	<-done
}
