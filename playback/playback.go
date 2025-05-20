package playback

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
	"github.com/ebitengine/oto/v3"
)

var (
	otoCtx *oto.Context
	format Format
	player *oto.Player
)

var Output aio.SampleWriter

// Format represents the output sample format.
type Format = oto.Format

const (
	// FormatFloat32LE is the format of 32-bit floating-point little endian.
	FormatFloat32LE Format = oto.FormatFloat32LE

	// FormatUnsignedInt8 is the format of 8-bit unsigned integer.
	FormatUnsignedInt8 Format = oto.FormatUnsignedInt8

	//FormatSignedInt16LE is the format of 16-bit signed integer little endian.
	FormatSignedInt16LE Format = oto.FormatSignedInt16LE
)

// Init initializes audio playback. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(audioFormat afmt.Format, form Format, bufferSize int) error {
	if otoCtx != nil {
		return errors.New("playback cannot be initialized more than once")
	}

	format = form

	var err error
	var readyChan chan struct{}
	otoCtx, readyChan, err = oto.NewContext(&oto.NewContextOptions{
		SampleRate:   int(audioFormat.SampleRate.Hertz()),
		ChannelCount: audioFormat.NumChannels,
		Format:       format,
		BufferSize:   afmt.NumSamplesToDuration(audioFormat.SampleRate, bufferSize),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize driver: %w", err)
	}
	<-readyChan // wait for driver to init on channel

	pr, pw := aio.Pipe()
	Output = pw

	otoCtx.NewPlayer(newPCMReader(pr, bufferSize*audioFormat.NumChannels, audioFormat.NumChannels))

	return nil
}

type pcmReader struct {
	r         aio.SampleReader
	sampleBuf []float64
	pcmBuf    []byte
}

func newPCMReader(r aio.SampleReader, bufferSize int, numChans int) *pcmReader {
	pcmr := &pcmReader{
		r:         r,
		sampleBuf: make([]float64, bufferSize),
	}

	switch format {
	case FormatFloat32LE:
		pcmr.pcmBuf = make([]byte, bufferSize*4) // float32 = 4 bytes
	case FormatUnsignedInt8:
		pcmr.pcmBuf = make([]byte, bufferSize*1) // uint8 = 1 byte
	case FormatSignedInt16LE:
		pcmr.pcmBuf = make([]byte, bufferSize*2) // int16 = 2 bytes
	}

	return pcmr
}

func (r *pcmReader) Read(p []byte) (n int, err error) {
	n, err = r.r.ReadSamples(r.sampleBuf)
	if err != nil {
		return 0, err
	}

	switch format {
	case FormatFloat32LE:
		want := n * 4
		if len(r.pcmBuf) < want {
			r.pcmBuf = make([]byte, want)
		}

		for i := range n {
			bits := math.Float32bits(float32(r.sampleBuf[i]))
			binary.LittleEndian.PutUint32(r.pcmBuf[i*4:], bits)
		}

		copy(p, r.pcmBuf[:n*4])

		return n * 4, nil
	case FormatUnsignedInt8:
	case FormatSignedInt16LE:
	default:
		return 0, fmt.Errorf("invalid format")
	}
}
