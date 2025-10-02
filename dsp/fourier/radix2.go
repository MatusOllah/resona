package fourier

import "math"

func radix2FFT(x []complex128) {
	N := len(x)

	// bit-reversal permutation
	for i, j := 0, 0; i < N; i++ {
		if i < j {
			x[i], x[j] = x[j], x[i]
		}
		m := N >> 1
		for m >= 1 && j >= m {
			j -= m
			m >>= 1
		}
		j += m
	}

	// Cooley-Tukey FFT
	for len := 2; len <= N; len <<= 1 {
		theta := -2 * math.Pi / float64(len)
		wlen := complex(math.Cos(theta), math.Sin(theta)) // twiddle factor
		for i := 0; i < N; i += len {
			w := complex(1, 0)
			for j := 0; j < len/2; j++ {
				u := x[i+j]
				v := x[i+j+len/2] * w
				x[i+j] = u + v
				x[i+j+len/2] = u - v
				w *= wlen
			}
		}
	}
}
