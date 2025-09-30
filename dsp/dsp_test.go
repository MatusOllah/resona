package dsp_test

import (
	"testing"

	"github.com/MatusOllah/resona/dsp"
	"github.com/MatusOllah/resona/internal/testutil"
)

func TestClamp(t *testing.T) {
	want := 1.0
	if got := dsp.Clamp(39.0); !testutil.EqualWithinTolerance(want, got, 1e-12) {
		t.Errorf("Clamp(39.0) = %v; want %v", got, want)
	}
}
