# withttp

Build http requests and parse their responses with fluent syntax and wit. This package aims
to quickly configure http roundtrips by covering common scenarios, while leaving all details
of http requests and responses open for the user to allow maximun flexibility.

Supported underlying http implementations are:

 - [net/http](https://pkg.go.dev/net/http)
 - [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp)
 - open an issue to include your preferred one!

#### Query Restful endpoints

```go
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
    WithHeaderFunc(func() (key, value string, override bool) {
      key = "X-Date"
      value = time.Now().String()
      override = true
      return
    }).
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
