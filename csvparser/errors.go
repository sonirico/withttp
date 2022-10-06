package csvparser

import "github.com/pkg/errors"

var (
	ErrColumnMismatch = errors.New("column mismatch")
	ErrQuoteExpected  = errors.New("quote was expected")
)
