package useragents

// Desktop user-agent strings — kept to major current browser versions.
// Update periodically as new browser versions release.

// Chrome on Windows
const (
	ChromeWin   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	ChromeWin11 = "Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

// Chrome on macOS
const (
	ChromeMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

// Chrome on Linux
const (
	ChromeLinux = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

// Firefox on Windows and macOS
const (
	FirefoxWin = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0"
	FirefoxMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.3; rv:123.0) Gecko/20100101 Firefox/123.0"
)

// Safari on macOS and iOS
const (
	SafariMac    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_3_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Safari/605.1.15"
	SafariIPad   = "Mozilla/5.0 (iPad; CPU OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Mobile/15E148 Safari/604.1"
	SafariIPhone = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Mobile/15E148 Safari/604.1"
)

// Edge on Windows
const (
	EdgeWin = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.0.0"
)

// Mobile — Chrome on Android
const (
	ChromeAndroid       = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.90 Mobile Safari/537.36"
	ChromeAndroidTablet = "Mozilla/5.0 (Linux; Android 14; Pixel Tablet) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.90 Safari/537.36"
)

// Bots — useful for sites that grant more access to known crawlers.
const (
	Googlebot       = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	GooglebotMobile = "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.90 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	Bingbot         = "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"
)

// Profile bundles a UA string with matching Accept-Language and platform hint headers.
// Use with PageRequest.RequestSettings.CustomHeaders for a fully consistent fingerprint.
type Profile struct {
	// UserAgent is the full User-Agent string.
	UserAgent string
	// Headers are commonly expected headers that accompany this UA.
	// Set these via RequestSettings.CustomHeaders for maximum believability.
	Headers map[string]string
}

// ChromeWindowsProfile returns a complete Chrome/Windows browser profile
// with a consistent user agent and realistic accompanying request headers.
func ChromeWindowsProfile() Profile {
	return Profile{
		UserAgent: ChromeWin,
		Headers: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-CH-UA":                 `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
		},
	}
}

// ChromeMacProfile returns a complete Chrome/Mac browser profile.
func ChromeMacProfile() Profile {
	return Profile{
		UserAgent: ChromeMac,
		Headers: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-CH-UA":                 `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"macOS"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
		},
	}
}

// FirefoxWindowsProfile returns a complete Firefox/Windows browser profile.
func FirefoxWindowsProfile() Profile {
	return Profile{
		UserAgent: FirefoxWin,
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.5",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"DNT":             "1",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
			"Sec-Fetch-User":  "?1",
		},
	}
}

// FirefoxMacProfile returns a complete Firefox/Mac browser profile.
func FirefoxMacProfile() Profile {
	return Profile{
		UserAgent: FirefoxMac,
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.5",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"DNT":             "1",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
			"Sec-Fetch-User":  "?1",
		},
	}
}

// SafariMacProfile returns a complete Safari/Mac browser profile.
func SafariMacProfile() Profile {
	return Profile{
		UserAgent: SafariMac,
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
		},
	}
}

// SafariIPadProfile returns a complete Safari/iPad browser profile.
func SafariIPadProfile() Profile {
	return Profile{
		UserAgent: SafariIPad,
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
		},
	}
}

// SafariIPhoneProfile returns a complete Safari/iPhone browser profile.
func SafariIPhoneProfile() Profile {
	return Profile{
		UserAgent: SafariIPhone,
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
		},
	}
}

// EdgeWindowsProfile returns a complete Edge/Windows browser profile.
func EdgeWindowsProfile() Profile {
	return Profile{
		UserAgent: EdgeWin,
		Headers: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-CH-UA":                 `"Chromium";v="122", "Not(A:Brand";v="24", "Microsoft Edge";v="122"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
		},
	}
}

// ChromeAndroidProfile returns a complete Chrome/Android browser profile.
func ChromeAndroidProfile() Profile {
	return Profile{
		UserAgent: ChromeAndroid,
		Headers: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-CH-UA":                 `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
			"Sec-CH-UA-Mobile":          "?1",
			"Sec-CH-UA-Platform":        `"Android"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
		},
	}
}

// ChromeAndroidTabletProfile returns a complete Chrome/Android Tablet browser profile.
func ChromeAndroidTabletProfile() Profile {
	return Profile{
		UserAgent: ChromeAndroidTablet,
		Headers: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-CH-UA":                 `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"Android"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
		},
	}
}
