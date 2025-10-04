package generator_test

import (
	"testing"

	"github.com/MatusOllah/resona/generator"
	"github.com/MatusOllah/resona/internal/testutil"
)

func TestConstant(t *testing.T) {
	want := []float32{0.39}

	c := generator.NewConstant(0.39)
	got := make([]float32, len(want))
	_, err := c.ReadSamples(got)
	if err != nil {
		t.Fatal(err)
	}

	if !testutil.EqualSliceWithinTolerance(got, want, 1e-12) {
		t.Errorf("constant: got %v, want %v", got, want)
	}
}
