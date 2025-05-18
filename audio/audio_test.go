package audio_test

import (
	"reflect"
	"testing"

	"github.com/MatusOllah/resona/audio"
)

func TestInterleave(t *testing.T) {
	tests := []struct {
		name  string
		input [][]float64
		want  []float64
	}{
		{
			name:  "Stereo",
			input: [][]float64{{0.1, 0.2}, {0.3, 0.4}, {0.5, 0.6}},
			want:  []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
		},
		{
			name:  "Mono",
			input: [][]float64{{0.1}, {0.2}, {0.3}},
			want:  []float64{0.1, 0.2, 0.3},
		},
		{
			name:  "EmptyInput",
			input: [][]float64{},
			want:  []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := audio.Interleave(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Interleave() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeinterleave(t *testing.T) {
	tests := []struct {
		name        string
		input       []float64
		numChannels int
		want        [][]float64
		shouldPanic bool
	}{
		{
			name:        "Stereo",
			input:       []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			numChannels: 2,
			want: [][]float64{
				{0.1, 0.3, 0.5}, //audio
				{0.2, 0.4, 0.6}, // Right
			},
		},
		{
			name:        "Mono",
			input:       []float64{0.1, 0.2, 0.3},
			numChannels: 1,
			want: [][]float64{
				{0.1, 0.2, 0.3},
			},
		},
		{
			name:        "EmptyInput",
			input:       []float64{},
			numChannels: 2,
			want: [][]float64{
				{}, {},
			},
		},
		{
			name:        "InvalidChannelCount",
			input:       []float64{0.1, 0.2},
			numChannels: 0,
			shouldPanic: true,
		},
		{
			name:        "LengthNotDivisibleByNumChannels",
			input:       []float64{0.1, 0.2, 0.3},
			numChannels: 2,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Deinterleave() did not panic as expected")
					}
				}()
			}

			got := audio.Deinterleave(tt.input, tt.numChannels)
			if !tt.shouldPanic && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deinterleave() = %v, want %v", got, tt.want)
			}
		})
	}
}
