package withttp

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
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
