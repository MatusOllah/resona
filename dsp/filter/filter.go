// Package filter provides fundamental DSP filter primitives.
//
// Filters here operate on a sample-by-sample basis and are intended as
// low-level building blocks for signal processing. They do not perform
// buffer-based processing, error handling, or higher-level musical effects.
//
// For buffer-oriented processing and effect chaining, see the effects package.
package filter

// Filter represents a filter.
type Filter interface {
	// ProcessSingle processes a single input sample and returns the filtered output.
	ProcessSingle(x float32) float32

	// Reset resets internal state.
	Reset()
}
