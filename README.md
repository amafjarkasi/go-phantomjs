# go-phantomjs

`go-phantomjs` is a robust and easy-to-use Golang client library for the [PhantomJsCloud](https://phantomjscloud.com/) API. It provides a structured way to interact with both the standard JSON API and the advanced Automation API (`overseerScript`), allowing you to fetch pages, take screenshots, generate PDFs, and run complex browser automation scripts right from Go.

## Features

- **Rapid Convenience Fetchers**: Directly extract final payloads for PDFs, Images, and PlainText via `FetchPDF()`, `FetchScreenshot()`, and `FetchPlainText()` without unpacking complicated JSON wrappers.
- **Full API Coverage**: Strongly-typed Go structs for all major PhantomJsCloud parameters including `PageRequest`, `RequestSettings`, `RenderSettings`, and more.
- **Automation API Builder**: A fluent `OverseerScriptBuilder` to dynamically generate complex automation scripts (clicking, typing, waiting for selectors, evaluating JavaScript) without manually concatenating strings. Use `FetchWithAutomation()` to extract native structured script results directly.
- **Smart Waits & Performance**: easily configure `doneWhen` events, `waitInterval: 0`, and `ResourceModifier` rules to speed up your scraping operations.
- **Response Metadata**: Automatically parses `pjsc-*` HTTP headers into a structured `ResponseMetadata` object to track billing costs, status codes, and done events.
- **Seamless Proxy Support**: Use predefined proxy constants (e.g., `phantomjscloud.ProxyAnonUS`, `phantomjscloud.ProxyGeoUK`) or supply your own strings.

## Installation

```bash
go get github.com/jbdt/go-phantomjs
```

## Quick Start

### Convenience Fetchers (PDFs, Images, Scripts)

If you don't care about the full API metadata, you can use the built in convenience fetchers:

```go
package main

import (
 "os"
 "log"
 "github.com/jbdt/go-phantomjs"
)

func main() {
    client := phantomjscloud.NewClient("") // demo key
    
    // Fetch purely the stripped text of a webpage (Great for LLMs!)
    text, err := client.FetchPlainText("https://example.com")
    if err == nil {
        log.Println(text)
    }

    // Fetch a base64-decoded PDF instantly
    pdfBytes, err := client.FetchPDF("https://example.com", nil)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    os.WriteFile("output.pdf", pdfBytes, 0644)
    
    // Evaluate a script block and cleanly parse exactly what it returns
    builder := phantomjscloud.NewOverseerScriptBuilder().
        WaitForNavigation().
        Evaluate("() => { return { title: document.title }; }")
        
    result, err := client.FetchWithAutomation("https://example.com", builder)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Prints map[title:Example Domain]
    log.Println(result) 
}
```

### Advanced HTML Extraction

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

### Comprehensive Automation Scripting

The `OverseerScriptBuilder` supports nearly all primary PhantomJsCloud automation flags. Check out this comprehensive usage example that demonstrates navigating, injecting custom scripts, taking screenshots during execution, and manually resolving the queue:

```go
 script := phantomjscloud.NewOverseerScriptBuilder().
  Goto("https://example.com").
  AddScriptTag("https://example.com/utility.js").
  Evaluate("() => { console.log('Injected custom DOM manipulation!'); }").
  WaitForSelector("body").
  ScrollBy(0, 500).            // Scroll down exactly 500px
  Hover("button#menu").        // Trigger CSS hover states natively
  Focus("input#search").       // Focus input fields
  ClearInput("input#search").  // Empty inputs reliably
  Type("input#search", "hello world", 100). 
  Select("select#country", "US", "UK"). // Support for array Multi-select!
  KeyboardPress("Enter", 1).   // Send keystrokes natively
  SetCookie("session", "abc", "example.com"). // Drop cookies directly
  AddStyleTag("body { background: red; }"). // Inject styles
  SetViewport(1920, 1080).     // Dynamically override bounds
  WaitForFunction("window.ready === true"). // Native JS waiting
  WaitForDelay(2000).          // Pause for UI transitions
  ManualWait().                // Take control over the completion event
  RenderScreenshot(true).      // Take a sync screenshot mid-execution!
  Reload().                    // Native browser refresh
  DeleteCookie("session", "example.com"). // Remove cookies natively
  Done().                      // Tell the engine we're finished manually
  Build()
```

### Advanced Automation Workflows

#### Auto-Login & Navigation

You can seamlessly automate filling out authentication forms, clicking submit, and waiting for the backend to redirect your headless browser securely.

```go
 script := phantomjscloud.NewOverseerScriptBuilder().
  Type("input#username", "USER@EXAMPLE.COM", 50).
  Type("input#password", "PASSWORD", 50).
  Click("button[type=submit]").
  WaitForNavigation(). // Wait for the form submission to redirect the page
  Build()

 req := &phantomjscloud.PageRequest{
  URL:            "https://www.linkedin.com/uas/login",
  RenderType:     "jpeg",
  OverseerScript: script,
 }
```

#### Speeding up Long Requests (DOM Content Loaded)

If a page has heavy ad trackers or infinite lazy loading, `PhantomJsCloud` might timeout waiting for the network idle state. You can override this to finish rendering as soon as the DOM is available or use `DoneWhen` in `RequestSettings`.

```go
 // Method 1: Inject a manual wait and exit specifically on the domcontentloaded event
 script := phantomjscloud.NewOverseerScriptBuilder().
  WaitForNavigationEvent("domcontentloaded").
  Done().
  Build()

 // Method 2: Configure it declaratively in RequestSettings natively
 reqSettings := phantomjscloud.RequestSettings{
  DoneWhen: []phantomjscloud.DoneWhen{
   {Event: "domReady"},
  },
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

#### Intercept and Modify Requests (Change URL & Blacklist)

You can use the `ResourceModifier` to change domains on the fly, or blacklist certain requests completely to save bandwidth (like CSS files).

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://www.highcharts.com",
  RenderType: "jpg",
  RequestSettings: phantomjscloud.RequestSettings{
   ClearCache: true, // Forces re-requesting css to be caught by the blacklist
   ResourceModifier: []phantomjscloud.ResourceModifier{
    {
     Regex:     ".*highcharts.com.*",
     ChangeUrl: "$$protocol:$$port//en.wikipedia.org/wiki$$path",
    },
    {
     Regex:         ".*css.*",
     IsBlacklisted: true,
    },
   },
  },
 }
```

#### Render Thumbnails and Zooming

Combine `Viewport`, `ClipRectangle`, and `ZoomFactor` to capture perfect thumbnails.

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://cnn.com",
  RenderType: "jpeg",
  RenderSettings: phantomjscloud.RenderSettings{
   ZoomFactor: 0.45,
   Viewport:      &phantomjscloud.Viewport{Width: 640, Height: 500},
   ClipRectangle: &phantomjscloud.ClipRectangle{Width: 640, Height: 500},
  },
 }
```

#### Uploading POST Data and JSONP

To submit POST data to a target URL natively, use `UrlSettings`:

```go
 req := &phantomjscloud.PageRequest{
  URL: "https://example.com/api",
  UrlSettings: &phantomjscloud.UrlSettings{
   Operation: "POST",
   Data:      `{"my_key":"my_value"}`,
  },
 }
```

#### HTTP Basic Auth & Reduced JSON Verbosity

Using `OutputAsJson: true` will return a massive payload. You can suppress fields using `SuppressJson`. Also, bypass HTTP Basic Auth natively using `Authentication`.

```go
 req := &phantomjscloud.PageRequest{
  URL:          "http://httpbin.org/basic-auth/user/pass",
  OutputAsJson: true,
  SuppressJson: []string{"pageResponses", "originalRequest"},
  RequestSettings: phantomjscloud.RequestSettings{
   Authentication: &phantomjscloud.Authentication{
    UserName: "user",
    Password: "pass",
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
