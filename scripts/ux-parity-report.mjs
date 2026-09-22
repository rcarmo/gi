import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { loadCorpus, mappedIds, sharedMappedIds } from '../tests/ux/support/catalogue.mjs';

const classic = loadCorpus();
const shared = loadCorpus('shared');
const expectedProjects = ['chromium-phone','chromium-tablet','chromium-desktop','webkit-phone','webkit-tablet','webkit-desktop'];
const rows = new Map();
for (const item of classic) {
  if (!rows.has(item.id)) rows.set(item.id, { id: item.id, name: item.name, source: `${item.uri}:${item.line}`, expandedCases: 0, status: 'unmapped', results: [] });
  rows.get(item.id).expandedCases++;
}
const sharedRows = new Map(shared.map(item => [item.id, { id: item.id, name: item.name, source: `${item.uri}:${item.line}`, expandedCases: 1, status: 'unmapped', results: [] }]));
// Explicit result paths only: do not silently reuse a stale fixture run.
for (const resultPath of process.argv.slice(2)) {
  const result = JSON.parse(readFileSync(resultPath, 'utf8'));
  function walk(suites) {
    for (const suite of suites || []) {
      for (const spec of suite.specs || []) {
        const id = spec.title.match(/@(?:ux|shared)-[\w-]+/)?.[0];
        const row = rows.get(id) || sharedRows.get(id);
        if (!row) continue;
        for (const test of spec.tests || []) {
          const last = test.results?.at(-1);
          row.results.push({ project: test.projectName, status: last?.status || 'not-run', errors: (last?.errors || []).map(error => error.message) });
        }
      }
      walk(suite.suites);
    }
  }
  walk(result.suites);
}
for (const row of [...rows.values(), ...sharedRows.values()]) {
  if (!mappedIds.has(row.id) && !sharedMappedIds.has(row.id)) continue;
  row.status = !row.results.length ? 'not-run' : row.results.some(test => test.status !== 'passed') ? 'fail' : expectedProjects.every(project => row.results.filter(test => test.project === project && test.status === 'passed').length === row.expandedCases) ? 'pass' : 'partial-matrix';
}
const counts = {};
for (const row of rows.values()) counts[row.status] = (counts[row.status] || 0) + 1;
const sharedCounts = {};
for (const row of sharedRows.values()) sharedCounts[row.status] = (sharedCounts[row.status] || 0) + 1;
const report = {
  sourceCommit: '70d33bc93ab540845bbcf5f80503ca8125c71594',
  scenarios: rows.size, expandedCases: classic.length, sharedContractCases: shared.length,
  expectedProjects, counts, rows: [...rows.values()], sharedCounts, sharedRows: [...sharedRows.values()],
};
mkdirSync('test-results/ux-parity', { recursive: true });
writeFileSync('test-results/ux-parity/matrix.json', JSON.stringify(report, null, 2) + '\n');
writeFileSync('test-results/ux-parity/matrix.md', `# Gi Piclaw Classic parity\n\nSource: \`${report.sourceCommit}\`. Frozen corpus: ${rows.size} scenarios / ${classic.length} expanded cases. Shared Vibes/Tau contract: ${shared.length} expanded cases, reported separately below.\n\nCounts: ${JSON.stringify(counts)}. A pass requires all six browser/viewport projects; unmapped is not a pass or skip.\n\n| ID | Status | Scenario |\n|---|---|---|\n${[...rows.values()].map(row => `| ${row.id} | ${row.status} | ${row.name.replaceAll('|', '\\|')} |`).join('\n')}\n`);
const sharedMarkdown = `\n## Shared contract (separate evidence)\n\nCounts: ${JSON.stringify(sharedCounts)}.\n\n| ID | Status | Scenario |\n|---|---|---|\n${[...sharedRows.values()].map(row => `| ${row.id} | ${row.status} | ${row.name.replaceAll('|', '\\|')} |`).join('\n')}\n`;
const matrixPath = 'test-results/ux-parity/matrix.md';
writeFileSync(matrixPath, readFileSync(matrixPath,'utf8') + sharedMarkdown);
console.log(JSON.stringify({ scenarios: rows.size, expandedCases: classic.length, sharedContractCases: shared.length, counts, sharedCounts }));
