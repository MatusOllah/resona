// Package qoa implements decoding of QOA (Quite OK Audio) files.
package qoa

// https://qoaformat.org/qoa-specification.pdf

const (
	lmsLen         = 4
	sliceLen       = 20
	maxChannels    = 8
	slicesPerFrame = 256
	magic          = "qoaf"
)

var reciprocalTab = [16]int{
	65536, 9363, 3121, 1457, 781, 475, 311, 216, 156, 117, 90, 71, 57, 47, 39, 32,
}

var quantTab = [17]int8{
	7, 7, 7, 5, 5, 3, 3, 1, /* -8..-1 */
	0,                      /*  0     */
	0, 2, 2, 4, 4, 6, 6, 6, /*  1.. 8 */
}

var dequantTab = [16][8]int16{
	{1, -1, 3, -3, 5, -5, 7, -7},
	{5, -5, 18, -18, 32, -32, 49, -49},
	{16, -16, 53, -53, 95, -95, 147, -147},
	{34, -34, 113, -113, 203, -203, 315, -315},
	{63, -63, 210, -210, 378, -378, 588, -588},
	{104, -104, 345, -345, 621, -621, 966, -966},
	{158, -158, 528, -528, 950, -950, 1477, -1477},
	{228, -228, 760, -760, 1368, -1368, 2128, -2128},
	{316, -316, 1053, -1053, 1895, -1895, 2947, -2947},
	{422, -422, 1405, -1405, 2529, -2529, 3934, -3934},
	{548, -548, 1828, -1828, 3290, -3290, 5117, -5117},
	{696, -696, 2320, -2320, 4176, -4176, 6496, -6496},
	{868, -868, 2893, -2893, 5207, -5207, 8099, -8099},
	{1064, -1064, 3548, -3548, 6386, -6386, 9933, -9933},
	{1286, -1286, 4288, -4288, 7718, -7718, 12005, -12005},
	{1536, -1536, 5120, -5120, 9216, -9216, 14336, -14336},
}

func frameSize(channels, slices uint32) uint32 {
	return 8 + lmsLen*4*channels + 8*slices*channels
}

type lms struct {
	history [lmsLen]int16
	weights [lmsLen]int16
}

func (l *lms) predict() int {
	var p int
	for n := range lmsLen {
		p += int(l.history[n]) * int(l.weights[n])
	}
	p >>= 13
	return p
}

func (l *lms) update(sample, residual int16) {
	delta := residual >> 4
	for i := range lmsLen {
		if l.history[i] < 0 {
			l.weights[i] -= delta
		} else {
			l.weights[i] += delta
		}
	}
	for i := range lmsLen - 1 {
		l.history[i] = l.history[i+1]
	}
	l.history[lmsLen-1] = sample
}

func clamp(v, min, max int) int {
	if v <= min {
		return min
	}
	if v >= max {
		return max
	}
	return v
}

func clampS16(v int) int16 {
	if uint(v+32768) > 65535 {
		if v < -32768 {
			return -32768
		}
		if v > 32767 {
			return 32767
		}
	}
	return int16(v)
}

func div(v, scaleFactor int) int {
	reciprocal := reciprocalTab[scaleFactor]
	n := (v*reciprocal + (1 << 15)) >> 16
	n += (btoi(v > 0) - btoi(v < 0)) - (btoi(n > 0) - btoi(n < 0)) // round away from 0
	return n
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
