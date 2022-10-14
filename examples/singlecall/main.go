package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/sonirico/withttp"
)

type GithubRepoInfo struct {
	ID  int    `json:"id"`
	URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (GithubRepoInfo, error) {
	l := zerolog.New(os.Stdout)

	call := withttp.NewCall[GithubRepoInfo](withttp.WithFasthttp()).
		WithURL(fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)).
		WithLogger(&l).
		WithMethod(http.MethodGet).
		WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
		WithParseJSON().
		WithExpectedStatusCodes(http.StatusOK)

	err := call.Call(context.Background())

	return call.BodyParsed, err
}

func main() {
	info, _ := GetRepoInfo("sonirico", "withttp")
	log.Println(info)
}
