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

    itest.RunTests("http://example.com", []*itest.TestCase{

        itest.GET("/endpoint").
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

    itest.RunTestsT(t, "http://example.com", []*itest.TestCase{

        itest.GET("/endpoint").
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

### Create a test runner

```go
runner := itest.NewTestRunner("http://example.com")
runner.RunTests(...)
```

### Define tests using chainable methods

```go
tests := []*itest.TestCase{

    itest.GET("/endpoint").
       ExpectStatus(200).
       ExpectBody(String("Hello, World!")),
    
    itest.POST("/endpoint").
       WithBody(Object{
         "name": "Burt Macklin",
         "age":  32,
       }).
       ExpectStatus(201),
    
    itest.DELETE("/endpoint/42").
       ExpectStatus(204),
}
```

### Define tests using structs

```go
tests := []*itest.TestCase{

    {
        Method: "GET",
        Path: "/endpoint",
        WantStatus: 200,
        WantBody: String("Hello, World!"),
    },
    {
        Method: "POST",
        Path: "/endpoint",
        RequestBody: Object{
            "name": "Burt Macklin",
            "age":  32,
        },
        WantStatus: 201,
    },
    {
        Method: "DELETE",
        Path: "/endpoint/42",
        WantStatus: 204,
    },
}
```

### Specify a custom HTTP client for requests

```go
client, err := &http.Client{}
runner := itest.NewRunner("http://example.com").WithHTTPClient(client)
```

### Specify a custom timeout for all tests

```go
runner := itest.NewRunner("http://example.com").WithTimeout(5 * time.Second)
```

### Specify a timeout for a specific test

```go
itest.GET("/endpoint").
    WithTimeout(5 * time.Second).
    ExpectStatus(200).
```

### Allow or disallow further tests to run after a failure

```go
runner := itest.NewRunner("http://example.com").WithContinueOnFailure(true)
```

### Define a test case with a custom HTTP request

```go
req, err := http.NewRequest("GET", "http://example.com/endpoint", nil)

itest.DO(req).
    ExpectStatus(200)
```

## Planned Features

- Output test results in different formats (e.g. JSON, XML, YAML)
- Standalone tool for running tests defined in text files
- Support for sourcing response expectations from golden files
- Support for running external commands before and after test cases

## Contributing

Please [open an issue](https://github.com/jefflinse/go-itest/issues) if you find a bug or have a feature request.

## License

MIT License (MIT) - see [`LICENSE.md`](https://github.com/jefflinse/go-itest/blob/master/LICENSE.md) for details.
