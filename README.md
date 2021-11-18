# melatonin

melatonin is a fluent, flexible REST API testing library for Go. It provides many of the benefits of a domain-specific test language but with the flexibililty of writing pure Go. Use it to write unit tests that test your `http.Handler`s routes directly, or E2E tests that target routes on a running service written in any language.

See the full [user guide](./USERGUIDE.md) and the [API documentation](https://pkg.go.dev/github.com/jefflinse/melatonin/mt) for more information.

## Installation

    go get github.com/jefflinse/melatonin/mt

## Usage

When built and run as a standalone binary, a melatonin app will print a formatted table of test results to stdout.

```go
package main

import "github.com/jefflinse/melatonin/mt"

func main() {

    myAPI := mt.NewURLContext("http://example.com")
    mt.RunTests([]mt.TestCase{

        myAPI.GET("/resource", "Fetch a record successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

    $ go run example.go
    running 1 tests for http://example.com
    ✔  Fetch a record successfully      GET   /foo  3.9252ms

    1 passed, 0 failed, 0 skipped in 3.9252ms

When run as a regular Go test, results will be reported through the standard `testing.T` context.

```go
package mypackage_test

import (
    "testing"
    "github.com/jefflinse/melatonin/mt"
)

func TestAPI(t *testing.T) {

    myAPI := mt.NewURLContext("http://example.com")
    mt.RunTestsT(t, []mt.TestCase{
        myAPI.GET("/resource", "Fetch a record successfully").
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

### Test a service running locally or remotely (E2E tests)

```go
myAPI := mt.NewURLContext("http://example.com")
mt.RunTests(...)
```

### Test a Go HTTP handler directly (unit tests)

```go
myAPI := mt.NewHandlerContext(http.NewServeMux())
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

See the full [V1 milestone](https://github.com/jefflinse/melatonin/milestone/1) for more.

## Contributing

Please [open an issue](https://github.com/jefflinse/melatonin/issues) if you find a bug or have a feature request.

## License

MIT License (MIT) - see [`LICENSE`](./LICENSE) for details.
