package pcm_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"slices"
	"testing"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/encoding/pcm"
)

func TestPCMRoundTrip(t *testing.T) {
	samples := []float64{0.0, 0.5, -0.5, 1.0, -1.0}

	tests := []struct {
		name         string
		sampleFormat afmt.SampleFormat
	}{
		{"Int8", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 8}},
		{"Int16LE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 16, Endian: binary.LittleEndian}},
		{"Int16BE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 16, Endian: binary.BigEndian}},
		{"Int24LE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 24, Endian: binary.LittleEndian}},
		{"Int24BE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 24, Endian: binary.BigEndian}},
		{"Int32LE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 32, Endian: binary.LittleEndian}},
		{"Int32BE", afmt.SampleFormat{Encoding: afmt.SampleEncodingInt, BitDepth: 32, Endian: binary.BigEndian}},
		{"Uint8", afmt.SampleFormat{Encoding: afmt.SampleEncodingUint, BitDepth: 8}},
		{"Float32LE", afmt.SampleFormat{Encoding: afmt.SampleEncodingFloat, BitDepth: 32, Endian: binary.LittleEndian}},
		{"Float32BE", afmt.SampleFormat{Encoding: afmt.SampleEncodingFloat, BitDepth: 32, Endian: binary.BigEndian}},
		{"Float64LE", afmt.SampleFormat{Encoding: afmt.SampleEncodingFloat, BitDepth: 64, Endian: binary.LittleEndian}},
		{"Float64BE", afmt.SampleFormat{Encoding: afmt.SampleEncodingFloat, BitDepth: 64, Endian: binary.BigEndian}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			var buf bytes.Buffer
			encoder := pcm.NewEncoder(&buf, tt.sampleFormat)

			n, err := encoder.WriteSamples(samples)
			if err != nil {
				t.Fatalf("Write failed: %v", err)
			}
			if n != len(samples) {
				t.Fatalf("Expected to encode %d samples, got %d", len(samples), n)
			}

			// Decode
			decoder := pcm.NewDecoder(&buf, tt.sampleFormat)

			decodedSamples := make([]float64, len(samples))
			n, err = decoder.ReadSamples(decodedSamples)
			if err != nil {
				t.Fatalf("ReadSamples failed: %v", err)
			}
			if n != len(samples) {
				t.Fatalf("Expected to decode %d samples, got %d", len(samples), n)
			}

			// Verify
			if !slices.EqualFunc(samples, decodedSamples, func(a, b float64) bool {
				return equalWithinTolerance(a, b, 1e-2)
			}) {
				t.Errorf("Decoded samples do not match original samples: got %v, want %v", decodedSamples, samples)
			}
		})
	}
}

func equalWithinTolerance(a, b float64, epsilon float64) bool {
	if a == b {
		return true
	}

	return math.Abs(a-b) <= epsilon
}
