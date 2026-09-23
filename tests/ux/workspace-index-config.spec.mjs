import {test,expect} from '@playwright/test';
const inputName='Message (Enter to send, Shift+Enter for newline)...';
test('Gi configured index roots/extensions accept initial absence and reject disappearance',async({page,request},info)=>{
 const runtime=await (await request.get('/api/runtime/config')).json();
 expect(runtime.workspace_index).toEqual({extraRoots:['docs'],extraExtensions:['nim'],optionalRoots:['notes','.pi/skills']});
 const quote=s=>"'"+s.replaceAll("'","'\\''")+"'";
 const shell=async command=>{const res=await request.post('/api/tools/execute',{data:{tool:'shell',input:{command:`cd ${quote(runtime.workspace_root)} && ${command}`}}});expect((await res.json()).error).toBeFalsy();};
 const write=async(path,content)=>{const res=await request.post('/api/tools/execute',{data:{tool:'write',input:{path,content}}});expect((await res.json()).error).toBeFalsy();};
 const token=info.project.name.replaceAll('-','');
 // Each project follows a completed prior one; notes may exist and must be kept.
 // Separate notes-scope cleanup occurs only after files are deliberately removed.
 await shell('mkdir -p notes; find notes -type f -name "fixture-*.md" -delete');
 expect((await request.post('/api/workspace/index?scope=all')).status()).toBe(200);
 await shell('rmdir notes');
 await write(`docs/${token}.nim`,`nim${token} source`);
 const session=await (await request.post('/api/sessions',{data:{title:token,agent_id:token}})).json();await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),session.id);
 await page.goto('/');const input=page.getByRole('textbox',{name:inputName,exact:true});await input.fill('optional roots draft');await page.locator('.compose-box input[type=file]').setInputFiles({name:'retain.txt',mimeType:'text/plain',buffer:Buffer.from('unsent')});
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Show workspace',exact:true}).click();
 const run=async()=>{await page.getByTestId('hamburger').click();const result=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/workspace/index');await page.getByRole('menuitem',{name:'Reindex workspace',exact:true}).click();return result;};
 const state=async()=>(await (await request.get('/api/workspace/index')).json());
 let response=await run();expect(response.status()).toBe(200);const ready=await response.json();expect(ready.required_roots).toBe(false);expect(ready.optional_roots).toEqual(['.pi/skills','notes']);expect(ready.roots).toEqual(['.pi/skills','docs','notes']);
 const hits=await (await request.get(`/api/workspace/search?q=nim${token}`)).json();expect(hits.hits.map(h=>h.path)).toEqual([`docs/${token}.nim`]);
 await write(`notes/fixture-${token}.md`,`kept${token} content`);response=await run();expect(response.status()).toBe(200);const populated=await response.json();
 await shell('mv notes notes-config-held');
 try{
  response=await run();expect(response.status()).toBe(500);expect((await response.json()).error).toContain('disappeared');
  await expect(page.locator('.workspace-index-status-row')).toContainText('Workspace index failed');
  const failed=await state();expect(failed.generation).toBe(populated.generation);expect(failed.indexed_file_count).toBe(populated.indexed_file_count);expect(failed.last_indexed_at).toBe(populated.last_indexed_at);
  expect((await (await request.get(`/api/workspace/search?q=kept${token}`)).json()).hits.map(h=>h.path)).toEqual([`notes/fixture-${token}.md`]);
  await expect(input).toHaveValue('optional roots draft');await expect(page.locator('.compose-file-pill').filter({hasText:'retain.txt'})).toBeVisible();await page.screenshot({path:info.outputPath('index-optional-root-failure.png')});
  await page.reload();await expect(input).toHaveValue('optional roots draft');expect((await state()).state).toBe('failed');
 }finally{await shell('mv notes-config-held notes')}
 expect((await request.post('/api/workspace/index')).status()).toBe(200);expect((await state()).generation).toBe(populated.generation+1);
 expect((await (await request.get(`/api/sessions/${session.id}/messages`)).json()).messages??[]).toEqual([]);
});
