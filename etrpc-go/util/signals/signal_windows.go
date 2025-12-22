// Package signals implements helpers for signal handling.
package signals

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
