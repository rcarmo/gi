import {expect,test} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchUploadCancel} from '../../../scripts/patch-upload-cancel.mjs';
import {patchComposePopupKeys} from '../../../scripts/patch-popup-keys.mjs';
import {patchModelPicker} from '../../../scripts/patch-model-picker.mjs';
test('batch cancel adapter guards capture, each upload, pre-dispatch and cleanup without changing supplied source',()=>{
 const raw=readFileSync('web/src/components/compose-box.ts','utf8'),source=patchModelPicker(patchComposePopupKeys(raw)),patched=patchUploadCancel(source);
 expect(patched).toContain('composeTransfers.beginUploadBatch');expect(patched).toContain('await uploadMedia(file, capturedChatJid, { signal: uploadBatch?.signal })');expect(patched.split('uploadBatch?.signal.throwIfAborted()')).toHaveLength(3);
 expect(patched.indexOf('uploadBatch?.end(); // No cancellation')).toBeLessThan(patched.indexOf('requestDispatched = true;'));
 expect(patched).toContain('Draft and attachments retained; send again to retry.');expect(patched).toContain('} finally {\n                uploadBatch?.end();');
 expect(()=>patchUploadCancel(source.replace('const capturedChatJid = currentChatJid;','drift'))).toThrow('anchor changed');expect(()=>patchUploadCancel(patched)).toThrow();
});
