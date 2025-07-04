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
		logger logger

		client Client

		reqOptions []ReqOption // TODO: Linked Lists
		resOptions []ResOption

		Req Request
		Res Response

		BodyRaw    []byte
		BodyParsed T

		ReqContentType string
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

func NewCall[T any](client Client) *Call[T] {
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
	return c.callEndpoint(ctx, nil)
}

func (c *Call[T]) CallEndpoint(ctx context.Context, e *Endpoint) (err error) {
	return c.callEndpoint(ctx, e)
}

func (c *Call[T]) callEndpoint(ctx context.Context, e *Endpoint) (err error) {
	req, err := c.client.Request(ctx)
	defer func() { c.Req = req }()

	if err != nil {
		return
	}

	if e != nil {
		for _, opt := range e.requestOpts {
			if err = opt.Configure(req); err != nil {
				return err
			}
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

	c.log("[withttp] %s %s", req.Method(), req.URL().String())

	res, err := c.client.Do(ctx, req)

	if c.ReqIsStream {
		wg.Wait()
	}

	if err != nil {
		return
	}

	if res != nil {
		c.log("[withttp] server returned status code %d", res.Status())
	}

	defer func() { c.Res = res }()

	if e != nil {
		for _, opt := range e.responseOpts {
			if err = opt.Parse(res); err != nil {
				return
			}
		}
	}

	if err = c.parseRes(res); err != nil {
		return
	}

	return
}

func (c *Call[T]) log(tpl string, args ...any) {
	if c.logger == nil {
		return
	}

	c.logger.Printf(tpl, args...)
}
