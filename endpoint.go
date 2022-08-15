package withttp

import (
	"context"
	"io"
	"net/url"
)

type (
	Request interface {
		SetMethod(string)
		SetHeader(k, v string)
		AddHeader(k, v string)
		SetURL(*url.URL)
		SetBody(rc io.ReadCloser)
	}

	Response interface {
		Status() int
		StatusText() string
		Body() io.ReadCloser

		SetBody(rc io.ReadCloser)
		SetStatus(status int)
	}

	client interface {
		Request() (Request, error)
		Do(ctx context.Context, req Request) (Response, error)
	}

	Endpoint struct {
		name string

		requestOpts []ReqOption

		responseOpts []ResOption
	}

	MockEndpoint struct {
		inner *Endpoint
	}

	ReqOption interface {
		Configure(r Request) error
	}

	ReqOptionFunc func(req Request) error

	ResOption interface {
		Parse(r Response) error
	}

	ResOptionFunc func(res Response) error
)

func (f ResOptionFunc) Parse(res Response) error {
	return f(res)
}

func (f ReqOptionFunc) Configure(req Request) error {
	return f(req)
}

func New(name string, opts ...ResOption) *Endpoint {
	e := &Endpoint{name: name}
	for _, opt := range opts {
		e.responseOpts = append(e.responseOpts, opt)
	}
	return e
}
