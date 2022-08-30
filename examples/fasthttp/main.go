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
		Request(withttp.WithBaseURL("https://api.github.com/"))
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
		WithParseJSON().
		WithExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), githubApi)

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

	call := withttp.NewCall[GithubCreateIssueResponse](
		withttp.NewDefaultFastHttpHttpClientAdapter(),
	).
		WithURI(fmt.Sprintf("repos/%s/%s/issues", user, repo)).
		WithMethod(http.MethodPost).
		WithContentType("application/vnd+github+json").
		WithBody(p).
		WithHeaderFunc(func() (key, value string, override bool) {
			key = "Authorization"
			value = fmt.Sprintf("Bearer %s", "S3cret")
			override = true
			return
		}).
		WithExpectedStatusCodes(http.StatusCreated)

	err := call.CallEndpoint(context.Background(), githubApi)

	log.Println("req body", string(call.Req.Body()))

	return call.BodyParsed, err
}

func main() {
	// Fetch repo info
	info, _ := GetRepoInfo("sonirico", "withttp")
	log.Println(info)

	// Create an issue
	res, err := CreateRepoIssue("sonirico", "withttp", "test",
		"This is a test", "sonirico")
	log.Println(res, err)
}
