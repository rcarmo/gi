import { readFileSync, readdirSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { generateMessages } from '@cucumber/gherkin';
import { IdGenerator, SourceMediaType } from '@cucumber/messages';

export const uxRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
export const mappedIds = new Set(['@ux-original-001', '@ux-original-002', '@ux-original-013', '@ux-original-014', '@ux-original-015', '@ux-original-016', '@ux-original-018', '@ux-original-021', '@ux-compaction-001', '@ux-compaction-002', '@ux-compaction-003', '@ux-compaction-004', '@ux-compaction-005', '@ux-compaction-006', '@ux-compaction-007', '@ux-compaction-008', '@ux-context-001', '@ux-context-002', '@ux-context-003', '@ux-context-004', '@ux-context-005', '@ux-reconnect-001', '@ux-reconnect-002', '@ux-reconnect-003', '@ux-reconnect-004', '@ux-reconnect-005', '@ux-compose-001', '@ux-compose-002', '@ux-compose-003', '@ux-compose-006', '@ux-compose-007', '@ux-compose-008', '@ux-compose-009', '@ux-compose-010', '@ux-compose-011', '@ux-timeline-023', '@ux-timeline-024', '@ux-original-029', '@ux-original-026', '@ux-timeline-013', '@ux-timeline-014', '@ux-timeline-015', '@ux-timeline-016', '@ux-workspace-008', '@ux-workspace-004']);
export const sharedMappedIds = new Set(['@shared-28', '@shared-30']);
const sha256 = data => createHash('sha256').update(data).digest('hex');

export function verifySources() {
  const lines = readFileSync(join(uxRoot, 'upstream/SHA256SUMS'), 'utf8').trim().split('\n');
  for (const line of lines) {
    const [, hash, source] = line.match(/^([a-f0-9]{64})\s+(.+)$/) || [];
    if (!hash || !source.startsWith('tests/e2e/features/classic/')) throw new Error(`Invalid source manifest line: ${line}`);
    const destination = join(uxRoot, 'features/classic', source.slice('tests/e2e/features/classic/'.length));
    if (sha256(readFileSync(destination)) !== hash) throw new Error(`Frozen Gherkin changed: ${destination}`);
  }
  const sharedHash = sha256(readFileSync(join(uxRoot, 'features/shared-canonical-ux.feature')));
  if (sharedHash !== 'a08a623880c6f327bc051edc51bb2bbff2959aed86421b5227e61d5a92fc2441') {
    throw new Error('Shared Vibes/Tau Gherkin changed');
  }
}

function featureFiles(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap(entry => {
    const path = join(dir, entry.name);
    return entry.isDirectory() ? featureFiles(path) : entry.name.endsWith('.feature') ? [path] : [];
  }).sort();
}

export function loadCorpus(kind = 'classic') {
  verifySources();
  const paths = kind === 'classic' ? featureFiles(join(uxRoot, 'features/classic')) : [join(uxRoot, 'features/shared-canonical-ux.feature')];
  const cases = [];
  for (const path of paths) {
    const uri = relative(uxRoot, path);
    const messages = generateMessages(readFileSync(path, 'utf8'), uri, SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN, {
      newId: IdGenerator.incrementing(), includeSource: false, includeGherkinDocument: true, includePickles: true,
    });
    const errors = messages.filter(message => message.parseError);
    if (errors.length) throw new Error(JSON.stringify(errors));
    for (const { pickle } of messages) {
      if (!pickle) continue;
      const id = pickle.tags.map(tag => tag.name).find(tag => /^@ux-/.test(tag)) || `@shared-${cases.length + 1}`;
      cases.push({ id, name: pickle.name, uri, line: pickle.location.line, steps: pickle.steps.map(step => step.text), tags: pickle.tags.map(tag => tag.name) });
    }
  }
  return cases;
}
