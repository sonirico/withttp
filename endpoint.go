package withttp

import (
	"context"
	"io"
	"net/url"
)

type (
	header interface {
		SetHeader(k, v string)
		AddHeader(k, v string)
		Header(k string) (string, bool)
		RangeHeaders(func(string, string))
	}

	Request interface {
		header

		Method() string
		SetMethod(string)

		SetURL(*url.URL)
		// SetBodyStream sets the stream of body data belonging to a request. bodySize parameter is needed
		// when using fasthttp implementation.
		SetBodyStream(rc io.ReadWriteCloser, bodySize int)
		SetBody([]byte)

		Body() []byte
		BodyStream() io.ReadWriteCloser

		URL() *url.URL
	}

	Response interface {
		header

		Status() int
		StatusText() string
		Body() io.ReadCloser

		SetBody(rc io.ReadCloser)
		SetStatus(status int)
	}

	Client interface {
		Request(ctx context.Context) (Request, error)
		Do(ctx context.Context, req Request) (Response, error)
	}

	Endpoint struct {
		name string

		requestOpts []ReqOption

		responseOpts []ResOption
	}

	MockEndpoint struct{}

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

func (e *Endpoint) Request(opts ...ReqOption) *Endpoint {
	e.requestOpts = append(e.requestOpts, opts...)
	return e
}

func (e *Endpoint) Response(opts ...ResOption) *Endpoint {
	e.responseOpts = append(e.responseOpts, opts...)
	return e
}

func NewEndpoint(name string) *Endpoint {
	return &Endpoint{name: name}
}
