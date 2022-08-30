package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sonirico/withttp/codec"

	"github.com/sonirico/withttp"
)

var (
	githubApi = withttp.NewEndpoint("GithubAPI").
		Request(withttp.WithURL("https://api.github.com/"))
)

type GithubRepoInfo struct {
	ID  int    `json:"id"`
	URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (GithubRepoInfo, error) {
	call := withttp.NewCall[GithubRepoInfo](withttp.NewDefaultFastHttpHttpClientAdapter()).
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

type GithubCreateIssueResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

func CreateRepoIssue(user, repo, title, body, assignee string) (GithubCreateIssueResponse, error) {
	type payload struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		Assignee string `json:"assignee"`
	}

	p := payload{
		Title:    title,
		Body:     body,
		Assignee: assignee,
	}

	data, err := codec.NewNativeJsonCodec().Encode(p)
	if err != nil {
		panic(err)
	}

	call := withttp.NewCall[GithubCreateIssueResponse](
		withttp.NewDefaultFastHttpHttpClientAdapter(),
	).
		//WithURI(fmt.Sprintf("repos/%s/%s/issues", user, repo)).
		WithURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb").
		WithMethod(http.MethodPost).
		WithRawBody(data).
		WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
		WithHeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		WithExpectedStatusCodes(http.StatusOK)

	err = call.Call(context.Background(), githubApi)

	return call.BodyParsed, err
}

func main() {
	//info, _ := GetRepoInfo("sonirico", "withttp")
	//log.Println(info)
	res, err := CreateRepoIssue("sonirico", "withttp", "test",
		"This is a test", "sonirico")
	log.Println(res, err)
}
