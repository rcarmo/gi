import {test,expect} from '@playwright/test';
import {loadCorpus} from './support/catalogue.mjs';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function evidence(info,id){const s=loadCorpus().find(s=>s.id===id);await info.attach('gherkin',{body:s.steps.join('\n'),contentType:'text/plain'});}
async function fixture(page,request,info,markdown){
 const token=`render-${info.project.name}-${Date.now()}`;
 const session=await(await request.post('/api/sessions',{data:{agent_id:token,title:token}})).json();
 const submitted=await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:'Render fixture:\n\n'+markdown,model:'test-model'}});expect(submitted.status()).toBe(202);
 const {turn_id}=await submitted.json();
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns.find(t=>t.id===turn_id)?.status)).toBe('completed');
 const messages=(await(await request.get(`/api/sessions/${session.id}/messages`)).json()).messages;
 const stored=messages.find(m=>m.role==='assistant');expect(stored.content).toContain(markdown);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);await page.goto('/');
 const post=page.locator(`#post-${stored.id}`);await expect(post).toBeVisible();
 const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('rendering draft retained');
 return {post,input,session,stored};
}
// Observe what the browser's real copy event receives, without replacing the
// clipboard API, execCommand, copy handlers or their success/failure results.
async function observeClipboard(page){
 await page.addInitScript(()=>{
  window.__nativeCopies=[];
  document.addEventListener('copy',event=>{
   const data=event.clipboardData;
   window.__nativeCopies.push({trusted:event.isTrusted,text:data?.getData('text/plain'),html:data?.getData('text/html')});
  });
 });
}

test('@ux-timeline-023 Markdown table spans the content area with automatic table layout',async({page,request},info)=>{
 await evidence(info,'@ux-timeline-023');
 const markdown='| Column | Description |\n| --- | --- |\n| Alpha | Native stored Markdown |\n| Beta | 第二行 |';
 const {post,input}=await fixture(page,request,info,markdown);
 const table=post.locator('.post-content table');await expect(table).toBeVisible();
 const layout=await table.evaluate(el=>{const css=getComputedStyle(el);return {display:css.display,layout:css.tableLayout,width:el.getBoundingClientRect().width,parent:el.parentElement.getBoundingClientRect().width}});
 expect(layout.display).toBe('table');expect(layout.layout).toBe('auto');expect(Math.abs(layout.width-layout.parent)).toBeLessThanOrEqual(1);
 await expect(table.locator('tbody tr')).toHaveCount(2);await expect(table).toContainText('第二行');await expect(input).toHaveValue('rendering draft retained');
});

test('@ux-timeline-024 Code-copy control copies native code text from the top-right of its block',async({page,request},info)=>{
 await evidence(info,'@ux-timeline-024');await observeClipboard(page);
 const code='const answer = "<tag> & 中文🙂";\nconsole.log(answer);\n';
 const {post,input}=await fixture(page,request,info,'```javascript\n'+code+'```');
 const block=post.locator('.post-code-block'),button=block.getByRole('button',{name:'Copy code',exact:true});
 await expect(button).toBeVisible();await expect(block.locator('pre code')).toHaveText(code);
 const position=await button.evaluate(el=>{const a=el.getBoundingClientRect(),b=el.closest('.post-code-block').getBoundingClientRect();return {top:a.top-b.top,right:b.right-a.right}});
 expect(position.top).toBeGreaterThanOrEqual(0);expect(position.top).toBeLessThanOrEqual(12);expect(position.right).toBeGreaterThanOrEqual(0);expect(position.right).toBeLessThanOrEqual(12);
 await button.click();
 await expect.poll(()=>page.evaluate(()=>window.__nativeCopies.at(-1))).toMatchObject({trusted:true,text:code});
 expect((await page.evaluate(()=>window.__nativeCopies.at(-1))).text).not.toContain('<span');
 await expect(input).toHaveValue('rendering draft retained');
});

test('@ux-original-029 Fenced SVG remains source code and copies without becoming a diagram',async({page,request},info)=>{
 await evidence(info,'@ux-original-029');await observeClipboard(page);
 const svg='<svg xmlns="http://www.w3.org/2000/svg" aria-label="source-only"><text x="2" y="12">SVG & text</text></svg>\n';
 const {post,input}=await fixture(page,request,info,'```svg\n'+svg+'```');
 const block=post.locator('.post-code-block');await expect(block.locator('pre code')).toHaveText(svg);
 await expect(post.locator('.post-content svg[aria-label="source-only"]')).toHaveCount(0);
 await block.getByRole('button',{name:'Copy code',exact:true}).click();
 await expect.poll(()=>page.evaluate(()=>window.__nativeCopies.at(-1))).toMatchObject({trusted:true,text:svg});
 await expect(input).toHaveValue('rendering draft retained');
});

test('Gi wide Markdown tables stay inside the timeline and preserve the draft after reload',async({page,request},info)=>{
 const headers=Array.from({length:12},(_,i)=>`Column ${i+1}`);
 const markdown='| '+headers.join(' | ')+' |\n| '+headers.map(()=> '---').join(' | ')+' |\n| '+headers.map((_,i)=>`payload_${i}_${'abcdef'.repeat(8)}`).join(' | ')+' |';
 const {post,input,stored,session}=await fixture(page,request,info,markdown);
 const table=post.locator('.post-content table');await expect(table).toBeVisible();
 const measurements=async()=>table.evaluate(el=>({display:getComputedStyle(el).display,layout:getComputedStyle(el).tableLayout,body:document.body.scrollWidth,viewport:innerWidth,content:el.closest('.post-content').getBoundingClientRect().width,table:el.getBoundingClientRect().width}));
 expect((await measurements()).body).toBeLessThanOrEqual((await measurements()).viewport);
 expect((await measurements()).display).toBe('table');expect((await measurements()).layout).toBe('auto');
 const overflow=await table.evaluate(el=>{const content=el.closest('.post-content');return {client:content.clientWidth,scroll:content.scrollWidth,overflow:getComputedStyle(content).overflowX}});
 expect(overflow.scroll).toBeGreaterThan(overflow.client);expect(overflow.overflow).toBe('auto');
 // Native wheel input can reach the final column; it must not be clipped by
 // the supplied post-body overflow boundary.
 await table.locator('td').first().hover();await page.mouse.wheel(10000,0);
 await expect.poll(()=>table.evaluate(el=>el.closest('.post-content').scrollLeft)).toBeGreaterThan(0);
 await expect.poll(()=>table.locator('td').last().evaluate(el=>{const cell=el.getBoundingClientRect(),content=el.closest('.post-content').getBoundingClientRect();return cell.right<=content.right+1&&cell.left<content.right})).toBe(true);
 expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);
 // Native edge gestures also belong to this table, not session navigation.
 await table.locator('td').last().hover();await page.mouse.wheel(10000,0);await page.waitForTimeout(500);await page.mouse.wheel(10000,0);await page.waitForTimeout(500);
 expect(await page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(session.id);
 await expect(table.locator('td')).toHaveCount(12);await expect(input).toHaveValue('rendering draft retained');
 await page.reload();await expect(page.locator(`#post-${stored.id} table td`)).toHaveCount(12);await expect(input).toHaveValue('rendering draft retained');
});
