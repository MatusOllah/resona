package au

// Encoding types.
const (
	Ulaw        uint32 = 1  // G.711 μ-law 8-bit
	LPCMInt8    uint32 = 2  // Linear PCM 8-bit integer
	LPCMInt16   uint32 = 3  // Linear PCM 16-bit integer
	LPCMInt24   uint32 = 4  // Linear PCM 24-bit integer
	LPCMInt32   uint32 = 5  // Linear PCM 32-bit integer
	LPCMFloat32 uint32 = 6  // Linear PCM 32-bit float
	LPCMFloat64 uint32 = 7  // Linear PCM 64-bit float
	Alaw        uint32 = 27 // G.711 A-law 8-bit
)
