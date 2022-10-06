package csvparser

import (
	"bytes"
)

var (
	quote = []byte{byte('"')}

	QuoteDouble byte = '"'
	QuoteSimple byte = '\''
	QuoteNone   byte = 0

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
	data = bytes.TrimSpace(data) // cleanup phase
	sepLen := 1                  // len(p.separator)

	for i, col := range p.columns {
		var read int
		read, err = col.Parse(data, item)
		if err != nil {
			return
		}

		// TODO: handle read =0
		_ = i

		if read > len(data) {
			break
		}

		// create a cursor to have better readability under the fact the column types will only parse
		// its desired data, letting the parser have the liability to advance de cursor.
		cursor := read
		if read+sepLen <= len(data) {
			cursor += sepLen
		}

		data = data[cursor:]
	}
	return nil
}

func New[T any](sep byte, cols ...ColFactory[T]) Parser[T] {
	columns := make([]Col[T], len(cols))
	opt := opts{sep: sep}

	for i, c := range cols {
		columns[i] = c(opt)
	}
	return Parser[T]{separator: sep, columns: columns}
}
