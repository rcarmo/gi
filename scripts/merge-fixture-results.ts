import { readdirSync, readFileSync, mkdirSync, writeFileSync } from 'node:fs';
import { join, dirname } from 'node:path';

export const fixtureProjects = ['chromium-phone','chromium-tablet','chromium-desktop','webkit-phone','webkit-tablet','webkit-desktop'];
export function mergeFixtureResults(reports: any[]) {
  if (reports.length !== fixtureProjects.length) throw Error('Expected exactly six shard reports');
  const seen = new Set<string>();
  const suites: any[] = [];
  for (const report of reports) {
    const projects = report.config?.projects;
    // Playwright includes the selected project in each test result even when
    // config.projects lists all configured projects.
    const actual = new Set<string>();
    let count = 0;
    const visit = (suite: any) => {
      for (const spec of suite.specs || []) for (const test of spec.tests || []) {
        actual.add(test.projectName); count++;
        if (!['expected','skipped'].includes(test.status)) throw Error(`Unsuccessful test: ${spec.title}`);
      }
      for (const child of suite.suites || []) visit(child);
    };
    for (const suite of report.suites || []) visit(suite);
    if (!count || actual.size !== 1 || report.errors?.length) throw Error('Empty, mixed or errored shard');
    const project = [...actual][0];
    if (!fixtureProjects.includes(project) || seen.has(project)) throw Error(`Invalid/duplicate project: ${project}`);
    if (projects && !projects.some((p: any) => p.name === project)) throw Error('Project missing from config');
    seen.add(project); suites.push(...report.suites);
  }
  if (fixtureProjects.some(p => !seen.has(p))) throw Error('Missing browser/device coverage');
  return { ...reports[0], suites, errors: [], stats: reports.reduce((out, r) => {
    for (const key of ['expected','unexpected','flaky','skipped']) out[key] = (out[key] || 0) + (r.stats?.[key] || 0);
    return out;
  }, {} as Record<string, number>) };
}
if (import.meta.main) {
  const [root, output] = process.argv.slice(2);
  if (!root || !output) throw Error('Usage: merge-fixture-results.ts shard-dir output.json');
  const reports = readdirSync(root, { withFileTypes: true }).filter(e => e.isDirectory())
    .map(e => JSON.parse(readFileSync(join(root,e.name,'results.json'),'utf8')));
  const merged = mergeFixtureResults(reports);
  mkdirSync(dirname(output), { recursive: true });
  writeFileSync(output, JSON.stringify(merged));
}
