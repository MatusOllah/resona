package generator

import (
	"math"

	"github.com/MatusOllah/resona/dsp/lutmath"
	"github.com/MatusOllah/resona/freq"
)

const tau = 2 * math.Pi

// OscilWaveform is a function that maps a waveform value for a given input x in the range [0, 1).
type OscilWaveform func(x float64) float64

// SineWaveform generates a sine wave.
func SineWaveform(x float64) float64 {
	return math.Sin(tau * x)
}

// LUTSineWaveform generates a sine wave using a LUT (see lutmath package).
func LUTSineWaveform(x float64) float64 {
	return lutmath.Sin(tau * x)
}

// TriangleWaveform generates a triangle wave.
func TriangleWaveform(x float64) float64 {
	return 1 - 4*math.Abs(math.Round(x-0.25)-x+0.25)
}

// SawtoothWaveform generates a sawtooth wave.
func SawtoothWaveform(x float64) float64 {
	return 2 * (x - math.Floor(x+0.5))
}

// SquareWaveform generates a square wave.
func SquareWaveform(x float64) float64 {
	return math.Copysign(1, math.Sin(tau*x))
}

// LUTSquareWaveform generates a square wave using a LUT (see lutmath package).
func LUTSquareWaveform(x float64) float64 {
	return math.Copysign(1, lutmath.Sin(tau*x))
}

// Oscillator is a simple oscillator that generates a waveform at a specified frequency and sample rate.
type Oscillator struct {
	Frequency  freq.Frequency
	sampleRate freq.Frequency
	waveform   OscilWaveform
	t          float64
}

// NewOscillator creates a new [Oscillator].
func NewOscillator(f freq.Frequency, sampleRate freq.Frequency, waveform OscilWaveform) *Oscillator {
	return &Oscillator{
		Frequency:  f,
		sampleRate: sampleRate,
		waveform:   waveform,
	}
}

func (o *Oscillator) ReadSamples(p []float64) (int, error) {
	f := o.Frequency.Hertz() / 2 / o.sampleRate.Hertz()
	for i := range p {
		p[i] = o.waveform(o.t * f)
		o.t++
	}
	return len(p), nil
}
