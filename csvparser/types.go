package csvparser

import (
	"io"
	"strconv"
)

type (
	Type[T any] interface {
		Parse(data []byte) (T, error)
		//Compile(x T, writer io.Writer) error
	}

	StringType struct {
		quoted bool
	}

	IntegerType struct {
		inner StringType
	}
)

func (s StringType) Parse(data []byte) (string, error) {
	if s.quoted {
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
	if s.quoted {

		n, err := w.Write(quote)
		if err != nil || n < 1 {
			// todo: handle
		}
	}

	n, err := w.Write(data)

	if err != nil || n < 1 {
		// todo: handle
	}

	if s.quoted {
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

func StrType(quoted bool) StringType {
	return StringType{quoted: quoted}
}

func IntType(quoted bool) IntegerType {
	return IntegerType{inner: StrType(quoted)}
}
