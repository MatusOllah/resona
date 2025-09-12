package filter_test

import (
	"testing"

	"github.com/MatusOllah/resona/dsp/filter"
	"github.com/MatusOllah/resona/internal/testutil"
)

func TestFIRImpulseResponse(t *testing.T) {
	coeffs := []float64{0.1, 0.2, 0.3, 0.5, 0.6}
	fir := filter.NewFIR(coeffs)

	input := []float64{1, 0, 0, 0, 0} // impulse
	got := make([]float64, len(input))
	for i := range input {
		got[i] = fir.ProcessSingle(input[i])
	}

	if !testutil.EqualSliceWithinTolerance(coeffs, got, 1e-12) {
		t.Errorf("output does not equal to coefficients: want %v, got %v", coeffs, got)
	}
}
