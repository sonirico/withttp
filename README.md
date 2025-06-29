<div align="center">

# 🎯 withttp

**Build HTTP requests and parse responses with fluent syntax and wit**

[![Build Status](https://github.com/sonirico/withttp/actions/workflows/go.yml/badge.svg)](https://github.com/sonirico/withttp/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonirico/withttp)](https://goreportcard.com/report/github.com/sonirico/withttp)
[![GoDoc](https://godoc.org/github.com/sonirico/withttp?status.svg)](https://godoc.org/github.com/sonirico/withttp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)

*A fluent HTTP client library that covers common scenarios while maintaining maximum flexibility*

</div>

---

## 🚀 Features

- **🔄 Fluent API** - Chain methods for intuitive request building
- **📡 Multiple HTTP Backends** - Support for `net/http` and `fasthttp`
- **🎯 Type-Safe Responses** - Generic-based response parsing
- **📊 Streaming Support** - Stream data from slices, channels, or readers
- **🧪 Mock-Friendly** - Built-in mocking capabilities for testing
- **⚡ High Performance** - Optimized for speed and low allocations

## 📦 Installation

```bash
go get github.com/sonirico/withttp
```

## 🎛️ Supported HTTP Implementations

| Implementation                                             | Description                                                                                     |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| [net/http](https://pkg.go.dev/net/http)                    | Go's standard HTTP client                                                                       |
| [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp) | High-performance HTTP client                                                                    |
| Custom Client                                              | Implement the [Client interface](https://github.com/sonirico/withttp/blob/main/endpoint.go#L43) |

> 💡 Missing your preferred HTTP client? [Open an issue](https://github.com/sonirico/withttp/issues/new) and let us know!

## 📚 Table of Contents

- [🎯 withttp](#-withttp)
  - [🚀 Features](#-features)
  - [📦 Installation](#-installation)
  - [🎛️ Supported HTTP Implementations](#️-supported-http-implementations)
  - [📚 Table of Contents](#-table-of-contents)
  - [🏁 Quick Start](#-quick-start)
  - [💡 Examples](#-examples)
    - [RESTful API Queries](#restful-api-queries)
    - [Streaming Data](#streaming-data)
      - [📄 Stream from Slice](#-stream-from-slice)
      - [📡 Stream from Channel](#-stream-from-channel)
      - [📖 Stream from Reader](#-stream-from-reader)
    - [Multiple Endpoints](#multiple-endpoints)
    - [Testing with Mocks](#testing-with-mocks)
  - [🗺️ Roadmap](#️-roadmap)
  - [🤝 Contributing](#-contributing)
  - [📄 License](#-license)
  - [⭐ Show Your Support](#-show-your-support)

## 🏁 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/sonirico/withttp"
)

type GithubRepo struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    URL  string `json:"html_url"`
}

func main() {
    call := withttp.NewCall[GithubRepo](withttp.Fasthttp()).
        URL("https://api.github.com/repos/sonirico/withttp").
        Method(http.MethodGet).
        Header("User-Agent", "withttp-example/1.0", false).
        ParseJSON().
        ExpectedStatusCodes(http.StatusOK)

    err := call.Call(context.Background())
    if err != nil {
        panic(err)
    }

    fmt.Printf("Repository: %s (ID: %d)\n", call.BodyParsed.Name, call.BodyParsed.ID)
}
```

## 💡 Examples

### RESTful API Queries

<details>
<summary>Click to expand</summary>

```go
type GithubRepoInfo struct {
  ID  int    `json:"id"`
  URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (GithubRepoInfo, error) {
  call := withttp.NewCall[GithubRepoInfo](withttp.Fasthttp()).
    URL(fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)).
    Method(http.MethodGet).
    Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
    ParseJSON().
    ExpectedStatusCodes(http.StatusOK)

  err := call.Call(context.Background())
  return call.BodyParsed, err
}

func main() {
  info, _ := GetRepoInfo("sonirico", "withttp")
  log.Println(info)
}
```

</details>

### Streaming Data

#### 📄 Stream from Slice

<details>
<summary>View example</summary>

[See full example](https://github.com/sonirico/withttp/blob/main/examples/request_stream/main.go)

```go
type metric struct {
  Time time.Time `json:"t"`
  Temp float32   `json:"T"`
}

func CreateStream() error {
  points := []metric{
    {Time: time.Unix(time.Now().Unix()-1, 0), Temp: 39},
    {Time: time.Now(), Temp: 40},
  }

  stream := withttp.Slice[metric](points)
  testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
    Request(withttp.BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"))

  call := withttp.NewCall[any](withttp.Fasthttp()).
    Method(http.MethodPost).
    ContentType(withttp.ContentTypeJSONEachRow).
    RequestSniffed(func(data []byte, err error) {
      fmt.Printf("recv: '%s', err: %v", string(data), err)
    }).
    RequestStreamBody(withttp.RequestStreamBody[any, metric](stream)).
    ExpectedStatusCodes(http.StatusOK)

  return call.CallEndpoint(context.Background(), testEndpoint)
}
```

</details>

#### 📡 Stream from Channel

<details>
<summary>View example</summary>

[See full example](https://github.com/sonirico/withttp/blob/main/examples/request_stream/main.go)

```go
func CreateStreamChannel() error {
  points := make(chan metric, 2)

  go func() {
    points <- metric{Time: time.Unix(time.Now().Unix()-1, 0), Temp: 39}
    points <- metric{Time: time.Now(), Temp: 40}
    close(points)
  }()

  stream := withttp.Channel[metric](points)
  testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
    Request(withttp.BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"))

  call := withttp.NewCall[any](withttp.Fasthttp()).
    Method(http.MethodPost).
    ContentType(withttp.ContentTypeJSONEachRow).
    RequestSniffed(func(data []byte, err error) {
      fmt.Printf("recv: '%s', err: %v", string(data), err)
    }).
    RequestStreamBody(withttp.RequestStreamBody[any, metric](stream)).
    ExpectedStatusCodes(http.StatusOK)

  return call.CallEndpoint(context.Background(), testEndpoint)
}
```

</details>

#### 📖 Stream from Reader

<details>
<summary>View example</summary>

[See full example](https://github.com/sonirico/withttp/blob/main/examples/request_stream/main.go)

```go
func CreateStreamReader() error {
  buf := bytes.NewBuffer(nil)

  go func() {
    buf.WriteString("{\"t\":\"2022-09-01T00:58:15+02:00\"")
    buf.WriteString(",\"T\":39}\n{\"t\":\"2022-09-01T00:59:15+02:00\",\"T\":40}\n")
  }()

  streamFactory := withttp.NewProxyStreamFactory(1 << 10)
  stream := withttp.NewStreamFromReader(buf, streamFactory)
  testEndpoint := withttp.NewEndpoint("webhook-site-request-stream-example").
    Request(withttp.BaseURL("https://webhook.site/24e84e8f-75cf-4239-828e-8bed244c0afb"))

  call := withttp.NewCall[any](withttp.NetHttp()).
    Method(http.MethodPost).
    RequestSniffed(func(data []byte, err error) {
      fmt.Printf("recv: '%s', err: %v", string(data), err)
    }).
    ContentType(withttp.ContentTypeJSONEachRow).
    RequestStreamBody(withttp.RequestStreamBody[any, []byte](stream)).
    ExpectedStatusCodes(http.StatusOK)

  return call.CallEndpoint(context.Background(), testEndpoint)
}
```

</details>

### Multiple Endpoints

<details>
<summary>Click to expand</summary>

Define reusable endpoint configurations for API consistency:

```go
var (
  githubApi = withttp.NewEndpoint("GithubAPI").
    Request(withttp.BaseURL("https://api.github.com/"))
)

type GithubRepoInfo struct {
  ID  int    `json:"id"`
  URL string `json:"html_url"`
}

func GetRepoInfo(user, repo string) (GithubRepoInfo, error) {
  call := withttp.NewCall[GithubRepoInfo](withttp.Fasthttp()).
    URI(fmt.Sprintf("repos/%s/%s", user, repo)).
    Method(http.MethodGet).
    Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
    HeaderFunc(func() (key, value string, override bool) {
      return "X-Date", time.Now().String(), true
    }).
    ParseJSON().
    ExpectedStatusCodes(http.StatusOK)

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

  p := payload{Title: title, Body: body, Assignee: assignee}

  call := withttp.NewCall[GithubCreateIssueResponse](withttp.Fasthttp()).
    URI(fmt.Sprintf("repos/%s/%s/issues", user, repo)).
    Method(http.MethodPost).
    ContentType("application/vnd+github+json").
    Body(p).
    HeaderFunc(func() (key, value string, override bool) {
      return "Authorization", fmt.Sprintf("Bearer %s", "S3cret"), true
    }).
    ExpectedStatusCodes(http.StatusCreated)

  err := call.CallEndpoint(context.Background(), githubApi)
  log.Println("req body", string(call.Req.Body()))

  return call.BodyParsed, err
}

func main() {
  // Fetch repo info
  info, _ := GetRepoInfo("sonirico", "withttp")
  log.Println(info)

  // Create an issue
  res, err := CreateRepoIssue("sonirico", "withttp", "test", "This is a test", "sonirico")
  log.Println(res, err)
}
```

</details>

### Testing with Mocks

<details>
<summary>Click to expand</summary>

Easily test your HTTP calls with built-in mocking:

```go
var (
  exchangeListOrders = withttp.NewEndpoint("ListOrders").
    Request(withttp.BaseURL("http://example.com")).
    Response(
      withttp.MockedRes(func(res withttp.Response) {
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

  call := withttp.NewCall[Order](withttp.Fasthttp()).
    URL("https://github.com/").
    Method(http.MethodGet).
    Header("User-Agent", "withttp/0.5.1 See https://github.com/sonirico/withttp", false).
    ParseJSONEachRowChan(res).
    ExpectedStatusCodes(http.StatusOK)

  go func() {
    for order := range res {
      log.Println(order)
    }
  }()

  err := call.CallEndpoint(context.Background(), exchangeListOrders)
  if err != nil {
    panic(err)
  }
}
```

</details>

## 🗺️ Roadmap

| Feature                       | Status        |
| ----------------------------- | ------------- |
| Form-data content type codecs | 🔄 In Progress |
| Enhanced auth methods         | 📋 Planned     |
| XML parsing support           | 📋 Planned     |
| Tabular data support          | 📋 Planned     |
| gRPC integration              | 🤔 Considering |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⭐ Show Your Support

If this project helped you, please give it a ⭐! It helps others discover the project.

---

<div align="center">

**[Documentation](https://godoc.org/github.com/sonirico/withttp)** • 
**[Examples](https://github.com/sonirico/withttp/tree/main/examples)** • 
**[Issues](https://github.com/sonirico/withttp/issues)** • 
**[Discussions](https://github.com/sonirico/withttp/discussions)**

Made with ❤️ by [sonirico](https://github.com/sonirico)

</div>
