package afmt_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/freq"
)

func TestNumSamplesToDuration(t *testing.T) {
	sr := 48 * freq.KiloHertz

	tests := []struct {
		name string
		n    int
		want time.Duration
	}{
		{"ZeroSamples", 0, 0},
		{"OneSample", 1, time.Second / 48000},
		{"48000Samples", 48000, time.Second},
		{"24000Samples", 24000, time.Second / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := afmt.NumSamplesToDuration(sr, tt.n)
			if got != tt.want {
				t.Errorf("NumSamplesToDuration(%v, %d) = %v; want %v", sr, tt.n, got, tt.want)
			}
		})
	}
}

func TestDurationToNumSamples(t *testing.T) {
	sr := 48 * freq.KiloHertz

	tests := []struct {
		name string
		d    time.Duration
		want int
	}{
		{"ZeroDuration", 0, 0},
		{"OneSecond", time.Second, 48000},
		{"HalfSecond", time.Second / 2, 24000},
		{"1500Microseconds", 1500 * time.Microsecond, 72}, // 48000 * 0.0015 = 72
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := afmt.DurationToNumSamples(sr, tt.d)
			if got != tt.want {
				t.Errorf("DurationToNumSamples(%v, %v) = %d; want %d", sr, tt.d, got, tt.want)
			}
		})
	}
}

func TestSampleEncoding_IsSigned(t *testing.T) {
	tests := []struct {
		name string
		e    afmt.SampleEncoding
		want bool
	}{
		{"SampleEncodingUnknown", afmt.SampleEncodingUnknown, false},
		{"SampleEncodingInt", afmt.SampleEncodingInt, true},
		{"SampleEncodingUint", afmt.SampleEncodingUint, false},
		{"SampleEncodingFloat", afmt.SampleEncodingFloat, false},
		{"SampleEncodingDSD", afmt.SampleEncodingDSD, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsSigned(); got != tt.want {
				t.Errorf("IsSigned() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestSampleEncoding_IsFloat(t *testing.T) {
	tests := []struct {
		name string
		e    afmt.SampleEncoding
		want bool
	}{
		{"SampleEncodingUnknown", afmt.SampleEncodingUnknown, false},
		{"SampleEncodingInt", afmt.SampleEncodingInt, false},
		{"SampleEncodingUint", afmt.SampleEncodingUint, false},
		{"SampleEncodingFloat", afmt.SampleEncodingFloat, true},
		{"SampleEncodingDSD", afmt.SampleEncodingDSD, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsFloat(); got != tt.want {
				t.Errorf("IsFloat() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestSampleEncoding_IsInt(t *testing.T) {
	tests := []struct {
		name string
		e    afmt.SampleEncoding
		want bool
	}{
		{"SampleEncodingUnknown", afmt.SampleEncodingUnknown, false},
		{"SampleEncodingInt", afmt.SampleEncodingInt, true},
		{"SampleEncodingUint", afmt.SampleEncodingUint, true},
		{"SampleEncodingFloat", afmt.SampleEncodingFloat, false},
		{"SampleEncodingDSD", afmt.SampleEncodingDSD, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsInt(); got != tt.want {
				t.Errorf("IsInt() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestSampleFormat_BytesPerSample(t *testing.T) {
	tests := []struct {
		name     string
		format   afmt.SampleFormat
		expected int
	}{
		{"Int16", afmt.SampleFormat{BitDepth: 16, Encoding: afmt.SampleEncodingInt, Endian: binary.LittleEndian}, 2},
		{"Int24", afmt.SampleFormat{BitDepth: 24, Encoding: afmt.SampleEncodingInt, Endian: binary.BigEndian}, 3},
		{"Uint8", afmt.SampleFormat{BitDepth: 8, Encoding: afmt.SampleEncodingUint, Endian: nil}, 1},
		{"Float32", afmt.SampleFormat{BitDepth: 32, Encoding: afmt.SampleEncodingFloat, Endian: binary.LittleEndian}, 4},
		{"DSD", afmt.SampleFormat{BitDepth: 1, Encoding: afmt.SampleEncodingDSD, Endian: nil}, 1},
		{"ZeroBitDepth", afmt.SampleFormat{BitDepth: 0, Encoding: afmt.SampleEncodingInt}, 0},
		{"NegativeBitDepth", afmt.SampleFormat{BitDepth: -8, Encoding: afmt.SampleEncodingInt}, 0},
		{"UnknownEncoding", afmt.SampleFormat{BitDepth: 16, Encoding: afmt.SampleEncodingUnknown}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.BytesPerSample()
			if got != tt.expected {
				t.Errorf("BytesPerSample() = %d; want %d", got, tt.expected)
			}
		})
	}
}

func TestSampleFormat_BytesPerFrame(t *testing.T) {
	format := afmt.SampleFormat{BitDepth: 16, Encoding: afmt.SampleEncodingInt, Endian: binary.LittleEndian}

	tests := []struct {
		name        string
		numChannels int
		expected    int
	}{
		{"Mono", 1, 2},
		{"Stereo", 2, 4},
		{"5_1Surround", 6, 12},
		{"ZeroChannels", 0, 0},
		{"NegativeChannels", -1, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.BytesPerFrame(tt.numChannels)
			if got != tt.expected {
				t.Errorf("BytesPerFrame(%d) = %d; want %d", tt.numChannels, got, tt.expected)
			}
		})
	}
}
