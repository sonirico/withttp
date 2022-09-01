package withttp

import (
	"bytes"
	"context"
	"io"
	"sync"
)

type (
	CallResOption[T any] interface {
		Parse(c *Call[T], r Response) error
	}

	CallReqOption[T any] interface {
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
		ReqIsStream    bool

		ReqStreamWriter  func(ctx context.Context, c *Call[T], res Request, wg *sync.WaitGroup) error
		ReqStreamSniffer func([]byte, error)
		ReqShouldSniff   bool
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

func (c *Call[T]) withReq(fn CallReqOption[T]) *Call[T] {
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

func (c *Call[T]) Call(ctx context.Context) (err error) {
	req, err := c.client.Request()
	defer func() { c.Req = req }()

	if err != nil {
		return
	}

	if err = c.configureReq(req); err != nil {
		return
	}

	var wg *sync.WaitGroup

	if c.ReqIsStream {
		wg = &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			_ = c.ReqStreamWriter(ctx, c, req, wg)
		}()
	}

	res, err := c.client.Do(ctx, req)

	if c.ReqIsStream {
		wg.Wait()
	}

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

	var wg *sync.WaitGroup

	if c.ReqIsStream {
		wg = &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			_ = c.ReqStreamWriter(ctx, c, req, wg)
		}()
	}

	res, err := c.client.Do(ctx, req)

	if c.ReqIsStream {
		wg.Wait()
	}

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
