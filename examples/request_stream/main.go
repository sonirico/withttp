package main

import (
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

	slice := withttp.Slice[metric](points)

	testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
		Request(
			withttp.WithBaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"),
		)

	call := withttp.NewCall[any](withttp.NewDefaultFastHttpHttpClientAdapter()).
		WithMethod(http.MethodPost).
		WithContentType(withttp.ContentTypeJSONEachRow).
		WithRequestStreamBody(
			withttp.WithRequestStreamBody[any, metric](slice),
		).
		WithExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), testEndpoint)

	received := call.Req.Body()

	fmt.Printf("recv: '%s'", string(received))
	/*
		{"t":"2022-09-01T00:58:15+02:00","T":39}
		{"t":"2022-09-01T00:58:16.15846898+02:00","T":40}
	*/
	return err
}

func main() {
	_ = CreateStream()
}
