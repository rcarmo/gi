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

mkdirSync(distDir, { recursive: true });
mkdirSync(vendorDir, { recursive: true });

run([
  'bun', 'build',
  `${webSrc}/vendor/preact-htm-entry.ts`,
  '--target=browser',
  '--format=esm',
  '--minify',
  '--sourcemap',
]);
if (existsSync(`${vendorDir}/preact-htm.js`)) rmSync(`${vendorDir}/preact-htm.js`);
if (existsSync(`${vendorDir}/preact-htm.js.map`)) rmSync(`${vendorDir}/preact-htm.js.map`);
renameSync(`${webSrc}/vendor/preact-htm-entry.js`, `${vendorDir}/preact-htm.js`);
renameSync(`${webSrc}/vendor/preact-htm-entry.js.map`, `${vendorDir}/preact-htm.js.map`);

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
