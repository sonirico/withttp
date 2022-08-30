package withttp

import "io"

type (
	closableReaderWriter struct {
		io.ReadWriter
	}
)

func (b closableReaderWriter) Close() error {
	return nil
}
