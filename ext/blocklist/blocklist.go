package blocklist

import phantomjscloud "github.com/amafjarkasi/go-phantomjs"

// ResourceModifier returns a PhantomJsCloud ResourceModifier that blacklists the given URL regex.
func block(regex string) phantomjscloud.ResourceModifier {
	return phantomjscloud.ResourceModifier{Regex: regex, IsBlacklisted: true}
}

// Ads returns ResourceModifiers that block the most common advertising networks.
// Reduces page load time and billing cost significantly on ad-heavy pages.
func Ads() []phantomjscloud.ResourceModifier {
	return []phantomjscloud.ResourceModifier{
		block(`.*doubleclick\.net.*`),
		block(`.*googlesyndication\.com.*`),
		block(`.*googleadservices\.com.*`),
		block(`.*adnxs\.com.*`),
		block(`.*serving-sys\.com.*`),
		block(`.*amazon-adsystem\.com.*`),
		block(`.*ads\.twitter\.com.*`),
		block(`.*advertising\.com.*`),
		block(`.*ads\.linkedin\.com.*`),
		block(`.*adsymptotic\.com.*`),
		block(`.*moatads\.com.*`),
		block(`.*criteo\.com.*`),
		block(`.*taboola\.com.*`),
		block(`.*outbrain\.com.*`),
		block(`.*revcontent\.com.*`),
		block(`.*adroll\.com.*`),
		block(`.*pubmatic\.com.*`),
		block(`.*openx\.net.*`),
		block(`.*rubiconproject\.com.*`),
		block(`.*smartadserver\.com.*`),
	}
}

// Trackers returns ResourceModifiers that block analytics and tracking beacons.
// Prevents session fingerprinting and reduces unnecessary request volume.
func Trackers() []phantomjscloud.ResourceModifier {
	return []phantomjscloud.ResourceModifier{
		block(`.*google-analytics\.com.*`),
		block(`.*googletagmanager\.com.*`),
		block(`.*segment\.com.*`),
		block(`.*segment\.io.*`),
		block(`.*mixpanel\.com.*`),
		block(`.*fullstory\.com.*`),
		block(`.*hotjar\.com.*`),
		block(`.*mouseflow\.com.*`),
		block(`.*crazyegg\.com.*`),
		block(`.*heap\.io.*`),
		block(`.*amplitude\.com.*`),
		block(`.*clarity\.ms.*`),
		block(`.*intercom\.io.*`),
		block(`.*intercomcdn\.com.*`),
		block(`.*drift\.com.*`),
		block(`.*hubspot\.com.*`),
		block(`.*marketo\.net.*`),
		block(`.*pardot\.com.*`),
		block(`.*sentry\.io.*`),
		block(`.*newrelic\.com.*`),
		block(`.*nr-data\.net.*`),
		block(`.*datadog-browser-agent\.com.*`),
		block(`.*facebook\.net.*`),
		block(`.*connect\.facebook\.net.*`),
		block(`.*ads\.facebook\.com.*`),
		block(`.*bat\.bing\.com.*`),
		block(`.*sc-static\.net.*`),
		block(`.*tiktok\.com.*`),
	}
}

// Media returns ResourceModifiers that block images and video assets.
// Use when you only need page text or DOM structure — can cut load time in half.
func Media() []phantomjscloud.ResourceModifier {
	return []phantomjscloud.ResourceModifier{
		block(`.*\.(jpg|jpeg|png|gif|webp|svg|ico|avif|bmp|tiff)(\?.*)?$`),
		block(`.*\.(mp4|webm|ogg|avi|mov|mkv|flv)(\?.*)?$`),
		block(`.*\.(mp3|wav|flac|aac|m4a)(\?.*)?$`),
	}
}

// Fonts returns ResourceModifiers that block web font requests.
// Saves bandwidth without affecting scraping accuracy on most sites.
func Fonts() []phantomjscloud.ResourceModifier {
	return []phantomjscloud.ResourceModifier{
		block(`.*fonts\.googleapis\.com.*`),
		block(`.*fonts\.gstatic\.com.*`),
		block(`.*use\.typekit\.net.*`),
		block(`.*fast\.fonts\.net.*`),
		block(`.*cloud\.typography\.com.*`),
		block(`.*\.(woff|woff2|ttf|eot|otf)(\?.*)?$`),
	}
}

// Lightweight returns a combined blocklist of Ads + Trackers + Fonts.
// The recommended default for most scraping tasks — minimal footprint with no
// impact on content accuracy.
func Lightweight() []phantomjscloud.ResourceModifier {
	rules := make([]phantomjscloud.ResourceModifier, 0, 60)
	rules = append(rules, Ads()...)
	rules = append(rules, Trackers()...)
	rules = append(rules, Fonts()...)
	return rules
}

// Full returns the complete combined blocklist of Ads + Trackers + Fonts + Media.
// Use when you only need plain text or DOM structure and don't care about rendering.
func Full() []phantomjscloud.ResourceModifier {
	rules := make([]phantomjscloud.ResourceModifier, 0, 70)
	rules = append(rules, Ads()...)
	rules = append(rules, Trackers()...)
	rules = append(rules, Fonts()...)
	rules = append(rules, Media()...)
	return rules
}
