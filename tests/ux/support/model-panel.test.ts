import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {modelPanelRows,modelPanelContext,modelPanelName} from '../../../web/src/gi-model-panel.ts';
import {patchModelPanel} from '../../../scripts/patch-model-panel.mjs';
import {patchComposeSurface} from '../../../scripts/patch-compose-surface.mjs';
import {patchComposeCommands} from '../../../scripts/patch-compose-commands.mjs';
import {patchPickerGeometry} from '../../../scripts/patch-picker-geometry.mjs';
import {patchSkillPrefill} from '../../../scripts/patch-skill-prefill.mjs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';

test('current-first projection preserves option identity and remaining catalogue order',()=>{
 const rows=[{label:'a'},{label:'b'},{label:'c'}];const result=modelPanelRows(rows,'b');expect(result).toEqual([rows[1],rows[0],rows[2]]);expect(result[0]).toBe(rows[1]);expect(rows.map(r=>r.label)).toEqual(['a','b','c']);
 expect(modelPanelRows(rows,'missing')).toEqual(rows);
});
test('metadata formatting follows pinned reference without invented values',()=>{
 expect(modelPanelName({name:'Known',id:'id',label:'p/id'})).toBe('Known');expect(modelPanelName({id:'id',label:'p/id'})).toBe('id');expect(modelPanelContext(65536)).toBe('65.5K context');expect(modelPanelContext(1000000)).toBe('1M context');expect(modelPanelContext(1050000)).toBe('1.1M context');expect(modelPanelContext(undefined)).toBe('');expect(modelPanelContext(-1)).toBe('');
});
test('guarded model panel adapter composes without changing supplied bytes or mutation ownership',()=>{
 const path='web/src/components/compose-box.ts',source=readFileSync(path,'utf8');
 const previous=patchComposeSurface(patchComposeCommands(patchPickerGeometry(patchSkillPrefill(patchUploadCancel(patchModelPicker(patchComposePopupKeys(source)))))));
 const adapted=patchModelPanel(previous);expect(adapted).toContain('openModelSettings(modelHintRef.current)');expect(adapted).toContain('void handleSelectModel(modelOption)');expect(adapted).toContain('disabled=${switchingModel || blocked}');expect(adapted).toContain('Thinking level (read-only)');expect(adapted).not.toContain('Next model\n');
 expect(()=>patchModelPanel(adapted)).toThrow();expect(()=>patchModelPanel('drift')).toThrow();expect(readFileSync(path,'utf8')).toBe(source);
});
