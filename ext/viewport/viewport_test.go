package viewport_test

import (
	"testing"

	"github.com/jbdt/go-phantomjs/ext/viewport"
)

func TestDesktopPresetDimensions(t *testing.T) {
	cases := []struct {
		name         string
		preset       viewport.Preset
		wantW, wantH int
	}{
		{"HD", viewport.HD, 1280, 720},
		{"FHD", viewport.FHD, 1920, 1080},
		{"QHD", viewport.QHD, 2560, 1440},
		{"UHD", viewport.UHD, 3840, 2160},
		{"Laptop", viewport.Laptop, 1366, 768},
	}
	for _, tc := range cases {
		rs := tc.preset.AsRenderSettings()
		if rs.Viewport == nil {
			t.Errorf("%s: AsRenderSettings().Viewport is nil", tc.name)
			continue
		}
		if rs.Viewport.Width != tc.wantW || rs.Viewport.Height != tc.wantH {
			t.Errorf("%s: want %dx%d, got %dx%d",
				tc.name, tc.wantW, tc.wantH, rs.Viewport.Width, rs.Viewport.Height)
		}
		if rs.Viewport.IsMobile {
			t.Errorf("%s: desktop preset should not have IsMobile=true", tc.name)
		}
	}
}

func TestMobilePresetsFlags(t *testing.T) {
	cases := []struct {
		name          string
		preset        viewport.Preset
		wantLandscape bool
	}{
		{"MobilePortrait", viewport.MobilePortrait, false},
		{"MobileLandscape", viewport.MobileLandscape, true},
		{"TabletPortrait", viewport.TabletPortrait, false},
		{"TabletLandscape", viewport.TabletLandscape, true},
	}
	for _, tc := range cases {
		rs := tc.preset.AsRenderSettings()
		if rs.Viewport == nil {
			t.Errorf("%s: Viewport is nil", tc.name)
			continue
		}
		if !rs.Viewport.IsMobile {
			t.Errorf("%s: IsMobile should be true", tc.name)
		}
		if !rs.Viewport.HasTouch {
			t.Errorf("%s: HasTouch should be true", tc.name)
		}
		if rs.Viewport.IsLandscape != tc.wantLandscape {
			t.Errorf("%s: IsLandscape want %v, got %v", tc.name, tc.wantLandscape, rs.Viewport.IsLandscape)
		}
	}
}

func TestThumbnailPresetsHaveClipRectangle(t *testing.T) {
	cases := []struct {
		name   string
		preset viewport.Preset
	}{
		{"Thumbnail640", viewport.Thumbnail640},
		{"Thumbnail1200", viewport.Thumbnail1200},
	}
	for _, tc := range cases {
		rs := tc.preset.AsRenderSettings()
		if rs.ClipRectangle == nil {
			t.Errorf("%s: ClipRectangle should not be nil for thumbnail presets", tc.name)
		}
	}
}

func TestCustomPreset(t *testing.T) {
	p := viewport.Custom(800, 600)
	rs := p.AsRenderSettings()
	if rs.Viewport == nil {
		t.Fatal("Custom: Viewport is nil")
	}
	if rs.Viewport.Width != 800 || rs.Viewport.Height != 600 {
		t.Errorf("Custom: want 800x600, got %dx%d", rs.Viewport.Width, rs.Viewport.Height)
	}
	if rs.ClipRectangle != nil {
		t.Errorf("Custom: ClipRectangle should be nil, got %+v", rs.ClipRectangle)
	}
}

func TestZoomFactorMapping(t *testing.T) {
	rs := viewport.Thumbnail640.AsRenderSettings()
	if rs.ZoomFactor == 0 {
		t.Errorf("Thumbnail640: ZoomFactor should be non-zero")
	}
	// FHD has no explicit zoom
	rs = viewport.FHD.AsRenderSettings()
	if rs.ZoomFactor != 0 {
		t.Errorf("FHD: ZoomFactor should be 0 (use default), got %v", rs.ZoomFactor)
	}
}
