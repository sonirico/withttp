package withttp

import (
	"io"
	"strconv"
)

type (
	StreamCallReqOption[T any] interface {
		CallReqOption[T]

		stream()
	}

	StreamCallReqOptionFunc[T any] func(c *Call[T], req Request) error
)

func (s StreamCallReqOptionFunc[T]) stream() {}

func (s StreamCallReqOptionFunc[T]) Configure(c *Call[T], req Request) error {
	return s(c, req)
}

func (c *Call[T]) Request(opts ...ReqOption) *Call[T] {
	c.reqOptions = append(c.reqOptions, opts...)
	return c
}

func (c *Call[T]) URL(raw string) *Call[T] {
	return c.withReq(URL[T](raw))
}

func (c *Call[T]) URI(raw string) *Call[T] {
	return c.withReq(URI[T](raw))
}

func (c *Call[T]) Method(method string) *Call[T] {
	return c.withReq(Method[T](method))
}

func (c *Call[T]) Query(k, v string) *Call[T] {
	return c.withReq(Query[T](k, v))
}

func (c *Call[T]) RequestSniffed(fn func([]byte, error)) *Call[T] {
	return c.withReq(RequestSniffer[T](fn))
}

func (c *Call[T]) RequestStreamBody(opt StreamCallReqOptionFunc[T]) *Call[T] {
	return c.withReq(opt)
}

// BodyStream receives a stream of data to set on the request. Second parameter `bodySize` indicates
// the estimated content-length of this stream. Required when employing fasthttp http client.
func (c *Call[T]) BodyStream(rc io.ReadWriteCloser, bodySize int) *Call[T] {
	return c.withReq(BodyStream[T](rc, bodySize))
}

func (c *Call[T]) Body(payload any) *Call[T] {
	return c.withReq(Body[T](payload))
}

func (c *Call[T]) RawBody(payload []byte) *Call[T] {
	return c.withReq(RawBody[T](payload))
}

func (c *Call[T]) ContentLength(length int) *Call[T] {
	return c.Header("content-length", strconv.FormatInt(int64(length), 10), true)
}

func (c *Call[T]) Header(key, value string, override bool) *Call[T] {
	return c.withReq(Header[T](key, value, override))
}

func (c *Call[T]) HeaderFunc(fn func() (key, value string, override bool)) *Call[T] {
	return c.withReq(HeaderFunc[T](fn))
}

func (c *Call[T]) BasicAuth(user, pass string) *Call[T] {
	return c.withReq(BasicAuth[T](user, pass))
}

func (c *Call[T]) ContentType(ct string) *Call[T] {
	return c.withReq(ContentType[T](ct))
}
