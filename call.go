package withttp

import (
	"bytes"
	"context"
	"io"
	"strconv"
)

type (
	CallResOption[T any] interface {
		Parse(c *Call[T], r Response) error
	}

	CallResOptionFunc[T any] func(c *Call[T], res Response) error

	Call[T any] struct {
		client client

		reqOptions []ReqOption
		resOptions []ResOption

		Req Request
		Res Response

		BodyRaw    []byte
		BodyParsed T

		ReqBodyRaw []byte
	}
)

func (f CallResOptionFunc[T]) Parse(c *Call[T], res Response) error {
	return f(c, res)
}

func NewCall[T any](client client) *Call[T] {
	return &Call[T]{client: client}
}

func (c *Call[T]) bodyReader(res Response) (rc io.ReadCloser) {
	if c.BodyRaw != nil {
		rc = io.NopCloser(bytes.NewReader(c.BodyRaw))
	} else {
		rc = res.Body()
	}
	return
}

func (c *Call[T]) withRes(fn CallResOption[T]) *Call[T] {
	c.resOptions = append(
		c.resOptions,
		ResOptionFunc(func(res Response) (err error) {
			return fn.Parse(c, res)
		}),
	)
	return c
}

func (c *Call[T]) withReq(fn ReqOption) *Call[T] {
	c.reqOptions = append(
		c.reqOptions,
		fn,
	)
	return c
}

func (c *Call[T]) parseRes(res Response) error {
	for _, opt := range c.resOptions {
		if err := opt.Parse(res); err != nil {
			return err
		}
	}
	return nil
}

func (c *Call[T]) configureReq(req Request) error {
	for _, opt := range c.reqOptions {
		if err := opt.Configure(req); err != nil {
			return err
		}
	}
	return nil
}

func (c *Call[T]) Request(opts ...ReqOption) *Call[T] {
	for _, opt := range opts {
		c.withReq(opt)
	}
	return c
}

func (c *Call[T]) Response(opts ...ResOption) *Call[T] {
	c.resOptions = append(c.resOptions, opts...)
	return c
}

func (c *Call[T]) Call(ctx context.Context, e *Endpoint) (err error) {
	req, err := c.client.Request()

	if err != nil {
		return
	}

	for _, opt := range e.requestOpts {
		if err = opt.Configure(req); err != nil {
			return err
		}
	}

	if err = c.configureReq(req); err != nil {
		return
	}

	res, err := c.client.Do(ctx, req)

	if err != nil {
		return
	}

	for _, opt := range e.responseOpts {
		if err = opt.Parse(res); err != nil {
			return
		}
	}

	if err = c.parseRes(res); err != nil {
		return
	}

	return
}

func (c *Call[T]) WithURL(raw string) *Call[T] {
	return c.withReq(WithURL(raw))
}

func (c *Call[T]) WithURI(raw string) *Call[T] {
	return c.withReq(WithURI(raw))
}

func (c *Call[T]) WithMethod(method string) *Call[T] {
	return c.withReq(
		ReqOptionFunc(func(req Request) error {
			req.SetMethod(method)
			return nil
		}),
	)
}

// WithBodyStream receives a stream of data to set on the request. Second parameter `bodySize` indicates
// the estimated content-length of this stream. Required when employing fasthttp http client.
func (c *Call[T]) WithBodyStream(rc io.ReadCloser, bodySize int) *Call[T] {
	return c.withReq(
		ReqOptionFunc(func(req Request) error {
			req.SetBodyStream(rc, bodySize)
			return nil
		}),
	)
}

func (c *Call[T]) WithRawBody(payload []byte) *Call[T] {
	return c.withReq(
		ReqOptionFunc(func(req Request) error {
			req.SetBody(payload)
			return nil
		}),
	)
}

func (c *Call[T]) WithContentLength(length int) *Call[T] {
	return c.WithHeader("content-length", strconv.FormatInt(int64(length), 10), true)
}

func (c *Call[T]) WithHeader(key, value string, override bool) *Call[T] {
	return c.withReq(
		ReqOptionFunc(func(req Request) error {
			return ConfigureHeader(req, key, value, override)
		}),
	)
}

func (c *Call[T]) WithHeaderFunc(fn func() (key, value string, override bool)) *Call[T] {
	return c.withReq(
		ReqOptionFunc(func(req Request) error {
			key, value, override := fn()
			return ConfigureHeader(req, key, value, override)
		}),
	)
}

func (c *Call[T]) WithReadBody() *Call[T] {
	return c.withRes(WithRawBody[T]())
}

func (c *Call[T]) WithStreamChan(factory StreamFactory[T], ch chan<- T) *Call[T] {
	return c.withRes(WithStreamChan[T](factory, ch))
}

func (c *Call[T]) WithStream(factory StreamFactory[T], fn func(T) bool) *Call[T] {
	return c.withRes(WithStream[T](factory, fn))
}

func (c *Call[T]) WithJSONEachRowChan(out chan<- T) *Call[T] {
	return c.WithStreamChan(NewJSONEachRowStreamFactory[T](), out)
}

func (c *Call[T]) WithJSONEachRow(fn func(T) bool) *Call[T] {
	return c.WithStream(NewJSONEachRowStreamFactory[T](), fn)
}

func (c *Call[T]) WithIgnoreBody() *Call[T] {
	return c.withRes(WithIgnoredBody[T]())
}

func (c *Call[T]) WithJSON() *Call[T] {
	return c.withRes(WithJSON[T]())
}

func (c *Call[T]) WithAssert(fn func(req Response) error) *Call[T] {
	return c.withRes(WithAssertion[T](fn))
}

func (c *Call[T]) WithExpectedStatusCodes(states ...int) *Call[T] {
	return c.withRes(WithExpectedStatusCodes[T](states...))
}
