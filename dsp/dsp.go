// Package dsp provides digital signal processing (DSP) math primitives.
package dsp

import "math"

// Clamp clamps the value x to the range [-1, 1].
func Clamp(x float64) float64 {
	return math.Max(-1, math.Min(1, x))
}
