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
	githubApi = NewEndpoint("GithubAPI").
		Request(BaseURL("https://api.github.com/"))
)

type githubRepoInfoFast struct {
	ID  int    `json:"id"`
	URL string `json:"html_url"`
}

type githubCreateIssueResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

func getRepoInfoFast(user, repo string) (githubRepoInfoFast, error) {
	call := NewCall[githubRepoInfoFast](Fasthttp()).
		URI(fmt.Sprintf("repos/%s/%s", user, repo)).
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), githubApi)

	return call.BodyParsed, err
}

func createRepoIssue(user, repo, title, body, assignee string) (githubCreateIssueResponse, error) {
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

	call := NewCall[githubCreateIssueResponse](Fasthttp()).
		URI(fmt.Sprintf("repos/%s/%s/issues", user, repo)).
		Method(http.MethodPost).
		ContentType(ContentTypeJSON).
		Body(p).
		HeaderFunc(func() (key, value string, override bool) {
			key = "Authorization"
			value = fmt.Sprintf("Bearer %s", "S3cret")
			override = true
			return
		}).
		ExpectedStatusCodes(http.StatusCreated)

	err := call.CallEndpoint(context.Background(), githubApi)

	return call.BodyParsed, err
}

func TestFastHTTP_GetRepoInfo(t *testing.T) {
	t.Skip("Skipping live API test - enable manually for integration testing")

	info, err := getRepoInfoFast("sonirico", "withttp")
	if err != nil {
		t.Fatalf("Failed to get repo info: %v", err)
	}

	if info.ID == 0 {
		t.Error("Expected repo ID to be non-zero")
	}

	if info.URL == "" {
		t.Error("Expected repo URL to be non-empty")
	}

	t.Logf("Repo info: ID=%d, URL=%s", info.ID, info.URL)
}

func TestFastHTTP_MockedExample(t *testing.T) {
	mockEndpoint := NewEndpoint("MockGithubAPI").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(
					io.NopCloser(
						strings.NewReader(
							`{"id": 67890, "html_url": "https://github.com/sonirico/withttp"}`,
						),
					),
				)
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[githubRepoInfoFast](Fasthttp()).
		URI("repos/sonirico/withttp").
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Date"
			value = time.Now().String()
			override = true
			return
		}).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		t.Fatalf("Failed to call mocked endpoint: %v", err)
	}

	if call.BodyParsed.ID != 67890 {
		t.Errorf("Expected ID 67890, got %d", call.BodyParsed.ID)
	}

	if call.BodyParsed.URL != "https://github.com/sonirico/withttp" {
		t.Errorf("Expected URL 'https://github.com/sonirico/withttp', got %s", call.BodyParsed.URL)
	}
}

func TestFastHTTP_CreateIssue(t *testing.T) {
	t.Skip("Skipping live API test - enable manually for integration testing")

	res, err := createRepoIssue("sonirico", "withttp", "test", "This is a test", "sonirico")
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	if res.ID == 0 {
		t.Error("Expected issue ID to be non-zero")
	}

	t.Logf("Issue created: ID=%d, URL=%s", res.ID, res.URL)
}

func TestFastHTTP_MockedCreateIssue(t *testing.T) {
	mockEndpoint := NewEndpoint("MockGithubAPI").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(
					io.NopCloser(
						strings.NewReader(
							`{"id": 999, "url": "https://api.github.com/repos/sonirico/withttp/issues/999"}`,
						),
					),
				)
				res.SetStatus(http.StatusCreated)
			}),
		)

	call := NewCall[githubCreateIssueResponse](Fasthttp()).
		URI("repos/sonirico/withttp/issues").
		Method(http.MethodPost).
		ContentType(ContentTypeJSON).
		Body(map[string]string{
			"title":    "test issue",
			"body":     "test body",
			"assignee": "sonirico",
		}).
		HeaderFunc(func() (key, value string, override bool) {
			key = "Authorization"
			value = "Bearer S3cret"
			override = true
			return
		}).
		ParseJSON().
		ExpectedStatusCodes(http.StatusCreated)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		t.Fatalf("Failed to call mocked endpoint: %v", err)
	}

	if call.BodyParsed.ID != 999 {
		t.Errorf("Expected ID 999, got %d", call.BodyParsed.ID)
	}

	expectedURL := "https://api.github.com/repos/sonirico/withttp/issues/999"
	if call.BodyParsed.URL != expectedURL {
		t.Errorf("Expected URL '%s', got %s", expectedURL, call.BodyParsed.URL)
	}
}

// Example_fasthttp_basicCall demonstrates how to make a simple HTTP call using the FastHTTP adapter.
func Example_fasthttp_basicCall() {
	type RepoInfo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Create a mocked endpoint for the example
	endpoint := NewEndpoint("GithubAPI").
		Request(BaseURL("https://api.github.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(`{"id": 12345, "name": "withttp"}`)))
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[RepoInfo](NewMockHttpClientAdapter()).
		URI("/repos/user/repo").
		Method(http.MethodGet).
		Header("User-Agent", "withttp-example", false).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Repo ID: %d, Name: %s\n", call.BodyParsed.ID, call.BodyParsed.Name)
	// Output: Repo ID: 12345, Name: withttp
}

// Example_fasthttp_postRequest demonstrates how to make a POST request with JSON body using FastHTTP.
func Example_fasthttp_postRequest() {
	type CreateRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	type CreateResponse struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}

	// Create a mocked endpoint for the example
	endpoint := NewEndpoint("API").
		Request(BaseURL("https://api.example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(
					io.NopCloser(
						strings.NewReader(`{"id": 42, "url": "https://api.example.com/items/42"}`),
					),
				)
				res.SetStatus(http.StatusCreated)
			}),
		)

	payload := CreateRequest{
		Title: "Example Item",
		Body:  "This is an example",
	}

	call := NewCall[CreateResponse](NewMockHttpClientAdapter()).
		URI("/items").
		Method(http.MethodPost).
		ContentType(ContentTypeJSON).
		Body(payload).
		Header("Authorization", "Bearer token", true).
		ParseJSON().
		ExpectedStatusCodes(http.StatusCreated)

	err := call.CallEndpoint(context.Background(), endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created item ID: %d\n", call.BodyParsed.ID)
	// Output: Created item ID: 42
}

// Example_fasthttp_withHeaderFunc demonstrates using header functions with FastHTTP.
func Example_fasthttp_withHeaderFunc() {
	type APIResponse struct {
		Message string `json:"message"`
	}

	endpoint := NewEndpoint("API").
		Request(BaseURL("https://api.example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(`{"message": "Hello"}`)))
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[APIResponse](NewMockHttpClientAdapter()).
		URI("/hello").
		Method(http.MethodGet).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Timestamp"
			value = "2025-06-29T12:00:00Z"
			override = true
			return
		}).
		HeaderFunc(func() (key, value string, override bool) {
			key = "X-Request-ID"
			value = "req-12345"
			override = true
			return
		}).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", call.BodyParsed.Message)
	// Output: Response: Hello
}
