package lutmath

import "math"

const (
	sinLength = 1 << 13
	sinStep   = (2 * math.Pi) / sinLength
	sinMask   = sinLength - 1
)

var sinTable [sinLength]float64

func init() {
	for i := range sinTable {
		sinTable[i] = math.Sin(float64(i) * sinStep)
	}
}

// SineLUTSize returns the size of the sine LUT in bytes.
func SineLUTSize() int {
	return sinLength * 8
}

// Sin returns the sine of the radian argument x using a LUT with linear interpolation.
func Sin(x float64) float64 {
	// convert radians to table index
	pos := x / sinStep

	i := int(pos)
	frac := pos - float64(i)
	if frac < 0 {
		frac += 1
		i--
	}

	i &= sinMask
	iNext := (i + 1) & sinMask

	return sinTable[i]*(1-frac) + sinTable[iNext]*frac // linear interpolation
}

// Sincos returns Sin(x), Cos(x).
func Sincos(x float64) (sin, cos float64) {
	return Sin(x), Cos(x)
}

// Cos returns the cosine of the radian argument x.
func Cos(x float64) float64 {
	return Sin(x + (math.Pi / 2))
}

// Tan returns the tangent of the radian argument x.
func Tan(x float64) float64 {
	return Sin(x) / Cos(x)
}
