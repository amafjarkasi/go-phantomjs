package viewport

import phantomjscloud "github.com/amafjarkasi/go-phantomjs"

// Preset is a named viewport configuration.
type Preset struct {
	Viewport      phantomjscloud.Viewport
	ClipRectangle *phantomjscloud.ClipRectangle // nil means no clip (full page)
	ZoomFactor    float64                       // 0 means use default (1.0)
}

// AsRenderSettings converts the preset into a RenderSettings ready to drop into PageRequest.
func (p Preset) AsRenderSettings() phantomjscloud.RenderSettings {
	rs := phantomjscloud.RenderSettings{
		Viewport:      &p.Viewport,
		ClipRectangle: p.ClipRectangle,
	}
	if p.ZoomFactor != 0 {
		rs.ZoomFactor = p.ZoomFactor
	}
	return rs
}

// ── Desktop Presets ──────────────────────────────────────────────────────────

// HD is a standard 1280×720 desktop viewport.
var HD = Preset{
	Viewport: phantomjscloud.Viewport{Width: 1280, Height: 720},
}

// FHD is a 1920×1080 Full HD desktop viewport — the most common monitor resolution.
var FHD = Preset{
	Viewport: phantomjscloud.Viewport{Width: 1920, Height: 1080},
}

// QHD is a 2560×1440 Quad HD desktop viewport.
var QHD = Preset{
	Viewport: phantomjscloud.Viewport{Width: 2560, Height: 1440},
}

// UHD is a 3840×2160 4K desktop viewport.
var UHD = Preset{
	Viewport: phantomjscloud.Viewport{Width: 3840, Height: 2160},
}

// Laptop is a common laptop viewport (1366×768).
var Laptop = Preset{
	Viewport: phantomjscloud.Viewport{Width: 1366, Height: 768},
}

// ── Mobile Presets ───────────────────────────────────────────────────────────

// MobilePortrait emulates a standard smartphone in portrait orientation (390×844, Pixel 8 / iPhone 15 Pro).
var MobilePortrait = Preset{
	Viewport: phantomjscloud.Viewport{
		Width:             390,
		Height:            844,
		DeviceScaleFactor: 3,
		IsMobile:          true,
		HasTouch:          true,
	},
}

// MobileLandscape emulates a standard smartphone in landscape orientation.
var MobileLandscape = Preset{
	Viewport: phantomjscloud.Viewport{
		Width:             844,
		Height:            390,
		DeviceScaleFactor: 3,
		IsMobile:          true,
		HasTouch:          true,
		IsLandscape:       true,
	},
}

// TabletPortrait emulates a tablet in portrait orientation (768×1024, iPad-ish).
var TabletPortrait = Preset{
	Viewport: phantomjscloud.Viewport{
		Width:             768,
		Height:            1024,
		DeviceScaleFactor: 2,
		IsMobile:          true,
		HasTouch:          true,
	},
}

// TabletLandscape emulates a tablet in landscape orientation.
var TabletLandscape = Preset{
	Viewport: phantomjscloud.Viewport{
		Width:             1024,
		Height:            768,
		DeviceScaleFactor: 2,
		IsMobile:          true,
		HasTouch:          true,
		IsLandscape:       true,
	},
}

// ── Thumbnail Presets ────────────────────────────────────────────────────────

// Thumbnail640 renders a full-width screenshot then clips to a 640×480 thumbnail.
// Ideal for link previews and card images.
var Thumbnail640 = Preset{
	Viewport: phantomjscloud.Viewport{Width: 1280, Height: 800},
	ClipRectangle: &phantomjscloud.ClipRectangle{
		Top: 0, Left: 0, Width: 640, Height: 480,
	},
	ZoomFactor: 0.5,
}

// Thumbnail1200 renders a standard 1200×630 Open Graph thumbnail.
// Use with renderType "jpeg" for maximum compatibility.
var Thumbnail1200 = Preset{
	Viewport: phantomjscloud.Viewport{Width: 1200, Height: 630},
	ClipRectangle: &phantomjscloud.ClipRectangle{
		Top: 0, Left: 0, Width: 1200, Height: 630,
	},
}

// Custom returns a simple Preset from a width and height with no clip or zoom.
func Custom(width, height int) Preset {
	return Preset{
		Viewport: phantomjscloud.Viewport{Width: width, Height: height},
	}
}
