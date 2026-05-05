import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { resolve, join } from 'node:path';
import { chromium } from 'playwright';

const root = resolve(process.argv[2] || '.');
const outDir = resolve(process.argv[3] || join(root, 'artifacts', 'tui-audit-rendered'));
mkdirSync(outDir, { recursive: true });

const captures = [
  { key: 'gi-wide', title: 'Gi TUI — wide layout (100x22)', path: join(root, 'gi-wide.txt') },
  { key: 'gi-narrow', title: 'Gi TUI — narrow layout (60x18)', path: join(root, 'gi-narrow.txt') },
  { key: 'gi-markdown', title: 'Gi TUI — markdown transcript rendering', path: join(root, 'gi-markdown.txt') },
  { key: 'pi-wide', title: 'Pi — wide layout reference (100x22)', path: join(root, 'pi-wide.txt') },
];

function esc(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

const sections = captures.map((cap) => {
  const text = existsSync(cap.path) ? readFileSync(cap.path, 'utf8') : '[missing capture]';
  return { ...cap, text };
});

const html = `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<title>Gi TUI feature report</title>
<style>
  :root {
    color-scheme: dark;
    --bg: #0b1020;
    --panel: #11182d;
    --panel2: #0e1527;
    --text: #dbe7ff;
    --muted: #94a3b8;
    --accent: #7dd3fc;
    --border: #23314f;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    padding: 32px;
    background: var(--bg);
    color: var(--text);
    font: 16px/1.45 Inter, system-ui, sans-serif;
  }
  h1,h2,h3 { margin: 0 0 12px; }
  p, li { color: var(--text); }
  .muted { color: var(--muted); }
  .section {
    margin: 0 0 32px;
    padding: 24px;
    border: 1px solid var(--border);
    border-radius: 14px;
    background: linear-gradient(180deg, var(--panel), var(--panel2));
    break-inside: avoid;
  }
  .terminal {
    margin-top: 16px;
    padding: 16px 18px;
    border-radius: 12px;
    background: #050816;
    border: 1px solid #1f2a44;
    box-shadow: 0 20px 40px rgba(0,0,0,0.35);
  }
  pre {
    margin: 0;
    color: #e2e8f0;
    white-space: pre-wrap;
    word-break: break-word;
    font: 13px/1.4 "SFMono-Regular", Consolas, Monaco, monospace;
  }
  .compare-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
  }
  .shot {
    margin-top: 12px;
    width: 100%;
    border-radius: 12px;
    border: 1px solid var(--border);
  }
  ul { margin-top: 8px; }
  @media print {
    body { padding: 18px; }
    .section { page-break-inside: avoid; }
  }
</style>
</head>
<body>
  <div class="section">
    <h1>Gi TUI feature report</h1>
    <p class="muted">Generated on ${new Date().toISOString()}</p>
    <p>This report summarizes the implemented TUI features, compares the current Gi terminal layout against Pi as a reference, and embeds tmux-captured screenshots rendered into a printable PDF.</p>
    <h2>Implemented feature summary</h2>
    <ul>
      <li>Streaming model output</li>
      <li>Default model/provider readiness with persisted selection</li>
      <li>Expanding multiline input with simplified horizontal-rule chrome</li>
      <li>Settable transcript scrollback limit</li>
      <li>Responsive narrow-terminal layout</li>
      <li>Markdown transcript rendering with responsive table fallback</li>
      <li>Tool, skill, settings, approval, session, tree, and plugin visibility commands</li>
      <li>tmux-backed Gherkin regression coverage and CI-friendly artifacts</li>
    </ul>
    <h2>Remaining gaps</h2>
    <ul>
      <li>Bracketed/ANSI text paste parity appears blocked by deeper terminal parser support in go-tui.</li>
      <li>Pi/Claude Code parity remains an informed approximation rather than a pixel-identical clone.</li>
    </ul>
  </div>

  <div class="section">
    <h2>Layout parity audit</h2>
    <p>The current Gi layout now keeps the same broad information hierarchy as Pi: a prominent status line, a session/context metadata block, the main transcript, and an input/footer region. Pi still differs in startup chrome and some placement/details, but the main pieces are now in comparable positions for terminal use.</p>
    <div class="compare-grid">
      <div>
        <h3>${esc(sections.find(s => s.key === 'gi-wide').title)}</h3>
        <div class="terminal"><pre>${esc(sections.find(s => s.key === 'gi-wide').text)}</pre></div>
      </div>
      <div>
        <h3>${esc(sections.find(s => s.key === 'pi-wide').title)}</h3>
        <div class="terminal"><pre>${esc(sections.find(s => s.key === 'pi-wide').text)}</pre></div>
      </div>
    </div>
  </div>

  ${sections.filter(s => s.key !== 'pi-wide').map((s) => `
  <div class="section capture" id="capture-${s.key}">
    <h2>${esc(s.title)}</h2>
    <div class="terminal"><pre>${esc(s.text)}</pre></div>
  </div>
  `).join('')}

  <div class="section">
    <h2>Evidence checklist</h2>
    <ul>
      <li><strong>Layout evidence:</strong> wide and narrow tmux captures for Gi, plus a Pi reference capture.</li>
      <li><strong>Markdown evidence:</strong> a seeded assistant transcript with headings, lists, blockquote, and responsive table rendering.</li>
      <li><strong>Automated tests:</strong> unit coverage for markdown/layout/input behavior and tmux Gherkin regression coverage.</li>
      <li><strong>Source docs:</strong> <code>docs/internal/tui-ux-user-stories.md</code>, <code>docs/internal/tui-ux-report.md</code>, <code>docs/internal/tui-paste-analysis.md</code>, and <code>docs/internal/topic-system.md</code>.</li>
    </ul>
  </div>
</body>
</html>`;

const htmlPath = join(outDir, 'tui-feature-report.html');
writeFileSync(htmlPath, html);

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 1600 }, colorScheme: 'dark' });
await page.goto('file://' + htmlPath, { waitUntil: 'load' });

for (const s of sections.filter(s => s.key !== 'pi-wide')) {
  const locator = page.locator('#capture-' + s.key);
  await locator.screenshot({ path: join(outDir, `${s.key}.png`) });
}
await page.pdf({ path: join(outDir, 'tui-feature-report.pdf'), format: 'A4', printBackground: true, margin: { top: '12mm', right: '12mm', bottom: '12mm', left: '12mm' } });
await browser.close();
console.log(outDir);
