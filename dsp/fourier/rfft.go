package fourier

import "math/cmplx"

// RFFT computes the Real-input Fast Fourier Transform of the input slice x
// and returns a new slice containing the result.
//
// The length of x must be a power of two. If not, the function will panic.
func RFFT(x []float32) []complex64 {
	N := len(x)
	if !isPowerOfTwo(N) {
		panic("fourier RFFT: input length is not a power of two")
	}

	// convert real => complex
	c := make([]complex64, N)
	for i := range x {
		c[i] = complex(x[i], 0)
	}

	FFTInPlace(c)

	return c[:N/2+1] // return only the good bins
}

// IRFFT computes the Inverse Real-input Fast Fourier Transform of the input slice x
// and returns a new slice containing the result.
//
// The length of x must be (N/2)+1 where N is a power of two. If not, the function will panic.
func IRFFT(x []complex64) []float32 {
	N := (len(x) - 1) * 2
	if !isPowerOfTwo(N) {
		panic("fourier IRFFT: input length is not a power of two")
	}

	// rebuild full Hermitian spectrum
	full := make([]complex64, N)
	copy(full[:len(x)], x)
	for i := 1; i < N/2; i++ {
		full[N-i] = complex64(cmplx.Conj(complex128(x[i])))
	}

	IFFTInPlace(full)

	// extract real part
	out := make([]float32, N)
	for i := range out {
		out[i] = real(full[i])
	}
	return out
}
