// The native catalogue is the only source of autocomplete commands. Preserve
// order, but reject malformed payloads rather than expose bundled fallbacks.
export function normaliseComposeCommands(data: unknown): {name:string,description:string}[] {
    const commands=(data as any)?.commands;
    if(!Array.isArray(commands))throw new Error('Invalid command catalogue');
    const seen=new Set<string>();
    return commands.map(command=>{
        if(typeof command?.name!=='string' || !/^\/[a-z][a-z0-9:_-]*$/i.test(command.name)
            || (command.description!==undefined && typeof command.description!=='string'))throw new Error('Invalid command catalogue');
        return {name:command.name,description:command.description||''};
    }).filter(command=>{if(seen.has(command.name))return false;seen.add(command.name);return true;});
}

export function declineComposeKey(event: Pick<KeyboardEvent,'defaultPrevented'|'isComposing'|'keyCode'|'repeat'|'key'>): boolean {
    return event.defaultPrevented || event.isComposing || event.keyCode===229 || (event.repeat && (event.key==='Enter'||event.key==='Tab'));
}
