package freq_test

import (
	"testing"

	"github.com/MatusOllah/resona/freq"
)

func TestFrequency_String(t *testing.T) {
	tests := []struct {
		str string
		f   freq.Frequency
	}{
		{"0nHz", 0},
		{"1nHz", 1 * freq.NanoHertz},
		{"1.100000µHz", 1100 * freq.NanoHertz},
		{"2.200000mHz", 2200 * freq.MicroHertz},
		{"3.300Hz", 3300 * freq.MilliHertz},
		{"4.005kHz", 4*freq.KiloHertz + 5*freq.Hertz},
		{"4.005001kHz", 4*freq.KiloHertz + 5001*freq.MilliHertz},
		{"5.006MHz", 5*freq.MegaHertz + 6*freq.KiloHertz},
		{"8.000001MHz", 8*freq.MegaHertz + 1*freq.Hertz},
		{"2.400GHz", 2400 * freq.MegaHertz},
		{"Inf Hz", 1<<63 - 1},
		{"-Inf Hz", -1 << 63},
	}

	for _, test := range tests {
		if str := test.f.String(); str != test.str {
			t.Errorf("Frequency(%d).String() = %s, want %s", int64(test.f), str, test.str)
		}
		if test.f > 0 {
			if str := (-test.f).String(); str != "-"+test.str {
				t.Errorf("Frequency(%d).String() = %s, want %s", int64(-test.f), str, "-"+test.str)
			}
		}
	}
}
