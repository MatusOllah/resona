package effect

// Invert inverts the audio signal.
//
// The zero value for Invert is ready to use.
type Invert struct{}

func (i *Invert) Process(p []float32) error {
	for i := range p {
		p[i] = 1 / p[i]
	}
	return nil
}
