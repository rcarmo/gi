import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
import cases from './fixtures/recovery-placeholders.json' with {type:'json'};
test.skip(!process.env.GI_UX_RECOVERY_PLACEHOLDERS,'requires isolated native placeholder fixtures');
test('@ux-extra-013 Empty info recovery is omitted without hiding text, attachments, cards, submissions or warnings',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(r=>r.id==='@ux-extra-013').steps.join('\n'),contentType:'text/plain'});
 const id='recovery-placeholders-fixture',messages=async()=> (await(await request.get(`/api/sessions/${id}/messages`)).json()).messages;
 const initial=await messages();expect(initial).toHaveLength(cases.length);
 for(const c of cases){const m=initial.find(m=>m.id===`recovery-placeholder-${c.id}`);expect(m.role).toBe(c.role);expect(m.content).toBe(c.content);expect(m.payload.content_blocks).toEqual(c.blocks);}
 const media=initial.find(m=>m.id==='recovery-placeholder-file').payload.media[0];expect(await(await request.get(`/api/media/${media.media_id}/raw`)).text()).toBe('recovery attachment bytes');
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);const writes=[],errors=[];
 page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/')&&!['GET','HEAD'].includes(r.method()))writes.push([r.method(),new URL(r.url()).pathname]);});
 await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('placeholder keep draft β');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'placeholder-kept.txt',mimeType:'text/plain',buffer:Buffer.from('retained placeholder bytes')});
 const draft=()=>page.evaluate(async id=>{const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:[...new Uint8Array(f.bytes)]}))}}:null);};tx.onerror=()=>reject(tx.error);});},id);
 await expect.poll(draft).toMatchObject({draft:{text:'placeholder keep draft β',media:[{name:'placeholder-kept.txt',bytes:[...Buffer.from('retained placeholder bytes')]}]},pending:[]});const saved=await draft();
 const check=async()=>{
  for(const c of cases)await expect(page.locator(`#post-recovery-placeholder-${c.id}`)).toHaveCount(c.hidden?0:1);
  await expect(page.locator('#post-recovery-placeholder-authored-text')).toContainText('Recovery authored text stays visible');
  await expect(page.locator('#post-recovery-placeholder-file .file-attachment')).toContainText('recovery-retained.txt');
  for(const [key,text]of [['card','Recovery retained card'],['card-fallback','Recovery fallback retained card']])await expect(page.locator(`#post-recovery-placeholder-${key} .post-adaptive-cards`)).toContainText(text);
  await expect(page.locator('#post-recovery-placeholder-submission')).toContainText('Keep recovery answer');
  await expect(page.locator('#post-recovery-placeholder-file-ref')).toContainText('keep.md');
  await expect(page.locator('#post-recovery-placeholder-message-ref .post-msg-pill-link')).toHaveCount(1);
  await expect(page.locator('#post-recovery-placeholder-attachment-ref')).toContainText('retained-name.txt');
  const resource=page.locator('#post-recovery-placeholder-resource');await expect(resource).toContainText('text://recovery');if(!(await resource.locator('.resource-embed-content').count()))await resource.locator('.resource-embed-toggle').click();await expect(resource).toContainText('Resource bytes stay visible');
  await expect(page.locator('#post-recovery-placeholder-text-annotation')).toContainText('Priority: 0.5');
  expect(await draft()).toEqual(saved);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
 };
 await check();await page.reload();await expect(input).toHaveValue('placeholder keep draft β');await check();
 // Authored recovery text remains searchable; the empty stored row stays intact.
 await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await search.fill('Recovery authored text');const found=page.waitForResponse(r=>r.url().includes('/search?')&&r.status()===200);await search.press('Enter');expect((await(await found).json()).messages.map(m=>m.id)).toContain('recovery-placeholder-authored-text');await expect(page.locator('#post-recovery-placeholder-authored-text')).toHaveCount(1);
 await page.getByRole('button',{name:'Close search',exact:true}).click();await check();
 await expect(input).toHaveValue('placeholder keep draft β');expect(await messages()).toEqual(initial);expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);expect(writes).toEqual([]);expect(errors).toEqual([]);
 await info.attach('recovery-placeholder-survivors',{body:await page.screenshot(),contentType:'image/png'});
});
