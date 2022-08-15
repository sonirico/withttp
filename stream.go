package withttp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
)

type (
	Stream[T any] interface {
		Next(ctx context.Context) bool
		Data() T
		Err() error
	}

	StreamFactory[T any] interface {
		Get(r io.Reader) Stream[T]
	}

	StreamFactoryFunc[T any] func(reader io.Reader) Stream[T]
)

func (f StreamFactoryFunc[T]) Get(r io.Reader) Stream[T] {
	return f(r)
}

type (
	JSONEachRowStream[T any] struct {
		current T

		scanner *bufio.Scanner

		err error
	}
)

func (s *JSONEachRowStream[T]) Next(_ context.Context) bool {
	if !s.scanner.Scan() {
		return false
	}

	s.err = json.Unmarshal(s.scanner.Bytes(), &s.current)

	return true
}

func (s *JSONEachRowStream[T]) Data() T {
	return s.current
}

func (s *JSONEachRowStream[T]) Err() error {
	if s.err != nil {
		return s.err
	}

	return s.scanner.Err()
}

func NewJSONEachRowStream[T any](r io.Reader) Stream[T] {
	return &JSONEachRowStream[T]{
		scanner: bufio.NewScanner(r),
	}
}

func NewJSONEachRowStreamFactory[T any]() StreamFactory[T] {
	return StreamFactoryFunc[T](func(r io.Reader) Stream[T] {
		return NewJSONEachRowStream[T](r)
	})
}
