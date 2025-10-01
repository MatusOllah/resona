// Package dsp provides digital signal processing (DSP) math primitives.
package dsp

import "math"

// Clamp clamps the value x to the range [-1, 1].
func Clamp(x float64) float64 {
	return math.Max(-1, math.Min(1, x))
}

// ToComplexSlice converts a slice of float64 to a slice of complex128 with zero imaginary parts.
func ToComplexSlice(f []float64) []complex128 {
	c := make([]complex128, len(f))
	for i := range f {
		c[i] = complex(f[i], 0)
	}
	return c
}

// ToFloatSlice converts a slice of complex128 to a slice of float64 by taking the real parts.
func ToFloatSlice(c []complex128) []float64 {
	f := make([]float64, len(c))
	for i := range c {
		f[i] = real(c[i])
	}
	return f
}
