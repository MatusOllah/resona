package effect

// Mute mutes the audio signal. While muted, it outputs silence.
type Mute struct {
	// Mute, if true, silences the audio signal.
	Mute bool
}

func (m *Mute) Process(p []float64) error {
	if m.Mute {
		clear(p)
	}
	return nil
}
