package scraper

import (
	"context"
	"errors"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
)

// AdaptiveAttempt stores one attempt in an adaptive block-policy run.
type AdaptiveAttempt struct {
	Level    blockpolicy.Level
	Response *phantomjscloud.UserResponseWithMeta
	Err      error
}

// DoPageWithAdaptiveBlockPolicy executes a request with progressive block-policy fallback.
// It starts at startLevel and moves toward less restrictive levels after block/challenge results.
func DoPageWithAdaptiveBlockPolicy(
	ctx context.Context,
	client *phantomjscloud.Client,
	baseReq *phantomjscloud.PageRequest,
	startLevel blockpolicy.Level,
	maxAttempts int,
) (*phantomjscloud.UserResponseWithMeta, []AdaptiveAttempt, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	level := startLevel
	attempts := make([]AdaptiveAttempt, 0, maxAttempts)

	for i := 0; i < maxAttempts; i++ {
		req := *baseReq
		blockpolicy.Apply(&req, level)

		resp, err := client.DoPageContext(ctx, &req)
		attempts = append(attempts, AdaptiveAttempt{
			Level:    level,
			Response: resp,
			Err:      err,
		})

		if err == nil && !blockpolicy.LooksBlocked(resp) {
			return resp, attempts, nil
		}

		if i == maxAttempts-1 {
			if err != nil {
				return nil, attempts, err
			}
			return resp, attempts, errors.New("request remained blocked after adaptive block-policy retries")
		}

		next := blockpolicy.NextLevel(level)
		if next == level {
			if err != nil {
				return nil, attempts, err
			}
			return resp, attempts, errors.New("request failed and no further block-policy fallback is available")
		}
		level = next
	}

	return nil, attempts, errors.New("adaptive block-policy ended unexpectedly")
}

// DoPageWithRoutingAndAdaptivePolicy combines host-aware proxy fallback routing
// with progressive block-policy retries in one call.
func DoPageWithRoutingAndAdaptivePolicy(
	ctx context.Context,
	client *phantomjscloud.Client,
	baseReq *phantomjscloud.PageRequest,
	router proxy.URLProxyFallbackProvider,
	startLevel blockpolicy.Level,
	maxAttempts int,
) (*phantomjscloud.UserResponseWithMeta, []AdaptiveAttempt, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	level := startLevel
	attempts := make([]AdaptiveAttempt, 0, maxAttempts)

	for i := 0; i < maxAttempts; i++ {
		req := *baseReq
		blockpolicy.Apply(&req, level)
		req.Proxy = router.GetProxyForURLAttempt(req.URL, i)

		resp, err := client.DoPageContext(ctx, &req)
		attempts = append(attempts, AdaptiveAttempt{
			Level:    level,
			Response: resp,
			Err:      err,
		})

		if err == nil && !blockpolicy.LooksBlocked(resp) {
			return resp, attempts, nil
		}

		if i == maxAttempts-1 {
			if err != nil {
				return nil, attempts, err
			}
			return resp, attempts, errors.New("request remained blocked after routing+adaptive retries")
		}

		next := blockpolicy.NextLevel(level)
		if next == level {
			if err != nil {
				return nil, attempts, err
			}
			return resp, attempts, errors.New("request failed and no further policy fallback is available")
		}
		level = next
	}

	return nil, attempts, errors.New("routing+adaptive policy ended unexpectedly")
}
