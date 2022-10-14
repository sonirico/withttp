package withttp

import (
	"bufio"
	"fmt"
	"io"
)

func (c *Call[T]) WithLogger(l logger) *Call[T] {
	c.logger = l
	return c
}

func (c *Call[T]) Log(w io.Writer) {
	buf := bufio.NewWriter(w)

	_, _ = buf.WriteString(c.Req.Method())
	_, _ = buf.WriteString(" ")
	_, _ = buf.WriteString(c.Req.URL().String())
	_, _ = buf.WriteString("\n")

	c.Req.RangeHeaders(func(key string, value string) {
		_, _ = buf.WriteString(key)
		_ = buf.WriteByte(':')
		_ = buf.WriteByte(' ')
		_, _ = buf.WriteString(value)
		_ = buf.WriteByte('\n')
	})

	if !c.ReqIsStream && len(c.ReqBodyRaw) > 0 {
		_ = buf.WriteByte('\n')
		_, _ = buf.Write(c.ReqBodyRaw)
		_ = buf.WriteByte('\n')
	}

	_ = buf.WriteByte('\n')

	// TODO: print text repr of status code
	_, _ = buf.WriteString(fmt.Sprintf("%d %s", c.Res.Status(), ""))
	_ = buf.WriteByte('\n')

	c.Res.RangeHeaders(func(key string, value string) {
		_, _ = buf.WriteString(key)
		_ = buf.WriteByte(':')
		_ = buf.WriteByte(' ')
		_, _ = buf.WriteString(value)
		_ = buf.WriteByte('\n')
	})

	if len(c.BodyRaw) > 0 {
		_, _ = buf.Write(c.BodyRaw)
	}

	_ = buf.WriteByte('\n')

	_ = buf.Flush()
}
