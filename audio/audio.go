// Package audio implements functions for the manipulation of sample slices and audio.
package audio

// Interleave takes a channel-separated audio sample slice, where each channel is a slice of samples,
// and combines them into a single interleaved slice.
//
// For example:
//
//	[][]float64{{L0, R0}, {L1, R1}, {L2, R2}, ...}
//
// becomes:
//
//	[]float64{L0, R0, L1, R1, L2, R2, ...}
func Interleave(samples [][]float64) []float64 {
	out := []float64{}
	for i := range samples {
		out = append(out, samples[i]...)
	}
	return out
}

// Deinterleave takes an interleaved audio slice and separates it into an individual channel-separated slice.
//
// For example:
//
//	[]float64{L0, R0, L1, R1, L2, R2, ...}
//
// becomes:
//
//	[][]float64{{L0, R0}, {L1, R1}, {L2, R2}, ...}
func Deinterleave(interleaved []float64, numChannels int) [][]float64 {
	if numChannels <= 0 {
		panic("audio: number of channels must be positive")
	}
	if len(interleaved)%numChannels != 0 {
		panic("audio: interleaved slice length is not divisible by number of channels")
	}

	totalFrames := len(interleaved) / numChannels
	out := make([][]float64, numChannels)
	for ch := range out {
		out[ch] = make([]float64, totalFrames)
	}

	for i := range totalFrames {
		for ch := range numChannels {
			out[ch][i] = interleaved[i*numChannels+ch]
		}
	}

	return out
}
