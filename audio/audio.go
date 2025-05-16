// Package audio implements functions for the manipulation of sample slices and audio.
package audio

func Interleave(samples [][]float64) []float64 {
	out := []float64{}
	for i := range samples {
		out = append(out, samples[i]...)
	}
	return out
}

func Deinterleave(interleaved []float64, numChannels int) [][]float64 {
	if numChannels <= 0 {
		panic("audio: number of channels must be positive")
	}

	out := make([][]float64, numChannels)
	totalFrames := len(interleaved) / numChannels

	for i := range out {
		out[i] = make([]float64, totalFrames)
	}

	for i := range totalFrames {
		for ch := range numChannels {
			out[ch][i] = interleaved[i*numChannels+ch]
		}
	}

	return out
}
