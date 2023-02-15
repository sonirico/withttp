package withttp

import (
	"context"
	"net/http"
)

type (
	MockHttpClientAdapter struct{}
)

func NewMockHttpClientAdapter() *MockHttpClientAdapter {
	return &MockHttpClientAdapter{}
}

func (a *MockHttpClientAdapter) Request(_ context.Context) (Request, error) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}
	return adaptReqMock(req), err
}

func (a *MockHttpClientAdapter) Do(_ context.Context, _ Request) (Response, error) {
	return adaptResMock(&http.Response{}), nil
}

func adaptResMock(res *http.Response) Response {
	return adaptResNative(res)
}

func adaptReqMock(req *http.Request) Request {
	return adaptReqNative(req)
}
