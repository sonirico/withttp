package withttp

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"sync"

	"github.com/sonirico/withttp/codec"
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

func WithRequestSniffer[T any](fn func([]byte, error)) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqShouldSniff = true
		c.ReqStreamSniffer = fn
		return nil
	}
}

func WithRequestStreamBody[T, U any](r rangeable[U]) StreamCallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqIsStream = true

		buf := closableReaderWriter{ReadWriter: bytes.NewBuffer(nil)} // TODO: pool buffer
		req.SetBodyStream(buf, -1)                                    // TODO: bodySize

		c.ReqStreamWriter = func(ctx context.Context, c *Call[T], req Request, wg *sync.WaitGroup) (err error) {
			defer func() { wg.Done() }()

			var encoder codec.Encoder
			if r.Serialize() {
				encoder, err = c.ReqContentType.Codec()

				if err != nil {
					return
				}
			} else {
				encoder = codec.ProxyBytesEncoder
			}

			var sniffer func([]byte, error)

			if c.ReqShouldSniff {
				sniffer = c.ReqStreamSniffer
			} else {
				sniffer = func(_ []byte, _ error) {}
			}

			err = EncodeStream(ctx, r, req, encoder, sniffer)

			return
		}
		return nil
	}
}

func WithBodyStream[T any](rc io.ReadWriteCloser, bodySize int) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) (err error) {
		req.SetBodyStream(rc, bodySize)
		return nil
	}
}

func EncodeBody(payload any, contentType ContentType) (bts []byte, err error) {
	encoder, err := contentType.Codec()
	if err != nil {
		return
	}
	bts, err = encoder.Encode(payload)
	return
}

func EncodeStream[T any](
	ctx context.Context,
	r rangeable[T],
	req Request,
	encoder codec.Encoder,
	sniffer func([]byte, error),
) (err error) {

	stream := req.BodyStream()

	defer func() { _ = stream.Close() }()

	var bts []byte

	r.Range(func(i int, x T) bool {
		defer func() {
			sniffer(bts, err)
		}()

		select {
		case <-ctx.Done():
			return false
		default:
			if bts, err = encoder.Encode(x); err != nil {
				return false
			}

			if _, err = stream.Write(bts); err != nil {
				return false
			}

			return true
		}
	})

	return
}
