package wav

import "fmt"

// GUID represents a GUID.
type GUID struct {
	Data1 int32
	Data2 int16
	Data3 int16
	Data4 [8]byte
}

// String returns the string representation of the GUID.
func (g GUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uint32(g.Data1),
		uint16(g.Data2),
		uint16(g.Data3),
		g.Data4[0], g.Data4[1],
		g.Data4[2], g.Data4[3], g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7],
	)
}

// WAVE formats.
const (
	FormatInt   = 1      // PCM Integer
	FormatFloat = 3      // IEEE Float
	FormatAlaw  = 6      // A-Law
	FormatUlaw  = 7      // U-Law
	FormatWAVEX = 0xFFFE // WAVE_FORMAT_EXTENSIBLE
)

// WAVEX subformat GUIDs.
var (
	GuidInt   = GUID{0x00000001, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}} // PCM Integer
	GuidFloat = GUID{0x00000003, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}} // IEEE Float
	GuidAlaw  = GUID{0x00000006, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}} // A-law
	GuidUlaw  = GUID{0x00000007, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}} // U-law
	GuidDFPWM = GUID{0x38fac13a, 0x1d81, 0x6143, [8]byte{0xa4, 0x0d, 0xce, 0x53, 0xca, 0x60, 0x7c, 0xd1}} // DFPWM
)
