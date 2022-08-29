package withttp

import (
	"bytes"
	"context"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"net/url"
)

type (
	fastHttpReqAdapter struct {
		req *fasthttp.Request
	}

	fastHttpResAdapter struct {
		statusText string
		res        *fasthttp.Response
	}

	FastHttpHttpClientAdapter struct {
		cli *fasthttp.Client
	}
)

var (
	fastClient = &fasthttp.Client{}
)

func (a *fastHttpReqAdapter) AddHeader(key, value string) {
	a.req.Header.Add(key, value)
}

func (a *fastHttpReqAdapter) SetHeader(key, value string) {
	a.req.Header.Set(key, value)
}

func (a *fastHttpReqAdapter) SetMethod(method string) {
	a.req.Header.SetMethod(method)
}

func (a *fastHttpReqAdapter) SetURL(u *url.URL) {
	uri := a.req.URI()
	uri.SetScheme(u.Scheme)
	uri.SetUsername(u.User.Username())
	uri.SetHost(u.Host)
	uri.SetPath(u.Path)
	uri.SetQueryString(u.RawQuery)
	uri.SetHash(string(uri.Hash()))
	if pass, ok := u.User.Password(); ok {
		uri.SetPassword(pass)
	}
}

func (a *fastHttpReqAdapter) SetBody(body io.ReadCloser) {
	a.req.SetBodyStream(body, 0)
}

func (a *fastHttpReqAdapter) URL() *url.URL {
	uri := a.req.URI()
	return &url.URL{
		Scheme:   string(uri.Scheme()),
		User:     url.UserPassword(string(uri.Username()), string(uri.Password())),
		Host:     string(uri.Host()),
		Path:     string(uri.Path()),
		RawPath:  string(uri.Path()),
		RawQuery: string(uri.QueryString()),
		Fragment: string(uri.Hash()),
	}
}

func adaptReqFastHttp(req *fasthttp.Request) Request {
	return &fastHttpReqAdapter{req: req}
}

func (a *FastHttpHttpClientAdapter) Request() (Request, error) {
	req := &fasthttp.Request{}
	req.Header.SetMethod(http.MethodGet)
	return adaptReqFastHttp(req), nil
}

func (a *FastHttpHttpClientAdapter) Do(_ context.Context, req Request) (Response, error) {
	res := &fasthttp.Response{} // TODO: Acquire/Release
	err := a.cli.Do(req.(*fastHttpReqAdapter).req, res)
	return adaptResFastHttp(res), err
}

func NewDefaultFastHttpHttpClientAdapter() *FastHttpHttpClientAdapter {
	return NewFastHttpHttpClientAdapter(fastClient)
}

func NewFastHttpHttpClientAdapter(cli *fasthttp.Client) *FastHttpHttpClientAdapter {
	return &FastHttpHttpClientAdapter{cli: cli}
}

func adaptResFastHttp(res *fasthttp.Response) Response {
	return &fastHttpResAdapter{res: res}
}

func (a *fastHttpResAdapter) SetBody(body io.ReadCloser) {
	a.res.SetBodyStream(body, 0)
}

func (a *fastHttpResAdapter) Status() int {
	return a.res.StatusCode()
}

func (a *fastHttpResAdapter) SetStatus(status int) {
	a.res.SetStatusCode(status)
}

func (a *fastHttpResAdapter) StatusText() string {
	return a.statusText
}

func (a *fastHttpResAdapter) Body() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(a.res.Body()))
}