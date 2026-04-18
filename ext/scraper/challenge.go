package scraper

import (
	"context"
	"errors"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/persona"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
	"github.com/amafjarkasi/go-phantomjs/ext/session"
)

// ChallengeOrchestrationOptions controls challenge retries and request mutation behavior.
type ChallengeOrchestrationOptions struct {
	Persona     persona.URLPersonaProvider
	Router      proxy.URLProxyFallbackProvider
	Session     *session.Store
	StartLevel  blockpolicy.Level
	MaxAttempts int
}

// ChallengeAttempt records one challenge orchestration attempt.
type ChallengeAttempt struct {
	Level    blockpolicy.Level
	Persona  string
	Response *phantomjscloud.UserResponseWithMeta
	Err      error
}

// DoPageWithChallengeOrchestration retries blocked requests while combining:
// adaptive block-policy fallback, optional host-aware personas, optional proxy routing,
// and optional cookie session persistence between attempts.
func DoPageWithChallengeOrchestration(
	ctx context.Context,
	client *phantomjscloud.Client,
	baseReq *phantomjscloud.PageRequest,
	opts ChallengeOrchestrationOptions,
) (*phantomjscloud.UserResponseWithMeta, []ChallengeAttempt, error) {
	maxAttempts := opts.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	level := opts.StartLevel
	attempts := make([]ChallengeAttempt, 0, maxAttempts)

	for i := 0; i < maxAttempts; i++ {
		req := *baseReq
		blockpolicy.Apply(&req, level)

		personaName := ""
		if opts.Persona != nil {
			personaName = opts.Persona.ApplyForURLAttempt(&req, i)
		}
		if opts.Router != nil {
			req.Proxy = opts.Router.GetProxyForURLAttempt(req.URL, i)
		}
		if opts.Session != nil {
			cookies := opts.Session.CookiesForURL(req.URL)
			if len(cookies) > 0 {
				req.RequestSettings.Cookies = cookies
			}
		}

		resp, err := client.DoPageContext(ctx, &req)
		if opts.Session != nil {
			opts.Session.CaptureFromResponse(resp)
		}
		attempts = append(attempts, ChallengeAttempt{
			Level:    level,
			Persona:  personaName,
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
			return resp, attempts, errors.New("request remained blocked after challenge orchestration retries")
		}

		level = blockpolicy.NextLevel(level)
	}

	return nil, attempts, errors.New("challenge orchestration ended unexpectedly")
}

