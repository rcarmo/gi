type RGB = {r:number;g:number;b:number};
const hex = (value:string):RGB|null => {
 const input=String(value||'').trim();if(!input.startsWith('#'))return null;
 const raw=input.slice(1);
 if(!/^(?:[0-9a-f]{3}|[0-9a-f]{6})$/i.test(raw))return null;
 const full=raw.length===3?[...raw].map(c=>c+c).join(''):raw;
 const n=parseInt(full,16);return {r:n>>16&255,g:n>>8&255,b:n&255};
};
const colour = (value:string):RGB|null => {
 const direct=hex(value);if(direct)return direct;
 const m=String(value||'').trim().match(/^rgba?\(\s*([\d.]+)[, ]+([\d.]+)[, ]+([\d.]+)(?:\s*[,/]\s*[\d.]+)?\s*\)$/i);
 if(!m)return null;
 const [r,g,b]=m.slice(1,4).map(Number);return {r,g,b};
};
const mix=(a:RGB,b:RGB,t:number):RGB=>({r:Math.round(a.r*(1-t)+b.r*t),g:Math.round(a.g*(1-t)+b.g*t),b:Math.round(a.b*(1-t)+b.b*t)});
const encode=(c:RGB)=>`#${[c.r,c.g,c.b].map(v=>Math.round(v).toString(16).padStart(2,'0')).join('')}`;
const luminance=(c:RGB)=>{
 const linear=(v:number)=>{v/=255;return v<=0.04045?v/12.92:((v+0.055)/1.055)**2.4;};
 return 0.2126*linear(c.r)+0.7152*linear(c.g)+0.0722*linear(c.b);
};
export function themeTextContrastRatio(a:RGB,b:RGB){const x=luminance(a),y=luminance(b);return (Math.max(x,y)+0.05)/(Math.min(x,y)+0.05);}
// Pinned Classic gc(): first 1% rounded RGB step meeting all backgrounds.
// Normalise valid colours to hex at step zero; malformed inputs remain unchanged.
export function readableThemeText(value:string,target:string,backgrounds:string[],minimum=4.5):string {
 const base=colour(value),end=colour(target),surfaces=backgrounds.map(colour);
 if(!base||!end||surfaces.some(c=>!c))return value;
 const meets=(c:RGB)=>surfaces.every(bg=>themeTextContrastRatio(c,bg!)>=minimum);
 for(let step=0;step<=100;step++){const candidate=mix(base,end,step/100);if(meets(candidate))return encode(candidate);}
 return target;
}
export function themeTextColours(palette:any,mode:string){
 const backgrounds=[palette.bgPrimary,palette.bgSecondary,palette.bgHover||palette.bgSecondary];
 const primary=readableThemeText(palette.textPrimary,palette.monochrome?palette.textPrimary:mode==='dark'?'#ffffff':'#000000',backgrounds);
 return {primary,secondary:readableThemeText(palette.textSecondary,primary,backgrounds)};
}
