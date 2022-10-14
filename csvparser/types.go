package csvparser

import (
	"io"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sonirico/stadio/slices"
)

type (
	Type[T any] interface {
		Parse(data []byte) (T, int, error)
		//Compile(x T, writer io.Writer) error
	}

	StringType struct {
		sep   byte
		quote byte
	}

	IntegerType struct {
		inner StringType
	}
)

// Parse parses `data`, which is ensured to be non-nil and its length greater than zero
func (s StringType) Parse(data []byte) (string, int, error) {
	if s.quote != QuoteNone {
		if data[0] != s.quote {
			return "", 0, errors.Wrapf(ErrQuoteExpected, "<%s>", string(s.quote))
		}

		i := 3
		// Find the next non-escaped quote
		for i < len(data) {
			prev := data[i-2]
			middle := data[i-1]
			next := data[i]
			if middle == s.quote && prev != '\\' && next == s.sep {
				break
			}
			i++
		}

		payload := data[1 : i-1]
		return string(payload), len(payload), nil
	}

	payload := data

	idx := slices.IndexOf(data, func(x byte) bool { return x == s.sep })
	if idx > -1 {
		// next separator has not been found. End of line?
		payload = data[:idx]
	}

	return string(payload), len(payload), nil
}

func (s StringType) Compile(data []byte, w io.Writer) error {
	if s.quote != QuoteNone {

		n, err := w.Write(quote)
		if err != nil || n < 1 {
			// todo: handle
		}
	}

	n, err := w.Write(data)

	if err != nil || n < 1 {
		// todo: handle
	}

	if s.quote != QuoteNone {
		n, err := w.Write(quote)
		if err != nil || n < 1 {
			// todo: handle
		}
	}
	return nil
}

func (i IntegerType) Parse(data []byte) (int, int, error) {
	val, n, err := i.inner.Parse(data)
	if err != nil {
		return 0, n, err
	}

	var res int64
	res, err = strconv.ParseInt(val, 10, 64)
	return int(res), n, err
}

func StrType(quote, sep byte) StringType {
	return StringType{quote: quote, sep: sep}
}

func IntType(quote, sep byte) IntegerType {
	return IntegerType{inner: StrType(quote, sep)}
}
