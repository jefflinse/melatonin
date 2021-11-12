# melatonin

Hassle-free REST API testing for Go.

melatonin is a fluent, flexible REST API testing library for Go. It provides many of the benefits of domain-specific test language but with the flexibililty of writing pure Go. Use it to write unit tests that test your `http.Handler`s directly, or target actual local or remote service endpoints to perform E2E tests of your API service.

See the full [user guide](./USERGUIDE.md) and the [API documentation](https://pkg.go.dev/github.com/jefflinse/melatonin/mt) for more information.

melatonin is very usable in its current state but has not yet reached its V1 release milestone. As such, the API surface may change without notice until then. See the roadmap for more information.

## Installation

    go get github.com/jefflinse/melatonin/mt

## Usage

melatonin can run as a standalone binary built with go build`. When run in this manner, the program will output a formatted table of test results to stdout.

melatonin can also run as a set of regular Go tests, in which case results will be reported through the usual `testing.T` context.

**As A Go Program**

```go
package main

import "github.com/jefflinse/melatonin/mt"

func main() {

    mt.TestEndpoint("http://example.com", []*mt.TestCase{

        mt.GET("/resource", "Fetch a record successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

    $ go run example.go
    running 1 tests for http://example.com
    ✔  Fetch a record successfully      GET   /foo  3.9252ms

    1 passed, 0 failed, 0 skipped in 3.9252ms

**As Go Tests**

```go
package mypackage_test

import (
    "testing"
    "github.com/jefflinse/melatonin/mt"
)

func TestAPI(t *testing.T) {

    mt.TestEndpointT(t, "http://example.com", []*mt.TestCase{

        mt.GET("/resource", "Fetch a record successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

    $ go test
    running 1 tests for http://localhost:8080
    1 passed, 0 failed, 0 skipped in 2.373352ms
    PASS
    ok      github.com/my/api    0.144s

## Examples

Check out the [examples](examples/README.md) directory for more examples.

### Test a service runnnig locally or remotely (E2E tests)

```go
runner := mt.NewEndpointTester("http://example.com")
runner.RunTests(...)
```

### Test an HTTP handler directly (unit tests)

```go
runner := mt.NewHandlerTester(http.NewServeMux())
runner.RunTests(...)
```

### Define tests using chainable methods

```go
tests := []*mt.TestCase{

    mt.GET("/resource").
       ExpectStatus(200).
       ExpectBody(String("Hello, World!")),
    
    mt.POST("/resource").
       WithBody(Object{
         "name": "Burt Macklin",
         "age":  32,
       }).
       ExpectStatus(201),
    
    mt.DELETE("/resource/42").
       ExpectStatus(204),
}
```

### Define tests using structs

```go
tests := []*mt.TestCase{

    {
        Method: "GET",
        Path: "/resource",
        WantStatus: 200,
        WantBody: String("Hello, World!"),
    },
    {
        Method: "POST",
        Path: "/resource",
        RequestBody: Object{
            "name": "Burt Macklin",
            "age":  32,
        },
        WantStatus: 201,
    },
    {
        Method: "DELETE",
        Path: "/resource/42",
        WantStatus: 204,
    },
}
```

### Use a custom HTTP client for requests

```go
client, err := &http.Client{}
runner := mt.NewEndpointTester("http://example.com").WithHTTPClient(client)
```

### Use a custom timeout for all tests

```go
runner := mt.NewEndpointTester("http://example.com").WithTimeout(5 * time.Second)
```

### Specify a timeout for a specific test

```go
mt.GET("/resource").
    WithTimeout(5 * time.Second).
    ExpectStatus(200).
```

### Specify query parameters for a test

Inline:

```go
mt.GET("/resource?first=foo&second=bar")
```

Individually:

```go
mt.GET("/resource").
    WithQueryParam("first", "foo").
    WithQueryParam("second", "bar")
```

All At Once:

```go
mt.GET("/resource").
    WithQueryParams(url.Values{
        "first": []string{"foo"},
        "second": []string{"bar"},
    })
```

### Allow or disallow further tests to run after a failure

```go
runner := mt.NewEndpointTester("http://example.com").WithContinueOnFailure(true)
```

### Create a test case with a custom HTTP request

```go
req, err := http.NewRequest("GET", "http://example.com/resource", nil)

mt.DO(req).
    ExpectStatus(200)
```

### Expect exact headers and JSON body content

Any unexpected headers or JSON keys or values present in the response will cause the test case to fail.

```go
mt.GET("/resource").
    ExpectExactHeaders(http.Header{
        "Content-Type": []string{"application/json"},
    }).
    ExpectExactBody(mt.Object{
        "foo": "bar",
    })
```

### Load expectations for a test case from a golden file

```go
mt.GET("/resource").
    ExpectGolden("path/to/file.golden")
```

Golden files keep your test definitions short and concise by storing expectations in a file. See the [golden file format specification](./golden/README.md).

## Planned Features

- Output test results in different formats (e.g. JSON, XML, YAML)
- Standalone tool for running tests defined in text files
- Support for running external commands before and after test cases
- Interfaces to allow wrapping of test cases with custom logic (e.g. AWS Lambda, etc.)
- Generate test cases from an OpenAPI specification

## Contributing

Please [open an issue](https://github.com/jefflinse/melatonin/issues) if you find a bug or have a feature request.

## License

MIT License (MIT) - see [`LICENSE`](./LICENSE) for details.
