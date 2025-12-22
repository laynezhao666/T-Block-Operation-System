// Package errcode defines error code
package errcode

const (
	ErrGetConfigTimeout = iota + 270100
	ErrGetConfigFail

	ErrNorthTBoxNtp
	ErrNorthTBoxTime

	ErrGwMode

	ErrMissLocalFile
	ErrServerLogic
	ErrBadRequest
)
