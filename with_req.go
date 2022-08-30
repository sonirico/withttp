package withttp

import (
	"io"
	"net/url"
)

func WithHeader[T any](k, v string, override bool) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) error {
		return ConfigureHeader(req, k, v, override)
	}
}

func WithHeaderFunc[T any](fn func() (string, string, bool)) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) error {
		k, v, override := fn()
		return ConfigureHeader(req, k, v, override)
	}
}

func WithContentType[T any](ct ContentType) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqContentType = ct
		return ConfigureHeader(req, "content-type", ct.String(), true)
	}
}

func ConfigureHeader(req Request, key, value string, override bool) error {
	if override {
		req.SetHeader(key, value)
	} else {
		req.AddHeader(key, value)
	}
	return nil
}

func WithMethod[T any](method string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetMethod(method)
		return
	}
}

func WithURL[T any](raw string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}

		req.SetURL(u)

		return
	}
}

func WithURI[T any](raw string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetURL(req.URL().JoinPath(raw))
		return
	}
}

func WithRawBody[T any](payload []byte) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetBody(payload)
		return nil
	}
}

func WithBody[T any](payload any) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) (err error) {
		data, err := EncodeBody(payload, c.ReqContentType)
		if err != nil {
			return err
		}
		req.SetBody(data)
		return nil
	}
}

func WithBodyStream[T any](rc io.ReadCloser, bodySize int) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) (err error) {
		req.SetBodyStream(rc, bodySize)
		return nil
	}
}

func EncodeBody(payload any, contentType ContentType) (bts []byte, err error) {
	codec, err := contentType.Codec()
	if err != nil {
		return
	}
	bts, err = codec.Encode(payload)
	return
}
