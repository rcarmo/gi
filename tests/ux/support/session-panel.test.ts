import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchSessionPanel} from '../../../scripts/patch-session-panel.mjs';
import {patchModelPanel} from '../../../scripts/patch-model-panel.mjs';
import {patchComposeSurface} from '../../../scripts/patch-compose-surface.mjs';
import {patchComposeCommands} from '../../../scripts/patch-compose-commands.mjs';
import {patchPickerGeometry} from '../../../scripts/patch-picker-geometry.mjs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';

test('session presentation adapter retains guarded native mutation and focus handlers',()=>{
 const path='web/src/components/compose-box.ts',source=readFileSync(path,'utf8');
 const previous=patchModelPanel(patchComposeSurface(patchComposeCommands(patchPickerGeometry(patchSkillPrefill(patchUploadCancel(patchModelPicker(patchComposePopupKeys(source))))))));
 const adapted=patchSessionPanel(previous);
 expect(adapted).toContain("import { normalizeHandle }");expect(adapted).toContain('class="compose-session-popup-header"');expect(adapted).toContain('aria-label=${label}');expect(adapted).toContain('aria-pressed=${chat.pinned');
 expect(adapted.match(/runSessionMutation\(chat, 'pin', !chat.pinned\)/g)).toHaveLength(1);
 for(const handler of ["beginSessionEdit(chat, 'rename')","beginSessionEdit(chat, 'archive')","runSessionMutation(chat, 'restore')","closeSessionPopup(true)","handleSessionSwitch(chat.chat_jid)","handleRestoreSession(chat.chat_jid)"])expect(adapted).toContain(handler);
 expect(adapted).toContain('useLayoutEffect(() => {\n        if (!showSessionPopup) return;\n        const preferred');
 expect(adapted).toContain('data-session-entry-key=${`session:${chat.chat_jid}`}');expect(adapted).toContain('ref=${sessionSearchRef}');expect(adapted).toContain("from '../ui/branch-lifecycle.js'");
 expect(()=>patchSessionPanel(adapted)).toThrow();expect(()=>patchSessionPanel('drift')).toThrow();expect(readFileSync(path,'utf8')).toBe(source);
});
