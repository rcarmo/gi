// @ts-nocheck
const THEME_STORAGE_KEY = 'piclaw_theme';
const TINT_STORAGE_KEY = 'piclaw_tint';

const DEFAULT_LIGHT = {
  bgPrimary: '#ffffff', bgSecondary: '#f7f9fa', bgHover: '#e8ebed', textPrimary: '#0f1419', textSecondary: '#536471', borderColor: '#eff3f4', accent: '#1d9bf0', accentHover: '#1a8cd8', warning: '#f0b429', danger: '#f4212e', success: '#00ba7c',
};
const DEFAULT_DARK = {
  bgPrimary: '#000000', bgSecondary: '#16181c', bgHover: '#1d1f23', textPrimary: '#e7e9ea', textSecondary: '#71767b', borderColor: '#2f3336', accent: '#1d9bf0', accentHover: '#1a8cd8', warning: '#f0b429', danger: '#f4212e', success: '#00ba7c',
};
export const THEME_PRESETS = {
  default: { label: 'Default', mode: 'auto', light: DEFAULT_LIGHT, dark: DEFAULT_DARK },
  dracula: { label: 'Dracula', mode: 'dark', dark: { bgPrimary:'#282a36', bgSecondary:'#303445', bgHover:'#3a3f52', textPrimary:'#f8f8f2', textSecondary:'#c5c8d6', borderColor:'#44475a', accent:'#bd93f9', accentHover:'#a87ded', danger:'#ff5555', success:'#50fa7b' } },
  nord: { label: 'Nord', mode: 'dark', dark: { bgPrimary:'#2e3440', bgSecondary:'#3b4252', bgHover:'#434c5e', textPrimary:'#eceff4', textSecondary:'#d8dee9', borderColor:'#4c566a', accent:'#88c0d0', accentHover:'#78a9c0', danger:'#bf616a', success:'#a3be8c' } },
  github: { label: 'GitHub', mode: 'auto', light: { bgPrimary:'#ffffff', bgSecondary:'#f6f8fa', bgHover:'#eaeef2', textPrimary:'#24292f', textSecondary:'#57606a', borderColor:'#d0d7de', accent:'#0969da', accentHover:'#0550ae', danger:'#cf222e', success:'#1a7f37' }, dark: { bgPrimary:'#0d1117', bgSecondary:'#161b22', bgHover:'#1f242d', textPrimary:'#e6edf3', textSecondary:'#8b949e', borderColor:'#30363d', accent:'#2f81f7', accentHover:'#1f6feb', danger:'#f85149', success:'#3fb950' } },
};
let currentTheme = { theme: 'default', tint: null };
let currentMode = 'light';
let mediaListenerAttached = false;
function getLocalStorageItem(key) { try { return localStorage.getItem(key); } catch { return null; } }
function setLocalStorageItem(key, value) { try { localStorage.setItem(key, value); } catch {} }
function normalizeThemeName(name) { return THEME_PRESETS[name] ? name : 'default'; }
function resolveSystemMode() { return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'; }
function resolvePreset(name) { return THEME_PRESETS[name] || THEME_PRESETS.default; }
function resolveModeForPreset(preset) { return preset.mode === 'auto' ? resolveSystemMode() : preset.mode; }
function resolvePalette(themeName, mode) { const preset = resolvePreset(themeName); return preset[mode] || preset.light || preset.dark || DEFAULT_LIGHT; }
function applyCssVariables(palette) { const root = document.documentElement; const vars = { '--bg-primary': palette.bgPrimary, '--bg-secondary': palette.bgSecondary, '--bg-hover': palette.bgHover, '--text-primary': palette.textPrimary, '--text-secondary': palette.textSecondary, '--border-color': palette.borderColor, '--accent-color': palette.accent, '--accent-hover': palette.accentHover, '--warning-color': palette.warning || DEFAULT_LIGHT.warning, '--danger-color': palette.danger, '--success-color': palette.success }; Object.entries(vars).forEach(([k,v]) => root.style.setProperty(k, v)); }
function applyThemeState(nextTheme, options = {}) { const themeName = normalizeThemeName(nextTheme?.theme || 'default'); currentTheme = { theme: themeName, tint: null }; currentMode = resolveModeForPreset(resolvePreset(themeName)); const root = document.documentElement; root.dataset.theme = currentMode; root.dataset.colorTheme = themeName; root.style.colorScheme = currentMode; if (currentMode === 'dark') { root.classList.add('dark'); root.classList.remove('light'); } else { root.classList.add('light'); root.classList.remove('dark'); } applyCssVariables(resolvePalette(themeName, currentMode)); if (options.persist !== false) setLocalStorageItem(THEME_STORAGE_KEY, themeName); window.dispatchEvent(new CustomEvent('piclaw-theme-change', { detail: { ...currentTheme, mode: currentMode } })); }
function handleSystemThemeChange() { if (resolvePreset(currentTheme.theme).mode !== 'auto') return; applyThemeState(currentTheme, { persist: false }); }
export function initTheme() { const storedTheme = normalizeThemeName(getLocalStorageItem(THEME_STORAGE_KEY) || 'default'); const storedTint = getLocalStorageItem(TINT_STORAGE_KEY) || null; applyThemeState({ theme: storedTheme, tint: storedTint }, { persist: false }); if (window.matchMedia && !mediaListenerAttached) { const media = window.matchMedia('(prefers-color-scheme: dark)'); const fn = () => handleSystemThemeChange(); if (media.addEventListener) media.addEventListener('change', fn); else if (media.addListener) media.addListener(fn); mediaListenerAttached = true; return () => { if (media.removeEventListener) media.removeEventListener('change', fn); else if (media.removeListener) media.removeListener(fn); mediaListenerAttached = false; }; } return () => {}; }
export function cycleThemePreset() { const names = Object.keys(THEME_PRESETS); const currentIndex = names.indexOf(currentTheme.theme); const next = names[(currentIndex + 1) % names.length] || 'default'; applyThemeState({ theme: next }, { persist: true }); return next; }
export function getCurrentThemeLabel() { return resolvePreset(currentTheme.theme).label; }
