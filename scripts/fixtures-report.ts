// Run the pinned report gate without writing disposable reports into the
// submodule. Only paths change; all gate logic remains upstream code.
import { resolve, join } from 'node:path';
import { mkdirSync, writeFileSync } from 'node:fs';
const run = process.env.GI_TEST_RUN_ROOT;
if (!run || !process.env.PROJECT_TMP_ROOT || !run.startsWith(process.env.PROJECT_TMP_ROOT + '/runs/')) throw Error('Use make fixtures-vibes-report');
const repo = resolve('references/fixtures-vibes');
let source = await Bun.file(join(repo, 'suite/report.ts')).text();
for (const anchor of ["const root = resolve(import.meta.dir, '..');", "const out = resolve(root, 'test-results');"]) {
  if (!source.includes(anchor)) throw Error('Pinned report path anchor changed: ' + anchor);
}
source = source.replace("const root = resolve(import.meta.dir, '..');", `const root = ${JSON.stringify(repo)};`)
  .replace("const out = resolve(root, 'test-results');", `const out = ${JSON.stringify(join(run, 'results/fixtures'))};`)
  .replace("from './catalogue'", `from ${JSON.stringify(join(repo, 'suite/catalogue.ts'))}`)
  .replace("from '../tools/manifest'", `from ${JSON.stringify(join(repo, 'tools/manifest.ts'))}`);
mkdirSync(join(run,'report'), {recursive:true});
const path = join(run,'report/report.ts');
writeFileSync(path, source);
const proc = Bun.spawn([process.execPath, path], {env:{...process.env,FIXTURES_RESULTS:process.env.FIXTURES_RESULTS || join(run,'results/fixtures/results.json')},stdout:'inherit',stderr:'inherit'});
process.exit(await proc.exited);
