package playback

import (
	"github.com/MatusOllah/resona/aio"
)

type pcmReader struct {
	r   aio.SampleReader
	buf []float64
}

func (r *pcmReader) Read(p []byte) (int, error) {
	//TODO: samples to float32 le

	return 0, nil
}
