// Package driver provides the interface for playback drivers.
package driver

import (
	"io"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
)

// Driver is the interface that playback drivers must implement.
type Driver interface {
	// Init initializes the driver with the given format and buffer size.
	Init(format afmt.Format, src aio.SampleReader) error

	io.Closer
}
