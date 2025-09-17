package effect

// Gain amplifies the audio signal. The output gets multiplies by (1 + Gain).
//
// Note that gain is not equivalent to volume. Human perception of volume is
// roughly exponential, while gain only amplifies linearly.
// To adjust volume, use [Volume] instead.
//
// The zero value for Gain is ready to use.
type Gain struct {
	// Gain is the linear gain factor.
	// A value of 0 means no change, negative values attenuate the signal,
	// and positive values amplify it.
	Gain float64
}

// NewGain creates a new [Gain] effect using gain as its initial gain.
//
// In most cases, new([Gain]) (or just declaring a [Gain] variable) is sufficient
// to create a new [Gain].
func NewGain(gain float64) *Gain {
	return &Gain{Gain: gain}
}

func (g *Gain) Process(p []float64) error {
	for i := range p {
		p[i] *= (1 + g.Gain)
	}
	return nil
}
