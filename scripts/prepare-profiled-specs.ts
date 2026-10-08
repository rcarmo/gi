// Copy shared tests to owned scratch and add capture instrumentation. Test
// assertions remain unchanged; the pinned frontend/suite checkout stays clean.
import { resolve, join } from 'node:path';
import { readdirSync, mkdirSync, writeFileSync } from 'node:fs';
import { relocateSpecImports } from './profiled-spec-imports';
const run = process.env.GI_TEST_RUN_ROOT;
if (!run || !process.env.PROJECT_TMP_ROOT || !run.startsWith(process.env.PROJECT_TMP_ROOT + '/runs/')) throw Error('Use the project Makefile');
const guard = Bun.spawn(['bash','-c','source scripts/project-test-env.sh && gi_test_path "$GI_TEST_RUN_ROOT/specs"'], {stdout:'inherit',stderr:'inherit'});
if (await guard.exited) process.exit(1);
const suite = resolve('references/fixtures-vibes/suite');
const out = join(run,'specs');
mkdirSync(out,{recursive:true});
writeFileSync(join(out,'package.json'),JSON.stringify({type:'module'}));
for (const name of readdirSync(join(suite,'specs')).filter(n=>n.endsWith('.ts'))) {
  const source = relocateSpecImports(await Bun.file(join(suite,'specs',name)).text(), suite,
    resolve('references/fixtures-vibes/node_modules/@playwright/test/index.mjs'),
    resolve('tests/fixtures-vibes/profiled-fixtures.ts'));
  writeFileSync(join(out,name),source);
}
