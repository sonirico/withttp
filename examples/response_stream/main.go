package main

import (
	"bytes"
	"context"
	"github.com/sonirico/withttp"
	"github.com/sonirico/withttp/csvparser"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	getRepoStatsEndpoint = withttp.NewEndpoint("GetRepoStats").
				Request(withttp.WithBaseURL("http://example.com")).
				Response(
			withttp.WithMockedRes(func(res withttp.Response) {
				res.SetBody(io.NopCloser(bytes.NewReader(mockResponse)))
				res.SetStatus(http.StatusOK)
			}),
		)
	mockResponse = []byte(strings.TrimSpace(`
rank,repo_name,stars
1,freeCodeCamp,341271
2,996.ICU,261139
`))
)

func parseCSVResponse() {
	type Repo struct {
		Rank  int
		Name  string
		Stars int
	}

	ignoreLines := 1 // in order to ignore header

	parser := csvparser.NewParser[Repo](
		csvparser.SeparatorComma,
		csvparser.IntCol[Repo](
			false,
			nil,
			func(x *Repo, rank int) { x.Rank = rank },
		),
		csvparser.StringCol[Repo](
			false,
			nil,
			func(x *Repo, name string) { x.Name = name },
		),
		csvparser.IntCol[Repo](
			false,
			nil,
			func(x *Repo, stars int) { x.Stars = stars },
		),
	)

	call := withttp.NewCall[Repo](withttp.WithFasthttp()).
		WithMethod(http.MethodGet).
		WithHeader("User-Agent", "withttp/0.6.0 See https://github.com/sonirico/withttp", false).
		WithParseCSV(ignoreLines, parser, func(r Repo) bool {
			log.Println("repo: ", r)
			return true
		}).
		WithExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), getRepoStatsEndpoint)

	if err != nil {
		panic(err)
	}
}

func main() {
	parseCSVResponse()
}
