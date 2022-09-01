package withttp

import (
	"context"
	"io"
)

type (
	rangeable[T any] interface {
		Range(func(int, T) bool)
		Serialize() bool
	}

	Slice[T any] []T

	Channel[T any] chan T

	StreamFromReader struct {
		io.Reader
		streamFactory StreamFactory[[]byte]
	}
)

func NewStreamFromReader(r io.Reader, sf StreamFactory[[]byte]) StreamFromReader {
	return StreamFromReader{
		Reader:        r,
		streamFactory: sf,
	}
}

func (s Slice[T]) Range(fn func(int, T) bool) {
	for i, x := range s {
		if !fn(i, x) {
			return
		}
	}
}

func (s Slice[T]) Serialize() bool { return true }

func (c Channel[T]) Range(fn func(int, T) bool) {
	i := 0
	for {
		x, ok := <-c
		if !ok {
			return
		}

		fn(i, x)

		i++
	}
}

func (c Channel[T]) Serialize() bool { return true }

func (r StreamFromReader) Range(fn func(int, []byte) bool) {
	stream := r.streamFactory.Get(r)
	i := 0
	for stream.Next(context.TODO()) {
		if stream.Err() != nil {
			return
		}

		fn(i, stream.Data())
		i++
	}
}

func (r StreamFromReader) Serialize() bool { return false }
