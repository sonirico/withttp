package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sonirico/withttp"
)

func main() {
	type Order struct {
		Amount float64 `json:"amount"`
		Pair   string  `json:"pair"`
	}

	mockResponse := strings.TrimSpace(`
	{"amount": 234, "pair": "BTC/USDT"}
	{"amount": 123, "pair": "ETH/USDT"}`)

	endpoint := withttp.New(
		"ListOrders",
		withttp.WithResMock(func(response withttp.Response) {
			response.SetBody(io.NopCloser(bytes.NewReader([]byte(mockResponse))))
			response.SetStatus(http.StatusOK)
		}),
	)

	call := withttp.NewCall[Order](withttp.NewMockHttpClientAdapter()).
		WithMethod(http.MethodGet).
		WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
		WithHeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		WithJSONEachRow(func(m Order) bool {
			log.Println(m)
			return true
		}).
		WithExpectedStatusCodes(http.StatusOK)

	err := call.Call(context.Background(), endpoint)

	if err != nil {
		panic(err)
	}
}
