package stealth

import _ "embed"

// JS is the combined stealth evasion script from puppeteer-extra-plugin-stealth.
// Inject via page.evaluateOnNewDocument to spoof browser fingerprinting.
// Regenerate: node scripts/gen_stealth.js
//
//go:embed evasions.js
var JS string
