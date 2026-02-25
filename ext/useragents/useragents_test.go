package useragents_test

import (
	"strings"
	"testing"

	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

var allUAConstants = []struct {
	name string
	ua   string
}{
	{"ChromeWin", useragents.ChromeWin},
	{"ChromeWin11", useragents.ChromeWin11},
	{"ChromeMac", useragents.ChromeMac},
	{"ChromeLinux", useragents.ChromeLinux},
	{"FirefoxWin", useragents.FirefoxWin},
	{"FirefoxMac", useragents.FirefoxMac},
	{"SafariMac", useragents.SafariMac},
	{"SafariIPad", useragents.SafariIPad},
	{"SafariIPhone", useragents.SafariIPhone},
	{"EdgeWin", useragents.EdgeWin},
	{"ChromeAndroid", useragents.ChromeAndroid},
	{"ChromeAndroidTablet", useragents.ChromeAndroidTablet},
	{"Googlebot", useragents.Googlebot},
	{"GooglebotMobile", useragents.GooglebotMobile},
	{"Bingbot", useragents.Bingbot},
}

func TestUAConstantsNonEmpty(t *testing.T) {
	for _, tc := range allUAConstants {
		if tc.ua == "" {
			t.Errorf("%s: user agent string is empty", tc.name)
		}
		if !strings.HasPrefix(tc.ua, "Mozilla/") {
			t.Errorf("%s: user agent should start with 'Mozilla/', got %q", tc.name, tc.ua[:20])
		}
	}
}

func TestProfileRequiredFields(t *testing.T) {
	profiles := []struct {
		name    string
		profile useragents.Profile
	}{
		{"ChromeWindowsProfile", useragents.ChromeWindowsProfile()},
		{"ChromeMacProfile", useragents.ChromeMacProfile()},
		{"FirefoxWindowsProfile", useragents.FirefoxWindowsProfile()},
		{"FirefoxMacProfile", useragents.FirefoxMacProfile()},
		{"SafariMacProfile", useragents.SafariMacProfile()},
		{"SafariIPadProfile", useragents.SafariIPadProfile()},
		{"SafariIPhoneProfile", useragents.SafariIPhoneProfile()},
		{"EdgeWindowsProfile", useragents.EdgeWindowsProfile()},
		{"ChromeAndroidProfile", useragents.ChromeAndroidProfile()},
		{"ChromeAndroidTabletProfile", useragents.ChromeAndroidTabletProfile()},
	}

	requiredKeys := []string{"Accept-Language", "Accept-Encoding", "Accept"}

	for _, tc := range profiles {
		if tc.profile.UserAgent == "" {
			t.Errorf("%s: UserAgent is empty", tc.name)
		}
		if tc.profile.Headers == nil {
			t.Errorf("%s: Headers map is nil", tc.name)
			continue
		}
		for _, key := range requiredKeys {
			if v := tc.profile.Headers[key]; v == "" {
				t.Errorf("%s: missing required header %q", tc.name, key)
			}
		}
	}
}

func TestChromiumProfilesHaveSecCHUA(t *testing.T) {
	chromiumProfiles := []struct {
		name    string
		profile useragents.Profile
	}{
		{"ChromeWindowsProfile", useragents.ChromeWindowsProfile()},
		{"ChromeMacProfile", useragents.ChromeMacProfile()},
		{"EdgeWindowsProfile", useragents.EdgeWindowsProfile()},
		{"ChromeAndroidProfile", useragents.ChromeAndroidProfile()},
		{"ChromeAndroidTabletProfile", useragents.ChromeAndroidTabletProfile()},
	}
	for _, tc := range chromiumProfiles {
		if v := tc.profile.Headers["Sec-CH-UA"]; v == "" {
			t.Errorf("%s: missing Sec-CH-UA header (required for Chromium profiles)", tc.name)
		}
	}
}
