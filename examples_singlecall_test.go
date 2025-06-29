package withttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type githubRepoInfo struct {
	ID  int    `json:"id"`
	URL string `json:"html_url"`
}

func getRepoInfo(user, repo string) (githubRepoInfo, error) {
	call := NewCall[githubRepoInfo](NetHttp()).
		URL(fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)).
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.Call(context.Background())

	return call.BodyParsed, err
}

func TestSingleCall_Example(t *testing.T) {
	// This is primarily an example test - it makes a real API call
	// In production tests, you might want to mock this
	t.Skip("Skipping live API test - enable manually for integration testing")

	info, err := getRepoInfo("sonirico", "withttp")
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

func TestSingleCall_MockedExample(t *testing.T) {
	// This test uses a mocked endpoint to test the functionality without making real API calls
	mockEndpoint := NewEndpoint("MockGithubAPI").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(
					io.NopCloser(
						strings.NewReader(
							`{"id": 12345, "html_url": "https://github.com/sonirico/withttp"}`,
						),
					),
				)
				res.SetStatus(http.StatusOK)
			}),
		)

	call := NewCall[githubRepoInfo](NetHttp()).
		URI("repos/sonirico/withttp").
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), mockEndpoint)
	if err != nil {
		t.Fatalf("Failed to call mocked endpoint: %v", err)
	}

	if call.BodyParsed.ID != 12345 {
		t.Errorf("Expected ID 12345, got %d", call.BodyParsed.ID)
	}

	if call.BodyParsed.URL != "https://github.com/sonirico/withttp" {
		t.Errorf("Expected URL 'https://github.com/sonirico/withttp', got %s", call.BodyParsed.URL)
	}
}

// Example_singleCall demonstrates making a simple HTTP GET request.
func Example_singleCall() {
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

// Example_singleCall_directURL demonstrates making a request directly to a URL without an endpoint.
func Example_singleCall_directURL() {
	type SimpleResponse struct {
		Status string `json:"status"`
	}

	// For this example, we'll simulate a call but skip it in practice
	// In real usage, you would just call without the skip
	call := NewCall[SimpleResponse](NetHttp()).
		URL("https://httpbin.org/json").
		Method(http.MethodGet).
		Header("User-Agent", "withttp-example", false).
		ParseJSON().
		ExpectedStatusCodes(http.StatusOK)

	// For documentation purposes, we'll show what the call would look like
	_ = call // Normally: err := call.Call(context.Background())

	fmt.Println("This would make a direct HTTP call to the specified URL")
	// Output: This would make a direct HTTP call to the specified URL
}
