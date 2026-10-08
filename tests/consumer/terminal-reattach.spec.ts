import { test, expect } from '../fixtures-vibes/profiled-fixtures';

// Independent Gi continuity acceptance beyond the shared tab-visibility gate.
test('@ux-terminal-006 Gi consumer: terminal popout restores the existing PTY and unsent shell input', async ({ page, context, runtime, sel }) => {
  test.setTimeout(90000);
  await page.goto((await runtime.newSession()).url);
  const composer=page.locator(sel('composeInput'));await composer.fill('keep composer draft');
  await page.getByRole('button',{name:'Menu',exact:true}).click();await page.getByRole('menuitem',{name:/open terminal in tab|open terminal/i}).first().click();
  const terminal=(p:any)=>p.locator(sel('terminal')).filter({visible:true}).first();
  const command=async(p:any,text:string,output:string)=>{const t=terminal(p);await t.click();await p.keyboard.type(text,{delay:15});await p.keyboard.press('Enter');await expect(t).toContainText(output,{timeout:15000});};
  await command(page,"export GI_REATTACH_MARK=preserved; printf 'MARK:%s PID:%s\\n' \"$GI_REATTACH_MARK\" \"$$\"",'MARK:preserved');
  const before=await terminal(page).innerText();const pid=/PID:(\d+)/.exec(before)?.[1];expect(pid).toBeTruthy();
  await terminal(page).click();await page.keyboard.type("printf 'DRAFT:%s\\n' \"$GI_REATTACH_MARK\"",{delay:15});
  const popup=context.waitForEvent('page');
  await page.getByRole('tab',{name:/terminal/i}).first().click({button:'right'});
  await page.getByRole('button',{name:/open.*window|pop out/i}).filter({visible:true}).first().click();
  const detached=await popup;await detached.waitForLoadState('domcontentloaded');
  await expect(terminal(detached)).toBeVisible({timeout:15000});
  // A same-origin stale sender with otherwise matching identifiers must not
  // claim the current detached pane. Keep the real popup/PTY alive.
  const params=new URL(detached.url()).searchParams;
  await page.evaluate(({path,id})=>window.dispatchEvent(new MessageEvent('message',{origin:location.origin,source:window,data:{type:'piclaw-pane-reattach-request',panePath:path,paneInstanceId:id}})),{path:params.get('pane_path'),id:params.get('pane_instance_id')});
  await expect(page.getByRole('tab',{name:/terminal/i}).first()).toHaveCount(0);
  await detached.close();
  await expect(page.getByRole('tab',{name:/terminal/i}).first()).toBeVisible({timeout:15000});
  await expect.poll(async()=> (await terminal(page).innerText()).includes('MARK:preserved'),{timeout:15000}).toBe(true);
  await terminal(page).click();await page.keyboard.press('Enter');
  await expect(terminal(page)).toContainText('DRAFT:preserved',{timeout:15000});
  await command(page,"printf 'RETURN:%s PID:%s\\n' \"$GI_REATTACH_MARK\" \"$$\"",`RETURN:preserved PID:${pid}`);
  await expect(composer).toHaveValue('keep composer draft');
  const status=await page.evaluate(async()=>{const token=localStorage.getItem('piclaw_terminal_client');const r=await fetch('/terminal/session',{headers:token?{'x-piclaw-terminal-client':token}:{}});if(!r.ok)throw Error(`status ${r.status}`);return r.json();});
  expect(status.connected_clients).toBe(1);
});
