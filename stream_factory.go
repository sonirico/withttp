package withttp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/sonirico/withttp/csvparser"
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

	CSVStream[T any] struct {
		current T

		inner Stream[[]byte]

		err error

		parser csvparser.Parser[T]

		ignoreLines int
		ignoreErr   bool

		rowCount int
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

func (s *CSVStream[T]) next(ctx context.Context, shouldParse bool) bool {
	if !s.inner.Next(ctx) {
		return false
	}

	if !shouldParse {
		return true
	}

	var zeroed T // TODO: json too?
	line := s.inner.Data()
	s.current = zeroed
	s.err = s.parser.Parse(line, &s.current)

	if s.err == nil || s.ignoreErr {
		s.rowCount++
	}

	return true
}

func (s *CSVStream[T]) Next(ctx context.Context) bool {
	for s.ignoreLines > 0 {
		_ = s.next(ctx, false)
		s.ignoreLines--
	}
	return s.next(ctx, true)
}

func (s *CSVStream[T]) Data() T {
	return s.current
}

func (s *CSVStream[T]) Err() error {
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

func NewCSVStream[T any](r io.Reader, ignoreLines int, parser csvparser.Parser[T]) Stream[T] {
	return &CSVStream[T]{
		inner:       NewNewLineStream(r),
		parser:      parser,
		ignoreLines: ignoreLines,
		ignoreErr:   false,
	}
}

func NewJSONEachRowStreamFactory[T any]() StreamFactory[T] {
	return StreamFactoryFunc[T](func(r io.Reader) Stream[T] {
		return NewJSONEachRowStream[T](r)
	})
}

func NewCSVStreamFactory[T any](ignoreLines int, parser csvparser.Parser[T]) StreamFactory[T] {
	return StreamFactoryFunc[T](func(r io.Reader) Stream[T] {
		return NewCSVStream[T](r, ignoreLines, parser)
	})
}
