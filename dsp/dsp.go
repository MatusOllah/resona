// Package dsp provides digital signal processing (DSP) math primitives.
package dsp

// Clamp clamps the value x to the range [-1, 1].
func Clamp(x float32) float32 {
	return max(-1, min(1, x))
}

// ToComplexSlice converts a slice of float32 to a slice of complex64 with zero imaginary parts.
func ToComplexSlice(f []float32) []complex64 {
	c := make([]complex64, len(f))
	for i := range f {
		c[i] = complex(f[i], 0)
	}
	return c
}

// ToFloatSlice converts a slice of complex64 to a slice of float32 by taking the real parts.
func ToFloatSlice(c []complex64) []float32 {
	f := make([]float32, len(c))
	for i := range c {
		f[i] = real(c[i])
	}
	return f
}
