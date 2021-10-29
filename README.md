# go-itest

Hassle-free REST API testing for Go.

## Installation

    go get github.com/jefflinse/go-itest

## Usage

Create tests for your API endpoints and run them as a standalone binary or as part of your usual Go tests.

**Standalone**

```go
package main

import "github.com/jefflinse/go-itest"

func main() {

    itest.TestEndpoint("http://example.com", []*itest.TestCase{

        itest.GET("/resource").
            Describe("Fetch a record successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

    $ go run example.go
    running 1 test for http://example.com
    ✔  Fetch a record successfully      GET   /foo  3.9252ms

    1 passed, 0 failed, 0 skipped in 3.9252ms

**Go Test**

```go
package mypackage_test

import "github.com/jefflinse/go-itest"

func TestAPI(t *testing.T) {

    itest.TestEndpointT(t, "http://example.com", []*itest.TestCase{

        itest.GET("/resource").
            Describe("Fetch a record successfully").
            ExpectStatus(200).
            ExpectBody("Hello, world!"),
    })
}
```

    $ go test
    running 1 test for http://example.com
    ✔  Fetch a record successfully second      GET   /foo  2.876222ms
    
    1 passed, 0 failed, 0 skipped in 2.876222ms
    PASS
    ok      github.com/jefflinse/go-itest/examples    0.135s

## Examples

### Test a service runnnig locally or remotely (E2E tests)

```go
runner := itest.NewEndpointTester("http://example.com")
runner.RunTests(...)
```

### Test an HTTP handler directly (unit tests)

```go
runner := itest.NewHandlerTester(http.NewServeMux())
runner.RunTests(...)
```

### Define tests using chainable methods

```go
tests := []*itest.TestCase{

    itest.GET("/resource").
       ExpectStatus(200).
       ExpectBody(String("Hello, World!")),
    
    itest.POST("/resource").
       WithBody(Object{
         "name": "Burt Macklin",
         "age":  32,
       }).
       ExpectStatus(201),
    
    itest.DELETE("/resource/42").
       ExpectStatus(204),
}
```

### Define tests using structs

```go
tests := []*itest.TestCase{

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
runner := itest.NewEndpointTester("http://example.com").WithHTTPClient(client)
```

### Use a custom timeout for all tests

```go
runner := itest.NewEndpointTester("http://example.com").WithTimeout(5 * time.Second)
```

### Specify a timeout for a specific test

```go
itest.GET("/resource").
    WithTimeout(5 * time.Second).
    ExpectStatus(200).
```

### Specify query parameters for a test

Inline:

```go
itest.GET("/resource?first=foo&second=bar")
```

Individually:

```go
itest.GET("/resource").
    WithQueryParam("first", "foo").
    WithQueryParam("second", "bar")
```

All At Once:

```go
itest.GET("/resource").
    WithQueryParams(url.Values{
        "first": []string{"foo"},
        "second": []string{"bar"},
    })
```

### Allow or disallow further tests to run after a failure

```go
runner := itest.NewEndpointTester("http://example.com").WithContinueOnFailure(true)
```

### Create a test case with a custom HTTP request

```go
req, err := http.NewRequest("GET", "http://example.com/resource", nil)

itest.DO(req).
    ExpectStatus(200)
```

### Expect exact headers and JSON body content

Any unexpected headers or JSON keys or values present in the response will cause the test case to fail.

```go
itest.GET("/resource").
    ExpectExactHeaders(http.Header{
        "Content-Type": []string{"application/json"},
    }).
    ExpectExactBody(itest.Object{
        "foo": "bar",
    })
```

### Load expectations for a test case from a golden file

```go
itest.GET("/resource").
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

Please [open an issue](https://github.com/jefflinse/go-itest/issues) if you find a bug or have a feature request.

## License

MIT License (MIT) - see [`LICENSE`](./LICENSE) for details.
