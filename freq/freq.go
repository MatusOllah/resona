// Package freq provides functionality for measuring and displaying frequency.
package freq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// A Frequency represents the frequency
// as an int64 nanohertz count. The representation limits the
// largest representable frequency to approximately 9.223GHz.
type Frequency int64

const (
	minFrequency Frequency = -1 << 63
	maxFrequency Frequency = 1<<63 - 1
)

// Common frequencies. There is no definition for terahertz or larger
// because of the signed 64-bit integer type's limitations.
//
// To count the number of units in a [Frequency], divide:
//
//	khz := freq.KiloHertz
//	fmt.Print(int64(khz/freq.Hertz)) // prints 1000
//
// To convert an integer number of units to a Frequency, multiply:
//
//	hertz := 10
//	fmt.Print(freq.Frequency(hertz)*freq.Hertz) // prints 10Hz
const (
	NanoHertz  Frequency = 1
	MicroHertz           = 1000 * NanoHertz
	MilliHertz           = 1000 * MicroHertz
	CentiHertz           = 10 * MilliHertz
	DeciHertz            = 100 * MilliHertz
	Hertz                = 1000 * MilliHertz
	DecaHertz            = 10 * Hertz
	HectoHertz           = 100 * Hertz
	KiloHertz            = 1000 * Hertz
	MegaHertz            = 1000 * KiloHertz
	GigaHertz            = 1000 * MegaHertz
)

// String returns a string representing the frequency in the form "12.3kHz".
func (f Frequency) String() string {
	switch {
	case f < 0:
		if -f >= 0 {
			return "-" + (-f).String()
		}
		return "-Inf Hz"
	case f < MicroHertz:
		return fmt.Sprintf("%dnHz", f)
	case f < MilliHertz:
		if f%MicroHertz == 0 {
			return fmt.Sprintf("%.3fµHz", float64(f)/float64(MicroHertz))
		}
		return fmt.Sprintf("%.6fµHz", float64(f)/float64(MicroHertz))
	case f < Hertz:
		if f%MilliHertz == 0 {
			return fmt.Sprintf("%.3fmHz", float64(f)/float64(MilliHertz))
		} else if f%MicroHertz == 0 {
			return fmt.Sprintf("%.6fmHz", float64(f)/float64(MilliHertz))
		}
		return fmt.Sprintf("%.9fmHz", float64(f)/float64(MilliHertz))
	case f < KiloHertz:
		if f%Hertz == 0 {
			return fmt.Sprintf("%.0fHz", float64(f)/float64(Hertz))
		} else if f%MilliHertz == 0 {
			return fmt.Sprintf("%.3fHz", float64(f)/float64(Hertz))
		} else if f%MicroHertz == 0 {
			return fmt.Sprintf("%.6fHz", float64(f)/float64(Hertz))
		}
		return fmt.Sprintf("%.9fHz", float64(f)/float64(Hertz))
	case f < MegaHertz:
		if f%KiloHertz == 0 {
			return fmt.Sprintf("%.0fkHz", float64(f)/float64(KiloHertz))
		} else if f%Hertz == 0 {
			return fmt.Sprintf("%.3fkHz", float64(f)/float64(KiloHertz))
		} else if f%MilliHertz == 0 {
			return fmt.Sprintf("%.6fkHz", float64(f)/float64(KiloHertz))
		} else if f%MicroHertz == 0 {
			return fmt.Sprintf("%.9fkHz", float64(f)/float64(KiloHertz))
		}
		return fmt.Sprintf("%.12fkHz", float64(f)/float64(KiloHertz))
	case f < GigaHertz:
		if f%MegaHertz == 0 {
			return fmt.Sprintf("%.0fMHz", float64(f)/float64(MegaHertz))
		} else if f%KiloHertz == 0 {
			return fmt.Sprintf("%.3fMHz", float64(f)/float64(MegaHertz))
		} else if f%Hertz == 0 {
			return fmt.Sprintf("%.6fMHz", float64(f)/float64(MegaHertz))
		} else if f%MilliHertz == 0 {
			return fmt.Sprintf("%.9fMHz", float64(f)/float64(MegaHertz))
		} else if f%MicroHertz == 0 {
			return fmt.Sprintf("%.12fMHz", float64(f)/float64(MegaHertz))
		}
		return fmt.Sprintf("%.15fGHz", float64(f)/float64(MegaHertz))
	case f < maxFrequency:
		if f%GigaHertz == 0 {
			return fmt.Sprintf("%.0fGHz", float64(f)/float64(GigaHertz))
		} else if f%MegaHertz == 0 {
			return fmt.Sprintf("%.3fGHz", float64(f)/float64(GigaHertz))
		} else if f%KiloHertz == 0 {
			return fmt.Sprintf("%.6fGHz", float64(f)/float64(GigaHertz))
		} else if f%Hertz == 0 {
			return fmt.Sprintf("%.9fGHz", float64(f)/float64(GigaHertz))
		} else if f%MilliHertz == 0 {
			return fmt.Sprintf("%.12fGHz", float64(f)/float64(GigaHertz))
		} else if f%MicroHertz == 0 {
			return fmt.Sprintf("%.15fGHz", float64(f)/float64(GigaHertz))
		}
		return fmt.Sprintf("%.18fGHz", float64(f)/float64(GigaHertz))
	case f >= maxFrequency:
		return "Inf Hz"
	}
	return fmt.Sprintf("%dnHz", f)
}

// NanoHertz returns the frequency as an integer nanohertz count.
func (f Frequency) NanoHertz() int64 {
	return int64(f)
}

// MicroHertz returns the frequency as an integer microhertz count.
func (f Frequency) MicroHertz() int64 {
	return int64(f) / 1e3
}

// MilliHertz returns the frequency as an integer millihertz count.
func (f Frequency) MilliHertz() int64 {
	return int64(f) / 1e6
}

// These methods return float64 because the dominant
// use case is for printing a floating point number like 1.5Hz, and
// a truncation to integer would make them not useful in those cases.
// Splitting the integer and fraction ourselves guarantees that
// converting the returned float64 to an integer rounds the same
// way that a pure integer conversion would have, even in cases
// where, say, float64(f.NanoHertz())/1e9 would have rounded
// differently.

// Hertz returns the frequency as a floating point number of hertz.
func (f Frequency) Hertz() float64 {
	hz := f / Hertz
	nhz := f % Hertz
	return float64(hz) + float64(nhz)/1e9
}

// KiloHertz returns the frequency as a floating point number of kilohertz.
func (f Frequency) KiloHertz() float64 {
	khz := f / KiloHertz
	nhz := f % KiloHertz
	return float64(khz) + float64(nhz)/1e12
}

// MegaHertz returns the frequency as a floating point number of megahertz.
func (f Frequency) MegaHertz() float64 {
	mhz := f / MegaHertz
	nhz := f % MegaHertz
	return float64(mhz) + float64(nhz)/1e15
}

// GigaHertz returns the frequency as a floating point number of gigahertz.
func (f Frequency) GigaHertz() float64 {
	ghz := f / GigaHertz
	nhz := f % GigaHertz
	return float64(ghz) + float64(nhz)/1e18
}

// Truncate returns the result of rounding f toward zero to a multiple of m.
// If m <= 0, Truncate returns f unchanged.
func (f Frequency) Truncate(m Frequency) Frequency {
	if m <= 0 {
		return f
	}
	return f - f%m
}

// lessThanHalf reports whether x+x < y but avoids overflow,
// assuming x and y are both positive (Frequency is signed).
func lessThanHalf(x, y Frequency) bool {
	return uint64(x)+uint64(x) < uint64(y)
}

// Round returns the result of rounding f to the nearest multiple of m.
// The rounding behavior for halfway values is to round away from zero.
// If the result exceeds the maximum (or minimum)
// value that can be stored in a [Frequency],
// Round returns the maximum (or minimum) frequency.
// If m <= 0, Round returns f unchanged.
func (f Frequency) Round(m Frequency) Frequency {
	if m <= 0 {
		return f
	}
	r := f % m
	if f < 0 {
		r = -r
		if lessThanHalf(r, m) {
			return f + r
		}
		if f1 := f - m + r; f1 < f {
			return f1
		}
		return minFrequency // overflow
	}
	if lessThanHalf(r, m) {
		return f - r
	}
	if f1 := f + m - r; f1 > f {
		return f1
	}
	return maxFrequency // overflow
}

// Abs returns the absolute value of f.
// As a special case, Frequency([math.MinInt64]) is converted to Frequency([math.MaxInt64]),
// reducing its magnitude by 1 nanohertz.
func (f Frequency) Abs() Frequency {
	switch {
	case f >= 0:
		return f
	case f == minFrequency:
		return maxFrequency
	default:
		return -f
	}
}

// FromPeriod returns the frequency whose period is p.
func FromPeriod(p time.Duration) Frequency {
	return Frequency((1e9 * time.Second) / p)
}

// Period returns the period of time of 1 cycle at frequency f.
func (f Frequency) Period() time.Duration {
	return (1e9 * time.Second) / time.Duration(f)
}

func (f Frequency) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(f))
}

func (f *Frequency) UnmarshalJSON(b []byte) error {
	var i int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*f = Frequency(i)
	return nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, rem, errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, rem, errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var unitMap = map[string]uint64{
	"nHz":  uint64(NanoHertz),
	"uHz":  uint64(MicroHertz),
	"µHz":  uint64(MicroHertz), // U+00B5 = micro symbol
	"μHz":  uint64(MicroHertz), // U+03BC = Greek letter mu
	"mHz":  uint64(MilliHertz),
	"cHz":  uint64(CentiHertz),
	"dHz":  uint64(DeciHertz),
	"Hz":   uint64(Hertz),
	"daHz": uint64(DecaHertz),
	"hHz":  uint64(HectoHertz),
	"kHz":  uint64(KiloHertz),
	"MHz":  uint64(MegaHertz),
	"GHz":  uint64(GigaHertz),
}

// Parse parses a frequency string.
// A frequency string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300Hz", "-1.5kHz" or "2kHz45Hz".
// Valid frequency units are "nHz", "µHz" (and friends), "mHz", "cHz", "dHz", "Hz", "daHz", "hHz", "kHz", "MHz", and "GHz".
func Parse(s string) (Frequency, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var frq uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("freq: invalid frequency " + strconv.Quote(orig))
	}
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if s[0] != '.' && (s[0] < '0' || s[0] > '9') {
			return 0, errors.New("freq: invalid frequency " + strconv.Quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("freq: invalid frequency " + strconv.Quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".Hz" or "-.Hz")
			return 0, errors.New("freq: invalid frequency " + strconv.Quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("freq: missing unit in frequency " + strconv.Quote(orig))
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("freq: unknown unit " + strconv.Quote(u) + " in frequency " + strconv.Quote(orig))
		}
		if v > 1<<63/unit {
			// overflow
			return 0, errors.New("freq: missing unit in frequency " + strconv.Quote(orig))
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanohertz accurate for fractions of gigahertz.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (nHz/GHz, GHz is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, errors.New("freq: missing unit in frequency " + strconv.Quote(orig))
			}
		}
		frq += v
		if frq > 1<<63 {
			return 0, errors.New("freq: missing unit in frequency " + strconv.Quote(orig))
		}
	}
	if neg {
		return -Frequency(frq), nil
	}
	if frq > 1<<63-1 {
		return 0, errors.New("freq: missing unit in frequency " + strconv.Quote(orig))
	}
	return Frequency(frq), nil
}
