# go-phantomjs

`go-phantomjs` is a robust and easy-to-use Golang client library for the [PhantomJsCloud](https://phantomjscloud.com/) API. It provides a structured way to interact with both the standard JSON API and the advanced Automation API (`overseerScript`), allowing you to fetch pages, take screenshots, generate PDFs, and run complex browser automation scripts right from Go.

## Features

- **Full API Coverage**: Strongly-typed Go structs for all major PhantomJsCloud parameters including `PageRequest`, `RequestSettings`, `RenderSettings`, and more.
- **Automation API Builder**: A fluent `OverseerScriptBuilder` to dynamically generate complex automation scripts (clicking, typing, waiting for selectors, evaluating JavaScript) without manually concatenating strings.
- **Smart Waits & Performance**: easily configure `doneWhen` events, `waitInterval: 0`, and `ResourceModifier` rules to speed up your scraping operations.
- **Response Metadata**: Automatically parses `pjsc-*` HTTP headers into a structured `ResponseMetadata` object to track billing costs, status codes, and done events.
- **Seamless Proxy Support**: Use predefined proxy constants (e.g., `phantomjscloud.ProxyAnonUS`, `phantomjscloud.ProxyGeoUK`) or supply your own strings.

## Installation

```bash
go get github.com/jbdt/go-phantomjs
```

## Quick Start

### Basic HTML Extraction

```go
package main

import (
 "fmt"
 "log"
 "github.com/jbdt/go-phantomjs"
)

func main() {
 // Passing an empty string uses the free demo key (low quota).
 // Replace with your actual PhantomJsCloud API key.
 client := phantomjscloud.NewClient("")

 req := &phantomjscloud.PageRequest{
  URL:        "https://example.com",
  RenderType: "html",
 }

 resp, err := client.DoPage(req)
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 fmt.Printf("Cost: %f credits\n", resp.Metadata.BillingCreditCost)
 fmt.Printf("Content: %s\n", resp.PageResponses[0].Content)
}
```

### Advanced Automation and Browser Scripting

Use the `OverseerScriptBuilder` to construct scripts that navigate, interact, and manipulate the DOM before returning the final rendered output.

```go
package main

import (
 "fmt"
 "log"
 "github.com/jbdt/go-phantomjs"
)

func main() {
 client := phantomjscloud.NewClient("YOUR-API-KEY")

 // Build an automation script dynamically
 script := phantomjscloud.NewOverseerScriptBuilder().
  WaitForSelector("body").
  Type("input#search", "golang", 100).
  Click("button#submit").
  WaitForNavigation().
  Build()

 req := &phantomjscloud.PageRequest{
  URL:            "https://example.com",
  RenderType:     "png", // Take a screenshot after the script runs
  OutputAsJson:   true,
  OverseerScript: script,
  RequestSettings: phantomjscloud.RequestSettings{
   // Connect through a US residential proxy
   Proxy: phantomjscloud.ProxyGeoUS,
            // Skip image loading for faster performance if we just want data
   ResourceModifier: []phantomjscloud.ResourceModifier{
    {Type: "image", IsBlacklisted: true},
   },
  },
  RenderSettings: phantomjscloud.RenderSettings{
   Viewport: phantomjscloud.Viewport{Width: 1280, Height: 720},
  },
 }

 resp, err := client.DoPage(req)
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 // Assuming success, resp.PageResponses[0].Content contains the base64 encoded PNG.
 fmt.Printf("Captured screenshot length: %d bytes\n", len(resp.PageResponses[0].Content))
}
```

### Advanced Features

#### Emulate Print Media for PDF Generation

Use the `EmulateMedia` parameter to generate a PDF exactly as it would look when printed.

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://example.com/invoice.html",
  RenderType: "pdf",
  RenderSettings: phantomjscloud.RenderSettings{
   EmulateMedia: "print", // Generate PDF using the @media:print CSS rules
  },
 }
```

#### Intercept and Modify Requests (Change URL)

You can use the `ResourceModifier` to change domains on the fly. Here we route a domain to another domain while keeping the path the same.

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://www.highcharts.com/demo/pie-donut",
  RenderType: "jpg",
  RequestSettings: phantomjscloud.RequestSettings{
   ResourceModifier: []phantomjscloud.ResourceModifier{
    {
     Regex:     ".*highcharts.com.*",
     ChangeUrl: "$$protocol:$$port//en.wikipedia.org/wiki$$path",
    },
   },
  },
 }
```

#### Full JSON Metadata Response and Cookies

To extract cookies, headers, and extensive metadata, use `OutputAsJson: true` and specify `Cookies` in the request settings.

```go
 req := &phantomjscloud.PageRequest{
  URL:          "http://example.com",
  RenderType:   "plainText",
  OutputAsJson: true,
  RequestSettings: phantomjscloud.RequestSettings{
   Cookies: []phantomjscloud.Cookie{
    {Domain: "example.com", Name: "myCookie1", Value: "value1"},
   },
  },
 }
```

## Structure

- `types.go`: Contains all JSON mappings aligning exactly with PhantomJsCloud interfaces (e.g. `RenderSettings`, `RequestSettings`, `ClipRectangle`).
- `client.go`: `Client` struct, request execution (`Do`, `DoPage`), and parsing of PJSC metadata headers.
- `automation.go`: `OverseerScriptBuilder` and Proxy constants.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Make sure to run the tests before submitting: `go test ./...`

## License

[MIT](https://choosealicense.com/licenses/mit/)
