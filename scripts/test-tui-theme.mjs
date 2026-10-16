/** Real controlling-PTY colour negotiation, using Bun and util-linux script. */
import {mkdirSync, mkdtempSync, writeFileSync, rmSync} from 'node:fs';
import {resolve, join} from 'node:path';
import {homedir} from 'node:os';
import {requirePi130} from './pi-reference.mjs';
import {existsSync as piModulesExist} from 'node:fs';
// Pi's packages: PI_NODE_MODULES, the global bun install, or the reference install.
const PI_MODULES = [process.env.PI_NODE_MODULES, `${homedir()}/.bun/install/global/node_modules`, '/workspace/.cache/pi-ref/node_modules'].find((dir) => dir && piModulesExist(`${dir}/@earendil-works/pi-coding-agent`));
if (!PI_MODULES) throw Error('Provide Pi 1.3.0 via PI_NODE_MODULES');
requirePi130(PI_MODULES, ['@earendil-works/pi-coding-agent']);
const {generateSystemThemeColors} = await import(`${PI_MODULES}/@earendil-works/pi-coding-agent/dist/modes/interactive/theme/system-theme.js`);
const bin = process.env.GI_THEME_TEST_BIN || resolve('bin/gi-theme-test');
if (!process.env.GI_TEST_RUN_ROOT) throw Error('Use make test-env and source scripts/project-test-env.sh');
const out = join(process.env.GI_TEST_RUN_ROOT, 'tests/tui-theme');
mkdirSync(out, {recursive:true});
const plain = s => s.replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '').replace(/\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)/g, '');
const quote = s => `'${s.replaceAll("'", "'\\''")}'`;
const sleep = ms => Bun.sleep(ms);
const light = '\x1b]10;rgb:0000/0000/0000\x07\x1b]11;rgb:ffff/ffff/ffff\x1b\\';
const dark = '\x1b]10;#ffffff\x07\x1b]11;#101010\x07';
const da = '\x1b[?62;22c';
const frappe = ['#51576d','#e78284','#a6d189','#e5c890','#8caaee','#f4b8e4','#81c8be','#b5bfe2','#626880','#e78284','#a6d189','#e5c890','#8caaee','#f4b8e4','#81c8be','#a5adce'];
const palette = frappe.map((c,i)=>`\x1b]4;${i};${c}\x07`).join('')+'\x1b]10;#c6d0f5\x07\x1b]11;#303446\x07';
const hex = h => ({r:parseInt(h.slice(1,3),16),g:parseInt(h.slice(3,5),16),b:parseInt(h.slice(5,7),16)});
// Pi's system theme for a reply: the SGR colours gi must emit.
const systemColors = (fg, bg, pal) => generateSystemThemeColors({foreground:hex(fg), background:hex(bg), palette:pal?.map(hex), saturation:1}).colors;
const cases = [
 {name:'light-osc', reply:light+da, want:'system', ui:true, system:systemColors('#000000','#ffffff')},
 {name:'dark-osc', reply:dark+da, want:'system', ui:true, system:systemColors('#ffffff','#101010')},
 {name:'palette-osc', reply:palette+da, want:'system', ui:true, system:systemColors('#c6d0f5','#303446',frappe)},
 {name:'fragmented-light', reply:light+da, fragmented:true, want:'system'},
 {name:'scheme-light', reply:'\x1b[?997;2n'+da, want:'light'},
 {name:'bg-overrides-scheme', reply:dark+'\x1b[?997;2n'+da, fgbg:'0;15', setting:'light/dark', want:'dark'},
 {name:'colorfgbg-light', reply:da, fgbg:'0;15', want:'light'},
 {name:'silent-fallback', reply:'', fgbg:'15;0', want:'dark'},
 {name:'malformed-fallback', reply:'\x1b]11;rgb:gg/ff/ff\x07'+da, fgbg:'0;15', want:'light'},
 {name:'explicit-dark', reply:light+da, setting:'dark', want:'dark', ui:true},
 {name:'auto-pair-light', reply:light+da, setting:'light/dark', want:'light'},
 {name:'auto-pair-dark', reply:dark+da, setting:'light/dark', want:'dark'},
 {name:'query-disabled', reply:'', disabled:true, fgbg:'0;15', want:'light'},
];
const results=[];
for (const tc of cases) {
 const dir=mkdtempSync(resolve('.gi-theme-pty-'));
 const env={...process.env, HOME:dir, TERM:'xterm-256color', COLORTERM:'truecolor', PI_TRUE_COLOR:'1', COLORFGBG:tc.fgbg||'', GI_THEME_PTY_DIR:dir, GI_THEME_SETTING:tc.setting||'', GI_THEME_PROBE_ONLY:tc.ui?'':'1', GI_TUI_NO_THEME_QUERY:tc.disabled?'1':''};
 const child=Bun.spawn(['script','-q','-e','-f','-c', `stty rows 40 cols 100; ${quote(bin)} -test.run '^TestTerminalThemePTYFixture$' -test.timeout 20s`, '/dev/null'], {env, stdin:'pipe', stdout:'pipe', stderr:'pipe'});
 let raw='', answered=false, readError=null;
 const reading=(async()=>{for await(const bytes of child.stdout){ raw+=new TextDecoder().decode(bytes); if(!answered && raw.includes('\x1b[c')){answered=true; if(tc.reply){if(tc.fragmented){for(const c of tc.reply){child.stdin.write(c); child.stdin.flush(); await sleep(1);}}else{child.stdin.write(tc.reply);child.stdin.flush();}}}}})().catch(e=>{readError=e;});
 const wait=async(fn,label)=>{const end=Date.now()+12000;while(Date.now()<end){if(fn())return;await sleep(20);}throw Error(`${tc.name}: timeout ${label}`);};
 try {
  await wait(()=>raw.includes('THEME-RESULT:'),'startup');
  const m=raw.match(/THEME-RESULT:(\w+) elapsed=(\d+)/);
  if(!m || m[1]!==tc.want)throw Error(`${tc.name}: wrong theme ${m?.[1]}`);
  if(Number(m[2])>1000)throw Error(`${tc.name}: unbounded startup ${m[2]}ms`);
  if(tc.disabled && answered)throw Error('query-disable ignored');
  if(!tc.disabled && !answered)throw Error('query was not sent');
  if(tc.ui){
   await wait(()=>raw.includes('%/'),'editor ready');
   child.stdin.write('/config\r');child.stdin.flush();
   await wait(()=>plain(raw).includes('settings: peering'),'settings output');
   child.stdin.write('\x1b[H');child.stdin.flush();
   await wait(()=>plain(raw).includes(`theme: ${tc.want}`),'settings theme');
   if(tc.system){
    // Generated colours reach the terminal; the built-in dark text does not.
    const emitted=Object.entries(tc.system).filter(([,v])=>v&&raw.includes('8;2;'+[hex(v).r,hex(v).g,hex(v).b].join(';')));
    if(emitted.length<3)throw Error(`${tc.name}: generated colours not emitted (${emitted.map(e=>e[0])})`);
    if(raw.includes('38;2;222;224;225'))throw Error(`${tc.name}: built-in dark text emitted`);
   } else {
    const rgb=tc.want==='light'?'59;63;65':'222;224;225';
    if(!raw.includes('38;2;'+rgb))throw Error(`${tc.name}: text RGB not emitted`);
   }
   child.stdin.write('\x03');child.stdin.flush();await sleep(50);child.stdin.write('\x03');child.stdin.flush();
  }
  const code=await Promise.race([child.exited,sleep(3000).then(()=>{throw Error(`${tc.name}: exit timeout`);})]);
  await reading;
  if(code!==0||readError)throw Error(`${tc.name}: exit ${code}, ${readError}`);
  writeFileSync(join(out,tc.name+'.ansi'),raw);
  results.push({name:tc.name,theme:tc.want,elapsedMs:Number(m[2]),ui:!!tc.ui});
 } catch(e) {writeFileSync(join(out,tc.name+'-failed.ansi'),raw);throw e;}
 finally {child.kill();child.stdin.end();await child.exited;rmSync(dir,{recursive:true,force:true});}
}
writeFileSync(join(out,'results.json'),JSON.stringify(results,null,2));
console.log(`PASS: ${results.length} real PTY theme scenarios (including light/dark native UI)`);
