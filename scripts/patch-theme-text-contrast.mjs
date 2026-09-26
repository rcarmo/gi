// Match pinned palette text contrast while preserving supplied theme sources.
export function patchThemeTextContrast(source){
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Theme text contrast anchor changed: ${from.slice(0,80)}`);result=result.replace(from,to);};
 replace('    const vars = {',`    const textColours = themeTextColours(palette, mode);
    const vars = {`);
 replace("        '--text-primary': palette.textPrimary,\n        '--text-secondary': palette.textSecondary,","        '--text-primary': textColours.primary,\n        '--text-secondary': textColours.secondary,");
 replace(`    if (themeName === 'default' && !tint) {
        clearCssVariables();
    } else {`, `    if (themeName === 'default' && !tint) {
        clearCssVariables();
        const textColours = themeTextColours(palette, mode);
        root.style.setProperty('--text-primary', textColours.primary);
        root.style.setProperty('--text-secondary', textColours.secondary);
    } else {`);
 return `import { themeTextColours } from '../gi-theme-text-contrast.js';\n${result}`;
}
