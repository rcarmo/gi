import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {patchPostOutcomes} from '../../../scripts/patch-post-outcomes.mjs';
import {patchPostSpeech} from '../../../scripts/patch-post-speech.mjs';
test('guarded outcome adapter moves existing chips after timestamp without modifying supplied source',()=>{
 const original=readFileSync('web/src/components/post.ts','utf8'),source=patchPostSpeech(original),out=patchPostOutcomes(source);
 expect(out.indexOf('<a class="post-time"')).toBeLessThan(out.indexOf('${recoveryMarker && html`'));
 expect(out.indexOf('${recoveryMarker && html`')).toBeLessThan(out.indexOf('${timeoutMarker && html`'));
 expect(out.match(/class="post-time"/g)).toHaveLength(1);expect(out.match(/formatRecoveryChipTooltip\(recoveryMarker\)/g)).toHaveLength(1);
 expect(()=>patchPostOutcomes(out)).toThrow();expect(()=>patchPostOutcomes(source.replace('class="post-time"','class="changed"'))).toThrow();
 expect(readFileSync('web/src/components/post.ts','utf8')).toBe(original);
});
