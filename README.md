![build](https://github.com/sonirico/withttp/actions/workflows/go.yml/badge.svg)

# withttp

Build http requests and parse their responses with fluent syntax and wit. This package aims
to quickly configure http roundtrips by covering common scenarios, while leaving all details
of http requests and responses open for developers to allow maximum flexibility.

Supported underlying http implementations are:

 - [net/http](https://pkg.go.dev/net/http)
 - [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp)
 - open an issue to include your preferred one!

#### Query Restful endpoints

```go
type GithubRepoInfo struct {
  ID  int    `json:"id"`
  URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (GithubRepoInfo, error) {

  call := withttp.NewCall[GithubRepoInfo](withttp.NewDefaultFastHttpHttpClientAdapter()).
    WithURL(fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)).
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
```

In case of a wide range catalog of endpoints, predefined parameters and behaviours can be
defined by employing an endpoint definition.

```go
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
```

#### Test your calls again a mock endpoint

Quickly test your calls by creating a mock endpoint

```go
var (
  exchangeListOrders = withttp.NewEndpoint("ListOrders").
        Request(withttp.WithURL("http://example.com")).
        Response(
      withttp.WithResMock(func(res withttp.Response) {
        res.SetBody(io.NopCloser(bytes.NewReader(mockResponse)))
        res.SetStatus(http.StatusOK)
      }),
    )
  mockResponse = []byte(strings.TrimSpace(`
    {"amount": 234, "pair": "BTC/USDT"}
    {"amount": 123, "pair": "ETH/USDT"}`))
)

func main() {
  type Order struct {
    Amount float64 `json:"amount"`
    Pair   string  `json:"pair"`
  }

  res := make(chan Order)

  call := withttp.NewCall[Order](withttp.NewDefaultFastHttpHttpClientAdapter()).
    WithURL("https://github.com/").
    WithMethod(http.MethodGet).
    WithHeader("User-Agent", "withttp/0.1.0 See https://github.com/sonirico/withttp", false).
    WithJSONEachRowChan(res).
    WithExpectedStatusCodes(http.StatusOK)

  go func() {
    for order := range res {
      log.Println(order)
    }
  }()

  err := call.Call(context.Background(), exchangeListOrders)

  if err != nil {
    panic(err)
  }
}
```
