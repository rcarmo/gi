import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchComposeSurface} from '../../../scripts/patch-compose-surface.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
import {patchPickerGeometry} from '../../../scripts/patch-picker-geometry.mjs';
import {patchComposeCommands} from '../../../scripts/patch-compose-commands.mjs';
import {composeHeightBounds,clampComposeHeight,readComposeHeight} from '../../../web/src/gi-compose-height.ts';

test('compose height bounds follow pinned automatic/manual limits',()=>{
 expect(composeHeightBounds(390,844)).toEqual({min:50,autoMax:300,manualMax:422});
 expect(composeHeightBounds(1440,900)).toEqual({min:70,autoMax:300,manualMax:450});
 expect(composeHeightBounds(820,400)).toEqual({min:50,autoMax:160,manualMax:200});
 expect(clampComposeHeight(900,1440,900)).toBe(450);expect(clampComposeHeight(-2,1440,900)).toBe(70);
 expect(readComposeHeight(null)).toBeNull();expect(readComposeHeight('')).toBeNull();expect(readComposeHeight('NaN')).toBeNull();expect(readComposeHeight('0')).toBeNull();expect(readComposeHeight('80')).toBe(80);
});
test('guarded adapter preserves supplied source and existing handlers',()=>{
 const path='web/src/components/compose-box.ts',source=readFileSync(path,'utf8');
 const before=patchComposeCommands(patchPickerGeometry(patchSkillPrefill(patchUploadCancel(patchModelPicker(patchComposePopupKeys(source))))));
 const adapted=patchComposeSurface(before);
 expect(adapted).toContain('const resizeTextarea = giComposeSurface.resize;');
 expect(adapted.match(/class="compose-session-trigger-group compose-session-trigger-top"/g)).toHaveLength(1);
 expect(adapted).toContain('onClick=${toggleSessionPopup}');expect(adapted).toContain('onKeyDown=${handleKeyDown}');
 expect(adapted).toContain('onClick=${toggleModelPopup}');expect(adapted).toContain('onMouseDown=${giComposeSurface.onMouseDown}');
 expect(adapted.indexOf('compose-session-trigger-top')).toBeLessThan(adapted.indexOf('<div class="compose-input-main">'));
 expect(()=>patchComposeSurface(adapted)).toThrow();expect(()=>patchComposeSurface(before.replace('const textareaRef =','const renamedRef ='))).toThrow();
 expect(readFileSync(path,'utf8')).toBe(source);
});
