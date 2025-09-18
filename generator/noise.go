package generator

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand/v2"
)

// Noise represents a RNG noise generator.
type Noise struct {
	rng *rand.Rand
}

// NewNoise creates a new [Noise] generator with the default RNG source.
func NewNoise() *Noise {
	var seed1, seed2 uint64
	_ = binary.Read(crand.Reader, binary.BigEndian, &seed1)
	_ = binary.Read(crand.Reader, binary.BigEndian, &seed2)

	return &Noise{
		rng: rand.New(rand.NewPCG(seed1, seed2)),
	}
}

// NewNoiseWithSource creates a new [Noise] generator with the provided RNG source.
func NewNoiseWithSource(src rand.Source) *Noise {
	return &Noise{
		rng: rand.New(src),
	}
}

func (n *Noise) ReadSamples(p []float64) (int, error) {
	for i := range p {
		p[i] = n.rng.Float64()*2 - 1
	}
	return len(p), nil
}
