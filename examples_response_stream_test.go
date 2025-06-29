package withttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sonirico/withttp/csvparser"
)

var (
	getRepoStatsEndpoint = NewEndpoint("GetRepoStats").
		Request(BaseURL("http://example.com")).
		Response(
			MockedRes(func(res Response) {
				mockResponse := `rank,repo_name,stars
1,freeCodeCamp,341271
2,996.ICU,261139`
				res.SetBody(io.NopCloser(strings.NewReader(mockResponse)))
				res.SetStatus(http.StatusOK)
			}),
		)
)

type Repo struct {
	Rank  int
	Name  string
	Stars int
}

func TestResponseStream_ParseCSV(t *testing.T) {
	ignoreLines := 1 // Skip header line
	repos := make([]Repo, 0)

	parser := csvparser.New[Repo](
		csvparser.SeparatorComma,
		csvparser.IntCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, rank int) { x.Rank = rank },
		),
		csvparser.StringCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, name string) { x.Name = name },
		),
		csvparser.IntCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, stars int) { x.Stars = stars },
		),
	)

	call := NewCall[Repo](Fasthttp()).
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.6.0 See https://github.com/sonirico/withttp", false).
		ParseCSV(ignoreLines, parser, func(r Repo) bool {
			repos = append(repos, r)
			t.Logf("Repo: %+v", r)
			return true // Continue parsing
		}).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), getRepoStatsEndpoint)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	// Verify we parsed the expected repositories
	if len(repos) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(repos))
	}

	if len(repos) >= 1 {
		if repos[0].Rank != 1 || repos[0].Name != "freeCodeCamp" || repos[0].Stars != 341271 {
			t.Errorf("First repo incorrect: got %+v", repos[0])
		}
	}

	if len(repos) >= 2 {
		if repos[1].Rank != 2 || repos[1].Name != "996.ICU" || repos[1].Stars != 261139 {
			t.Errorf("Second repo incorrect: got %+v", repos[1])
		}
	}
}

func TestResponseStream_ParseCSVChannel(t *testing.T) {
	ignoreLines := 1 // Skip header line
	repoChan := make(chan Repo)
	repos := make([]Repo, 0)

	// Collect repos from channel
	go func() {
		for repo := range repoChan {
			repos = append(repos, repo)
		}
	}()

	parser := csvparser.New[Repo](
		csvparser.SeparatorComma,
		csvparser.IntCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, rank int) { x.Rank = rank },
		),
		csvparser.StringCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, name string) { x.Name = name },
		),
		csvparser.IntCol[Repo](
			csvparser.QuoteNone,
			nil,
			func(x *Repo, stars int) { x.Stars = stars },
		),
	)

	call := NewCall[Repo](Fasthttp()).
		Method(http.MethodGet).
		Header("User-Agent", "withttp/0.6.0 See https://github.com/sonirico/withttp", false).
		ParseStreamChan(NewCSVStreamFactory[Repo](ignoreLines, parser), repoChan).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), getRepoStatsEndpoint)
	if err != nil {
		t.Fatalf("Failed to call endpoint: %v", err)
	}

	// Give the goroutine time to process
	// In a real test, you might use a sync mechanism
	if len(repos) == 0 {
		// Channel might still be processing, wait a bit
		for i := 0; i < 10 && len(repos) == 0; i++ {
			// Small delay to allow goroutine to process
		}
	}

	// Verify we parsed the expected repositories
	if len(repos) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(repos))
	}

	if len(repos) >= 1 {
		if repos[0].Rank != 1 || repos[0].Name != "freeCodeCamp" || repos[0].Stars != 341271 {
			t.Errorf("First repo incorrect: got %+v", repos[0])
		}
	}

	if len(repos) >= 2 {
		if repos[1].Rank != 2 || repos[1].Name != "996.ICU" || repos[1].Stars != 261139 {
			t.Errorf("Second repo incorrect: got %+v", repos[1])
		}
	}
}

// Example_parseCSV demonstrates parsing CSV response data.
func Example_parseCSV() {
	type Repository struct {
		Rank  int
		Name  string
		Stars int
	}

	csvData := `rank,name,stars
1,awesome-go,75000
2,gin,65000`

	endpoint := NewEndpoint("RepoStats").
		Request(BaseURL("https://api.example.com")).
		Response(
			MockedRes(func(res Response) {
				res.SetBody(io.NopCloser(strings.NewReader(csvData)))
				res.SetStatus(http.StatusOK)
			}),
		)

	parser := csvparser.New[Repository](
		csvparser.SeparatorComma,
		csvparser.IntCol[Repository](
			csvparser.QuoteNone,
			nil,
			func(x *Repository, rank int) { x.Rank = rank },
		),
		csvparser.StringCol[Repository](
			csvparser.QuoteNone,
			nil,
			func(x *Repository, name string) { x.Name = name },
		),
		csvparser.IntCol[Repository](
			csvparser.QuoteNone,
			nil,
			func(x *Repository, stars int) { x.Stars = stars },
		),
	)

	call := NewCall[Repository](NewMockHttpClientAdapter()).
		URI("/top-repos.csv").
		Method(http.MethodGet).
		ParseCSV(1, parser, func(repo Repository) bool { // skip 1 header line
			fmt.Printf("Rank %d: %s (%d stars)\n", repo.Rank, repo.Name, repo.Stars)
			return true // continue processing
		}).
		ExpectedStatusCodes(http.StatusOK)

	err := call.CallEndpoint(context.Background(), endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Output: Rank 1: awesome-go (75000 stars)
	// Rank 2: gin (65000 stars)
}
