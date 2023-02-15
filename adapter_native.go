package withttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type (
	nativeReqAdapter struct {
		body io.ReadWriteCloser

		req *http.Request
	}

	nativeResAdapter struct {
		res *http.Response
	}

	NativeHttpClientAdapter struct {
		cli *http.Client
	}
)

func (a *nativeReqAdapter) AddHeader(key, value string) {
	a.req.Header.Add(key, value)
}

func (a *nativeReqAdapter) SetHeader(key, value string) {
	a.req.Header.Set(key, value)
}

func (a *nativeReqAdapter) Header(key string) (string, bool) {
	s := a.req.Header.Get(key)
	return s, len(s) > 0
}

func (a *nativeReqAdapter) Method() string {
	return a.req.Method
}

func (a *nativeReqAdapter) SetMethod(method string) {
	a.req.Method = method
}

func (a *nativeReqAdapter) SetURL(u *url.URL) {
	a.req.URL = u
}

func (a *nativeReqAdapter) SetBodyStream(body io.ReadWriteCloser, _ int) {
	a.body = body
	a.req.Body = body
}

func (a *nativeReqAdapter) SetBody(payload []byte) {
	// TODO: pool these readers
	a.body = closableReaderWriter{ReadWriter: bytes.NewBuffer(payload)}
	a.req.Body = a.body
	a.req.ContentLength = int64(len(payload))
}

func (a *nativeReqAdapter) Body() []byte {
	bts, _ := io.ReadAll(a.req.Body)
	return bts
}

func (a *nativeReqAdapter) RangeHeaders(fn func(string, string)) {
	for k, v := range a.req.Header {
		fn(k, strings.Join(v, ", "))
	}
}

func (a *nativeReqAdapter) BodyStream() io.ReadWriteCloser {
	return a.body
}

func (a *nativeReqAdapter) URL() *url.URL {
	return a.req.URL
}

func adaptReqNative(req *http.Request) Request {
	return &nativeReqAdapter{req: req}
}

func (a *NativeHttpClientAdapter) Request(ctx context.Context) (Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "", nil)
	if err != nil {
		return nil, err
	}
	return adaptReqNative(req), err
}

func (a *NativeHttpClientAdapter) Do(ctx context.Context, req Request) (Response, error) {
	res, err := a.cli.Do(req.(*nativeReqAdapter).req)
	return adaptResNative(res), err
}

func WithNetHttpClient(cli *http.Client) *NativeHttpClientAdapter {
	return &NativeHttpClientAdapter{cli: cli}
}

func WithNetHttp() *NativeHttpClientAdapter {
	return WithNetHttpClient(http.DefaultClient)
}

func adaptResNative(res *http.Response) Response {
	return &nativeResAdapter{res: res}
}

func (a *nativeResAdapter) SetBody(body io.ReadCloser) {
	a.res.Body = body
}

func (a *nativeResAdapter) Status() int {
	return a.res.StatusCode
}

func (a *nativeResAdapter) SetStatus(status int) {
	a.res.StatusCode = status
}

func (a *nativeResAdapter) StatusText() string {
	return a.res.Status
}

func (a *nativeResAdapter) Body() io.ReadCloser {
	return a.res.Body
}

func (a *nativeResAdapter) AddHeader(key, value string) {
	a.res.Header.Add(key, value)
}

func (a *nativeResAdapter) SetHeader(key, value string) {
	a.res.Header.Set(key, value)
}

func (a *nativeResAdapter) Header(key string) (string, bool) {
	s := a.res.Header.Get(key)
	return s, len(s) > 0
}

func (a *nativeResAdapter) RangeHeaders(fn func(string, string)) {
	for k, v := range a.res.Header {
		fn(k, strings.Join(v, ", "))
	}
}
