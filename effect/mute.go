package effect

// Mute mutes the audio signal. While muted, it outputs silence.
//
// The zero value for Mute is ready to use.
type Mute struct {
	// Mute, if true, silences the audio signal.
	Mute bool
}

// NewMute creates a new [Mute] effect with mute being its initial state.
//
// In most cases, new([Mute]) (or just declaring a [Mute] variable) is sufficient
// to create a new [Mute].
func NewMute(mute bool) *Mute {
	return &Mute{Mute: mute}
}

func (m *Mute) Process(p []float64) error {
	if m.Mute {
		clear(p)
	}
	return nil
}
