import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
test('canonical skill prefill preserves text only for skill commands and fences late focus',()=>{
 const source=patchUploadCancel(patchModelPicker(patchComposePopupKeys(readFileSync('web/src/components/compose-box.ts','utf8'))));const patched=patchSkillPrefill(source);
 expect(patched).toContain("resolved.text = resolved.text.trim() + ' ' + content;");expect(patched).toContain("skillPrefill && (!mountedRef.current || document.querySelector('.settings-dialog[aria-modal=\"true\"]'))");
 const pattern=/^\/skill:[A-Za-z0-9][A-Za-z0-9_-]{0,63}\s*$/;
 for(const text of ['/skill:proof ','/skill:Name_ok-2\t'])expect(pattern.test(text)).toBe(true);
 for(const text of ['/model ','/skill:','/skill:../bad ','/skill:proof argument'])expect(pattern.test(text)).toBe(false);
 expect(()=>patchSkillPrefill(source.replace('lastPrefillTokenRef.current = resolved.nextToken;','changed'))).toThrow();expect(()=>patchSkillPrefill(patched)).toThrow();
});
