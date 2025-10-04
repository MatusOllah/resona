package generator

// Silence is a generator that generates silence.
//
// The zero value for Silence is ready to use.
type Silence struct{}

func (s *Silence) ReadSamples(p []float32) (int, error) {
	clear(p)
	return len(p), nil
}
