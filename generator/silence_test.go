package generator_test

import (
	"testing"

	"github.com/MatusOllah/resona/generator"
	"github.com/MatusOllah/resona/internal/testutil"
)

func TestSilence(t *testing.T) {
	want := make([]float64, 4)

	silence := &generator.Silence{}

	got := make([]float64, 4)
	_, err := silence.ReadSamples(got)
	if err != nil {
		t.Fatal(err)
	}

	if !testutil.EqualSliceWithinTolerance(want, got, 1e-12) {
		t.Errorf("Silence.ReadSamples() = %v, want silence", got)
	}
}
