# User Guide

## Test Contexts

A test context used to create tests that target a particular service at a base URL or a Go `http.Handler`.

### URL Context

Create a URL context to create endpoint test cases that target a base URL. This enables you to write E2E tests that target any HTTP services running locally or remotely.

For example, you can test a service running locally on port 8080:

```go
ctx := mt.NewURLContext("http://localhost:8080")
```

You can just as easily test a service running remotely:

```go
ctx := mt.NewURLContext("http://example.com")
```

It's often useful to customize the HTTP client object that is used to make requests. For example, you may want to customize the transport or add cookies to the requests.

```go
client := &http.Client{
    // configure as needed
}

ctx := mt.NewURLContext("http://example.com")
    .WithHTTPClient(client)
```

### Handler Context

Create a handler context to create endpoint test cases that target a Go `http.Handler`. This enables you to write unit tests that test the handler logic directly without involving the network stack.

Any handler type that implements the `http.Handler` interface can be tested. This could be a basic `http.ServeMux`, a 3rd party router/mux such as [Gorilla](https://github.com/gorilla/mux) or [Gin](https://github.com/gin-gonic/gin), or a custom handler.

```go
mux := http.NewServeMux()
ctx := mt.NewHandlerContext(mux)
```

## Creating and Running Test Cases

The basic unit of a melatonin test is a test case. Test cases are created using test contexts. They can be run using a test runner, or manually by calling `Execute()`.

A bare minimum test case looks like this:

```go
ctx := mt.NewURLContext("http://localhost")
testcase := ctx.GET("/foo")
result, err := testcase.Execute()
```

Running this test will make a GET request to `http://localhost/foo`. This isn't very interesting, however, since we don't seem to care about what we get as a response. Let's add some assertions to the test case:

```go
ctx := mt.NewURLContext("http://localhost")
testcase := ctx.GET("/foo").
    ExpectStatus(http.StatusOK).
    ExpectBody("Hello, World!")
```

## Test Results

