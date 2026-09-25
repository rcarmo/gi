// Guarded build adaptation only; retain the supplied theme source verbatim.
export function patchAccentContrast(source){
 const from=`function contrastTextColor(bg) {
    return relativeLuminance(bg) > 0.4 ? '#000000' : '#ffffff';
}`;
 if(source.split(from).length!==2)throw Error('Accent contrast adapter anchor changed');
 return "import { accentForeground } from '../gi-accent-contrast.js';\n"+source.replace(from,`function contrastTextColor(bg) {
    return accentForeground(bg);
}`);
}
