package oggvorbis

import (
	"errors"
	"io"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/codec"
	"github.com/MatusOllah/resona/freq"
	"github.com/jfreymuth/oggvorbis"
)

// Decoder represents the decoder for the Ogg Vorbis file format.
// It implements codec.Decoder.
type Decoder struct {
	oggR   *oggvorbis.Reader
	f32buf []float32
}

// NewDecoder creates a new [Decoder] and decodes the headers.
func NewDecoder(r io.Reader) (_ codec.Decoder, err error) {
	d := &Decoder{}

	d.oggR, err = oggvorbis.NewReader(r)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Format returns the audio stream format.
func (d *Decoder) Format() afmt.Format {
	return afmt.Format{
		SampleRate:  freq.Frequency(d.oggR.SampleRate()) * freq.Hertz,
		NumChannels: d.oggR.Channels(),
	}
}

// SampleFormat returns the sample format that samples are being decoded to internally.
// Note that this isn't actually the audio stream's sample format, as it's compressed.
func (d *Decoder) SampleFormat() afmt.SampleFormat {
	return afmt.SampleFormat{
		BitDepth: 32,
		Encoding: afmt.SampleEncodingFloat,
	}
}

// Len returns the total number of frames.
func (d *Decoder) Len() int {
	return int(d.oggR.Length()) / d.oggR.Channels()
}

// ReadSamples reads float64 samples into p.
// It returns the number of samples read and/or an error.
func (d *Decoder) ReadSamples(p []float64) (int, error) {
	if len(d.f32buf) < len(p) {
		d.f32buf = make([]float32, len(p))
	} else {
		d.f32buf = d.f32buf[:len(p)]
	}

	n, err := d.oggR.Read(d.f32buf)
	if err != nil {
		return 0, err
	}

	for i := range n {
		p[i] = float64(d.f32buf[i])
	}

	return n, nil
}

// Seek seeks to the specified frame.
// It returns the new offset relative to the start and/or an error.
// It will return an error if the source is not an [io.Seeker].
func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	// return position
	if offset == 0 && whence == io.SeekCurrent {
		return d.oggR.Position() / int64(d.oggR.Channels()), nil
	}

	var target int64 = d.oggR.Position() / int64(d.oggR.Channels())
	switch whence {
	case io.SeekStart:
		target = offset
	case io.SeekCurrent:
		target += offset
	case io.SeekEnd:
		target = int64(d.Len()) + offset
	default:
		return 0, errors.New("oggvorbis: invalid seek whence")
	}

	if target < 0 || target > int64(d.Len()) {
		return 0, errors.New("oggvorbis: seek out of bounds")
	}

	if err := d.oggR.SetPosition(target); err != nil {
		return 0, err
	}

	return target, nil
}

func init() {
	codec.RegisterFormat("ogg", string([]byte{0x4F, 0x67, 0x67, 0x53}), NewDecoder)
}
