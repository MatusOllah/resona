package effect

import (
	"math"
)

// Volume adjusts the volume of the audio signal in decibels (dB).
type Volume struct {
	// Volume is the volume adjustment in decibels (dB).
	Volume float64

	// Mute, if true, silences the audio signal.
	Mute bool
}

func (v *Volume) Process(p []float64) error {
	var gain float64 = 0.0
	if !v.Mute {
		gain = math.Pow(10, v.Volume/20.0)
	}
	for i := range p {
		p[i] *= gain
	}
	return nil
}
