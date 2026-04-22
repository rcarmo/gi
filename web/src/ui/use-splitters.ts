import { useRef } from '../vendor/preact-htm.js';
import { setLocalStorageItem } from '../utils/storage.js';

export function useSplitters({ appShellRef, sidebarWidthRef, editorWidthRef, dockHeightRef }) {
  const handleSplitterMouseDown = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const startX = e.clientX;
    const startW = sidebarWidthRef.current || 280;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    shell.classList.add('sidebar-resizing');
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    let lastX = startX;
    const onMove = (me) => {
      lastX = me.clientX;
      const w = Math.min(Math.max(startW + (me.clientX - startX), 160), 600);
      shell.style.setProperty('--sidebar-width', `${w}px`);
      sidebarWidthRef.current = w;
    };
    const onUp = () => {
      const w = Math.min(Math.max(startW + (lastX - startX), 160), 600);
      sidebarWidthRef.current = w;
      splitter.classList.remove('dragging');
      shell.classList.remove('sidebar-resizing');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      setLocalStorageItem('sidebarWidth', String(Math.round(w)));
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }).current;

  const handleSplitterTouchStart = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const touch = e.touches[0];
    if (!touch) return;
    const startX = touch.clientX;
    const startW = sidebarWidthRef.current || 280;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    shell.classList.add('sidebar-resizing');
    document.body.style.userSelect = 'none';

    const onMove = (te) => {
      const t = te.touches[0];
      if (!t) return;
      te.preventDefault();
      const w = Math.min(Math.max(startW + (t.clientX - startX), 160), 600);
      shell.style.setProperty('--sidebar-width', `${w}px`);
      sidebarWidthRef.current = w;
    };
    const onUp = () => {
      splitter.classList.remove('dragging');
      shell.classList.remove('sidebar-resizing');
      document.body.style.userSelect = '';
      setLocalStorageItem('sidebarWidth', String(Math.round(sidebarWidthRef.current || startW)));
      document.removeEventListener('touchmove', onMove);
      document.removeEventListener('touchend', onUp);
      document.removeEventListener('touchcancel', onUp);
    };
    document.addEventListener('touchmove', onMove, { passive: false });
    document.addEventListener('touchend', onUp);
    document.addEventListener('touchcancel', onUp);
  }).current;

  const handleEditorSplitterMouseDown = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const startX = e.clientX;
    const startW = editorWidthRef.current || sidebarWidthRef.current || 280;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    let lastX = startX;
    const onMove = (me) => {
      lastX = me.clientX;
      const w = Math.min(Math.max(startW + (me.clientX - startX), 200), 800);
      shell.style.setProperty('--editor-width', `${w}px`);
      editorWidthRef.current = w;
    };
    const onUp = () => {
      const w = Math.min(Math.max(startW + (lastX - startX), 200), 800);
      editorWidthRef.current = w;
      splitter.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      setLocalStorageItem('editorWidth', String(Math.round(w)));
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }).current;

  const handleEditorSplitterTouchStart = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const touch = e.touches[0];
    if (!touch) return;
    const startX = touch.clientX;
    const startW = editorWidthRef.current || sidebarWidthRef.current || 280;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    document.body.style.userSelect = 'none';

    const onMove = (te) => {
      const t = te.touches[0];
      if (!t) return;
      te.preventDefault();
      const w = Math.min(Math.max(startW + (t.clientX - startX), 200), 800);
      shell.style.setProperty('--editor-width', `${w}px`);
      editorWidthRef.current = w;
    };
    const onUp = () => {
      splitter.classList.remove('dragging');
      document.body.style.userSelect = '';
      setLocalStorageItem('editorWidth', String(Math.round(editorWidthRef.current || startW)));
      document.removeEventListener('touchmove', onMove);
      document.removeEventListener('touchend', onUp);
      document.removeEventListener('touchcancel', onUp);
    };
    document.addEventListener('touchmove', onMove, { passive: false });
    document.addEventListener('touchend', onUp);
    document.addEventListener('touchcancel', onUp);
  }).current;

  const handleDockSplitterMouseDown = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const startY = e.clientY;
    const startH = dockHeightRef?.current || 200;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    document.body.style.cursor = 'row-resize';
    document.body.style.userSelect = 'none';

    let lastY = startY;
    const onMove = (me) => {
      lastY = me.clientY;
      // Dragging up increases dock height
      const h = Math.min(Math.max(startH - (me.clientY - startY), 100), window.innerHeight * 0.5);
      shell.style.setProperty('--dock-height', `${h}px`);
      if (dockHeightRef) dockHeightRef.current = h;
      window.dispatchEvent(new CustomEvent('dock-resize'));
    };
    const onUp = () => {
      const h = Math.min(Math.max(startH - (lastY - startY), 100), window.innerHeight * 0.5);
      if (dockHeightRef) dockHeightRef.current = h;
      splitter.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      setLocalStorageItem('dockHeight', String(Math.round(h)));
      // Notify dock pane of resize
      window.dispatchEvent(new CustomEvent('dock-resize'));
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }).current;

  const handleDockSplitterTouchStart = useRef((e) => {
    e.preventDefault();
    const shell = appShellRef.current;
    if (!shell) return;
    const touch = e.touches[0];
    if (!touch) return;
    const startY = touch.clientY;
    const startH = dockHeightRef?.current || 200;
    const splitter = e.currentTarget;
    splitter.classList.add('dragging');
    document.body.style.userSelect = 'none';

    const onMove = (te) => {
      const t = te.touches[0];
      if (!t) return;
      te.preventDefault();
      const h = Math.min(Math.max(startH - (t.clientY - startY), 100), window.innerHeight * 0.5);
      shell.style.setProperty('--dock-height', `${h}px`);
      if (dockHeightRef) dockHeightRef.current = h;
      window.dispatchEvent(new CustomEvent('dock-resize'));
    };
    const onUp = () => {
      splitter.classList.remove('dragging');
      document.body.style.userSelect = '';
      setLocalStorageItem('dockHeight', String(Math.round(dockHeightRef?.current || startH)));
      window.dispatchEvent(new CustomEvent('dock-resize'));
      document.removeEventListener('touchmove', onMove);
      document.removeEventListener('touchend', onUp);
      document.removeEventListener('touchcancel', onUp);
    };
    document.addEventListener('touchmove', onMove, { passive: false });
    document.addEventListener('touchend', onUp);
    document.addEventListener('touchcancel', onUp);
  }).current;

  return {
    handleSplitterMouseDown,
    handleSplitterTouchStart,
    handleEditorSplitterMouseDown,
    handleEditorSplitterTouchStart,
    handleDockSplitterMouseDown,
    handleDockSplitterTouchStart,
  };
}
