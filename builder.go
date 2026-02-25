package phantomjscloud

import "github.com/amafjarkasi/go-phantomjs/ext/useragents"

// PageRequestBuilder constructs a PageRequest using a fluent API.
// It is the recommended way to compose requests from ext/ presets without
// manually nesting RequestSettings and RenderSettings structs.
//
// Use the ext/ packages to generate values:
//
//	import "github.com/amafjarkasi/go-phantomjs/ext/blocklist"
//	import "github.com/amafjarkasi/go-phantomjs/ext/useragents"
//	import "github.com/amafjarkasi/go-phantomjs/ext/viewport"
//
//	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
//	    WithRenderType("jpeg").
//	    WithProxy(phantomjscloud.ProxyAnonUS).
//	    WithRenderSettings(viewport.FHD.AsRenderSettings()).
//	    WithProfile(useragents.ChromeWindowsProfile()).
//	    WithBlocklist(blocklist.Lightweight()).
//	    WithOverseerScript(script).
//	    Build()
type PageRequestBuilder struct {
	req PageRequest
}

// NewPageRequestBuilder returns a builder for the given URL.
// RenderType defaults to "html" if never set.
func NewPageRequestBuilder(url string) *PageRequestBuilder {
	return &PageRequestBuilder{
		req: PageRequest{URL: url},
	}
}

// WithURL overrides the target URL.
func (b *PageRequestBuilder) WithURL(url string) *PageRequestBuilder {
	b.req.URL = url
	return b
}

// WithRenderType sets the output format: "html", "plainText", "jpeg", "png", "pdf", "json", etc.
func (b *PageRequestBuilder) WithRenderType(rt string) *PageRequestBuilder {
	b.req.RenderType = rt
	return b
}

// WithOutputAsJson wraps the response in a full JSON envelope (UserResponse).
func (b *PageRequestBuilder) WithOutputAsJson(v bool) *PageRequestBuilder {
	b.req.OutputAsJson = v
	return b
}

// WithProxy sets a proxy string (e.g. ProxyAnonUS) or a ProxyOptions struct.
// Accepts the same types as PageRequest.Proxy.
func (b *PageRequestBuilder) WithProxy(proxy interface{}) *PageRequestBuilder {
	b.req.Proxy = proxy
	return b
}

// WithUserAgent sets the User-Agent string. Prefer WithProfile for a complete fingerprint.
func (b *PageRequestBuilder) WithUserAgent(ua string) *PageRequestBuilder {
	b.req.RequestSettings.UserAgent = ua
	return b
}

// WithCustomHeaders sets (replaces) all custom request headers.
// To add individual headers without clearing existing ones use WithHeader.
func (b *PageRequestBuilder) WithCustomHeaders(headers map[string]string) *PageRequestBuilder {
	b.req.RequestSettings.CustomHeaders = headers
	return b
}

// WithHeader adds or overrides a single custom request header.
func (b *PageRequestBuilder) WithHeader(key, value string) *PageRequestBuilder {
	if b.req.RequestSettings.CustomHeaders == nil {
		b.req.RequestSettings.CustomHeaders = make(map[string]string)
	}
	b.req.RequestSettings.CustomHeaders[key] = value
	return b
}

// WithProfile sets the UserAgent and the full matching header set (Accept, Sec-CH-UA,
// Accept-Language, Sec-Fetch-* etc.) from an ext/useragents Profile in one call.
//
// Example:
//
//	.WithProfile(useragents.ChromeWindowsProfile())
func (b *PageRequestBuilder) WithProfile(p useragents.Profile) *PageRequestBuilder {
	b.req.RequestSettings.UserAgent = p.UserAgent
	if b.req.RequestSettings.CustomHeaders == nil {
		b.req.RequestSettings.CustomHeaders = make(map[string]string)
	}
	for k, v := range p.Headers {
		b.req.RequestSettings.CustomHeaders[k] = v
	}
	return b
}

// WithBlocklist appends ResourceModifier rules to the request, used to block or
// redirect network resources. Pass values from ext/blocklist:
//
//	.WithBlocklist(blocklist.Lightweight())
//	.WithBlocklist(append(blocklist.Ads(), blocklist.Fonts()...))
func (b *PageRequestBuilder) WithBlocklist(rules []ResourceModifier) *PageRequestBuilder {
	b.req.RequestSettings.ResourceModifier = append(
		b.req.RequestSettings.ResourceModifier, rules...,
	)
	return b
}

// WithResourceModifier is an alias for WithBlocklist for cases where you're
// setting arbitrary ResourceModifier rules (e.g. ChangeUrl), not just blacklists.
func (b *PageRequestBuilder) WithResourceModifier(rules []ResourceModifier) *PageRequestBuilder {
	return b.WithBlocklist(rules)
}

// WithDoneWhen sets the completion trigger conditions for the request.
func (b *PageRequestBuilder) WithDoneWhen(events []DoneWhen) *PageRequestBuilder {
	b.req.RequestSettings.DoneWhen = events
	return b
}

// WithWaitInterval sets the milliseconds to wait after the page fires its done event.
func (b *PageRequestBuilder) WithWaitInterval(ms int) *PageRequestBuilder {
	b.req.RequestSettings.WaitInterval = ms
	return b
}

// WithIgnoreImages skips loading all image resources, cutting request time and cost.
func (b *PageRequestBuilder) WithIgnoreImages(v bool) *PageRequestBuilder {
	b.req.RequestSettings.IgnoreImages = v
	return b
}

// WithClearCache forces a cache-busting re-fetch of all resources.
func (b *PageRequestBuilder) WithClearCache(v bool) *PageRequestBuilder {
	b.req.RequestSettings.ClearCache = v
	return b
}

// WithRequestSettings replaces the entire RequestSettings block. Use this when
// you have a pre-built settings struct. Individual With* methods will still
// work after this call and will overwrite individual fields.
func (b *PageRequestBuilder) WithRequestSettings(rs RequestSettings) *PageRequestBuilder {
	b.req.RequestSettings = rs
	return b
}

// WithViewport sets the viewport dimensions in RenderSettings.
// For full mobile emulation flags (isMobile, hasTouch) use WithRenderSettings
// or WithApplyViewport together with ApplyViewport in the OverseerScriptBuilder.
//
// To use a named preset:
//
//	.WithViewport(viewport.FHD.Viewport)
func (b *PageRequestBuilder) WithViewport(v Viewport) *PageRequestBuilder {
	b.req.RenderSettings.Viewport = &v
	return b
}

// WithClipRectangle sets the clip rectangle in RenderSettings.
func (b *PageRequestBuilder) WithClipRectangle(cr ClipRectangle) *PageRequestBuilder {
	b.req.RenderSettings.ClipRectangle = &cr
	return b
}

// WithZoomFactor sets the zoom factor in RenderSettings.
func (b *PageRequestBuilder) WithZoomFactor(z float64) *PageRequestBuilder {
	b.req.RenderSettings.ZoomFactor = z
	return b
}

// WithEmulateMedia sets the CSS media type to emulate (e.g. "print", "screen").
func (b *PageRequestBuilder) WithEmulateMedia(media string) *PageRequestBuilder {
	b.req.RenderSettings.EmulateMedia = media
	return b
}

// WithRenderSettings replaces the entire RenderSettings block. Use this when
// you have a pre-built settings struct, e.g. from ext/viewport:
//
//	.WithRenderSettings(viewport.FHD.AsRenderSettings())
//	.WithRenderSettings(viewport.Thumbnail1200.AsRenderSettings())
func (b *PageRequestBuilder) WithRenderSettings(rs RenderSettings) *PageRequestBuilder {
	b.req.RenderSettings = rs
	return b
}

// WithOverseerScript sets a raw automation script string.
// Use WithOverseerScriptBuilder for the fluent builder alternative.
func (b *PageRequestBuilder) WithOverseerScript(script string) *PageRequestBuilder {
	b.req.OverseerScript = script
	return b
}

// WithOverseerScriptBuilder calls Build() on the provided OverseerScriptBuilder
// and sets the result as the overseer script.
func (b *PageRequestBuilder) WithOverseerScriptBuilder(sb *OverseerScriptBuilder) *PageRequestBuilder {
	b.req.OverseerScript = sb.Build()
	return b
}

// WithContent sets inline HTML content to render instead of fetching a URL.
func (b *PageRequestBuilder) WithContent(html string) *PageRequestBuilder {
	b.req.Content = html
	return b
}

// Build returns the fully configured PageRequest. Safe to call multiple times;
// each call returns a value copy of the current state.
func (b *PageRequestBuilder) Build() *PageRequest {
	req := b.req
	return &req
}
