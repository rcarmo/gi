import { resolve, dirname } from 'path';
import { mkdirSync, readFileSync, writeFileSync, renameSync, existsSync, rmSync, copyFileSync, cpSync } from 'fs';
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
// Keep the renderer CSS/fonts on the same version as its JavaScript bundle.
mkdirSync('internal/web/static/fonts/katex', { recursive: true });
cpSync('node_modules/katex/dist/fonts', 'internal/web/static/fonts/katex', { recursive: true });
copyFileSync('node_modules/katex/LICENSE', 'internal/web/static/fonts/katex/LICENSE');
const katexCSS = readFileSync('node_modules/katex/dist/katex.min.css', 'utf8');
writeFileSync('internal/web/static/css/katex.min.css', katexCSS.replaceAll('url(fonts/', 'url(/fonts/katex/'));
buildVendor('mermaid-entry.ts',    vendorDir,  'beautiful-mermaid.js');
buildVendor('codemirror-entry.ts', editorVendorDir, 'codemirror.js');

// ── App bundle ────────────────────────────────────────────────────────────
const appBuild = await Bun.build({
  entrypoints: [`${webSrc}/gi-bootstrap.ts`], outdir: distDir,
  target: 'browser', format: 'esm', sourcemap: 'linked', splitting: true, modulePreload: false,
  naming: { entry: 'app.bundle.[ext]', chunk: 'chunks/[name]-[hash].[ext]', asset: 'assets/[name]-[hash].[ext]' },
  external: ['/editor-vendor/codemirror.js'],
  plugins: [{ name: 'gi-appearance-renderer', setup(build) {
    // Export only the existing catalogue and renderer; leave supplied source bytes unchanged.
    build.onLoad({ filter: /[\\/]ui[\\/]theme\.ts$/ }, async args => ({
      contents: await Bun.file(args.path).text() + '\nexport { THEME_PRESETS as giThemePresets, applyThemeState as giApplyThemeState };\n',
      loader: 'ts',
    }));
  } }, { name: 'gi-clipboard-safety', setup(build) {
    build.onResolve({filter: /^\.\/post-runtime-safety\.js$/}, args => {
      if (args.importer === resolve(webSrc, 'components/post.ts')) {
        return { path: resolve(webSrc, 'gi-clipboard-safety.ts') };
      }
    });
  } }],
});
if (!appBuild.success) { console.error(appBuild.logs); process.exit(1); }
// Retain only this build's hashed chunks. Old open clients receive a pane-load
// error after deployment rather than executing a mismatched module.
const outputPaths = new Set(appBuild.outputs.map(output => resolve(output.path)));
for (const entry of new Bun.Glob('chunks/**/*').scanSync({ cwd: distDir, onlyFiles: true })) {
  const path = resolve(distDir, entry);
  if (!outputPaths.has(path)) rmSync(path);
}

// No post-processing needed — vendor scripts are loaded as modules.

// Now clean up preact-htm alias
const phtmAlias = `${webSrc}/vendor/preact-htm.js`;
if (existsSync(phtmAlias)) rmSync(phtmAlias);

// Clean up vendor alias copies EXCEPT preact-htm.js (kept for app externals resolution)
['marked.min.js', 'katex.min.js', 'beautiful-mermaid.js', 'codemirror.js'].forEach((f) => {
  const p = `${webSrc}/vendor/${f}`;
  if (existsSync(p)) rmSync(p);
});

// ── CSS bundle ────────────────────────────────────────────────────────────
// CSS bundle — all Piclaw CSS is served from /css/styles.css (with @import partials).
// app.bundle.css is kept minimal — only Gi-specific overrides go here.
writeFileSync(`${distDir}/app.bundle.css`, '/* Gi app overrides */\n', 'utf-8');

console.log('Gi web build complete.');
