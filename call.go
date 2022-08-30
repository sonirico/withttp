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

	CalReqOption[T any] interface {
		Configure(c *Call[T], r Request) error
	}

	CallReqOptionFunc[T any] func(c *Call[T], res Request) error

	CallResOptionFunc[T any] func(c *Call[T], res Response) error

	Call[T any] struct {
		client client

		reqOptions []ReqOption // TODO: Linked Lists
		resOptions []ResOption

		Req Request
		Res Response

		BodyRaw    []byte
		BodyParsed T

		ReqContentType ContentType
		ReqBodyRaw     []byte
	}
)

func (f CallResOptionFunc[T]) Parse(c *Call[T], res Response) error {
	return f(c, res)
}

func (f CallReqOptionFunc[T]) Configure(c *Call[T], req Request) error {
	return f(c, req)
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

func (c *Call[T]) withReq(fn CallReqOptionFunc[T]) *Call[T] {
	c.reqOptions = append(
		c.reqOptions,
		ReqOptionFunc(func(req Request) error {
			return fn.Configure(c, req)
		}),
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
	c.reqOptions = append(c.reqOptions, opts...)
	return c
}

func (c *Call[T]) Response(opts ...ResOption) *Call[T] {
	c.resOptions = append(c.resOptions, opts...)
	return c
}

func (c *Call[T]) Call(ctx context.Context) (err error) {
	req, err := c.client.Request()
	defer func() { c.Req = req }()

	if err != nil {
		return
	}

	if err = c.configureReq(req); err != nil {
		return
	}

	res, err := c.client.Do(ctx, req)

	if err != nil {
		return
	}

	defer func() { c.Res = res }()

	if err = c.parseRes(res); err != nil {
		return
	}

	return
}

func (c *Call[T]) CallEndpoint(ctx context.Context, e *Endpoint) (err error) {
	req, err := c.client.Request()
	defer func() { c.Req = req }()

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

	defer func() { c.Res = res }()

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
	return c.withReq(WithURL[T](raw))
}

func (c *Call[T]) WithURI(raw string) *Call[T] {
	return c.withReq(WithURI[T](raw))
}

func (c *Call[T]) WithMethod(method string) *Call[T] {
	return c.withReq(WithMethod[T](method))
}

// WithBodyStream receives a stream of data to set on the request. Second parameter `bodySize` indicates
// the estimated content-length of this stream. Required when employing fasthttp http client.
func (c *Call[T]) WithBodyStream(rc io.ReadCloser, bodySize int) *Call[T] {
	return c.withReq(WithBodyStream[T](rc, bodySize))
}

func (c *Call[T]) WithBody(payload any) *Call[T] {
	return c.withReq(WithBody[T](payload))
}

func (c *Call[T]) WithRawBody(payload []byte) *Call[T] {
	return c.withReq(WithRawBody[T](payload))
}

func (c *Call[T]) WithContentLength(length int) *Call[T] {
	return c.WithHeader("content-length", strconv.FormatInt(int64(length), 10), true)
}

func (c *Call[T]) WithHeader(key, value string, override bool) *Call[T] {
	return c.withReq(WithHeader[T](key, value, override))
}

func (c *Call[T]) WithHeaderFunc(fn func() (key, value string, override bool)) *Call[T] {
	return c.withReq(WithHeaderFunc[T](fn))
}

func (c *Call[T]) WithContentType(ct ContentType) *Call[T] {
	return c.withReq(WithContentType[T](ct))
}

func (c *Call[T]) WithReadBody() *Call[T] {
	return c.withRes(WithParseBodyRaw[T]())
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

func (c *Call[T]) WithParseJSON() *Call[T] {
	return c.withRes(WithJSON[T]())
}

func (c *Call[T]) WithAssert(fn func(req Response) error) *Call[T] {
	return c.withRes(WithAssertion[T](fn))
}

func (c *Call[T]) WithExpectedStatusCodes(states ...int) *Call[T] {
	return c.withRes(WithExpectedStatusCodes[T](states...))
}
