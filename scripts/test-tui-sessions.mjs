#!/usr/bin/env bun
/** Real tmux acceptance: draft switching and bounded temporary selector. */
import { execFileSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';

const root = resolve(import.meta.dir, '..');
const artifacts = join(root, 'test-results/tui-sessions');
const temp = mkdtempSync(join(tmpdir(), 'gi-tui-sessions-'));
const db = join(temp, 'gi.db');
const session = `gi-session-parity-${process.pid}`;
const target = `${session}:0`;
const tmux = (...args) => execFileSync('tmux', args, { encoding: 'utf8' });
const sql = query => execFileSync('sqlite3', [db, query], { encoding: 'utf8' }).trim();
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
const capture = () => tmux('capture-pane', '-p', '-t', target);
const keys = (...keys) => tmux('send-keys', '-t', target, ...keys);
const type = text => tmux('send-keys', '-t', target, '-l', text);
async function waitFor(predicate, label) {
  const end = Date.now()+10000;
  while (Date.now()<end) {
    try { if (predicate()) return; } catch {}
    await sleep(80);
  }
  throw new Error(`Timed out: ${label}\n${capture()}`);
}
async function command(text) { type(text); keys('Enter'); await sleep(150); }
async function snapshot(name) {
  await sleep(180);
  const screen = capture();
  writeFileSync(join(artifacts, `${name}.txt`), screen);
  return screen;
}
function assert(condition, detail) { if (!condition) throw new Error(detail); }
function selectorRows(screen) { return screen.split('\n').filter(line => /^\s*[›*]?\s*\d+\. /.test(line)); }
function separators(screen) { return screen.split('\n').map((line,index)=>({line,index})).filter(({line})=>/^\s*[─━-]{10,}\s*$/.test(line)).map(({index})=>index); }
function withoutCursor(screen) { return screen.replaceAll('▌',' ').split('\n').map(line=>line.trimEnd()).join('\n'); }

try {
  mkdirSync(artifacts,{recursive:true});
  mkdirSync(join(temp,'.pi'));
  writeFileSync(join(temp,'.pi/settings.json'),JSON.stringify({defaultProvider:'test',defaultModel:'test-model',defaultThinkingLevel:'low',enabledModels:['test-model']}));
  const quote = text => `'${text.replaceAll("'", "'\\''")}'`;
  tmux('new-session','-d','-x','100','-y','22','-s',session,`cd ${quote(root)} && ${quote(join(root,'bin/gi'))} -tui -db ${quote(db)} -workspace ${quote(temp)}`);
  await waitFor(()=>capture().includes('m0/t0'),'startup');
  await command('/fork @other');
  await waitFor(()=>sql('select count(*) from sessions;')==='2','native child fork');
  await command('/switch @agent');
  // Populate genuine sessions through native /fork commands, not rendered rows.
  for(let i=0;i<7;i++) await command(`/fork @extra${i}`);
  await command('/switch @agent');
  const summaries=[];
  for(const [width,height] of [[60,18],[100,22],[140,36]]) {
    const label=`${width}x${height}`;
    tmux('resize-window','-t',target,'-x',String(width),'-y',String(height));
    await sleep(250);
    keys('Tab'); type(`A draft ${label}`);
    await waitFor(()=>capture().includes(`A draft ${label}`),'A draft');
    const before=await snapshot(`${label}-idle-draft`);
    keys('M-s');
    await waitFor(()=>capture().includes('Select session'),'open selector');
    const open=await snapshot(`${label}-picker`);
    assert(selectorRows(open).length>0 && selectorRows(open).length<=6,`${label}: unbounded results`);
    assert(open.includes(`A draft ${label}`),`${label}: editor hidden by selector`);
    assert(separators(open).length===2,`${label}: selector added box/separator chrome`);
    type('no-matching-session-xyz');
    await waitFor(()=>capture().includes('no matching sessions'),'empty-state noun');
    keys('Escape');
    await waitFor(()=>!capture().includes('Select session'),'close selector');
    const after=await snapshot(`${label}-cancel`);
    assert(withoutCursor(before)===withoutCursor(after),`${label}: cancel changed layout/transcript/draft`);
    keys('M-s'); await waitFor(()=>capture().includes('Select session'),'reopen');
    type('@other');
    await waitFor(()=>selectorRows(capture()).length===1,'filter other');
    keys('Enter');
    await waitFor(()=>!capture().includes('Select session') && !capture().includes(`A draft ${label}`),'switch to B');
    type(`B draft ${label}`);
    keys('M-s'); await waitFor(()=>capture().includes('Select session'),'B picker');
    type('@agent'); await waitFor(()=>selectorRows(capture()).length===1,'filter main'); keys('Enter');
    await waitFor(()=>capture().includes(`A draft ${label}`),'restore A');
    await snapshot(`${label}-restored-A`);
    keys('M-s'); await waitFor(()=>capture().includes('Select session'),'A picker');
    type('@other'); await waitFor(()=>selectorRows(capture()).length===1,'filter B'); keys('Enter');
    await waitFor(()=>capture().includes(`B draft ${label}`),'restore B');
    await snapshot(`${label}-restored-B`);
    keys('C-u'); // Clear B's unsent line, then return to A and clear it.
    keys('M-s'); await waitFor(()=>capture().includes('Select session'),'return main picker');
    type('@agent'); await waitFor(()=>selectorRows(capture()).length===1,'filter return'); keys('Enter');
    await waitFor(()=>capture().includes(`A draft ${label}`),'main restored'); keys('C-u');
    assert(sql('select count(*) from turns;')==='0',`${label}: picker submitted a prompt`);
    summaries.push(`${label}: <=6 results, exact cancel snapshot, A/B drafts, no submitted turns`);
  }
  // Resize while the popup is open: keep the selected tail row visible.
  keys('M-s'); await waitFor(()=>capture().includes('Select session'),'resize picker'); keys('End');
  for(const [width,height] of [[60,18],[100,22],[140,36]]) {
    tmux('resize-window','-t',target,'-x',String(width),'-y',String(height));
    const screen=await snapshot(`${width}x${height}-resize-open`);
    assert(selectorRows(screen).length<=6 && screen.includes('›'), 'resize lost visible selection');
  }
  keys('Escape');
  writeFileSync(join(artifacts,'summary.txt'),summaries.join('\n')+'\nLive resize: selected row remains visible at all sizes.\n');
  console.log(summaries.join('\n'));
} catch(error) {
  try { writeFileSync(join(artifacts,'failure.txt'),`${error.stack}\n${capture()}`); } catch {}
  throw error;
} finally {
  try { tmux('kill-session','-t',session); } catch {}
  rmSync(temp,{recursive:true,force:true});
}
