package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sonirico/withttp"
)

type metric struct {
	Time time.Time `json:"t"`
	Temp float32   `json:"T"`
}

func CreateStream() error {
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

	stream := withttp.Slice[metric](points)

	testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
		Request(
			withttp.WithBaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"),
		)

	call := withttp.NewCall[any](withttp.NewDefaultFastHttpHttpClientAdapter()).
		WithMethod(http.MethodPost).
		WithContentType(withttp.ContentTypeJSONEachRow).
		WithRequestSniffed(func(data []byte, err error) {
			fmt.Printf("recv: '%s', err: %v", string(data), err)
		}).
		WithRequestStreamBody(
			withttp.WithRequestStreamBody[any, metric](stream),
		).
		WithExpectedStatusCodes(http.StatusOK)

	return call.CallEndpoint(context.Background(), testEndpoint)
}

func CreateStreamChannel() error {
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

	stream := withttp.Channel[metric](points)

	testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
		Request(
			withttp.WithBaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"),
		)

	call := withttp.NewCall[any](withttp.NewDefaultFastHttpHttpClientAdapter()).
		WithMethod(http.MethodPost).
		WithContentType(withttp.ContentTypeJSONEachRow).
		WithRequestSniffed(func(data []byte, err error) {
			fmt.Printf("recv: '%s', err: %v", string(data), err)
		}).
		WithRequestStreamBody(
			withttp.WithRequestStreamBody[any, metric](stream),
		).
		WithExpectedStatusCodes(http.StatusOK)

	return call.CallEndpoint(context.Background(), testEndpoint)
}

func CreateStreamReader() error {
	buf := bytes.NewBuffer(nil)

	go func() {
		buf.WriteString("{\"t\":\"2022-09-01T00:58:15+02:00\"")
		buf.WriteString(",\"T\":39}\n{\"t\":\"2022-09-01T00:59:15+02:00\",\"T\":40}\n")
	}()

	streamFactory := withttp.NewProxyStreamFactory(1 << 10)

	stream := withttp.NewStreamFromReader(buf, streamFactory)

	testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
		Request(
			withttp.WithBaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"),
		)

	call := withttp.NewCall[any](withttp.NewDefaultNativeHttpClientAdapter()).
		WithMethod(http.MethodPost).
		WithRequestSniffed(func(data []byte, err error) {
			fmt.Printf("recv: '%s', err: %v", string(data), err)
		}).
		WithContentType(withttp.ContentTypeJSONEachRow).
		WithRequestStreamBody(
			withttp.WithRequestStreamBody[any, []byte](stream),
		).
		WithExpectedStatusCodes(http.StatusOK)

	return call.CallEndpoint(context.Background(), testEndpoint)
}

func main() {
	_ = CreateStream()
	_ = CreateStreamChannel()
	_ = CreateStreamReader()
}
