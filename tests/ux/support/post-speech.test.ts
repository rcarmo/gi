import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {createSpeechPlayback,buildSpeakablePostText} from '../../../web/src/gi-post-speech.ts';
import {patchPostSpeech} from '../../../scripts/patch-post-speech.mjs';
function fixture(){
 const spoken:any[]=[],calls:string[]=[];let throws=false;
 const win={SpeechSynthesisUtterance:class {text:string;constructor(text:string){this.text=text;}},speechSynthesis:{speak(u:any){calls.push('speak');if(throws)throw Error('denied');spoken.push(u);},cancel(){calls.push('cancel');spoken.at(-1)?.onend?.();}}};
 const playback=createSpeechPlayback(()=>win);playback.setScope('main');return{playback,spoken,calls,deny:()=>throws=true};
}
test('bounded speech text follows frozen Markdown normalisation and omits fenced code',()=>{
 expect(buildSpeakablePostText({data:{content:'# Heading\n\n[link](https://example.invalid) `inline`\n```js\nsecret();\n```'}})).toBe('Heading. link inline Code block omitted.');
 expect(buildSpeakablePostText({data:{content:' '.repeat(10)}})).toBe('');
 expect(buildSpeakablePostText({data:{content:'a'.repeat(3000)}})).toHaveLength(1600);
});
test('speech owns a mounted occurrence, fences synchronous cancel and late callbacks',()=>{
 const {playback:p,spoken,calls}=fixture(),a={},b={};let latest:any;const unsubscribe=p.subscribe(s=>latest=s);
 expect(p.speak(a,'first')).toBe(true);const old=spoken[0];expect(latest.owner).toBe(a);
 expect(p.speak(b,'second')).toBe(true);expect(calls).toEqual(['speak','cancel','speak']);expect(latest.owner).toBe(b);
 old.onend();old.onerror();expect(p.state().owner).toBe(b);p.stop(a);expect(p.state().owner).toBe(b);
 p.stop(b);expect(p.state().speaking).toBe(false);p.speak(b,'same id new occurrence');spoken[1].onend();expect(p.state().owner).toBe(b);
 p.setScope('research');expect(p.state().speaking).toBe(false);spoken.at(-1).onend();expect(latest.speaking).toBe(false);unsubscribe();
});
test('capability, empty input, exceptions and ended speech fail without sticky active state',()=>{
 const unavailable=createSpeechPlayback(()=>({speechSynthesis:{}}));unavailable.setScope('main');expect(unavailable.supported()).toBe(false);expect(unavailable.speak({},'text')).toBe(false);
 const {playback:p,spoken,deny}=fixture();expect(p.speak({},' ')).toBe(false);p.speak({},'text');spoken[0].onerror();expect(p.state().speaking).toBe(false);deny();expect(p.speak({},'denied')).toBe(false);expect(p.state().speaking).toBe(false);p.setScope(null);expect(p.speak({},'no selection')).toBe(false);
});
test('speech adapter keeps supplied Post unchanged and rejects missing or duplicate anchors',()=>{
 const source=readFileSync('web/src/components/post.ts','utf8'),patched=patchPostSpeech(source);
 expect(patched).toContain('isAgent && speechSupported && speakableText');expect(patched).toContain('aria-pressed=');expect(patched).toContain('speechPlayback.stop(speechOwner)');
 expect(()=>patchPostSpeech(patched)).toThrow();expect(()=>patchPostSpeech(source.replace('                <div class="post-actions">','changed'))).toThrow();
 expect(readFileSync('web/src/components/post.ts','utf8')).toBe(source);
});
