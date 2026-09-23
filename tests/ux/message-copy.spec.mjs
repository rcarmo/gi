import {test,expect} from '@playwright/test';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
async function fixture(page,request,info){
 const name=`copy-${info.project.name}-${Date.now()}`;
 const session=await(await request.post('/api/sessions',{data:{agent_id:name,title:name}})).json();
 const source='# Source Markdown\n\n**Bold** and _italic_ with [a link](https://example.invalid/safe).\n\n```js\nconst value = "<世界>";\n```';
 const accepted=await request.post(`/api/sessions/${session.id}/prompt`,{data:{prompt:source,model:'test-model'}});expect(accepted.status()).toBe(202);const run=await accepted.json();
 await expect.poll(async()=>((await(await request.get(`/api/sessions/${session.id}/turns`)).json()).turns??[]).find(t=>t.id===run.turn_id)?.status).toBe('completed');
 const {messages}=await(await request.get(`/api/sessions/${session.id}/messages`)).json();const stored=messages.find(m=>m.role==='assistant');
 await page.addInitScript(id=>{localStorage.setItem('gi_session_id',id);window.__copyEvents=[];document.addEventListener('copy',event=>window.__copyEvents.push({trusted:event.isTrusted,text:event.clipboardData?.getData('text/plain'),html:event.clipboardData?.getData('text/html')}));},session.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();
 await input.fill('newer draft 世界');await page.locator('.compose-box input[type=file]').setInputFiles({name:'kept.txt',mimeType:'text/plain',buffer:Buffer.from('draft bytes')});await input.evaluate(el=>el.setSelectionRange(4,4));
 const post=page.locator(`[id="post-${stored.id}"]`);await expect(post).toBeVisible();
 return {session,input,post,stored,messages};
}
test('Gi message copy uses stored source Markdown and real plain/rich clipboard data without changing draft or media',async({page,request},info)=>{
 const {session,input,post,stored,messages}=await fixture(page,request,info);
 await post.getByRole('button',{name:'Copy message',exact:true}).click();
 await expect(post.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
 await expect.poll(()=>page.evaluate(()=>window.__copyEvents.length)).toBeGreaterThan(0);
 const copied=await page.evaluate(()=>window.__copyEvents.at(-1));expect(copied.trusted).toBe(true);expect(copied.text).toBe(stored.content.replace(/\r\n?/g,'\n').trimEnd());expect(copied.html).toContain('<strong>Bold</strong>');expect(copied.html).toContain('<style>');expect(copied.text).toContain('```js');expect(copied.text).not.toContain('<strong>');
 await expect(input).toHaveValue('newer draft 世界');expect(await input.evaluate(el=>[el.selectionStart,el.selectionEnd])).toEqual([4,4]);await expect(page.locator('.compose-file-pill').filter({hasText:'kept.txt'})).toBeVisible();
 expect((await(await request.get(`/api/sessions/${session.id}/messages`)).json()).messages).toEqual(messages);await page.screenshot({path:info.outputPath('message-copy-success.png')});
 await expect(post.getByRole('button',{name:'Copy message',exact:true})).toBeVisible({timeout:5000});
 await page.reload();await expect(input).toHaveValue('newer draft 世界');await expect(page.locator('.compose-file-pill').filter({hasText:'kept.txt'})).toBeVisible();
});
test('Gi message copy cannot report success when all clipboard mechanisms are unavailable',async({page,request},info)=>{
 const {input,post}=await fixture(page,request,info);
 await page.evaluate(()=>{document.execCommand=()=>false;Object.defineProperty(navigator,'clipboard',{configurable:true,value:undefined});});
 await post.getByRole('button',{name:'Copy message',exact:true}).click();
 await expect(post.getByRole('button',{name:'Copy failed',exact:true})).toBeVisible();await expect(input).toHaveValue('newer draft 世界');await expect(post.getByRole('button',{name:'Copied',exact:true})).toHaveCount(0);
});

test('Gi message copy falls back from denied rich copy to a real plain-text copy event',async({page,request},info)=>{
 const {input,post,stored}=await fixture(page,request,info);
 await page.evaluate(()=>{
  const native=document.execCommand.bind(document);let attempts=0;window.__legacyAttempts=0;
  document.execCommand=(command,...args)=>{if(command==='copy'){window.__legacyAttempts=++attempts;if(attempts===1)return false;}return native(command,...args);};
  Object.defineProperty(navigator,'clipboard',{configurable:true,value:{write:async()=>{throw new DOMException('Denied for failure test','NotAllowedError')},writeText:async()=>{throw new DOMException('Denied for failure test','NotAllowedError')}}});
 });
 await post.getByRole('button',{name:'Copy message',exact:true}).click();await expect(post.getByRole('button',{name:'Copied',exact:true})).toBeVisible();
 expect(await page.evaluate(()=>window.__legacyAttempts)).toBe(2);
 const copied=await page.evaluate(()=>window.__copyEvents.at(-1));expect(copied.trusted).toBe(true);expect(copied.text).toBe(stored.content.trimEnd());expect(copied.html).toBe('');
 await expect(input).toHaveValue('newer draft 世界');expect(await input.evaluate(el=>el.selectionStart)).toBe(4);
});

test('Gi denied clipboard reports failure for both message and code and resets without editing the draft',async({page,request},info)=>{
 const {input,post}=await fixture(page,request,info);
 await page.evaluate(()=>{document.execCommand=()=>false;Object.defineProperty(navigator,'clipboard',{configurable:true,value:{write:async()=>{throw new DOMException('denied','NotAllowedError')},writeText:async()=>{throw new DOMException('denied','NotAllowedError')}}});});
 await post.getByRole('button',{name:'Copy message',exact:true}).click();await expect(post.getByRole('button',{name:'Copy failed',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('message-copy-denied.png')});
 await expect(post.getByRole('button',{name:'Copy message',exact:true})).toBeVisible({timeout:5000});
 await post.getByRole('button',{name:'Copy code',exact:true}).click();await expect(post.locator('.post-code-copy-btn')).toHaveAttribute('data-copy-state','error');
 await expect(input).toHaveValue('newer draft 世界');await expect(page.locator('.compose-file-pill').filter({hasText:'kept.txt'})).toBeVisible();
 expect(await page.evaluate(()=>window.__copyEvents.length)).toBe(0);
});

test('Gi late clipboard denial after session switch does not enter another post or draft',async({page,request},info)=>{
 const {session,input,post}=await fixture(page,request,info);
 const child=(await(await request.post(`/api/sessions/${session.id}/fork`,{data:{agent_id:'copy-child-'+Date.now(),title:'copy child'}})).json()).branch.chat_jid.slice(3);
 await page.reload();await expect(post).toBeVisible();
 await page.evaluate(()=>{window.__copyPending=false;document.execCommand=()=>false;Object.defineProperty(navigator,'clipboard',{configurable:true,value:{writeText:()=>new Promise((_,reject)=>{window.__copyPending=true;window.__rejectCopy=()=>reject(new Error('late denied'));})}});});
 await post.getByRole('button',{name:'Copy message',exact:true}).click();await expect.poll(()=>page.evaluate(()=>window.__copyPending)).toBe(true);
 await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${child}"]`).getByRole('menuitem').click();await expect(input).toHaveValue('');await input.fill('child draft');
 await page.evaluate(()=>window.__rejectCopy());await expect(page.getByRole('button',{name:'Copy failed',exact:true})).toHaveCount(0);await expect(input).toHaveValue('child draft');
 await page.getByRole('button',{name:/Manage sessions for/}).last().click();await page.locator(`[data-session-jid="gi:${session.id}"]`).getByRole('menuitem').click();await expect(input).toHaveValue('newer draft 世界');await expect(post.getByRole('button',{name:'Copy message',exact:true})).toBeVisible();
});
