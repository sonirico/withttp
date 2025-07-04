package withttp

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sonirico/vago/slices"

	"github.com/pkg/errors"
)

func CloseBody[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		return
	}
}

func IgnoredBody[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		_, err = io.Copy(io.Discard, rc)
		return
	}
}

func ParseBodyRaw[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		rc := c.bodyReader(res)
		defer func() { _ = rc.Close() }()
		c.BodyRaw, err = io.ReadAll(rc)
		return
	}
}

func ParseJSON[T any]() CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		c.BodyParsed, err = ReadJSON[T](c.bodyReader(res))
		return
	}
}

func ParseStream[T any](factory StreamFactory[T], fn func(T) bool) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		return ReadStream[T](c.bodyReader(res), factory, fn)
	}
}

func ParseStreamChan[T any](factory StreamFactory[T], out chan<- T) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) (err error) {
		return ReadStreamChan(c.bodyReader(res), factory, out)
	}
}

func ExpectedStatusCodes[T any](states ...int) CallResOptionFunc[T] {
	return Assertion[T](func(res Response) error {
		if slices.Includes(states, res.Status()) {
			return nil
		}
		return errors.Wrapf(ErrUnexpectedStatusCode, "want: %v, have: %d", states, res.Status())
	})
}

func Assertion[T any](fn func(res Response) error) CallResOptionFunc[T] {
	return func(c *Call[T], res Response) error {
		if err := fn(res); err != nil {
			return errors.Wrapf(ErrAssertion, err.Error())
		}

		return nil
	}
}

func MockedRes(fn func(response Response)) ResOption {
	return ResOptionFunc(func(res Response) (err error) {
		fn(res)
		return
	})
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

	for keep && stream.Next(context.TODO()) {
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
