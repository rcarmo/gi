import {chromium} from 'playwright';
import {readFile,writeFile,mkdir,readdir} from 'node:fs/promises';
import {resolve,dirname,relative} from 'node:path';
import {fileURLToPath} from 'node:url';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {installPixelHost} from '../tests/ux/support/pixel-adapter.mjs';
import {compareCaptureSet} from '../tests/ux/support/pixel-comparison.mjs';

const repo=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const load=async path=>JSON.parse(await readFile(resolve(repo,path),'utf8'));
const out=resolve(repo,process.env.PIXEL_OUTPUT||"test-results/compose-pixels");
await mkdir(out,{recursive:true});
// Keep previous captures intact, but give every invocation its own directory.
const runDir=resolve(out,`run-${Date.now()}-${process.pid}`);
await mkdir(runDir);
try{
const state=await load('tests/ux/fixtures/compose-pixel-state.json');
const reference=await load('tests/ux/fixtures/compose-pixel-reference.json');
const refRoot=process.env.PICLAW_PIXEL_ROOT;
if(!refRoot)throw Error('PICLAW_PIXEL_ROOT must point to the pinned Piclaw runtime directory');
const candidateRoot=resolve(repo,'internal/web/static');
const hash=bytes=>createHash('sha256').update(bytes).digest('hex');
const select=(env,allowed)=>{const result=(process.env[env]||allowed.join(',')).split(',');if(result.some(v=>!allowed.includes(v))||new Set(result).size!==result.length)throw Error(`Invalid ${env}`);return result;};
const viewports=select('PIXEL_VIEWPORTS',Object.keys(state.viewports));
const themes=select('PIXEL_THEMES',state.themes);
const scenarios=select('PIXEL_SCENARIOS',state.scenarios);

// Invalidate any old result before preflight, including missing pinned assets.
await writeFile(resolve(runDir,'comparison.json'),JSON.stringify({passed:false,captureComplete:false,reason:'Preflight in progress'})+'\n');
// Pin every reference input up front. Capture requested candidate hashes once
// and reject changes during the run, including changes between fresh contexts.
for(const [url,file]of Object.entries(reference.files)){
 const bytes=await readFile(resolve(refRoot,file.relativePath));
 if(hash(bytes)!==file.sha256||bytes.length!==file.bytes)throw Error(`Reference asset mismatch: ${url}`);
}
async function candidateManifest(dir){
 const files={};
 for(const entry of await readdir(dir,{withFileTypes:true})){
  const path=resolve(dir,entry.name);
  if(entry.isDirectory())Object.assign(files,await candidateManifest(path));
  else if(entry.isFile()){const bytes=await readFile(path);files[relative(candidateRoot,path)]={sha256:hash(bytes),bytes:bytes.length};}
 }
 return files;
}
const candidate=await candidateManifest(candidateRoot);
const records=[],failures=[];
const launchArgs=['--disable-gpu','--disable-gpu-compositing','--disable-gpu-rasterization','--disable-oop-rasterization','--force-color-profile=srgb'];
const metadataBrowser=await chromium.launch({headless:false,args:launchArgs});
const browserVersion=metadataBrowser.version();await metadataBrowser.close();
const environment={browser:browserVersion,freshBrowserPerCapture:true,warmupScreenshots:1,browserExecutable:chromium.executablePath(),browserSHA256:hash(await readFile(chromium.executablePath())),platform:process.platform,architecture:process.arch,launchArgs,headed:true,locale:state.locale,timezoneId:state.timezoneId,deviceScaleFactor:state.deviceScaleFactor,clock:state.now,repeats:2,reducedMotion:'reduce',animations:'disabled at screenshot after settle',caret:'hide',masks:[],fontconfig:execFileSync('fc-list',['--format','%{file}: %{family}: %{style}\n'],{encoding:'utf8'}).split('\n').sort()};
const captures={runDir,state,reference,candidate,environment,viewports,themes,scenarios,records,captureFinished:false,candidateUnchanged:false};
const save=()=>writeFile(resolve(runDir,'captures.json'),JSON.stringify(captures,null,2)+'\n');
await save();
await writeFile(resolve(runDir,'comparison.json'),JSON.stringify({passed:false,captureComplete:false,reason:'Capture in progress'})+'\n');
for(const viewportName of viewports)for(const theme of themes)for(const scenario of scenarios)for(const host of ['piclaw','gi'])for(let repeat=1;repeat<=2;repeat++){
  const viewport=state.viewports[viewportName],key=`${viewportName}-${theme}-${scenario}`;
  // Isolate renderer/Skia caches as well as storage between captures.
  const record={key,viewport:viewportName,theme,scenario,host,repeat,status:'failed'};records.push(record);
  let adapter,browser,context,page;
  try{
   browser=await chromium.launch({headless:false,args:launchArgs});
   context=await browser.newContext({viewport,locale:state.locale,timezoneId:state.timezoneId,deviceScaleFactor:state.deviceScaleFactor,colorScheme:theme,reducedMotion:"reduce",serviceWorkers:"block"});
   page=await context.newPage();page.setDefaultTimeout(10000);
   await page.clock.setFixedTime(new Date(state.now));
   adapter=await installPixelHost({page,host,root:host==='piclaw'?refRoot:candidateRoot,state:{...state,theme},reference});
   await page.goto(adapter.origin+`/?chat_jid=${encodeURIComponent(state.sessionId)}`);
   const input=page.locator('.compose-box textarea');await input.waitFor();
   const dismiss=page.getByRole('button',{name:'Dismiss',exact:true});if(await dismiss.isVisible())await dismiss.click();
   await adapter.connected();
   await page.locator('.compose-model-hint-btn').waitFor();
   await input.fill(state.compose);
   if(scenario==='models'){await page.locator('.compose-model-hint-btn').click();await page.locator('.compose-model-popup').waitFor();}
   if(scenario==='sessions'){await page.locator('button.compose-session-trigger').last().click();await page.locator('.compose-session-popup').waitFor();}
   await page.evaluate(async()=>{await document.fonts.ready;await Promise.all([...document.images].map(i=>i.decode()));document.activeElement?.blur?.();});
   await page.mouse.move(0,0);await page.waitForTimeout(500);
   await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
   record.dom=await page.evaluate(()=>{
    const inspect=selector=>{const e=document.querySelector(selector);if(!e)return null;const r=e.getBoundingClientRect(),c=getComputedStyle(e);return {x:r.x,y:r.y,width:r.width,height:r.height,font:c.font,lineHeight:c.lineHeight,padding:c.padding,color:c.color,background:c.backgroundColor,text:e.textContent};};
    return {compose:inspect('.compose-box'),textarea:inspect('.compose-box textarea'),panel:inspect('.compose-model-popup'),theme:document.documentElement.dataset.theme,fonts:[...document.fonts].map(f=>({family:f.family,status:f.status,weight:f.weight})),viewport:{width:innerWidth,height:innerHeight,dpr:devicePixelRatio}};
   });
   if(record.dom.theme!==theme||record.dom.viewport.width!==viewport.width||record.dom.viewport.height!==viewport.height||record.dom.viewport.dpr!==state.deviceScaleFactor)throw Error('Theme/viewport mismatch');
   if(await input.inputValue()!==state.compose)throw Error('Draft mismatch');
   if(scenario!=='compose'&&!record.dom.panel)throw Error('Missing panel');
   record.file=`${key}-${host}-${repeat}.png`;
   // Flush the same screenshot animation/caret handling for both hosts before
   // recording the final frame; never retry or select a matching screenshot.
   await page.screenshot({path:resolve(runDir,`${key}-${host}-${repeat}-warmup.png`),animations:'disabled',caret:'hide'});
   await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
   await page.screenshot({path:resolve(runDir,record.file),animations:'disabled',caret:'hide'});
   record.sha256=hash(await readFile(resolve(runDir,record.file)));
   adapter.assert();
   for(const [url,asset]of Object.entries(adapter.assets))if(host==='gi'&&candidate[relative(candidateRoot,asset.path)]?.sha256!==asset.sha256)throw Error(`Candidate asset changed: ${url}`);
   record.paint=await page.evaluate(()=>{const e=document.querySelector('.compose-input-wrapper'),c=getComputedStyle(e);return {active:document.activeElement?.outerHTML?.slice(0,180),border:c.border,background:c.background,transform:c.transform,zoom:c.zoom,animations:document.getAnimations().map(a=>({time:a.currentTime,state:a.playState,timing:a.effect?.getComputedTiming()}))};});
   record.status='captured';
  }catch(e){record.error=e.message;failures.push(`${key}/${host}/${repeat}: ${e.message}`);if(page)await page.screenshot({path:resolve(runDir,`${key}-${host}-${repeat}-failure.png`),animations:'disabled'}).catch(()=>{});console.error(failures.at(-1));}
  finally{
   if(adapter){record.assets=adapter.assets;record.calls=adapter.calls;record.failures=adapter.failures;record.streamAborts=adapter.streamAborts;await adapter.dispose().catch(e=>{record.status="failed";record.failures.push(e.message);});}
   try{await context?.close();}finally{await browser?.close();await save();}
  }
 }
captures.candidateUnchanged=JSON.stringify(candidate)===JSON.stringify(await candidateManifest(candidateRoot));
captures.captureFinished=true;await save();
const result=await compareCaptureSet(runDir,captures);
console.log(JSON.stringify(result,null,2));
process.exitCode=!result.captureComplete?2:result.passed?0:1;

}catch(error){
 const result={passed:false,captureComplete:false,reason:error.message,runDir};
 await writeFile(resolve(runDir,"comparison.json"),JSON.stringify(result,null,2)+"\n");
 console.error(error.message);process.exitCode=2;
}finally{console.log(`Pixel evidence: ${runDir}`);}
