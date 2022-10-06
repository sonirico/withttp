package csvparser

type (
	Col[T any] interface {
		Parse(data []byte, item *T) (int, error)
		//Compile(x T, writer io.Writer) error
	}

	opts struct {
		sep byte
	}

	ColFactory[T any] func(opts) Col[T]

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
)

func (s StringColumn[T]) Parse(data []byte, item *T) (int, error) {
	val, n, err := s.inner.Parse(data)
	if err != nil {
		return n, err
	}
	s.setter(item, val)
	return n, nil
}

func (c IntColumn[T]) Parse(data []byte, item *T) (int, error) {
	val, n, err := c.inner.Parse(data)
	if err != nil {
		return n, err
	}
	c.setter(item, val)
	return n, nil
}

func StringCol[T any](
	quote byte,
	getter func(T) string,
	setter func(*T, string),
) ColFactory[T] {
	return func(opts opts) Col[T] {
		return StringColumn[T]{
			inner:  StrType(quote, opts.sep),
			getter: getter, setter: setter,
		}
	}
}

func IntCol[T any](
	quote byte,
	getter func(T) int,
	setter func(*T, int),
) ColFactory[T] {
	return func(opts opts) Col[T] {
		return IntColumn[T]{
			inner:  IntType(quote, opts.sep),
			getter: getter, setter: setter,
		}
	}
}
