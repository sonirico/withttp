package withttp

import "github.com/sonirico/withttp/csvparser"

func (c *Call[T]) Response(opts ...ResOption) *Call[T] {
	c.resOptions = append(c.resOptions, opts...)
	return c
}

func (c *Call[T]) WithReadBody() *Call[T] {
	return c.withRes(WithParseBodyRaw[T]())
}

func (c *Call[T]) WithParseStreamChan(factory StreamFactory[T], ch chan<- T) *Call[T] {
	return c.withRes(WithParseStreamChan[T](factory, ch))
}

func (c *Call[T]) WithParseStream(factory StreamFactory[T], fn func(T) bool) *Call[T] {
	return c.withRes(WithParseStream[T](factory, fn))
}

func (c *Call[T]) WithParseJSONEachRowChan(out chan<- T) *Call[T] {
	return c.WithParseStreamChan(NewJSONEachRowStreamFactory[T](), out)
}

func (c *Call[T]) WithParseJSONEachRow(fn func(T) bool) *Call[T] {
	return c.WithParseStream(NewJSONEachRowStreamFactory[T](), fn)
}

func (c *Call[T]) WithParseCSV(
	ignoreLines int,
	parser csvparser.Parser[T],
	fn func(T) bool,
) *Call[T] {
	return c.WithParseStream(NewCSVStreamFactory[T](ignoreLines, parser), fn)
}

func (c *Call[T]) WithIgnoreResponseBody() *Call[T] {
	return c.withRes(WithIgnoredBody[T]())
}

func (c *Call[T]) WithParseJSON() *Call[T] {
	return c.withRes(WithParseJSON[T]())
}

func (c *Call[T]) WithAssert(fn func(req Response) error) *Call[T] {
	return c.withRes(WithAssertion[T](fn))
}

func (c *Call[T]) WithExpectedStatusCodes(states ...int) *Call[T] {
	return c.withRes(WithExpectedStatusCodes[T](states...))
}
