package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sonirico/withttp"
)

var (
	githubApi = withttp.NewEndpoint("GithubAPI").
		Request(withttp.WithURL("https://api.github.com/"))
)

type githubRepoInfo struct {
	ID  int    `json:"id"`
	URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (githubRepoInfo, error) {
	call := withttp.NewCall[githubRepoInfo](withttp.NewDefaultFastHttpHttpClientAdapter()).
		WithURI(fmt.Sprintf("repos/%s/%s", user, repo)).
		WithMethod(http.MethodGet).
		WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
		WithHeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		WithJSON().
		WithExpectedStatusCodes(http.StatusOK)

	err := call.Call(context.Background(), githubApi)

	return call.BodyParsed, err
}

func main() {
	info, _ := GetRepoInfo("sonirico", "withttp")
	log.Println(info)
}
