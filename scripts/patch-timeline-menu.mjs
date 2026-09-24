// Guarded dismissal-only adaptation; the supplied component stays verbatim.
export function patchTimelineMenu(source) {
  const from = `    useEffect(() => {
        if (!open) return;
        const onClick = (e) => {
            if (menuRef.current?.contains(e.target)) return;
            if (btnRef.current?.contains(e.target)) return;
            setOpen(false);
        };
        document.addEventListener('mousedown', onClick, true);
        return () => document.removeEventListener('mousedown', onClick, true);
    }, [open]);

    useEffect(() => {
        if (!open) return;
        const onKey = (e) => { if (e.key === 'Escape') setOpen(false); };
        document.addEventListener('keydown', onKey);
        return () => document.removeEventListener('keydown', onKey);
    }, [open]);`;
  // The portal render's layout effect must run first so refs point to the new
  // menu before listeners are installed, including its very first frame.
  const anchor = '    return null;';
  if (source.split(from).length !== 2 || source.split(anchor).length !== 2)
    throw new Error('Timeline menu adapter anchor changed');
  return "import { bindMenuDismissal } from '../gi-menu-dismissal.js';\n" + source.replace(from, '').replace(anchor, `    useLayoutEffect(() => {
        if (!open || !menuRef.current || !btnRef.current) return;
        return bindMenuDismissal(menuRef.current, btnRef.current, () => setOpen(false));
    }, [open]);

${anchor}`);
}
