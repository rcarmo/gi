import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {accentForeground} from '../../../web/src/gi-accent-contrast.ts';
import {patchAccentContrast} from '../../../scripts/patch-accent-contrast.mjs';

test('maximum contrast matches pinned black/white foreground across accents',()=>{
 expect(accentForeground({r:29,g:155,b:240})).toBe('#000000');
 expect(accentForeground({r:255,g:0,b:0})).toBe('#000000');
 expect(accentForeground({r:0,g:0,b:255})).toBe('#ffffff');
 expect(accentForeground({r:0,g:0,b:0})).toBe('#ffffff');
 expect(accentForeground({r:255,g:255,b:255})).toBe('#000000');
 const linear=(v:number)=>{const c=v/255;return c<=.03928?c/12.92:((c+.055)/1.055)**2.4;};
 for(let r=0;r<=255;r+=17)for(let g=0;g<=255;g+=17)for(let b=0;b<=255;b+=17){
  const l=.2126*linear(r)+.7152*linear(g)+.0722*linear(b),chosen=accentForeground({r,g,b});
  expect(chosen==='#000000'?(l+.05)/.05:1.05/(l+.05)).toBeGreaterThanOrEqual(4.5);
 }
});
test('contrast adapter is guarded and leaves supplied bytes unchanged',()=>{
 const path='web/src/ui/theme.ts',source=readFileSync(path,'utf8'),adapted=patchAccentContrast(source);
 expect(adapted).toContain('return accentForeground(bg);');expect(adapted).toContain("from '../gi-accent-contrast.js'");
 expect(()=>patchAccentContrast(adapted)).toThrow('anchor changed');expect(()=>patchAccentContrast('drift')).toThrow('anchor changed');
 expect(readFileSync(path,'utf8')).toBe(source);
});
