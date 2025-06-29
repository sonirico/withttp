package withttp

import "github.com/sonirico/withttp/csvparser"

func (c *Call[T]) Response(opts ...ResOption) *Call[T] {
	c.resOptions = append(c.resOptions, opts...)
	return c
}

func (c *Call[T]) ReadBody() *Call[T] {
	return c.withRes(ParseBodyRaw[T]())
}

func (c *Call[T]) ParseStreamChan(factory StreamFactory[T], ch chan<- T) *Call[T] {
	return c.withRes(ParseStreamChan[T](factory, ch))
}

func (c *Call[T]) ParseStream(factory StreamFactory[T], fn func(T) bool) *Call[T] {
	return c.withRes(ParseStream[T](factory, fn))
}

func (c *Call[T]) ParseJSONEachRowChan(out chan<- T) *Call[T] {
	return c.ParseStreamChan(NewJSONEachRowStreamFactory[T](), out)
}

func (c *Call[T]) ParseJSONEachRow(fn func(T) bool) *Call[T] {
	return c.ParseStream(NewJSONEachRowStreamFactory[T](), fn)
}

func (c *Call[T]) ParseCSV(
	ignoreLines int,
	parser csvparser.Parser[T],
	fn func(T) bool,
) *Call[T] {
	return c.ParseStream(NewCSVStreamFactory[T](ignoreLines, parser), fn)
}

func (c *Call[T]) IgnoreResponseBody() *Call[T] {
	return c.withRes(IgnoredBody[T]())
}

func (c *Call[T]) ParseJSON() *Call[T] {
	return c.withRes(ParseJSON[T]())
}

func (c *Call[T]) Assert(fn func(req Response) error) *Call[T] {
	return c.withRes(Assertion[T](fn))
}

func (c *Call[T]) ExpectedStatusCodes(states ...int) *Call[T] {
	return c.withRes(ExpectedStatusCodes[T](states...))
}
