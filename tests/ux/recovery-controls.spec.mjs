import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
import cases from './fixtures/recovery-controls.json' with {type:'json'};
test.skip(!process.env.GI_UX_RECOVERY_CONTROLS,'requires isolated native stored recovery controls');
test('@ux-extra-012 Validated recovery controls stay hidden while malformed native lookalikes remain visible',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(r=>r.id==='@ux-extra-012').steps.join('\n'),contentType:'text/plain'});
 const id='recovery-controls-fixture',messages=async()=> (await(await request.get(`/api/sessions/${id}/messages`)).json()).messages;
 const initial=await messages();expect(initial).toHaveLength(cases.length);
 for(const c of cases){const m=initial.find(m=>m.id===`recovery-control-${c.id}`);expect(m.content).toContain(`Recovery probe ${c.id}`);expect(m.payload.content_blocks).toEqual(c.fields===null?[]:[{type:'control_intent',intent:'protected_recovery_continuation',schema_version:1,source_message_id:'source',source_row_id:1,thread_id:1,...c.fields}]);}
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);const writes=[],errors=[];
 page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/')&&!['GET','HEAD'].includes(r.method()))writes.push([r.method(),new URL(r.url()).pathname]);});
 await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('recovery keep draft β');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'recovery-kept.txt',mimeType:'text/plain',buffer:Buffer.from('retained recovery bytes')});
 const draft=()=>page.evaluate(async id=>{const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts');r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:[...new Uint8Array(f.bytes)]}))}}:null);};tx.onerror=()=>reject(tx.error);});},id);
 await expect.poll(draft).toMatchObject({draft:{text:'recovery keep draft β',media:[{name:'recovery-kept.txt',bytes:[...Buffer.from('retained recovery bytes')]}]},pending:[]});const saved=await draft();
 const check=async()=>{for(const c of cases){const post=page.locator(`#post-recovery-control-${c.id}`);if(c.hidden)await expect(post).toHaveCount(0);else {await expect(post).toHaveCount(1);await expect(post).toContainText(`Recovery probe ${c.id}`);}}expect(await draft()).toEqual(saved);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);};
 await check();await page.reload();await expect(input).toHaveValue('recovery keep draft β');await check();
 await page.getByRole('button',{name:'Search',exact:true}).click();const search=page.getByRole('textbox',{name:'Search (Enter to run)...',exact:true});await search.fill('Recovery probe');const found=page.waitForResponse(r=>r.url().includes('/search?')&&r.status()===200);await search.press('Enter');const searchResult=await(await found).json();expect(searchResult.messages).toHaveLength(cases.length);await expect(page.getByRole('button',{name:'Close search',exact:true})).toBeVisible();await check();
 await page.getByRole('button',{name:'Close search',exact:true}).click();await check();await expect(input).toHaveValue('recovery keep draft β');
 expect(await messages()).toEqual(initial);expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);expect(writes).toEqual([]);expect(errors).toEqual([]);
 await info.attach('recovery-controls',{body:await page.screenshot(),contentType:'image/png'});
});
