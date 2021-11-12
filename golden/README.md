# Golden Files

Generally speaking, [golden files](https://softwareengineering.stackexchange.com/a/358792) contain the expected output of a program or process being tested. They are especially useful if your expected output is lengthy or complex.

melatonin uses a very simple golden file format to specify the expected status code, headers, and body content of a test result:

```
200
--- headers
Content-Type: application/json
--- body
{
  "message": "Hello, world!"
}
```

## Status Code

The first line of every golden file is required to be the expected status code. This is the minimum requirement for a valid golden file.

```
200
```

## Headers

An optional headers section may be defined after the status code but before any body section. The section must begin with the exact text `--- headers` and subsequent lines will be treated as headers defined as `key: value` pairs. Header lines are read until a `--- body` section or EOF is encountered. Headers values are appended to the specified key in the order they're read.

```
200
--- headers
Content-Type: application/json
My-Custom-Header: foo
My-Custom-Header: bar
```

By default, any headers present in an actual test result that are not present in the golden file will be ignored. To fail a test if an unexpected header is present in the response, use the `exact` directive in the header section declaration:

```
200
--- headers exact
Content-Type: application/json
My-Custom-Header: foo
My-Custom-Header: bar
```

## Body

An option body section may be defined after the status code and headers section. The section must begin with the exact text `--- body` and subsequent lines will be treated as the expected body content of the response. Body content is read until EOF is encountered.

```
200
--- body
Hello, world!
```

By default, a test result's body content is expected to match the content in the golden file exactly. This means that JSON content will also be matched using a simple string comparison, including whitespace. To match JSON semantically, use the `json` directive in the body section declaration:

```
200
--- body json
{
  "message:": "Hello, world!"
}
```

When comparing a JSON response to the expected content, melatonin will ensure that all keys or elements present in the expected content are present in the actual response and that their values match. This allows for one to specify just the subset of a JSON response of interest to the test. To fail a test if an unexpected key or element is present in the response (that is, to match the JSON content exactly, ignoring whitespace), use the `exact` directive in the body section declaration:

```
200
--- body json exact
{
  "message": "Hello, world!"
}
```
