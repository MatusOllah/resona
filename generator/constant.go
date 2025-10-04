package generator

// Constant generates a constant audio signal.
//
// The zero value for Constant is ready to use.
type Constant struct {
	// Value is the costant value to be outputted.
	Value float32
}

// NewConstant creates a new [Constant] generator.
//
// In most cases, new([Constant]) (or just declaring a [Constant] variable) is sufficient
// to create a new [Constant].
func NewConstant(val float32) *Constant {
	return &Constant{Value: val}
}

func (c *Constant) ReadSamples(p []float32) (n int, err error) {
	for i := range p {
		p[i] = c.Value
	}
	return len(p), nil
}
