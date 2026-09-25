import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchPickerGeometry} from '../../../scripts/patch-picker-geometry.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
import reference from '../fixtures/picker-geometry-reference.json';
const path='web/src/components/compose-box.ts';
test('picker adapter only inserts close controls after guarded existing adapters',()=>{
 const source=readFileSync(path,'utf8'),before=patchSkillPrefill(patchUploadCancel(patchModelPicker(patchComposePopupKeys(source)))),after=patchPickerGeometry(before);
 expect(after.replace(/^.*class="gi-picker-close".*\n/gm,'')).toBe(before);
 expect(after.match(/class="gi-picker-close"/g)).toHaveLength(2);expect(after).toContain('closeSessionPopup(true)');
 expect(readFileSync(path,'utf8')).toBe(source);
 for(const changed of [before.replace('Select model</div>','Pick model</div>'),before+'                            <div class="compose-model-popup-title">Select model</div>',after])expect(()=>patchPickerGeometry(changed)).toThrow();
});
test('picker reference pins geometry separately from frozen feature mappings',()=>{
 expect(reference.reference.commit).toBe('0afe5366ced9bca8246abd99a0feb1875a6ffbcc');expect(reference.reference.observedAsset).toBe('15958f2c3dc9');
 expect(reference.reference.chatSHA256).toBe('b6d3da5e2b0aff389efe095a20fe7388c37fe679bfa1ba04bc12ca4f4680597b');
 expect(reference.rules).toEqual({composerInlinePadding:0,mobileMaxWidth:639,mobileInset:8,desktopGap:6,modelMaxWidth:680,modelViewportGutter:24});
 expect(reference.adaptations.desktopModelMaxAnchorWidth).toBe(true);expect(reference.observed.map(r=>r.width)).toEqual([390,639,640,820,1440]);
});
