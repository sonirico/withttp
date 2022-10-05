package csvparser

import (
	"github.com/pkg/errors"
	"github.com/sonirico/stadio/slices"
)

var (
	quote = []byte{byte('"')}

	SeparatorComma     byte = ','
	SeparatorSemicolon byte = ';'
	SeparatorTab       byte = '\t'
)

type (
	Parser[T any] struct {
		separator byte
		columns   []Col[T]
	}
)

func (p Parser[T]) Parse(data []byte, item *T) (err error) {
	counter := 0
	for _, col := range p.columns {
		pos := slices.IndexOf[byte](
			data,
			func(x byte) bool {
				return x == p.separator
			},
		)

		lastCol := counter == len(p.columns)-1

		if pos == -1 && !lastCol {
			// Only if no more separators have been found, and current column is not the last one, yield error
			err = errors.Wrapf(
				ErrColumnMismatch,
				"want %d, have %d",
				len(p.columns),
				counter,
			)
		}

		payload := data
		if !lastCol {
			payload = data[:pos]
		}

		if err = col.Parse(payload, item); err != nil {
			return err
		}

		counter++

		if !lastCol {
			data = data[pos+1:]
		}
	}
	return nil
}

func NewParser[T any](sep byte, cols ...Col[T]) Parser[T] {
	return Parser[T]{separator: sep, columns: cols}
}
