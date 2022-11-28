# melatonin

[![Build Status](https://github.com/jefflinse/melatonin/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jefflinse/melatonin/actions/workflows/ci.yml)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/jefflinse/melatonin)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jefflinse/melatonin)
[![Go Report Card](https://goreportcard.com/badge/github.com/jefflinse/melatonin)](https://goreportcard.com/report/github.com/jefflinse/melatonin)
[![Go Reference](https://pkg.go.dev/badge/github.com/jefflinse/melatonin/mt.svg)](https://pkg.go.dev/github.com/jefflinse/melatonin/mt)
![License](https://img.shields.io/github/license/jefflinse/melatonin)

**Melatonin is a flexible API testing library for Go.**

It provides syntactic sugar for writing table-based API tests at any level of testing.

Use it to write:

- **Native Go tests** that test your `http.Handler`s routes directly. Mock out your dependencies and test your handler logic in isolation. [More »](#native-go-tests)

- **Component tests** that target any running service. Spin up your service with stubbed dependencies and test the API surface. [More »](#component-tests)

- **E2E test suites** that target APIs across multiple running services. Perform acceptance tests against your entire system. [More »](#e2e-test-suites)

See the full [user guide](./USERGUIDE.md) and the [API documentation](https://pkg.go.dev/github.com/jefflinse/melatonin/mt) for more information.

## Installation

    go get github.com/jefflinse/melatonin/mt

## Usage

### Native Go tests

A `HandlerContext` wraps a Go `http.Handler` (such as a mux/router) and provides methods for defining tests that run against the it. This is useful, for example, for testing the logic of your mux and individual handlers in isolation with mocked dependencies.

```go
func TestMyAPI(t *testing.T) {
    // myHandler can be anything implementing http.Handler
    myAPI := mt.NewHandlerContext(myHandler)
    mt.RunTestsT(t, []mt.TestCase{

        myAPI.GET("/resource", "Fetch a resource successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

Run these tests with `go test`, just like any other Go tests.

### Component tests

A `URLContext` wraps a base URL and provides methods for defining tests that run against the API at that URL. This is useful for blackbox testing the API surface of a service, either with real or stubbed external dependencies.

```go
func main() {
    // myURL can be any valid base URL parsable by url.Parse()
    myAPI := mt.NewURLContext(myURL)
    results := mt.RunTests([]mt.TestCase{

        myAPI.GET("/resource", "Fetch a resource successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })

    mt.PrintResults(results)
}
```

### E2E test suites

Similar to component tests, it's easy to create multiple test contexts (i.e. one per service) and define test suites that execute high-level user stories across your entire system.

```go
func main() {
    authAPI := mt.NewURLContext("https://myapi.example.com/auth")
    usersAPI := mt.NewURLContext("https://myapi.example.com/users")

    var uid, token string

    results := mt.RunTests([]mt.TestCase{

        authAPI.POST("/login", "Can log in").
            WithBody(json.Object{
                "username": "someone@example.com",
                "password": "password",
            }).
            ExpectStatus(200).
            ExpectBody(json.Object{
                "uid":           bind.String(&uid)
                "access_token":  bind.String(&token),
                "refresh_token": expect.String(),
            }),

        usersAPI.GET("/:id/profile}", "Can fetch own profile").
            WithHeader("Authorization", "Bearer " + &token).
            WithPathParam("id", &uid).
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })

    mt.PrintResults(results)
}
```

More on data binding and expectations can be found in the [user guide](./USERGUIDE.md).

## Examples

See the [examples](examples) directory for full, runnable examples.

### Test a Go HTTP handler

```go
myAPI := mt.NewHandlerContext(http.NewServeMux())
mt.RunTests(...)
```

### Test a base URL endpoint

```go
myAPI := mt.NewURLContext("http://example.com")
mt.RunTests(...)
```

### Define tests

```go
myAPI := mt.NewURLContext("http://example.com")
tests := []mt.TestCase{

    myAPI.GET("/resource").
       ExpectStatus(200).
       ExpectBody(String("Hello, World!")),
    
    myAPI.POST("/resource").
       WithBody(Object{
         "name": "Burt Macklin",
         "age":  32,
       }).
       ExpectStatus(201),
    
    myAPI.DELETE("/resource/42").
       ExpectStatus(204),
}
```

### Use a custom HTTP client for requests

```go
client := &http.Client{}
myAPI := mt.NewURLContext("http://example.com").WithHTTPClient(client)
```

### Use a custom timeout for all tests

```go
timeout := time.Duration(5 * time.Second)
myAPI := mt.NewURLContext("http://example.com").WithTimeout(timeout)
```

### Specify a timeout for a specific test

```go
myAPI.GET("/resource").
    WithTimeout(5 * time.Second).
    ExpectStatus(200).
```

### Specify query parameters for a test

Inline:

```go
myAPI.GET("/resource?first=foo&second=bar")
```

Individually:

```go
myAPI.GET("/resource").
    WithQueryParam("first", "foo").
    WithQueryParam("second", "bar")
```

All At Once:

```go
myAPI.GET("/resource").
    WithQueryParams(url.Values{
        "first": []string{"foo"},
        "second": []string{"bar"},
    })
```

### Allow or disallow further tests to run after a failure

```go
runner := mt.NewURLContext("http://example.com").WithContinueOnFailure(true)
```

### Create a test case with a custom HTTP request

```go
req, err := http.NewRequest("GET", "http://example.com/resource", nil)
myAPI.DO(req).
    ExpectStatus(200)
```

### Expect exact headers and JSON body content

Any unexpected headers or JSON keys or values present in the response will cause the test case to fail.

```go
myAPI.GET("/resource").
    ExpectExactHeaders(http.Header{
        "Content-Type": []string{"application/json"},
    }).
    ExpectExactBody(mt.Object{
        "foo": "bar",
    })
```

### Load expectations for a test case from a golden file

```go
myAPI.GET("/resource").
    ExpectGolden("path/to/file.golden")
```

Golden files keep your test definitions short and concise by storing expectations in a file. See the [golden file format specification](./golden/README.md).

## Planned Features

- Output test results in different formats (e.g. JSON, XML, YAML)
- Generate test cases from an OpenAPI specification
- Support for testing GraphQL APIs
- Support for testing gRPC APIs
- Support for testing websockets

See the full [V1 milestone](https://github.com/jefflinse/melatonin/milestone/1) for more.

## Contributing

Please [open an issue](https://github.com/jefflinse/melatonin/issues) if you find a bug or have a feature request.

## License

MIT License (MIT) - see [`LICENSE`](./LICENSE) for details.
