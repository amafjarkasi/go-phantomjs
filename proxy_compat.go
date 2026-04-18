package phantomjscloud

import "strings"

func normalizeUserRequestForAPI(req *UserRequest) *UserRequest {
	if req == nil {
		return nil
	}

	out := *req
	if len(req.Pages) == 0 {
		return &out
	}

	out.Pages = make([]PageRequest, len(req.Pages))
	for i := range req.Pages {
		p := req.Pages[i]
		p.Proxy = normalizePageProxyForAPI(p.Proxy)
		out.Pages[i] = p
	}
	return &out
}

func normalizePageProxyForAPI(v interface{}) interface{} {
	switch p := v.(type) {
	case ProxyBuiltin:
		return formatBuiltinProxyString(p)
	case *ProxyBuiltin:
		if p == nil {
			return nil
		}
		return formatBuiltinProxyString(*p)
	case ProxyOptions:
		if g := formatGeolocationProxyString(p.Geolocation); g != "" {
			return g
		}
		if p.Builtin != nil {
			return formatBuiltinProxyString(*p.Builtin)
		}
		if p.Custom != nil {
			return formatCustomProxyString(*p.Custom)
		}
		return v
	case *ProxyOptions:
		if p == nil {
			return nil
		}
		if g := formatGeolocationProxyString(p.Geolocation); g != "" {
			return g
		}
		if p.Builtin != nil {
			return formatBuiltinProxyString(*p.Builtin)
		}
		if p.Custom != nil {
			return formatCustomProxyString(*p.Custom)
		}
		return v
	default:
		return v
	}
}

func formatBuiltinProxyString(b ProxyBuiltin) string {
	location := strings.TrimSpace(strings.ToLower(b.Location))
	if location == "" {
		return ""
	}

	if strings.HasPrefix(location, "anon-") || strings.HasPrefix(location, "geo-") {
		return location
	}

	kind := strings.TrimSpace(strings.ToLower(b.Type))
	if kind == "" || kind == "anon" || kind == "anonymous" {
		return "anon-" + location
	}
	if kind == "geo" || kind == "geolocation" || kind == "static" {
		return "geo-" + location
	}

	return kind + "-" + location
}

func formatGeolocationProxyString(geo string) string {
	geo = strings.TrimSpace(strings.ToLower(geo))
	if geo == "" {
		return ""
	}
	if strings.HasPrefix(geo, "geo-") {
		return geo
	}
	return "geo-" + geo
}

func formatCustomProxyString(c ProxyCustom) string {
	host := strings.TrimSpace(c.Host)
	if host == "" {
		return ""
	}
	auth := strings.TrimSpace(c.Auth)
	if auth == "" {
		return "custom-" + host
	}
	return "custom-" + host + ":" + auth
}
