import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {themeTextColours,readableThemeText,themeTextContrastRatio} from '../../../web/src/gi-theme-text-contrast.ts';
import {patchThemeTextContrast} from '../../../scripts/patch-theme-text-contrast.mjs';
import {patchAccentContrast} from '../../../scripts/patch-accent-contrast.mjs';
const dark={bgPrimary:'#000000',bgSecondary:'#16181c',bgHover:'#1d1f23',textPrimary:'#e7e9ea',textSecondary:'#71767b'};
const light={bgPrimary:'#ffffff',bgSecondary:'#f7f9fa',bgHover:'#e8ebed',textPrimary:'#0f1419',textSecondary:'#536471'};
const rgb=(s:string)=>({r:parseInt(s.slice(1,3),16),g:parseInt(s.slice(3,5),16),b:parseInt(s.slice(5,7),16)});
test('pinned Classic text contrast uses all three surfaces and the first rounded 1% step',()=>{
 expect(themeTextColours(dark,'dark')).toEqual({primary:'#e7e9ea',secondary:'#82868b'});
 expect(themeTextColours(light,'light')).toEqual({primary:'#0f1419',secondary:'#536471'});
 const adjusted=rgb('#82868b'),previous=rgb('#808589');
 for(const bg of [dark.bgPrimary,dark.bgSecondary,dark.bgHover])expect(themeTextContrastRatio(adjusted,rgb(bg))).toBeGreaterThanOrEqual(4.5);
 expect(themeTextContrastRatio(previous,rgb(dark.bgHover))).toBeLessThan(4.5);
 expect(readableThemeText('#fff','#000',['#fff'])).toBe('#757575');
});
test('generated tint RGB, missing hover and invalid colours keep deterministic fallbacks',()=>{
 expect(readableThemeText('rgb(113 118 123)','#e7e9ea',['#000','#16181c','#1d1f23'])).toBe('#82868b');
 expect(readableThemeText('rgba(113, 118, 123, 1)','#e7e9ea',['#000','#16181c','#1d1f23'])).toBe('#82868b');
 expect(readableThemeText('invalid','#fff',['#000'])).toBe('invalid');
 expect(readableThemeText('#777','invalid',['#000'])).toBe('#777');
 expect(readableThemeText('#777','#fff',['invalid'])).toBe('#777');
 expect(readableThemeText('#777','#fff',['#000','#fff'])).toBe('#fff'); // No possible contrast: pinned target fallback.
 expect(themeTextColours({...light,bgHover:undefined},'light')).toEqual({primary:'#0f1419',secondary:'#536471'});
 expect(themeTextColours({...dark,monochrome:true},'dark').primary).toBe(dark.textPrimary);
});
test('text adapter composes with accent contrast and leaves supplied bytes intact',()=>{
 const path='web/src/ui/theme.ts',source=readFileSync(path,'utf8'),result=patchThemeTextContrast(patchAccentContrast(source));
 expect(result).toContain("'--text-secondary': textColours.secondary");expect(result).toContain("'--text-primary': textColours.primary");expect(result).toContain('return accentForeground(bg)');
 expect(()=>patchThemeTextContrast(result)).toThrow();expect(()=>patchThemeTextContrast('drift')).toThrow();expect(readFileSync(path,'utf8')).toBe(source);
});
