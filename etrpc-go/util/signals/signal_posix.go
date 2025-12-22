//go:build !windows
// +build !windows

// Package signals implements helpers for signal handling.
package signals

import (
	"os"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
