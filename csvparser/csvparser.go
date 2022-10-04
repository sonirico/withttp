package csvparser

import (
	"io"
	"strconv"

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
	// Header represents a set of columns definitions
	Header struct {
	}

	StringColumn[T any] struct {
		inner  StringType
		setter func(x *T, v string)
		getter func(x T) string
	}

	IntColumn[T any] struct {
		inner  IntegerType
		setter func(x *T, v int)
		getter func(x T) int
	}

	BoolColumn[T any] struct {
		inner  StringType
		setter func(x T, v bool)
		getter func(x T) bool
	}

	Type[T any] interface {
		Parse(data []byte) (T, error)
		//Compile(x T, writer io.Writer) error
	}

	Col[T any] interface {
		Parse(data []byte, item *T) error
		//Compile(x T, writer io.Writer) error
	}

	StringType struct {
		Quoted bool
	}

	IntegerType struct {
		inner StringType
	}

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

func (s StringColumn[T]) Parse(data []byte, item *T) error {
	val, err := s.inner.Parse(data)
	if err != nil {
		return err
	}
	s.setter(item, val)
	return nil
}

func (c IntColumn[T]) Parse(data []byte, item *T) error {
	val, err := c.inner.Parse(data)
	if err != nil {
		return err
	}
	c.setter(item, val)
	return nil
}

func (s StringType) Parse(data []byte) (string, error) {
	if s.Quoted {
		// ,"",
		if len(data) > 2 {
			return string(data[1 : len(data)-1]), nil // todo: nalloc
		}

		return "", nil
		// todo: assert data[0] == quote. Keep until other quote+separator are found
	}

	return string(data), nil
}

func (s StringType) Compile(data []byte, w io.Writer) error {
	if s.Quoted {

		n, err := w.Write(quote)
		if err != nil || n < 1 {
			// todo: handle
		}
	}

	n, err := w.Write(data)

	if err != nil || n < 1 {
		// todo: handle
	}

	if s.Quoted {
		n, err := w.Write(quote)
		if err != nil || n < 1 {
			// todo: handle
		}
	}
	return nil
}

func (i IntegerType) Parse(data []byte) (int, error) {
	val, err := i.inner.Parse(data)
	if err != nil {
		return 0, err
	}

	var res int64
	res, err = strconv.ParseInt(val, 10, 64)
	return int(res), err
}

func StringCol[T any](
	quoted bool,
	getter func(T) string,
	setter func(*T, string),
) Col[T] {
	return StringColumn[T]{
		inner:  StringType{Quoted: quoted},
		getter: getter, setter: setter,
	}
}

func IntCol[T any](
	quoted bool,
	getter func(T) int,
	setter func(*T, int),
) Col[T] {
	return IntColumn[T]{
		inner:  IntegerType{inner: StringType{Quoted: quoted}},
		getter: getter, setter: setter,
	}
}
