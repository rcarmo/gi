import { resolve, dirname } from 'path';
import { mkdirSync, readFileSync, writeFileSync, renameSync, existsSync, rmSync } from 'fs';
import { fileURLToPath } from 'url';

function run(cmd) {
  const proc = Bun.spawnSync(cmd, { stdout: 'inherit', stderr: 'inherit' });
  if (proc.exitCode !== 0) process.exit(proc.exitCode ?? 1);
}

const __dirname = dirname(fileURLToPath(import.meta.url));
process.chdir(__dirname);
const webSrc = 'web/src';
const staticDir = 'internal/web/static';
const distDir = 'internal/web/static/dist';
const vendorDir = 'internal/web/static/js/vendor';
const jsDir = 'internal/web/static/js';
const editorVendorDir = 'internal/web/static/editor-vendor';

mkdirSync(distDir, { recursive: true });
mkdirSync(vendorDir, { recursive: true });
mkdirSync(jsDir, { recursive: true });
mkdirSync(editorVendorDir, { recursive: true });

function vendorEntry(entry, outputDir, outputName) {
  run(['bun', 'build', `${webSrc}/vendor/${entry}.ts`, '--target=browser', '--format=esm', '--minify', '--sourcemap']);
  if (existsSync(`${outputDir}/${outputName}`)) rmSync(`${outputDir}/${outputName}`);
  if (existsSync(`${outputDir}/${outputName}.map`)) rmSync(`${outputDir}/${outputName}.map`);
  renameSync(`${webSrc}/vendor/${entry}.js`, `${outputDir}/${outputName}`);
  renameSync(`${webSrc}/vendor/${entry}.js.map`, `${outputDir}/${outputName}.map`);
}

vendorEntry('preact-htm-entry', vendorDir, 'preact-htm.js');
vendorEntry('marked-entry', jsDir, 'marked.min.js');
vendorEntry('katex-entry', vendorDir, 'katex.min.js');
vendorEntry('mermaid-entry', vendorDir, 'beautiful-mermaid.js');
vendorEntry('codemirror-entry', editorVendorDir, 'codemirror.js');

run([
  'bun', 'build',
  `${webSrc}/app.ts`,
  '--target=browser',
  '--format=esm',
  '--minify',
  '--sourcemap',
]);
if (existsSync(`${distDir}/app.bundle.js`)) rmSync(`${distDir}/app.bundle.js`);
if (existsSync(`${distDir}/app.bundle.js.map`)) rmSync(`${distDir}/app.bundle.js.map`);
renameSync(`${webSrc}/app.js`, `${distDir}/app.bundle.js`);
renameSync(`${webSrc}/app.js.map`, `${distDir}/app.bundle.js.map`);
if (existsSync(`${webSrc}/app.js`)) rmSync(`${webSrc}/app.js`);
if (existsSync(`${webSrc}/app.js.map`)) rmSync(`${webSrc}/app.js.map`);

const cssSources = [`${webSrc}/styles.css`];
const combined = cssSources.map((f) => readFileSync(f, 'utf-8')).join('\n');
const minified = combined
  .replace(/\/\*[\s\S]*?\*\//g, '')
  .replace(/\s*\n\s*/g, '\n')
  .replace(/\n+/g, '\n')
  .replace(/;\s*}/g, '}')
  .replace(/\s*{\s*/g, '{')
  .replace(/\s*}\s*/g, '}')
  .replace(/\s*:\s*/g, ':')
  .replace(/\s*;\s*/g, ';')
  .replace(/\s*,\s*/g, ',')
  .trim();
writeFileSync(`${distDir}/app.bundle.css`, minified, 'utf-8');
console.log('Built Gi web assets');
