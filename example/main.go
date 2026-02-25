package main

import (
	"fmt"
	"log"
	"os"

	phantomjscloud "github.com/jbdt/go-phantomjs"
)

func main() {
	// Initialize a PhantomJsCloud Client
	// We read the API key from the environment variable PHANTOMJSCLOUD_API_KEY.
	// Passing an empty string uses the free demo key (which has severe rate limits).
	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	client := phantomjscloud.NewClient(apiKey)

	// Create a simple script using the OverseerScriptBuilder
	script := phantomjscloud.NewOverseerScriptBuilder().
		WaitForNavigation().
		// Adding an external script, e.g., a hilitor
		AddScriptTag("http://phantomjscloud.com/examples/scripts/hilitor.js").
		// And evaluating a function in the page context
		Evaluate("() => { let _hilitor = new Hilitor(); _hilitor.apply('Example'); }").
		Build()

	fmt.Println("Generated Overseer Script:\n", script)

	// Build the PageRequest
	pageReq := &phantomjscloud.PageRequest{
		URL:            "https://example.com",
		RenderType:     "html", // Or "png", "pdf", "automation"
		OutputAsJson:   true,
		OverseerScript: script,
		RequestSettings: phantomjscloud.RequestSettings{
			WaitInterval: 0,
			DoneWhen: []phantomjscloud.DoneWhen{
				{Event: "domReady"},
			},
			ResourceModifier: []phantomjscloud.ResourceModifier{
				{
					Type:          "image",
					IsBlacklisted: true,
				},
			},
		},
	}

	fmt.Println("Sending request to PhantomJsCloud...")
	resp, err := client.DoPage(pageReq)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	fmt.Printf("Received Response!\n")
	fmt.Printf("Status Code: %d\n", resp.Metadata.ContentStatusCode)
	fmt.Printf("Billing Credit Cost: %f\n", resp.Metadata.BillingCreditCost)

	if len(resp.PageResponses) > 0 {
		fmt.Printf("Content length: %d\n", len(resp.PageResponses[0].Content))
	} else {
		fmt.Println("No page responses received.")
	}
}
