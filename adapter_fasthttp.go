package withttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

type (
	fastHttpReqAdapter struct {
		stream io.ReadWriteCloser

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

func (a *fastHttpReqAdapter) RangeHeaders(fn func(string, string)) {
	a.req.Header.VisitAll(func(key, value []byte) {
		fn(string(key), string(value))
	})
}

func (a *fastHttpReqAdapter) SetHeader(key, value string) {
	a.req.Header.Set(key, value)
}

func (a *fastHttpReqAdapter) Header(key string) (string, bool) {
	data := a.req.Header.Peek(key)
	if len(data) == 0 {
		return "", false
	}
	bts := make([]byte, len(data))
	copy(bts, data)
	return string(bts), true
}

func (a *fastHttpReqAdapter) Method() string {
	return string(a.req.Header.Method())
}

func (a *fastHttpReqAdapter) SetMethod(method string) {
	a.req.Header.SetMethod(method)
}

func (a *fastHttpReqAdapter) SetURL(u *url.URL) {
	uri := a.req.URI()
	uri.SetScheme(u.Scheme)
	uri.SetHost(u.Host)
	uri.SetPath(u.Path)
	uri.SetQueryString(u.RawQuery)
	uri.SetHash(string(uri.Hash()))

	username := u.User.Username()
	if StrIsset(username) {
		uri.SetUsername(username)
	}

	if pass, ok := u.User.Password(); ok {
		uri.SetPassword(pass)
	}
}

func (a *fastHttpReqAdapter) SetBodyStream(body io.ReadWriteCloser, bodySize int) {
	a.stream = body
	a.req.SetBodyStream(a.stream, bodySize)
}

func (a *fastHttpReqAdapter) SetBody(body []byte) {
	a.req.SetBody(body)
	a.req.Header.SetContentLength(len(body))
}

func (a *fastHttpReqAdapter) Body() (bts []byte) {
	bts, _ = io.ReadAll(a.stream)
	return
}

func (a *fastHttpReqAdapter) BodyStream() io.ReadWriteCloser {
	return a.stream
}

func (a *fastHttpReqAdapter) URL() *url.URL {
	uri := a.req.URI()

	var user *url.Userinfo
	if BtsIsset(uri.Username()) {
		user = url.UserPassword(string(uri.Username()), string(uri.Password()))
	}

	u := &url.URL{
		Scheme:   string(uri.Scheme()),
		User:     user,
		Host:     string(uri.Host()),
		Path:     string(uri.Path()),
		RawPath:  string(uri.Path()),
		RawQuery: string(uri.QueryString()),
		Fragment: string(uri.Hash()),
	}

	return u
}

func adaptReqFastHttp(req *fasthttp.Request) Request {
	return &fastHttpReqAdapter{req: req}
}

func (a *FastHttpHttpClientAdapter) Request(_ context.Context) (Request, error) {
	req := &fasthttp.Request{}
	req.Header.SetMethod(http.MethodGet)
	return adaptReqFastHttp(req), nil
}

func (a *FastHttpHttpClientAdapter) Do(_ context.Context, req Request) (Response, error) {
	res := &fasthttp.Response{} // TODO: Acquire/Release
	err := a.cli.Do(req.(*fastHttpReqAdapter).req, res)
	return adaptResFastHttp(res), err
}

func Fasthttp() *FastHttpHttpClientAdapter {
	return newFastHttpHttpClientAdapter(fastClient)
}

func FasthttpClient(cli *fasthttp.Client) *FastHttpHttpClientAdapter {
	return newFastHttpHttpClientAdapter(cli)
}

func newFastHttpHttpClientAdapter(cli *fasthttp.Client) *FastHttpHttpClientAdapter {
	return &FastHttpHttpClientAdapter{cli: cli}
}

func adaptResFastHttp(res *fasthttp.Response) Response {
	return &fastHttpResAdapter{res: res}
}

func (a *fastHttpResAdapter) SetBody(body io.ReadCloser) {
	a.res.SetBodyStream(body, -1)
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

func (a *fastHttpResAdapter) AddHeader(key, value string) {
	a.res.Header.Add(key, value)
}

func (a *fastHttpResAdapter) SetHeader(key, value string) {
	a.res.Header.Set(key, value)
}

func (a *fastHttpResAdapter) Header(key string) (string, bool) {
	data := a.res.Header.Peek(key)
	if len(data) == 0 {
		return "", false
	}
	bts := make([]byte, len(data))
	copy(bts, data)
	return string(bts), true
}

func (a *fastHttpResAdapter) RangeHeaders(fn func(string, string)) {
	a.res.Header.VisitAll(func(key, value []byte) {
		fn(string(key), string(value))
	})
}
