package withttp

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type (
	nativeReqAdapter struct {
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

func (a *nativeReqAdapter) SetMethod(method string) {
	a.req.Method = method
}

func (a *nativeReqAdapter) SetURL(u *url.URL) {
	a.req.URL = u
}

func (a *nativeReqAdapter) SetBody(body io.ReadCloser) {
	a.req.Body = body
}

func adaptReqNative(req *http.Request) Request {
	return &nativeReqAdapter{req: req}
}

func (a *NativeHttpClientAdapter) Request() (Request, error) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}
	return adaptReqNative(req), err
}

func (a *NativeHttpClientAdapter) Do(ctx context.Context, req Request) (Response, error) {
	res, err := a.cli.Do(req.(*nativeReqAdapter).req)
	return adaptResNative(res), err
}

func NewNativeHttpClientAdapter() *NativeHttpClientAdapter {
	return &NativeHttpClientAdapter{cli: http.DefaultClient}
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
