package withttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type metric struct {
	Time time.Time `json:"t"`
	Temp float32   `json:"T"`
}

func TestRequestStream_FromSlice(t *testing.T) {
	points := []metric{
		{
			Time: time.Unix(time.Now().Unix()-1, 0),
			Temp: 39,
		},
		{
			Time: time.Now(),
			Temp: 40,
		},
	}

	stream := Slice[metric](points)

	// Mock endpoint that captures the request body
	var receivedData []byte
	testEndpoint := NewEndpoint("webhook-site-request-stream-example").
		Request(BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb")).
		Response(
			MockedRes(func(res Response) {
				res.SetStatus(http.StatusOK)
				res.SetBody(io.NopCloser(strings.NewReader(`{"status": "ok"}`)))
			}),
		)

	call := NewCall[any](Fasthttp()).
		Method(http.MethodPost).
		ContentType(ContentTypeJSONEachRow).
		RequestSniffed(func(data []byte, err error) {
			receivedData = append(receivedData, data...)
			t.Logf("Received: '%s', err: %v", string(data), err)
		}).
		RequestStreamBody(
			RequestStreamBody[any, metric](stream),
		).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), testEndpoint)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	if len(receivedData) == 0 {
		t.Error("Expected to receive some data from request sniffer")
	}

	// Check that we received JSON data
	dataStr := string(receivedData)
	if !strings.Contains(dataStr, `"T":39`) && !strings.Contains(dataStr, `"T":40`) {
		t.Errorf("Expected JSON data containing temperature values, got: %s", dataStr)
	}
}

func TestRequestStream_FromChannel(t *testing.T) {
	points := make(chan metric, 2)

	go func() {
		points <- metric{
			Time: time.Unix(time.Now().Unix()-1, 0),
			Temp: 39,
		}

		points <- metric{
			Time: time.Now(),
			Temp: 40,
		}

		close(points)
	}()

	stream := Channel[metric](points)

	var receivedData []byte
	testEndpoint := NewEndpoint("webhook-site-request-stream-example").
		Request(BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb")).
		Response(
			MockedRes(func(res Response) {
				res.SetStatus(http.StatusOK)
				res.SetBody(io.NopCloser(strings.NewReader(`{"status": "ok"}`)))
			}),
		)

	call := NewCall[any](Fasthttp()).
		Method(http.MethodPost).
		ContentType(ContentTypeJSONEachRow).
		RequestSniffed(func(data []byte, err error) {
			receivedData = append(receivedData, data...)
			t.Logf("Received: '%s', err: %v", string(data), err)
		}).
		RequestStreamBody(
			RequestStreamBody[any, metric](stream),
		).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), testEndpoint)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	if len(receivedData) == 0 {
		t.Error("Expected to receive some data from request sniffer")
	}

	// Check that we received JSON data
	dataStr := string(receivedData)
	if !strings.Contains(dataStr, `"T":39`) && !strings.Contains(dataStr, `"T":40`) {
		t.Errorf("Expected JSON data containing temperature values, got: %s", dataStr)
	}
}

func TestRequestStream_FromReader(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	go func() {
		buf.WriteString(`{"t":"2022-09-01T00:58:15+02:00","T":39}`)
		buf.WriteByte('\n')
		buf.WriteString(`{"t":"2022-09-01T00:59:15+02:00","T":40}`)
		buf.WriteByte('\n')
	}()

	streamFactory := NewProxyStreamFactory(1 << 10)
	stream := NewStreamFromReader(buf, streamFactory)

	var receivedData []byte
	testEndpoint := NewEndpoint("webhook-site-request-stream-example").
		Request(BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb")).
		Response(
			MockedRes(func(res Response) {
				res.SetStatus(http.StatusOK)
				res.SetBody(io.NopCloser(strings.NewReader(`{"status": "ok"}`)))
			}),
		)

	call := NewCall[any](NetHttp()).
		Method(http.MethodPost).
		RequestSniffed(func(data []byte, err error) {
			receivedData = append(receivedData, data...)
			t.Logf("Received: '%s', err: %v", string(data), err)
		}).
		ContentType(ContentTypeJSONEachRow).
		RequestStreamBody(
			RequestStreamBody[any, []byte](stream),
		).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), testEndpoint)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	if len(receivedData) == 0 {
		t.Error("Expected to receive some data from request sniffer")
	}

	// Check that we received JSON data
	dataStr := string(receivedData)
	if !strings.Contains(dataStr, `"T":39`) && !strings.Contains(dataStr, `"T":40`) {
		t.Errorf("Expected JSON data containing temperature values, got: %s", dataStr)
	}
}

// Example_requestStream_fromSlice demonstrates streaming data from a slice to the server.
func Example_requestStream_fromSlice() {
	type DataPoint struct {
		Timestamp int64   `json:"ts"`
		Value     float64 `json:"val"`
	}

	// Sample data to stream
	data := []DataPoint{
		{Timestamp: 1672531200, Value: 25.5},
		{Timestamp: 1672531260, Value: 26.1},
		{Timestamp: 1672531320, Value: 24.8},
	}

	stream := Slice[DataPoint](data)

	mockEndpoint := NewEndpoint("DataIngestion").
		Request(BaseURL("https://api.example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(`{"status": "received"}`)))
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[any](NewMockHttpClientAdapter()).
		URI("/data").
		Method(http.MethodPost).
		ContentType(ContentTypeJSONEachRow).
		RequestStreamBody(
			RequestStreamBody[any, DataPoint](stream),
		).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Data streamed successfully")
	// Output: Data streamed successfully
}

// Example_requestStream_withSniffer demonstrates streaming with request sniffing.
func Example_requestStream_withSniffer() {
	type Event struct {
		Type string `json:"type"`
	}

	events := []Event{
		{Type: "click"},
		{Type: "view"},
	}

	stream := Slice[Event](events)

	mockEndpoint := NewEndpoint("Analytics").
		Request(BaseURL("https://api.example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(`{"processed": true}`)))
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[any](NewMockHttpClientAdapter()).
		URI("/events").
		Method(http.MethodPost).
		ContentType(ContentTypeJSONEachRow).
		RequestSniffed(func(data []byte, err error) {
			if err == nil {
				// Simplify output to avoid JSON formatting issues
				if strings.Contains(string(data), "click") {
					fmt.Println("Sending click event")
				} else if strings.Contains(string(data), "view") {
					fmt.Println("Sending view event")
				}
			}
		}).
		RequestStreamBody(
			RequestStreamBody[any, Event](stream),
		).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Events processed")
	// Output: Sending click event
	// Sending view event
	// Events processed
}
