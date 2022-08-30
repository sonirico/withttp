package withttp

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/url"
)

func WithCloseBody[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		return
	}
}

func WithIgnoredBody[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		_, err = io.Copy(io.Discard, rc)
		return
	}
}

func WithParseBodyRaw[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		c.BodyRaw, err = io.ReadAll(rc)
		return
	}
}

func WithJSON[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		c.BodyParsed, err = ReadJSON[T](c.bodyReader(res))
		return
	}
}

func WithStream[T any](factory StreamFactory[T], fn func(T) bool) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		return ReadStream[T](c.bodyReader(res), factory, fn)
	}
}

func WithStreamChan[T any](factory StreamFactory[T], out chan<- T) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		return ReadStreamChan(c.bodyReader(res), factory, out)
	}
}

func WithExpectedStatusCodes[T any](states ...int) CallResOptionFunc[T] {
	return WithAssertion[T](func(res Response) error {
		found := false
		for _, status := range states {
			if status == res.Status() {
				found = true
				break
			}
		}

		if !found {
			return errors.Wrapf(ErrUnexpectedStatusCode, "want: %v, have: %d", states, res.Status())
		}

		return nil
	})
}

func WithAssertion[T any](fn func(res Response) error) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) error {
		if err := fn(res); err != nil {
			return errors.Wrapf(ErrAssertion, err.Error())
		}

		return nil
	}
}

func WithMockedRes(fn func(response Response)) ResOption {
	return ResOptionFunc(func(res Response) (err error) {
		fn(res)
		return
	})
}

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

func ReadStreamChan[T any](rc io.ReadCloser, factory StreamFactory[T], out chan<- T) (err error) {
	defer func() {
		close(out)
	}()
	err = ReadStream[T](rc, factory, func(item T) bool {
		out <- item
		return true
	})

	return
}

func ReadStream[T any](rc io.ReadCloser, factory StreamFactory[T], fn func(T) bool) (err error) {
	defer func() { _ = rc.Close() }()

	stream := factory.Get(rc)
	keep := true

	for keep && stream.Next(nil) {
		if err = stream.Err(); err != nil {
			return
		}

		keep = fn(stream.Data())
	}

	return
}

func ReadJSON[T any](rc io.ReadCloser) (res T, err error) {
	defer func() { _ = rc.Close() }()

	if err = json.NewDecoder(rc).Decode(&res); err != nil {
		return
	}

	return
}
