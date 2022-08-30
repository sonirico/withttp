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

var (
	exchangeListOrders = withttp.NewEndpoint("ListOrders").
				Request(withttp.WithURL("http://example.com")).
				Response(
			withttp.WithMockedRes(func(res withttp.Response) {
				res.SetBody(io.NopCloser(bytes.NewReader(mockResponse)))
				res.SetStatus(http.StatusOK)
			}),
		)
	mockResponse = []byte(strings.TrimSpace(`
		{"amount": 234, "pair": "BTC/USDT"}
		{"amount": 123, "pair": "ETH/USDT"}`))
)

func main() {
	type Order struct {
		Amount float64 `json:"amount"`
		Pair   string  `json:"pair"`
	}

	res := make(chan Order)

	call := withttp.NewCall[Order](withttp.NewDefaultFastHttpHttpClientAdapter()).
		WithURL("https://github.com/").
		WithMethod(http.MethodGet).
		WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
		WithHeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		WithJSONEachRowChan(res).
		WithExpectedStatusCodes(http.StatusOK)

	go func() {
		for order := range res {
			log.Println(order)
		}
	}()

	err := call.Call(context.Background(), exchangeListOrders)

	if err != nil {
		panic(err)
	}
}
