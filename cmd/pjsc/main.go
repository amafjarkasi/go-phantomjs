package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func main() {
	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: PHANTOMJSCLOUD_API_KEY environment variable is required.")
		os.Exit(1)
	}

	renderCmd := flag.NewFlagSet("render", flag.ExitOnError)
	url := renderCmd.String("url", "", "Target URL to render")
	output := renderCmd.String("output", "html", "Output format (html, plainText, png, jpeg, pdf)")
	file := renderCmd.String("file", "", "File path to save the output (for images/pdf)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := phantomjscloud.NewClient(apiKey)

	switch os.Args[1] {
	case "render":
		renderCmd.Parse(os.Args[2:])
		if *url == "" {
			fmt.Println("Error: -url is required")
			renderCmd.Usage()
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		var result []byte
		var err error
		var textResult string

		switch *output {
		case "html":
			textResult, err = client.FetchPlainText(*url) // FetchPlainText is misleadingly named in current client, it fetches content. Actually client has RenderRawHTML etc.
			// Re-evaluating: client.DoPage is most versatile.
			req := phantomjscloud.NewPageRequestBuilder(*url).WithRenderType("html").Build()
			resp, err2 := client.DoPage(req)
			if err2 == nil {
				textResult = resp.PageResponses[0].Content
			}
			err = err2
		case "plainText":
			textResult, err = client.FetchPlainText(*url)
		case "png", "jpeg":
			result, err = client.FetchScreenshot(*url, *output, nil)
		case "pdf":
			result, err = client.FetchPDF(*url, nil)
		default:
			fmt.Printf("Unknown output format: %s\n", *output)
			os.Exit(1)
		}

		if err != nil {
			log.Fatalf("Render failed: %v", err)
		}

		if *file != "" && result != nil {
			if err := os.WriteFile(*file, result, 0644); err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}
			fmt.Printf("Saved output to %s\n", *file)
		} else if textResult != "" {
			fmt.Println(textResult)
		} else if result != nil {
			fmt.Printf("Received %d bytes of binary data. Use -file to save it.\n", len(result))
		}

	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("pjsc - PhantomJS Cloud CLI Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  pjsc <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  render    Fetch and render a URL")
	fmt.Println("  help      Show this help message")
	fmt.Println("\nExample:")
	fmt.Println("  pjsc render -url https://example.com -output png -file screenshot.png")
}
