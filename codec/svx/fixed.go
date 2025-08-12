package svx

type Fixed16_16 int32

func (f Fixed16_16) Float64() float64 {
	return float64(f) / float64((1<<16 - 1))
}
