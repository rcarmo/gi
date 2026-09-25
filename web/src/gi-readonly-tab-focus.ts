// Event ownership for Gi's read-only host; supplied tabs remain unchanged.
export function readonlyTabDestination(key:string,index:number,count:number):number|null {
    if(count<1||index<0)return null;
    switch(key){case 'Home':return 0;case 'End':return count-1;case 'ArrowRight':return (index+1)%count;case 'ArrowLeft':return (index-1+count)%count;default:return null;}
}
export function bindReadonlyTabKeys(root:HTMLElement,activate:(id:string)=>void,menu:(id:string,x:number,y:number)=>void){
    const onKey=(e:KeyboardEvent)=>{
        if(e.defaultPrevented||e.isComposing||e.keyCode===229||root.closest('[hidden]')||document.querySelector('.settings-dialog[aria-modal="true"]'))return;
        const target=e.target as HTMLElement;
        if(target?.getAttribute('role')!=='tab')return; // Close buttons keep Enter/Space.
        const tabs=Array.from(root.querySelectorAll<HTMLElement>('[role="tab"]')),index=tabs.indexOf(target);
        if(index<0)return;
        if((e.shiftKey&&e.key==='F10')||e.key==='ContextMenu'){
            e.preventDefault();e.stopPropagation();const r=target.getBoundingClientRect();menu(target.dataset.readonlyTabId!,r.left,r.bottom);return;
        }
        if(e.ctrlKey||e.metaKey||e.altKey||e.shiftKey)return;
        const next=readonlyTabDestination(e.key,index,tabs.length);
        if(next!==null){e.preventDefault();e.stopPropagation();activate(tabs[next].dataset.readonlyTabId!);tabs[next].focus({preventScroll:true});}
    };
    root.addEventListener('keydown',onKey);return()=>root.removeEventListener('keydown',onKey);
}

export function bindReadonlyTabMenu(menu:HTMLElement,strip:HTMLElement,id:string){
    const restore=()=>{
        if(document.querySelector('.settings-dialog[aria-modal="true"]')||strip.closest('[hidden]'))return;
        const active=document.activeElement;
        if(active!==document.body&&!menu.contains(active))return;
        const tabs=Array.from(strip.querySelectorAll<HTMLElement>('[role="tab"]'));
        const target=tabs.find(t=>t.dataset.readonlyTabId===id)||tabs.find(t=>t.getAttribute('aria-selected')==='true');
        target?.focus({preventScroll:true});
    };
    if(!document.querySelector('.settings-dialog[aria-modal="true"]'))menu.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus({preventScroll:true});
    // Disposal runs before/after host removal depending on the dismissal path.
    // Defer restoration so close-all can instead focus the visible composer.
    return()=>requestAnimationFrame(restore);
}
