package scraper

import (
	"context"
	"sync"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

// BatchProcessor manages high-volume scraping tasks.
// It automatically batches requests to PhantomJsCloud's multi-page API
// and handles concurrency, ensuring you stay within rate limits.
type BatchProcessor struct {
	client      *phantomjscloud.Client
	concurrency int
	batchSize   int
}

// Result is the outcome of a single URL scrape.
type Result struct {
	Request  phantomjscloud.PageRequest
	Response *phantomjscloud.PageResponse
	Metadata phantomjscloud.ResponseMetadata
	Error    error
}

// NewBatchProcessor creates a processor with the given client.
// Concurrency controls how many simultaneous HTTP requests are made.
// BatchSize controls how many pages are sent in each single HTTP request.
func NewBatchProcessor(client *phantomjscloud.Client, concurrency, batchSize int) *BatchProcessor {
	if batchSize < 1 {
		batchSize = 1
	}
	if batchSize > 100 {
		batchSize = 100 // PJSC limit
	}
	if concurrency < 1 {
		concurrency = 1
	}
	return &BatchProcessor{
		client:      client,
		concurrency: concurrency,
		batchSize:   batchSize,
	}
}

// Scrape concurrently processes all requests and streams results back through the returned channel.
func (p *BatchProcessor) Scrape(ctx context.Context, requests []phantomjscloud.PageRequest) <-chan Result {
	resultChan := make(chan Result, len(requests))

	go func() {
		defer close(resultChan)

		// Create batches
		var batches [][]phantomjscloud.PageRequest
		for i := 0; i < len(requests); i += p.batchSize {
			end := i + p.batchSize
			if end > len(requests) {
				end = len(requests)
			}
			batches = append(batches, requests[i:end])
		}

		// Process batches with limited concurrency
		var wg sync.WaitGroup
		sem := make(chan struct{}, p.concurrency)

		for _, batch := range batches {
			wg.Add(1)
			sem <- struct{}{} // Acquire

			go func(b []phantomjscloud.PageRequest) {
				defer wg.Done()
				defer func() { <-sem }() // Release

				userReq := &phantomjscloud.UserRequest{
					Pages: b,
				}

				res, err := p.client.DoContext(ctx, userReq)
				if err != nil {
					for _, req := range b {
						resultChan <- Result{Request: req, Error: err}
					}
					return
				}

				// Map results back to requests
				for i := range res.PageResponses {
					pageRes := res.PageResponses[i]
					resultChan <- Result{
						Request:  b[i],
						Response: &pageRes,
						Metadata: res.Metadata,
					}
				}
			}(batch)
		}

		wg.Wait()
	}()

	return resultChan
}

// ScrapeAll is like Scrape but blocks until all results are gathered.
func (p *BatchProcessor) ScrapeAll(ctx context.Context, requests []phantomjscloud.PageRequest) ([]Result, error) {
	results := make([]Result, 0, len(requests))
	for res := range p.Scrape(ctx, requests) {
		results = append(results, res)
	}
	return results, nil
}

// ScrapeSimple is a convenience method for scraping a list of URLs with default settings.
func (p *BatchProcessor) ScrapeSimple(ctx context.Context, urls []string) ([]Result, error) {
	reqs := make([]phantomjscloud.PageRequest, len(urls))
	for i, url := range urls {
		reqs[i] = phantomjscloud.PageRequest{URL: url}
	}
	return p.ScrapeAll(ctx, reqs)
}
