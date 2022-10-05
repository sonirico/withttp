package csvparser

type (
	Col[T any] interface {
		Parse(data []byte, item *T) error
		//Compile(x T, writer io.Writer) error
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
)

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

func StringCol[T any](
	quoted bool,
	getter func(T) string,
	setter func(*T, string),
) Col[T] {
	return StringColumn[T]{
		inner:  StrType(quoted),
		getter: getter, setter: setter,
	}
}

func IntCol[T any](
	quoted bool,
	getter func(T) int,
	setter func(*T, int),
) Col[T] {
	return IntColumn[T]{
		inner:  IntType(quoted),
		getter: getter, setter: setter,
	}
}
