import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
import card from './fixtures/card-rejection.json' with {type:'json'};
const errorText='Card submissions are not supported by Gi yet. Your inputs have not been submitted.';
test.skip(!process.env.GI_UX_CARD_REJECTION,'requires isolated native card fixture');
test('@ux-extra-003 Unsupported native card Submit rejects into its UI notice without success or data loss',async({page,request},info)=>{
 await info.attach('gherkin',{body:loadCorpus().find(r=>r.id==='@ux-extra-003').steps.join('\n'),contentType:'text/plain'});
 const id='card-rejection-main',other='card-rejection-other';
 const messages=async sid=>(await(await request.get(`/api/sessions/${sid}/messages`)).json()).messages;
 const before=await messages(id),otherBefore=await messages(other);expect(before[0].payload.content_blocks).toEqual([card]);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);const writes=[],pageErrors=[];
 page.on('pageerror',e=>pageErrors.push(e.message));page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/')&&!['GET','HEAD'].includes(r.method()))writes.push([r.method(),new URL(r.url()).pathname]);});
 await page.goto('/');const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('card retained composer β');await input.evaluate(el=>el.setSelectionRange(5,5));
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'card-kept.txt',mimeType:'text/plain',buffer:Buffer.from('retained card bytes')});
 const draft=()=>page.evaluate(async id=>{const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts');r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:[...new Uint8Array(f.bytes)]}))}}:null);};tx.onerror=()=>reject(tx.error);});},id);
 await expect.poll(draft).toMatchObject({draft:{text:'card retained composer β',media:[{name:'card-kept.txt',bytes:[...Buffer.from('retained card bytes')]}]},pending:[]});const saved=await draft();
 const post=page.locator(`#post-${id}-post`),container=post.locator('.adaptive-card-container'),answer=container.getByRole('textbox',{name:'Card answer',exact:true}),choice=container.getByRole('checkbox',{name:'Keep this choice',exact:true}),submit=container.getByRole('button',{name:'Submit answer',exact:true});
 await expect(submit).toBeVisible();await answer.fill('native card answer β');await choice.check();
 // Capture real native SDK DOM events/state. No submitted result, API or callback is replaced.
 await container.evaluate(el=>{window.__cardProof={events:[],busy:false,notices:0};for(const type of ['click','keydown'])el.addEventListener(type,e=>{if(e.target?.closest('button,[role="button"]'))window.__cardProof.events.push({type,key:e.key||'',trusted:e.isTrusted});},true);new MutationObserver(records=>{if(records.some(r=>r.type==='attributes'&&r.attributeName==='class'&&r.oldValue?.includes('adaptive-card-busy')))window.__cardProof.busy=true;for(const r of records)for(const n of r.addedNodes)if(n instanceof Element&&n.classList.contains('adaptive-card-notice'))window.__cardProof.notices++;}).observe(el,{attributes:true,attributeOldValue:true,childList:true});});
 const rejected=async()=>{await expect(container.locator('.adaptive-card-notice-error')).toHaveText(errorText);await expect(container).not.toHaveClass(/adaptive-card-busy|adaptive-card-finished|adaptive-card-readonly/);await expect(container.locator('.adaptive-card-status,.adaptive-card-submission-receipt')).toHaveCount(0);await expect(answer).toHaveValue('native card answer β');await expect(choice).toBeChecked();await expect(submit).toBeEnabled();await expect(input).toHaveValue('card retained composer β');expect(await draft()).toEqual(saved);};
 await submit.click();await rejected();
 await submit.focus();await page.keyboard.press('Enter');await rejected();await expect.poll(()=>container.evaluate(()=>window.__cardProof.notices)).toBe(2);
 const proof=await container.evaluate(()=>window.__cardProof);expect(proof.busy).toBe(true);expect(proof.events.some(e=>e.type==='click'&&e.trusted)).toBe(true);expect(proof.events.some(e=>e.type==='keydown'&&e.key==='Enter'&&e.trusted)).toBe(true);
 await info.attach('rejected-card',{body:await page.screenshot(),contentType:'image/png'});
 expect(await messages(id)).toEqual(before);expect(await messages(other)).toEqual(otherBefore);expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);
 await page.reload();await expect(input).toHaveValue('card retained composer β');await expect(submit).toBeVisible();await expect(container.locator('.adaptive-card-notice')).toHaveCount(0);expect(await draft()).toEqual(saved);
 // Card edits are mount-local and must not be presented as persisted submissions.
 await expect(answer).toHaveValue('');expect((await messages(id))[0].payload.content_blocks[0].state).toBe('active');
 expect(writes).toEqual([]);expect(pageErrors).toEqual([]);expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(id);
});
