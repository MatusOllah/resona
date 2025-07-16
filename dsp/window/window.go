// Package window implements window functions for digital signal processing (DSP).
package window

import (
	"errors"
	"math"
)

// WindowFunc represents a window function.
type WindowFunc func(int) []float64

// Apply applies a window function to s.
// It returns an error if the window function returns a slice of incorrect length.
func Apply(s []float64, fn WindowFunc) error {
	w := fn(len(s))
	if len(w) != len(s) {
		return errors.New("window: window function returned slice of incorrect length")
	}
	for i := range s {
		s[i] *= w[i]
	}
	return nil
}

// MustApply is like [Apply] but it panics on error.
// Use this only when you're sure the window function is valid.
func MustApply(s []float64, fn WindowFunc) {
	if err := Apply(s, fn); err != nil {
		panic(err)
	}
}

// ApplyTo applies a window function to src and writes the result to dst.
// It returns an error if dst and src are of different lengths, or the window function
// returns a slice of incorrect length.
func ApplyTo(dst, src []float64, fn WindowFunc) error {
	if len(dst) != len(src) {
		return errors.New("window: dst and src must have the same length")
	}

	w := fn(len(src))
	if len(w) != len(src) {
		return errors.New("window: window function returned slice of incorrect length")
	}

	for i := range src {
		dst[i] = src[i] * w[i]
	}
	return nil
}

// MustApplyTo is like ApplyTo but it panics on error.
// Use only when you're sure dst, src, and the window function are valid.
func MustApplyTo(dst, src []float64, fn WindowFunc) {
	if err := ApplyTo(dst, src, fn); err != nil {
		panic(err)
	}
}

// Rectangular returns an n-point rectangular window. All values here are 1.
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Rectangular_window
func Rectangular(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)
	for i := range w {
		w[i] = 1
	}
	return w
}

// Welch returns an n-point Welch window.
//
// Special case is:
//
//	Welch(1) = []float64{1.0}
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Welch_window
func Welch(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	for i := range w {
		w[i] = 1 - math.Pow((float64(i)-(float64(n-1)/2))/(float64(n)/2), 2)
	}
	return w
}

// Hann returns an n-point Hann window.
//
// Special case is:
//
//	Hann(1) = []float64{1.0}
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Hann_window
func Hann(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	for i := range w {
		w[i] = 0.5 * (1 - math.Cos((2*math.Pi*float64(i))/(float64(n-1))))
	}
	return w
}

// Hamming returns an n-point Hamming window.
//
// Special case is:
//
//	Hamming(1) = []float64{1.0}
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Hamming_window
func Hamming(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	for i := range w {
		w[i] = 0.54 - 0.46*math.Cos((2*math.Pi*float64(i))/(float64(n-1)))
	}
	return w
}

// Blackman returns an n-point Blackman window.
//
// Special case is:
//
//	Blackman(1) = []float64{1.0}
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Blackman_window
func Blackman(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	const (
		a0 float64 = 0.42
		a1 float64 = 0.5
		a2 float64 = 0.08
	)

	for i := range w {
		w[i] = a0 - a1*math.Cos((2*math.Pi*float64(i))/(float64(n-1))) + a2*math.Cos((4*math.Pi*float64(i))/(float64(n-1)))
	}
	return w
}

// Blackman returns an n-point exact Blackman window.
//
// Special case is:
//
//	Blackman(1) = []float64{1.0}
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Blackman_window
func ExactBlackman(n int) []float64 {
	if n <= 0 {
		return nil
	}

	w := make([]float64, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	const (
		a0 float64 = 7938.0 / 18608.0
		a1 float64 = 9240.0 / 18608.0
		a2 float64 = 1430.0 / 18608.0
	)

	for i := range w {
		w[i] = a0 - a1*math.Cos((2*math.Pi*float64(i))/(float64(n-1))) + a2*math.Cos((4*math.Pi*float64(i))/(float64(n-1)))
	}
	return w
}
