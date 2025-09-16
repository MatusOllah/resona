package effect

// Chain represents a sequence of effects that will be applied one after another.
// If any effect returns an error, processing stops and the error is returned.
type Chain []Effect

func (c Chain) Process(p []float64) error {
	for _, fx := range c {
		if err := fx.Process(p); err != nil {
			return err
		}
	}
	return nil
}
