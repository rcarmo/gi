import {test,expect} from '@playwright/test';
import {readFileSync,mkdirSync} from 'node:fs';
const secret='gi-fixture-provider-key',inputName='Message (Enter to send, Shift+Enter for newline)...';
const feature=readFileSync('tests/features/settings/gi-settings.feature','utf8');
async function metadata(request){const response=await request.get('/api/settings/providers');expect(response.status()).toBe(200);const text=await response.text();expect(text).not.toContain(secret);expect(text).not.toContain('fixture-only');return JSON.parse(text);}
async function clean(request){const current=await metadata(request);expect((await request.delete('/api/settings/providers',{data:{provider:'openai',revision:current.revision}})).status()).toBe(200);}
async function setup(page,request,info){
 await info.attach('gi-gherkin',{body:feature,contentType:'text/plain'});await clean(request);
 const main=await(await request.post('/api/sessions',{data:{agent_id:`keys-${info.project.name}-${Date.now()}`}})).json();
 await page.addInitScript(id=>{if(!localStorage.getItem('gi_session_id'))localStorage.setItem('gi_session_id',id);},main.id);await page.goto('/');
 const input=page.getByRole('textbox',{name:inputName,exact:true});await expect(input).toBeVisible();await input.fill('provider settings draft');
 const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});const open=async()=>{await page.keyboard.press('Control+,');await dialog.getByRole('button',{name:'Providers',exact:true}).click();await expect(dialog.getByRole('region',{name:'Provider openai',exact:true})).toBeVisible();};
 return{main,input,dialog,open,row:dialog.getByRole('region',{name:'Provider openai',exact:true})};
}

test('@gi-settings-020 @gi-settings-021 Metadata-only Providers and explicit native key consumption',async({page,request},info)=>{
 const{main,input,dialog,open,row}=await setup(page,request,info);const notices=[];page.on('console',msg=>notices.push(msg.text()));
 try{
  const before=await metadata(request);await open();await expect(dialog.getByText(/private plaintext/)).toBeVisible();
  const other=dialog.getByRole('region',{name:'Provider ux-local',exact:true});await expect(other.getByText('Read-only credential metadata.')).toBeVisible();await expect(other.locator('input')).toHaveCount(0);
  const key=row.getByLabel('API key for openai');await expect(key).toHaveAttribute('type','password');await key.fill(secret);expect((await metadata(request)).revision).toBe(before.revision);
  let release,held=false;const gate=new Promise(r=>release=r);await page.route('**/api/settings/providers',async route=>{if(route.request().method()!=='PATCH')return route.continue();const response=await route.fetch();expect((await response.text())).not.toContain(secret);held=true;await gate;await route.fulfill({response});});
  try{await row.getByRole('button',{name:'Save key',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(key).toBeDisabled();release();await expect(dialog.getByRole('status')).toContainText('stored, not verified');await expect(key).toHaveValue('');}finally{release();}
  await page.unroute('**/api/settings/providers');await expect(row).toContainText('Stored api_key');
  expect((await metadata(request)).providers.find(p=>p.id==='openai')).toMatchObject({stored:true,editable:true,kind:'api_key'});
  expect(await page.evaluate(()=>JSON.stringify({...localStorage,...sessionStorage}))).not.toContain(secret);expect(notices.join('\n')).not.toContain(secret);
  if(process.env.GI_SETTINGS_CAPTURE&&info.project.name.startsWith('chromium-')){mkdirSync('test-results/gi-settings-captures',{recursive:true});await page.screenshot({path:`test-results/gi-settings-captures/providers-${info.project.name}.png`});}
  await page.keyboard.press('Escape');await expect(input).toHaveValue('provider settings draft');
  expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).toHaveLength(0);
  const accepted=await request.post(`/api/sessions/${main.id}/prompt`,{data:{prompt:'Native provider accepted the configured fixture credential',model:'openai/gi-key-fixture'}});expect(accepted.status()).toBe(202);const run=await accepted.json();
  await expect.poll(async()=>((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).find(t=>t.id===run.turn_id)?.status).toBe('completed');
  const messages=(await(await request.get(`/api/sessions/${main.id}/messages`)).json()).messages;expect(messages.some(m=>m.role==='assistant'&&m.content.includes('Native provider accepted'))).toBe(true);
  await page.reload();await expect(input).toHaveValue('provider settings draft');await open();await expect(row).toContainText('Stored api_key');await expect(row.getByLabel('API key for openai')).toHaveValue('');
 }finally{await clean(request);}
});

test('@gi-settings-022 @gi-settings-023 Stale credentials reject, refresh retries and remove requires confirmation',async({page,request},info)=>{
 const{input,dialog,open,row}=await setup(page,request,info);
 try{
  await open();const key=row.getByLabel('API key for openai');await key.fill(secret);
  const current=await metadata(request);expect((await request.patch('/api/settings/providers',{data:{provider:'openai',revision:current.revision,key:'other-fixture-key'}})).status()).toBe(200);
  await row.getByRole('button',{name:'Save key',exact:true}).click();await expect(dialog.getByRole('alert')).toContainText('Credential store changed');await expect(key).toHaveValue('');
  await dialog.getByRole('button',{name:'Refresh providers'}).click();await expect(row).toContainText('Stored api_key');await key.fill(secret);
  await row.getByRole('button',{name:'Save key',exact:true}).click();await expect(dialog.getByRole('status')).toContainText('stored, not verified');
  let deletes=0;page.on('request',r=>{if(r.method()==='DELETE'&&r.url().endsWith('/api/settings/providers'))deletes++;});
  await row.getByRole('button',{name:'Remove key',exact:true}).click();await row.getByRole('button',{name:'Cancel removal'}).click();expect(deletes).toBe(0);
  await row.getByRole('button',{name:'Remove key',exact:true}).click();await row.getByRole('button',{name:'Confirm removal'}).click();await expect(row).toContainText('Not stored');expect(deletes).toBe(1);
  expect((await metadata(request)).providers.find(p=>p.id==='ux-local').stored).toBe(true);
  await page.keyboard.press('Escape');await expect(input).toHaveValue('provider settings draft');
 }finally{await clean(request);}
});

test('@gi-settings-023 Closed save responses cannot restore secrets or announce success in a new view',async({page,request},info)=>{
 const{dialog,open,row}=await setup(page,request,info);let release,held=false,done;const gate=new Promise(r=>release=r),delivered=new Promise(r=>done=r);
 await page.route('**/api/settings/providers',async route=>{if(route.request().method()!=='PATCH')return route.continue();const response=await route.fetch();held=true;await gate;await route.fulfill({response});done();});
 try{
  await open();await row.getByLabel('API key for openai').fill(secret);await row.getByRole('button',{name:'Save key',exact:true}).click();await expect.poll(()=>held).toBe(true);
  await page.keyboard.press('Escape');await open();await expect(row).toContainText('Stored api_key');release();await delivered;
  await expect(row.getByLabel('API key for openai')).toHaveValue('');await expect(dialog.getByRole('status').filter({hasText:'stored, not verified'})).toHaveCount(0);
  expect(await page.evaluate(()=>JSON.stringify({...localStorage,...sessionStorage}))).not.toContain(secret);
 }finally{release();await clean(request);}
});

test('@gi-settings-022 Native storage failure and invalid keys report no secret or false success',async({page,request},info)=>{
 const{dialog,open,row}=await setup(page,request,info);
 try{
  await open();const key=row.getByLabel('API key for openai');await key.fill('invalid key');await row.getByRole('button',{name:'Save key',exact:true}).click();
  await expect(dialog.getByRole('alert')).toContainText('Credential change failed');await expect(key).toHaveValue('');expect((await metadata(request)).providers.find(p=>p.id==='openai').stored).toBe(false);
  const cfg=await(await request.get('/api/runtime/config')).json();
  // The local provider fixture isolates HOME and workspace to this generated directory.
  expect(cfg.workspace_root).toMatch(/\/gi-steer-ux-[^/]+$/);
  const{join}=await import('node:path'),{rmSync}=await import('node:fs');const lock=join(cfg.workspace_root,'.pi','agent','.gi-auth.lock');
  rmSync(lock,{force:true});mkdirSync(lock);
  try{
   await key.fill(secret);await row.getByRole('button',{name:'Save key',exact:true}).click();await expect(dialog.getByRole('alert')).toContainText('Credential change failed');await expect(key).toHaveValue('');expect((await metadata(request)).providers.find(p=>p.id==='openai').stored).toBe(false);
   expect(await dialog.innerText()).not.toContain(secret);
  }finally{rmSync(lock,{recursive:true,force:true});}
  await dialog.getByRole('button',{name:'Refresh providers'}).click();await key.fill(secret);await row.getByRole('button',{name:'Save key',exact:true}).click();await expect(dialog.getByRole('status')).toContainText('stored, not verified');
 }finally{await clean(request);}
});

test('@gi-settings-020 @gi-settings-023 Held provider read stays out of a reopened dialog',async({page,request},info)=>{
 const{dialog,open,row}=await setup(page,request,info);let release,held=false,done;const gate=new Promise(r=>release=r),delivered=new Promise(r=>done=r);
 await page.route('**/api/settings/providers',async route=>{if(route.request().method()!=='GET'||held)return route.continue();const response=await route.fetch();held=true;await gate;await route.fulfill({response});done();});
 try{
  await page.keyboard.press('Control+,');await dialog.getByRole('button',{name:'Providers',exact:true}).click();await expect.poll(()=>held).toBe(true);await expect(dialog.getByRole('status')).toHaveText('Loading providers…');
  await page.keyboard.press('Escape');const current=await metadata(request);await request.patch('/api/settings/providers',{data:{provider:'openai',revision:current.revision,key:secret}});
  await open();await expect(row).toContainText('Stored api_key');release();await delivered;await expect(row).toContainText('Stored api_key');await expect(row.getByLabel('API key for openai')).toHaveValue('');
 }finally{release();await clean(request);}
});

test('@gi-settings-020 Remote HTTP metadata disables credential inputs and read failures offer retry',async({page,request},info)=>{
 const{dialog,open,row}=await setup(page,request,info);
 try{
  await page.route('**/api/settings/providers',route=>route.request().method()==='GET'?route.abort():route.continue());
  await page.keyboard.press('Control+,');await dialog.getByRole('button',{name:'Providers',exact:true}).click();
  await expect(dialog.getByRole('alert')).toContainText('Cannot read providers');await expect(dialog.locator('input[type="password"]')).toHaveCount(0);
  await page.unroute('**/api/settings/providers');
  // Preserve the native response, but exercise its untrusted public-host transport branch.
  await page.route('**/api/settings/providers',async route=>{
   const response=await route.fetch({headers:{...route.request().headers(),host:'untrusted-public.invalid'}});await route.fulfill({response});
  });
  await dialog.getByRole('button',{name:'Refresh providers'}).click();await expect(row).toBeVisible();await expect(dialog.getByText(/Use HTTPS or a loopback connection/)).toBeVisible();
  await expect(dialog.locator('input[type="password"]')).toHaveCount(0);await expect(dialog.getByRole('button',{name:'Save key',exact:true})).toHaveCount(0);
  const current=await metadata(request);const denied=await request.patch('/api/settings/providers',{headers:{Host:'untrusted-public.invalid'},data:{provider:'openai',revision:current.revision,key:secret}});expect(denied.status()).toBe(403);expect(await denied.text()).not.toContain(secret);
  await page.unroute('**/api/settings/providers');await dialog.getByRole('button',{name:'Refresh providers'}).click();await expect(row.getByLabel('API key for openai')).toBeVisible();
 }finally{await clean(request);}
});
