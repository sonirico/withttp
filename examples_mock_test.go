package withttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	exchangeListOrders = NewEndpoint("ListOrders").
				Request(BaseURL("http://example.com")).
				Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(mockResponse)))
				res.SetStatus(http.StatusOK)
			}),
		)
	mockResponse = `{"amount": 234, "pair": "BTC/USDT"}
{"amount": 123, "pair": "ETH/USDT"}`
)

type Order struct {
	Amount float64 `json:"amount"`
	Pair   string  `json:"pair"`
}

func TestMockEndpoint_ParseJSONEachRowChan(t *testing.T) {
	res := make(chan Order)
	orders := make([]Order, 0)

	// Collect orders from channel
	go func() {
		for order := range res {
			orders = append(orders, order)
		}
	}()

	call := NewCall[Order](Fasthttp()).
		Method(http.MethodGet).
		BasicAuth("pepito", "secret").
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		ParseJSONEachRowChan(res).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), exchangeListOrders)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	// Wait a bit for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	if len(orders) >= 1 {
		if orders[0].Amount != 234 || orders[0].Pair != "BTC/USDT" {
			t.Errorf("First order incorrect: got %+v", orders[0])
		}
	}

	if len(orders) >= 2 {
		if orders[1].Amount != 123 || orders[1].Pair != "ETH/USDT" {
			t.Errorf("Second order incorrect: got %+v", orders[1])
		}
	}

	// Check authorization header was set
	authHeader, hasAuth := call.Req.Header("Authorization")
	if !hasAuth {
		t.Error("Expected Authorization header to be set")
	} else {
		t.Logf("Authorization header: %s", authHeader)
	}
}

func TestMockEndpoint_ParseJSONEachRow(t *testing.T) {
	orders := make([]Order, 0)

	call := NewCall[Order](Fasthttp()).
		Method(http.MethodGet).
		BasicAuth("pepito", "secret").
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		ParseJSONEachRow(func(order Order) bool {
			orders = append(orders, order)
			return true // continue processing
		}).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), exchangeListOrders)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	if len(orders) >= 1 {
		if orders[0].Amount != 234 || orders[0].Pair != "BTC/USDT" {
			t.Errorf("First order incorrect: got %+v", orders[0])
		}
	}

	if len(orders) >= 2 {
		if orders[1].Amount != 123 || orders[1].Pair != "ETH/USDT" {
			t.Errorf("Second order incorrect: got %+v", orders[1])
		}
	}
}

// Example_mockEndpoint demonstrates how to create and use a mocked endpoint for testing.
func Example_mockEndpoint() {
	type Order struct {
		Amount float64 `json:"amount"`
		Pair   string  `json:"pair"`
	}

	// Create a mocked endpoint that returns test data
	mockEndpoint := NewEndpoint("MockExchange").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(`{"amount": 100.5, "pair": "BTC/USD"}`)))
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[Order](NewMockHttpClientAdapter()).
		Method(http.MethodGet).
		Header("Authorization", "Bearer token123", true).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Order amount: %.1f, Pair: %s\n", call.BodyParsed.Amount, call.BodyParsed.Pair)
	// Output: Order amount: 100.5, Pair: BTC/USD
}

// Example_parseJSONEachRow demonstrates parsing JSON-each-row format responses.
func Example_parseJSONEachRow() {
	type Trade struct {
		Price  float64 `json:"price"`
		Volume float64 `json:"volume"`
	}

	mockData := `{"price": 50000.0, "volume": 1.5}
{"price": 51000.0, "volume": 0.8}`

	mockEndpoint := NewEndpoint("TradesAPI").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(mockData)))
				res.SetStatus(http.StatusOK)
			}),
		)

	var trades []Trade
	call := NewCall[Trade](NewMockHttpClientAdapter()).
		Method(http.MethodGet).
		ParseJSONEachRow(func(trade Trade) bool {
			trades = append(trades, trade)
			return true // continue processing
		}).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Processed %d trades\n", len(trades))
	if len(trades) > 0 {
		fmt.Printf("First trade: $%.0f, volume: %.1f\n", trades[0].Price, trades[0].Volume)
	}
	// Output: Processed 2 trades
	// First trade: $50000, volume: 1.5
}
