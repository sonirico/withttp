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

		inner Stream[[]byte]

		err error
	}

	NewLineStream struct {
		current []byte

		scanner *bufio.Scanner

		err error
	}

	ProxyStream struct {
		err     error
		current []byte
		buffer  []byte
		reader  io.Reader
	}
)

func (s *ProxyStream) Next(_ context.Context) bool {
	s.err = nil
	bts := s.buffer[:cap(s.buffer)]
	read, err := s.reader.Read(bts)

	if err != nil || read == 0 {
		s.err = err
		return false
	}

	s.current = make([]byte, read)
	copy(s.current, bts)
	return true
}

func (s *ProxyStream) Data() []byte {
	return s.current
}

func (s *ProxyStream) Err() error {
	return s.err
}

func (s *NewLineStream) Next(_ context.Context) bool {
	if !s.scanner.Scan() {
		s.err = s.scanner.Err()
		return false
	}
	s.current = s.scanner.Bytes()
	return true
}

func (s *NewLineStream) Data() []byte {
	return s.current
}

func (s *NewLineStream) Err() error {
	if s.err != nil {
		return s.err
	}

	return s.scanner.Err()
}

func (s *JSONEachRowStream[T]) Next(ctx context.Context) bool {
	if !s.inner.Next(ctx) {
		return false
	}

	s.err = json.Unmarshal(s.inner.Data(), &s.current)

	return true
}

func (s *JSONEachRowStream[T]) Data() T {
	return s.current
}

func (s *JSONEachRowStream[T]) Err() error {
	if s.err != nil {
		return s.err
	}

	return s.inner.Err()
}

func NewNewLineStream(r io.Reader) Stream[[]byte] {
	return &NewLineStream{scanner: bufio.NewScanner(r)}
}

func NewNewLineStreamFactory() StreamFactory[[]byte] {
	return StreamFactoryFunc[[]byte](func(r io.Reader) Stream[[]byte] {
		return NewNewLineStream(r)
	})
}

func NewProxyStream(r io.Reader, bufferSize int) Stream[[]byte] {
	return &ProxyStream{reader: r, buffer: make([]byte, bufferSize)}
}

func NewProxyStreamFactory(bufferSize int) StreamFactory[[]byte] {
	return StreamFactoryFunc[[]byte](func(r io.Reader) Stream[[]byte] {
		return NewProxyStream(r, bufferSize)
	})
}

func NewJSONEachRowStream[T any](r io.Reader) Stream[T] {
	return &JSONEachRowStream[T]{
		inner: NewNewLineStream(r),
	}
}

func NewJSONEachRowStreamFactory[T any]() StreamFactory[T] {
	return StreamFactoryFunc[T](func(r io.Reader) Stream[T] {
		return NewJSONEachRowStream[T](r)
	})
}
