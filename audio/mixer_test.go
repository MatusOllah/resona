package audio_test

import (
	"testing"

	"github.com/MatusOllah/resona/audio"
	"github.com/MatusOllah/resona/internal/testutil"
)

func TestMixer(t *testing.T) {
	src1 := audio.NewReader([]float64{0.1, 0.2, 0.3, 0.0})
	src2 := audio.NewReader([]float64{0.2, 0.3, 0.4, 0.0})
	want := []float64{0.3, 0.5, 0.7, 0.0}

	mixer := audio.NewMixer(src1)
	mixer.Add(src2)
	got := make([]float64, len(want))
	_, err := mixer.ReadSamples(got)
	if err != nil {
		t.Fatal(err)
	}

	if !testutil.EqualSliceWithinTolerance(got, want, 1e-12) {
		t.Errorf("mixer: got %v, want %v", got, want)
	}
}
