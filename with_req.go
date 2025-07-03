package withttp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"sync"

	"github.com/pkg/errors"

	"github.com/sonirico/withttp/codec"
)

func Header[T any](k, v string, override bool) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) error {
		return ConfigureHeader(req, k, v, override)
	}
}

func BasicAuth[T any](user, pass string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) error {
		header, err := CreateAuthorizationHeader(authHeaderKindBasic, user, pass)
		if err != nil {
			return err
		}
		return ConfigureHeader(req, "authorization", header, true)
	}
}

func HeaderFunc[T any](fn func() (string, string, bool)) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) error {
		k, v, override := fn()
		return ConfigureHeader(req, k, v, override)
	}
}

func ContentType[T any](ct string) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqContentType = ct
		return ConfigureHeader(req, "content-type", ct, true)
	}
}

type authHeaderKind string

var (
	authHeaderKindBasic authHeaderKind = "Basic"
)

func (a authHeaderKind) Codec() func(...string) (string, error) {
	switch a {
	case authHeaderKindBasic:
		return func(s ...string) (string, error) {
			if len(s) < 2 {
				return "", errors.Wrapf(ErrAssertion, "header kind: %s", a)
			}
			user := s[0]
			pass := s[1]

			return base64.StdEncoding.EncodeToString(S2B(user + ":" + pass)), nil
		}
	default:
		panic("unknown auth header kind")
	}
}

func CreateAuthorizationHeader(kind authHeaderKind, user, pass string) (string, error) {
	fn := kind.Codec()
	header, err := fn(user, pass)
	if err != nil {
		return header, err
	}
	return fmt.Sprintf("%s %s", kind, header), nil
}

func ConfigureHeader(req Request, key, value string, override bool) error {
	if override {
		req.SetHeader(key, value)
	} else {
		req.AddHeader(key, value)
	}
	return nil
}

func Method[T any](method string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetMethod(method)
		return
	}
}

func Query[T any](k, v string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		u := req.URL()
		qs := u.Query()
		qs.Set(k, v)
		u.RawQuery = qs.Encode()
		req.SetURL(u)
		return
	}
}

func URL[T any](raw string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}

		req.SetURL(u)

		return
	}
}

func URI[T any](raw string) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetURL(req.URL().JoinPath(raw))
		return
	}
}

func RawBody[T any](payload []byte) CallReqOptionFunc[T] {
	return func(_ *Call[T], req Request) (err error) {
		req.SetBody(payload)
		return nil
	}
}

func Body[T any](payload any) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) (err error) {
		data, err := EncodeBody(payload, c.ReqContentType)
		if err != nil {
			return err
		}
		req.SetBody(data)
		return nil
	}
}

func RequestSniffer[T any](fn func([]byte, error)) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqShouldSniff = true
		c.ReqStreamSniffer = fn
		return nil
	}
}

func RequestStreamBody[T, U any](r rangeable[U]) StreamCallReqOptionFunc[T] {
	return func(c *Call[T], req Request) error {
		c.ReqIsStream = true

		buf := closableReaderWriter{ReadWriter: bytes.NewBuffer(nil)} // TODO: pool buffer
		req.SetBodyStream(buf, -1)                                    // TODO: bodySize

		c.ReqStreamWriter = func(ctx context.Context, c *Call[T], req Request, wg *sync.WaitGroup) (err error) {
			defer func() { wg.Done() }()

			var encoder codec.Encoder
			if r.Serialize() {
				encoder, err = ContentTypeCodec(c.ReqContentType)

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

func BodyStream[T any](rc io.ReadWriteCloser, bodySize int) CallReqOptionFunc[T] {
	return func(c *Call[T], req Request) (err error) {
		req.SetBodyStream(rc, bodySize)
		return nil
	}
}

func EncodeBody(payload any, contentType string) (bts []byte, err error) {
	encoder, err := ContentTypeCodec(contentType)
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
