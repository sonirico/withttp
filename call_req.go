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

func (c *Call[T]) WithURL(raw string) *Call[T] {
	return c.withReq(WithURL[T](raw))
}

func (c *Call[T]) WithURI(raw string) *Call[T] {
	return c.withReq(WithURI[T](raw))
}

func (c *Call[T]) WithMethod(method string) *Call[T] {
	return c.withReq(WithMethod[T](method))
}

type (
	rangeable[T any] interface {
		Range(func(int, T) bool)
	}

	Slice[T any] []T
)

func (s Slice[T]) Range(fn func(int, T) bool) {
	for i, x := range s {
		if !fn(i, x) {
			return
		}
	}
}

func (c *Call[T]) WithRequestStreamBody(opt StreamCallReqOptionFunc[T]) *Call[T] {
	return c.withReq(opt)
}

// WithBodyStream receives a stream of data to set on the request. Second parameter `bodySize` indicates
// the estimated content-length of this stream. Required when employing fasthttp http client.
func (c *Call[T]) WithBodyStream(rc io.ReadWriteCloser, bodySize int) *Call[T] {
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
