import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {normaliseComposeCommands,declineComposeKey} from '../../../web/src/gi-compose-commands';
import {patchComposeCommands} from '../../../scripts/patch-compose-commands.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchPickerGeometry} from '../../../scripts/patch-picker-geometry.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
test('native compose catalogue validates, deduplicates and preserves supported order',()=>{
 expect(normaliseComposeCommands({commands:[{name:'/model',description:'model'},{name:'/skill:proof'},{name:'/model'}]})).toEqual([{name:'/model',description:'model'},{name:'/skill:proof',description:''}]);
 expect(normaliseComposeCommands({commands:[]})).toEqual([]);
 for(const commands of [null,{},['/model'],[{name:'model'}],[{name:'/m arg'}],[{name:'/model',description:7}]])expect(()=>normaliseComposeCommands({commands})).toThrow();
});
test('compose key ownership declines consumed, repeat and IME without blocking ordinary keys',()=>{
 const key={key:'Enter',repeat:false,defaultPrevented:false,isComposing:false,keyCode:13};expect(declineComposeKey(key)).toBe(false);
 for(const extra of [{repeat:true},{defaultPrevented:true},{isComposing:true},{keyCode:229}])expect(declineComposeKey({...key,...extra})).toBe(true);
 expect(declineComposeKey({...key,key:'ArrowDown',repeat:true})).toBe(false);
 expect(declineComposeKey({...key,key:'Tab',repeat:true})).toBe(true);
});
test('compose command adapter preserves render markup and fails on drift or double application',()=>{
 const path='web/src/components/compose-box.ts',source=readFileSync(path,'utf8');
 const before=patchPickerGeometry(patchSkillPrefill(patchUploadCancel(patchModelPicker(patchComposePopupKeys(source))))),after=patchComposeCommands(before);
 const withoutNotice=after.replace(/^.*\$\{commandCatalogueError && html`.*\n/gm,'');
 const render='        <div class="compose-box">';
 expect(withoutNotice.slice(withoutNotice.indexOf(render))).toBe(before.slice(before.indexOf(render)));
 expect(before.indexOf(render)).toBeGreaterThan(0);
 expect(after).not.toContain('/agent/commands?chat_jid=');expect(after).not.toContain('dynamicCommandsRef.current || SLASH_COMMANDS');expect(after).toContain('getAgentCommands(currentChatJid)');expect(after).toContain('if (cancelled) return;');
 for(const changed of [before.replace('const dynamicCommandsRef = useRef(null);','const dynamicCommandsRef = useRef(unknown);'),before+'    const dynamicCommandsRef = useRef(null);',after])expect(()=>patchComposeCommands(changed)).toThrow();
 expect(readFileSync(path,'utf8')).toBe(source);
});
