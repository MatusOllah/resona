package fourier

import (
	"math/cmplx"
)

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// FFTInPlace computes the Fast Fourier Transform of the input slice x.
// This is done in-place (modifying the input slice).
//
// The length of x must be a power of two. If not, the function will panic.
func FFTInPlace(x []complex128) {
	if !isPowerOfTwo(len(x)) {
		panic("fourier FFTInPlace: input length is not a power of two")
	}

	radix2FFT(x)
}

// IFFTInPlace computes the Inverse Fast Fourier Transform of the input slice x.
// This is done in-place (modifying the input slice).
//
// The length of x must be a power of two.
func IFFTInPlace(x []complex128) {
	if !isPowerOfTwo(len(x)) {
		panic("fourier IFFTInPlace: input length is not a power of two")
	}
	N := len(x)
	for i := range x {
		x[i] = cmplx.Conj(x[i])
	}
	FFTInPlace(x)
	for i := range x {
		x[i] = cmplx.Conj(x[i]) / complex(float64(N), 0)
	}
}

// FFT computes the Fast Fourier Transform of the input slice x
// and returns a new slice containing the result.
//
// The length of x must be a power of two. If not, the function will panic.
func FFT(x []complex128) []complex128 {
	f := make([]complex128, len(x))
	copy(f, x)
	FFTInPlace(f)
	return f
}

// IFFT computes the Inverse Fast Fourier Transform of the input slice x
// and returns a new slice containing the result.
//
// The length of x must be a power of two. If not, the function will panic.
func IFFT(x []complex128) []complex128 {
	f := make([]complex128, len(x))
	copy(f, x)
	IFFTInPlace(f)
	return f
}

// Convolve computes the circular convolution of two input slices x and y.
// The lengths of x and y must be equal and a power of two.
func Convolve(x, y []complex128) []complex128 {
	if len(x) != len(y) {
		panic("fourier Convolve: input lengths do not match")
	}

	Fx := FFT(x)
	Fy := FFT(y)

	r := make([]complex128, len(x))
	for i := range r {
		r[i] = Fx[i] * Fy[i]
	}
	return IFFT(r)
}
