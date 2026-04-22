import { resolve, dirname } from 'path';
import { mkdirSync, readFileSync, writeFileSync, renameSync, existsSync, rmSync, copyFileSync } from 'fs';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
process.chdir(__dirname);

const webSrc = 'web/src';
const distDir = 'internal/web/static/dist';
const vendorDir = 'internal/web/static/js/vendor';
const jsDir = 'internal/web/static/js';
const editorVendorDir = 'internal/web/static/editor-vendor';

mkdirSync(distDir, { recursive: true });
mkdirSync(vendorDir, { recursive: true });
mkdirSync(jsDir, { recursive: true });
mkdirSync(editorVendorDir, { recursive: true });

function run(cmd) {
  const proc = Bun.spawnSync(cmd, { stdout: 'inherit', stderr: 'inherit' });
  if (proc.exitCode !== 0) process.exit(proc.exitCode ?? 1);
}

function move(src, dst) {
  if (existsSync(dst)) rmSync(dst);
  if (existsSync(src)) renameSync(src, dst);
}

// Build a vendor entry. Leaves a copy in web/src/vendor/<outputName> for
// app.ts import resolution, then moves the canonical output to static/.
function buildVendor(entryFile, finalDir, finalName) {
  const entry = `${webSrc}/vendor/${entryFile}`;
  const base = entryFile.replace(/\.ts$/, '');
  run(['bun', 'build', entry, '--target=browser', '--format=esm', '--minify', '--sourcemap']);
  const srcJs = `${webSrc}/vendor/${base}.js`;
  const srcMap = `${webSrc}/vendor/${base}.js.map`;
  // Leave a copy at the exact path components expect (e.g. preact-htm.js)
  const vendorAlias = `${webSrc}/vendor/${finalName}`;
  if (existsSync(srcJs) && vendorAlias !== srcJs) {
    copyFileSync(srcJs, vendorAlias);
  }
  // Move canonical build output to static/
  move(srcJs, `${finalDir}/${finalName}`);
  move(srcMap, `${finalDir}/${finalName}.map`);
}

// ── Vendor bundles ────────────────────────────────────────────────────────
buildVendor('preact-htm-entry.ts', vendorDir, 'preact-htm.js');
buildVendor('marked-entry.ts',     jsDir,      'marked.min.js');
buildVendor('katex-entry.ts',      vendorDir,  'katex.min.js');
buildVendor('mermaid-entry.ts',    vendorDir,  'beautiful-mermaid.js');
buildVendor('codemirror-entry.ts', editorVendorDir, 'codemirror.js');

// ── App bundle ────────────────────────────────────────────────────────────
run([
  'bun', 'build',
  `${webSrc}/app.ts`,
  '--target=browser',
  '--format=esm',
  '--sourcemap',
  '--external', '#editor-vendor/codemirror',
  '--external', './vendor/preact-htm.js',
  '--external', '../vendor/preact-htm.js',
]);
move(`${webSrc}/app.js`,     `${distDir}/app.bundle.js`);
move(`${webSrc}/app.js.map`, `${distDir}/app.bundle.js.map`);

// Post-process: wrap in IIFE to isolate var declarations from global scope.
// This prevents Safari's "duplicate variable that shadows a global property" 
// error without breaking var's re-declaration semantics.
const appBundle = readFileSync(`${distDir}/app.bundle.js`, 'utf-8');
writeFileSync(`${distDir}/app.bundle.js`, `(function(){\n${appBundle}\n})();\n`, 'utf-8');

// Now clean up preact-htm alias
const phtmAlias = `${webSrc}/vendor/preact-htm.js`;
if (existsSync(phtmAlias)) rmSync(phtmAlias);

// Clean up vendor alias copies EXCEPT preact-htm.js (kept for app externals resolution)
['marked.min.js', 'katex.min.js', 'beautiful-mermaid.js', 'codemirror.js'].forEach((f) => {
  const p = `${webSrc}/vendor/${f}`;
  if (existsSync(p)) rmSync(p);
});

// ── CSS bundle ────────────────────────────────────────────────────────────
const appCss = readFileSync(`${webSrc}/styles/app.css`, 'utf-8');
const minCss = appCss.replace(/\/\*[\s\S]*?\*\//g, '').replace(/[ \t]*\n[ \t]*/g, '\n').replace(/\n{2,}/g, '\n').trim();
writeFileSync(`${distDir}/app.bundle.css`, minCss, 'utf-8');

console.log('Gi web build complete.');
