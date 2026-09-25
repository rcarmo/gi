// Match pinned Classic themeForeground: select the greater WCAG black/white
// contrast, rather than treating mid-luminance accents as uniformly dark.
export function accentForeground({r,g,b}:{r:number;g:number;b:number}):string{
 const luminance=[r,g,b].map(value=>{const channel=value/255;return channel<=0.03928?channel/12.92:((channel+0.055)/1.055)**2.4;});
 const l=luminance[0]*0.2126+luminance[1]*0.7152+luminance[2]*0.0722;
 const black=(l+0.05)/0.05,white=1.05/(l+0.05);
 return black>white?'#000000':'#ffffff';
}
