package effect

import (
	"math"
)

// Volume adjusts the volume of the audio signal in decibels (dB).
//
// The zero value for Volume is ready to use.
type Volume struct {
	// Volume is the volume adjustment in decibels (dB).
	Volume float64
}

// NewVolume creates a new [Volume].
//
// In most cases, new([Volume]) (or just declaring a [Volume] variable) is sufficient
// to create a new [Volume].
func NewVolume(vol float64) *Volume {
	return &Volume{
		Volume: vol,
	}
}

func (v *Volume) Process(p []float32) error {
	var gain float32 = float32(math.Pow(10, v.Volume/20.0))
	for i := range p {
		p[i] *= gain
	}
	return nil
}
