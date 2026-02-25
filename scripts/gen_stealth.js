/**
 * gen_stealth.js
 *
 * Run from repo root:
 *   node scripts/gen_stealth.js
 *
 * Bundles evasions from puppeteer-extra-plugin-stealth into:
 *   ext/stealth/evasions.js   raw JS payload (regenerable, can be gitignored)
 *   ext/stealth/stealth.go    Go file that embeds evasions.js via go:embed
 *
 * Re-run whenever the plugin is updated:
 *   npm update puppeteer-extra-plugin-stealth && node scripts/gen_stealth.js
 */

'use strict';

const fs = require('fs');
const fsp = fs.promises;
const path = require('path');

const ROOT = path.join(__dirname, '..');
const EVASION_DIR = path.join(ROOT, 'node_modules', 'puppeteer-extra-plugin-stealth', 'evasions');
const UTILS_INDEX = path.join(EVASION_DIR, '_utils', 'index.js');
const OUT_DIR = path.join(ROOT, 'ext', 'stealth');
const OUT_JS = path.join(OUT_DIR, 'evasions.js');
const OUT_GO = path.join(OUT_DIR, 'stealth.go');

async function exists(path) {
    try {
        await fsp.access(path);
        return true;
    } catch {
        return false;
    }
}

// ── 2. Balanced-brace extractor ─────────────────────────────────────────────
function extractBalancedBraces(src, start) {
    let depth = 0;
    for (let i = start; i < src.length; i++) {
        if (src[i] === '{') depth++;
        else if (src[i] === '}') {
            depth--;
            if (depth === 0) return src.slice(start, i + 1);
        }
    }
    return null;
}

// ── 3. Payload extractor ────────────────────────────────────────────────────
// Each evasion calls`.evaluateOnNewDocument(fn)` where fn is an arrow function.
// We split on the marker, then grab params and body separately.
function extractEvalFn(src) {
    const MARKER = '.evaluateOnNewDocument(';
    const idx = src.indexOf(MARKER);
    if (idx === -1) return null;

    let pos = idx + MARKER.length;

    // Skip whitespace
    while (pos < src.length && /[\s\n\r]/.test(src[pos])) pos++;

    // Detect `function` keyword vs arrow function
    let paramStr;
    let bodyStart;

    if (src.slice(pos, pos + 8) === 'function') {
        // function([params]) { ... }
        const openP = src.indexOf('(', pos);
        const closeP = src.indexOf(')', openP + 1);
        paramStr = src.slice(openP + 1, closeP);
        bodyStart = src.indexOf('{', closeP + 1);
    } else {
        // Arrow: () =>  or  utils =>  or  (utils, opts) =>
        const arrowIdx = src.indexOf('=>', pos);
        if (arrowIdx === -1) return null;
        const rawParams = src.slice(pos, arrowIdx).trim();
        // Strip surrounding parens
        paramStr = rawParams.replace(/^\(|\)$/g, '');
        bodyStart = src.indexOf('{', arrowIdx + 2);
    }

    if (bodyStart === -1) return null;
    const body = extractBalancedBraces(src, bodyStart);
    if (!body) return null;

    return { params: paramStr.trim(), body };
}

async function run() {
    if (!(await exists(EVASION_DIR))) {
        console.error('ERROR: puppeteer-extra-plugin-stealth not found.');
        console.error('Run:   npm install puppeteer-extra-plugin-stealth');
        process.exit(1);
    }

    await fsp.mkdir(OUT_DIR, { recursive: true });

    // ── 1. Utils block ──────────────────────────────────────────────────────────
    let utilsSrc = await fsp.readFile(UTILS_INDEX, 'utf8');
    // Remove the CJS export line so we can inline the code.
    utilsSrc = utilsSrc.replace(/^module\.exports\s*=\s*utils\s*;?\s*$/m, '').trimEnd();

    // ── 4. Process each evasion directory ───────────────────────────────────────
    const allDirs = await fsp.readdir(EVASION_DIR);

    const evasionDirNames = (await Promise.all(
        allDirs
            .filter(d => !d.startsWith('_'))
            .map(async d => {
                const stat = await fsp.stat(path.join(EVASION_DIR, d));
                return stat.isDirectory() ? d : null;
            })
    )).filter(Boolean).sort();

    const results = await Promise.all(evasionDirNames.map(async name => {
        const indexPath = path.join(EVASION_DIR, name, 'index.js');
        if (!(await exists(indexPath))) return null;

        const src = await fsp.readFile(indexPath, 'utf8');
        const result = extractEvalFn(src);

        if (!result) {
            return { name, error: `WARN: could not extract payload from "${name}" — skipping.` };
        }

        const { params, body } = result;
        const needsUtils = /withUtils/.test(src);

        // Build the call expression — pass `utils` if the evasion accepts it
        const argList = needsUtils ? 'utils' : '';
        const fnLiteral = `(${params}) => ${body}`;
        const call = `(${fnLiteral})(${argList});`;

        return {
            name,
            params,
            needsUtils,
            block: `  // ── Evasion: ${name}\n` +
                   `  try {\n    ${call}\n  } catch (e) { console.warn('[stealth] ${name}:', e.message); }`
        };
    }));

    const finalEvasionBlocks = [];
    for (const res of results) {
        if (!res) continue;
        if (res.error) {
            console.warn(res.error);
            continue;
        }
        console.log(`  ✓  ${res.name}  (params: "${res.params}"${res.needsUtils ? ', with utils' : ''})`);
        finalEvasionBlocks.push(res.block);
    }

    if (finalEvasionBlocks.length === 0) {
        console.error('ERROR: No evasion payloads extracted. Check the plugin source.');
        process.exit(1);
    }

    // ── 5. Assemble the combined IIFE ───────────────────────────────────────────
    const lines = [
        '// AUTO-GENERATED — do not edit. Run: node scripts/gen_stealth.js',
        '(function () {',
        '',
        utilsSrc,
        '',
        'utils.init();',
        '',
        finalEvasionBlocks.join('\n\n'),
        '',
        '})();',
    ];
    const combinedJs = lines.join('\n');

    await fsp.writeFile(OUT_JS, combinedJs, 'utf8');
    console.log(`\nWrote ${OUT_JS}  (${(combinedJs.length / 1024).toFixed(1)} KB, ${finalEvasionBlocks.length} evasions)`);

    // ── 6. Write the Go embed wrapper ────────────────────────────────────────────
    const goLines = [
        'package stealth',
        '',
        'import _ "embed"',
        '',
        '// JS is the combined stealth evasion script from puppeteer-extra-plugin-stealth.',
        '// Inject via page.evaluateOnNewDocument to spoof browser fingerprinting.',
        '// Regenerate: node scripts/gen_stealth.js',
        '//',
        '//go:embed evasions.js',
        'var JS string',
        '',
    ];

    await fsp.writeFile(OUT_GO, goLines.join('\n'), 'utf8');
    console.log(`Wrote ${OUT_GO}`);
    console.log('Done.');
}

run().catch(err => {
    console.error(err);
    process.exit(1);
});
