package withttp

func (c *Call[T]) WithLogger(l logger) *Call[T] {
	c.logger = l
	return c
}
