package testutil

import (
	"math"
	"math/cmplx"
	"slices"
)

type number interface {
	float32 | float64 | complex64 | complex128
}

func EqualWithinTolerance[T number](a, b T, epsilon float64) bool {
	if a == b {
		return true
	}

	switch va := any(a).(type) {
	case float32:
		return math.Abs(float64(va)-float64(any(b).(float32))) <= epsilon
	case float64:
		return math.Abs(va-any(b).(float64)) <= epsilon
	case complex64:
		return cmplx.Abs(complex128(va)-complex128(any(b).(complex64))) <= epsilon
	case complex128:
		return cmplx.Abs(va-any(b).(complex128)) <= epsilon
	default:
		return false
	}
}

func EqualSliceWithinTolerance[T number](a, b []T, epsilon float64) bool {
	if a == nil && b == nil {
		return true
	}

	return slices.EqualFunc(a, b, func(aa, bb T) bool {
		return EqualWithinTolerance(aa, bb, epsilon)
	})
}
