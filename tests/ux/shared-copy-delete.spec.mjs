import {test,expect} from '@playwright/test';
import {mkdirSync,writeFileSync} from 'node:fs';
import {resolve} from 'node:path';
import {loadCorpus} from './support/catalogue.mjs';

const inputName='Message (Enter to send, Shift+Enter for newline)...';
const source='# Stored Markdown\n\n**Bold** and _italic_.\n\n```js\nconst value = "<世界>";\n```';
const code='const value = "<世界>";\n';

async function setup(page,request,info){
 const agent=`shared37-${info.project.name}-${Date.now()}`;
 const main=await(await request.post('/api/sessions',{data:{agent_id:agent,title:agent}})).json();
 const child=(await(await request.post(`/api/sessions/${main.id}/fork`,{data:{agent_id:`${agent}-child`,title:'child'}})).json()).branch.chat_jid.slice(3);
 const prompt=async(id,text)=>{
  const res=await request.post(`/api/sessions/${id}/prompt`,{data:{prompt:text,model:'test-model'}});expect(res.status()).toBe(202);
  const {turn_id}=await res.json();
  await expect.poll(async()=>((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns??[]).find(t=>t.id===turn_id)?.status).toBe('completed');
  await expect.poll(async()=> (await(await request.get(`/api/sessions/${id}/activity`)).json()).status).toBe('idle');
 };
 await prompt(main.id,source);await prompt(child,'child intact');
 const getMessages=async (id=main.id)=>((await(await request.get(`/api/sessions/${id}/messages`)).json()).messages??[]);
 const stored=await getMessages(),target=stored.find(m=>m.role==='user'&&m.content===source);
 const other=stored.find(m=>m.role==='assistant');expect(target).toBeTruthy();expect(other).toBeTruthy();
 const childBefore=await getMessages(child);
 await page.addInitScript(id=>{
  localStorage.setItem('gi_session_id',id);window.__nativeCopies=[];
  document.addEventListener('copy',event=>window.__nativeCopies.push({trusted:event.isTrusted,text:event.clipboardData?.getData('text/plain'),html:event.clipboardData?.getData('text/html')}));
 },main.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 const targetPost=page.locator(`[id="post-${target.id}"]`),otherPost=page.locator(`[id="post-${other.id}"]`);
 await expect(targetPost).toBeVisible();await expect(otherPost).toBeVisible();
 await input.fill('draft retained 世界');
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'pending.txt',mimeType:'text/plain',buffer:Buffer.from('pending bytes')});
 // A message reference to a *different* row must remain in the editor.
 await otherPost.locator('.post-time').click();
 const ref=page.locator('.compose-file-refs').getByTitle(`Message reference: ${other.id}`);
 await expect(ref).toBeVisible();
 return {main,child,input,target,other,source:stored,childBefore,targetPost,otherPost,ref,getMessages};
}

async function unchangedDraft(input,ref,page){
 await expect(input).toHaveValue('draft retained 世界');
 await expect(ref).toBeVisible();
 await expect(page.locator('.compose-file-pill').filter({hasText:'pending.txt'})).toBeVisible();
}

test('@shared-37 Copy and delete timeline messages through native actions',async({page,request},info)=>{
 const scenario=loadCorpus('shared').find(c=>c.id==='@shared-37');expect(scenario).toBeTruthy();
 await info.attach('gherkin',{body:scenario.steps.join('\n'),contentType:'text/plain'});
 const {main,child,input,target,other,source:stored,childBefore,targetPost,otherPost,ref,getMessages}=await setup(page,request,info);
 const targetCopy=targetPost.getByRole('button',{name:'Copy message',exact:true});
 const targetDelete=targetPost.getByRole('button',{name:'Delete message',exact:true});
 const otherCopy=otherPost.getByRole('button',{name:'Copy message',exact:true});
 const otherDelete=otherPost.getByRole('button',{name:'Delete message',exact:true});
 const codeCopy=targetPost.getByRole('button',{name:'Copy code',exact:true});
 for(const control of [targetCopy,targetDelete,otherCopy,otherDelete,codeCopy]){await expect(control).toBeVisible();await expect(control).toBeEnabled();}
 await targetCopy.focus();await page.keyboard.press('Enter');
 await expect(targetPost.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
 const first=await page.evaluate(()=>window.__nativeCopies.at(-1));expect(first).toMatchObject({trusted:true,text:source.trimEnd()});
 expect(first.text).toContain('```js');expect(first.text).not.toContain('<strong>');expect(first.html).toContain('<strong>Bold</strong>');
 await expect(targetCopy).toBeVisible({timeout:5000});
 await codeCopy.click();await expect(targetPost.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
 await expect.poll(()=>page.evaluate(()=>window.__nativeCopies.at(-1))).toMatchObject({trusted:true,text:code});
 await expect(codeCopy).toBeVisible({timeout:5000});await unchangedDraft(input,ref,page);
 // Native denial does not remove or animate a message; both copy controls
 // report failure and return to idle after their own timers.
 await page.evaluate(()=>{document.execCommand=()=>false;Object.defineProperty(navigator,'clipboard',{configurable:true,value:{write:async()=>{throw new DOMException('denied','NotAllowedError')},writeText:async()=>{throw new DOMException('denied','NotAllowedError')}}});});
 await targetCopy.click();await expect(targetPost.getByRole('button',{name:'Copy failed',exact:true})).toBeVisible();
 await expect(targetCopy).toBeVisible({timeout:5000});
 await codeCopy.click();await expect(targetPost.locator('.post-code-copy-btn')).toHaveAttribute('data-copy-state','error');
 await expect(codeCopy).toBeVisible({timeout:5000});
 await unchangedDraft(input,ref,page);
 // A real provider gate holds a running turn. Native DELETE must reject the
 // captured ID while busy, then accept the same ID after that turn settles.
 const token=`shared37-${info.project.name}-${Date.now()}`;
 const gate=resolve('test-results/ux-parity/queue-gates',token);mkdirSync(resolve(gate,'..'),{recursive:true});
 const accepted=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:`UX steer gate:${token}`,model:'ux-local/gate'}});
 expect(accepted.status()).toBe(202);const {turn_id}=await accepted.json();
 try{
  await expect.poll(async()=> (await(await request.get(`/api/sessions/${main.id}/activity`)).json()).status).toBe('running');
  const rejected=page.waitForResponse(res=>res.request().method()==='DELETE'&&res.url().includes(`/api/sessions/${main.id}/messages/${target.id}?cascade=false`));
  await targetDelete.click();expect((await rejected).status()).toBe(409);
  await expect(page.getByRole('alert').filter({hasText:'Delete failed:'})).toBeVisible();
  await expect(targetPost).toBeVisible();await expect(targetPost).not.toHaveClass(/removing/);
  const rejectedRows=await getMessages();for(const m of stored)expect(rejectedRows.some(row=>row.id===m.id&&row.content===m.content)).toBe(true);
  await unchangedDraft(input,ref,page);
 }finally{writeFileSync(gate,'release')}
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns??[]).find(t=>t.id===turn_id)?.status).toBe('completed');
 await expect.poll(async()=> (await(await request.get(`/api/sessions/${main.id}/activity`)).json()).status).toBe('idle');
 const acceptedDelete=page.waitForResponse(res=>res.request().method()==='DELETE'&&res.url().includes(`/api/sessions/${main.id}/messages/${target.id}?cascade=false`));
 await targetDelete.focus();await page.keyboard.press('Enter');
 const acceptedResponse=await acceptedDelete;expect(acceptedResponse.status()).toBe(200);
 expect((await acceptedResponse.json()).deleted).toEqual([target.id]);
 await expect(targetPost).toHaveCount(0);await expect(otherPost).toBeVisible();await unchangedDraft(input,ref,page);
 const after=await getMessages();expect(after.some(m=>m.id===target.id)).toBe(false);
 for(const m of stored.filter(m=>m.id!==target.id))expect(after.some(row=>row.id===m.id&&row.content===m.content)).toBe(true);
 expect(await getMessages(child)).toEqual(childBefore);
 await page.reload();await expect(targetPost).toHaveCount(0);await expect(otherPost).toBeVisible();await unchangedDraft(input,ref,page);
});
