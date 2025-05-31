package flac

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/codec"
	"github.com/MatusOllah/resona/freq"
	"github.com/mewkiz/flac"
)

const magic = "fLaC"

var ErrSeekingUnsupported = errors.New("flac: resource does not support seeking")

// Decoder represents the decoder for the FLAC file format.
// It implements codec.Decoder.
type Decoder struct {
	stream   *flac.Stream
	isSeeker bool
	pos      int
}

// NewDecoder creates a new [Decoder] and decodes the headers.
func NewDecoder(r io.Reader) (_ codec.Decoder, err error) {
	d := &Decoder{}

	rs, ok := r.(io.ReadSeeker)
	d.isSeeker = ok
	if ok {
		d.stream, err = flac.NewSeek(rs)
	} else {
		d.stream, err = flac.New(r)
	}
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Decoder) Format() afmt.Format {
	return afmt.Format{
		SampleRate:  freq.Frequency(d.stream.Info.SampleRate) * freq.Hertz,
		NumChannels: int(d.stream.Info.NChannels),
	}
}

func (d *Decoder) SampleFormat() afmt.SampleFormat {
	return afmt.SampleFormat{
		BitDepth: int(d.stream.Info.BitsPerSample),
		Encoding: afmt.SampleEncodingInt,
		Endian:   binary.LittleEndian,
	}
}

func (d *Decoder) Len() int {
	return int(d.stream.Info.NSamples)
}

func (d *Decoder) ReadSamples(p []float64) (int, error) {
	return 0, nil // TODO: refill buffer from FLAC frame and read samples
}

func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	if !d.isSeeker {
		return 0, ErrSeekingUnsupported
	}

	var target int
	switch whence {
	case io.SeekStart:
		target = int(offset)
	case io.SeekCurrent:
		target = d.pos + int(offset)
	case io.SeekEnd:
		target = int(d.stream.Info.NSamples) + int(offset)
	default:
		return 0, errors.New("flac: invalid seek whence")
	}

	if target < 0 || target >= int(d.stream.Info.NSamples) {
		return 0, errors.New("flac: seek out of bounds")
	}

	_pos, err := d.stream.Seek(uint64(target))
	if err != nil {
		return 0, err
	}
	d.pos = int(_pos)

	return int64(d.pos), nil
}

func init() {
	codec.RegisterFormat("flac", magic, NewDecoder)
}
