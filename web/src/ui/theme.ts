// @ts-nocheck
import { getLocalStorageItem, setLocalStorageItem } from '../utils/storage.js';

const THEME_STORAGE_KEY = 'piclaw_theme';
const TINT_STORAGE_KEY = 'piclaw_tint';
const CHAT_THEMES_STORAGE_KEY = 'piclaw_chat_themes';

const DEFAULT_LIGHT = {
    bgPrimary: '#ffffff',
    bgSecondary: '#f7f9fa',
    bgHover: '#e8ebed',
    textPrimary: '#0f1419',
    textSecondary: '#536471',
    borderColor: '#eff3f4',
    accent: '#1d9bf0',
    accentHover: '#1a8cd8',
    warning: '#f0b429',
    danger: '#f4212e',
    success: '#00ba7c',
};

const DEFAULT_DARK = {
    bgPrimary: '#000000',
    bgSecondary: '#16181c',
    bgHover: '#1d1f23',
    textPrimary: '#e7e9ea',
    textSecondary: '#71767b',
    borderColor: '#2f3336',
    accent: '#1d9bf0',
    accentHover: '#1a8cd8',
    warning: '#f0b429',
    danger: '#f4212e',
    success: '#00ba7c',
};

const THEME_PRESETS = {
    default: {
        label: 'Default',
        mode: 'auto',
        light: DEFAULT_LIGHT,
        dark: DEFAULT_DARK,
    },
    tango: {
        label: 'Tango',
        mode: 'light',
        light: {
            bgPrimary: '#f6f5f4',
            bgSecondary: '#efedeb',
            bgHover: '#e5e3e1',
            textPrimary: '#2e3436',
            textSecondary: '#5c6466',
            borderColor: '#d3d7cf',
            accent: '#3465a4',
            accentHover: '#2c5890',
            danger: '#cc0000',
            success: '#4e9a06',
        },
    },
    xterm: {
        label: 'XTerm',
        mode: 'dark',
        dark: {
            bgPrimary: '#000000',
            bgSecondary: '#0a0a0a',
            bgHover: '#121212',
            textPrimary: '#d0d0d0',
            textSecondary: '#8a8a8a',
            borderColor: '#1f1f1f',
            accent: '#00a2ff',
            accentHover: '#0086d1',
            danger: '#ff5f5f',
            success: '#5fff87',
        },
    },
    monokai: {
        label: 'Monokai',
        mode: 'dark',
        dark: {
            bgPrimary: '#272822',
            bgSecondary: '#2f2f2f',
            bgHover: '#3a3a3a',
            textPrimary: '#f8f8f2',
            textSecondary: '#cfcfc2',
            borderColor: '#3e3d32',
            accent: '#f92672',
            accentHover: '#e81560',
            danger: '#f92672',
            success: '#a6e22e',
        },
    },
    'monokai-pro': {
        label: 'Monokai Pro',
        mode: 'dark',
        dark: {
            bgPrimary: '#2d2a2e',
            bgSecondary: '#363237',
            bgHover: '#403a40',
            textPrimary: '#fcfcfa',
            textSecondary: '#c1c0c0',
            borderColor: '#444046',
            accent: '#ff6188',
            accentHover: '#f74f7e',
            danger: '#ff4f5e',
            success: '#a9dc76',
        },
    },
    ristretto: {
        label: 'Ristretto',
        mode: 'dark',
        dark: {
            bgPrimary: '#2c2525',
            bgSecondary: '#362d2d',
            bgHover: '#403535',
            textPrimary: '#f4f1ef',
            textSecondary: '#cbbdb8',
            borderColor: '#4a3c3c',
            accent: '#ff9f43',
            accentHover: '#f28a2e',
            danger: '#ff5f56',
            success: '#a9dc76',
        },
    },
    dracula: {
        label: 'Dracula',
        mode: 'dark',
        dark: {
            bgPrimary: '#282a36',
            bgSecondary: '#303445',
            bgHover: '#3a3f52',
            textPrimary: '#f8f8f2',
            textSecondary: '#c5c8d6',
            borderColor: '#44475a',
            accent: '#bd93f9',
            accentHover: '#a87ded',
            danger: '#ff5555',
            success: '#50fa7b',
        },
    },
    catppuccin: {
        label: 'Catppuccin',
        mode: 'dark',
        dark: {
            bgPrimary: '#1e1e2e',
            bgSecondary: '#24273a',
            bgHover: '#2c2f41',
            textPrimary: '#cdd6f4',
            textSecondary: '#a6adc8',
            borderColor: '#313244',
            accent: '#89b4fa',
            accentHover: '#74a0f5',
            danger: '#f38ba8',
            success: '#a6e3a1',
        },
    },
    nord: {
        label: 'Nord',
        mode: 'dark',
        dark: {
            bgPrimary: '#2e3440',
            bgSecondary: '#3b4252',
            bgHover: '#434c5e',
            textPrimary: '#eceff4',
            textSecondary: '#d8dee9',
            borderColor: '#4c566a',
            accent: '#88c0d0',
            accentHover: '#78a9c0',
            danger: '#bf616a',
            success: '#a3be8c',
        },
    },
    gruvbox: {
        label: 'Gruvbox',
        mode: 'dark',
        dark: {
            bgPrimary: '#282828',
            bgSecondary: '#32302f',
            bgHover: '#3c3836',
            textPrimary: '#ebdbb2',
            textSecondary: '#bdae93',
            borderColor: '#3c3836',
            accent: '#d79921',
            accentHover: '#c28515',
            danger: '#fb4934',
            success: '#b8bb26',
        },
    },
    solarized: {
        label: 'Solarized',
        mode: 'auto',
        light: {
            bgPrimary: '#fdf6e3',
            bgSecondary: '#f5efdc',
            bgHover: '#eee8d5',
            textPrimary: '#586e75',
            textSecondary: '#657b83',
            borderColor: '#e0d8c6',
            accent: '#268bd2',
            accentHover: '#1f78b3',
            danger: '#dc322f',
            success: '#859900',
        },
        dark: {
            bgPrimary: '#002b36',
            bgSecondary: '#073642',
            bgHover: '#0b3c4a',
            textPrimary: '#eee8d5',
            textSecondary: '#93a1a1',
            borderColor: '#18424a',
            accent: '#268bd2',
            accentHover: '#1f78b3',
            danger: '#dc322f',
            success: '#859900',
        },
    },
    tokyo: {
        label: 'Tokyo',
        mode: 'dark',
        dark: {
            bgPrimary: '#1a1b26',
            bgSecondary: '#24283b',
            bgHover: '#2f3549',
            textPrimary: '#c0caf5',
            textSecondary: '#9aa5ce',
            borderColor: '#414868',
            accent: '#7aa2f7',
            accentHover: '#6b92e6',
            danger: '#f7768e',
            success: '#9ece6a',
        },
    },
    miasma: {
        label: 'Miasma',
        mode: 'dark',
        dark: {
            bgPrimary: '#1f1f23',
            bgSecondary: '#29292f',
            bgHover: '#33333a',
            textPrimary: '#e5e5e5',
            textSecondary: '#b4b4b4',
            borderColor: '#3d3d45',
            accent: '#c9739c',
            accentHover: '#b8618c',
            danger: '#e06c75',
            success: '#98c379',
        },
    },
    github: {
        label: 'GitHub',
        mode: 'auto',
        light: {
            bgPrimary: '#ffffff',
            bgSecondary: '#f6f8fa',
            bgHover: '#eaeef2',
            textPrimary: '#24292f',
            textSecondary: '#57606a',
            borderColor: '#d0d7de',
            accent: '#0969da',
            accentHover: '#0550ae',
            danger: '#cf222e',
            success: '#1a7f37',
        },
        dark: {
            bgPrimary: '#0d1117',
            bgSecondary: '#161b22',
            bgHover: '#21262d',
            textPrimary: '#c9d1d9',
            textSecondary: '#8b949e',
            borderColor: '#30363d',
            accent: '#2f81f7',
            accentHover: '#1f6feb',
            danger: '#f85149',
            success: '#3fb950',
        },
    },
    gotham: {
        label: 'Gotham',
        mode: 'dark',
        dark: {
            bgPrimary: '#0b0f14',
            bgSecondary: '#111720',
            bgHover: '#18212b',
            textPrimary: '#cbd6e2',
            textSecondary: '#9bb0c3',
            borderColor: '#1f2a37',
            accent: '#5ccfe6',
            accentHover: '#48b8ce',
            danger: '#d26937',
            success: '#2aa889',
        },
    },
};

const THEME_VAR_KEYS = [
    '--bg-primary',
    '--bg-secondary',
    '--bg-hover',
    '--text-primary',
    '--text-secondary',
    '--border-color',
    '--accent-color',
    '--accent-hover',
    '--accent-color-alpha',
    '--accent-contrast-text',
    '--accent-soft',
    '--accent-soft-strong',
    '--warning-color',
    '--danger-color',
    '--success-color',
    '--search-highlight-color',
];

let currentTheme = {
    theme: 'default',
    tint: null,
};
let currentMode = 'light';
let mediaListenerAttached = false;

function normalizeThemeName(value) {
    const raw = String(value || '').trim().toLowerCase();
    if (!raw) return 'default';
    if (raw === 'solarized-dark' || raw === 'solarized-light') return 'solarized';
    if (raw === 'github-dark' || raw === 'github-light') return 'github';
    if (raw === 'tokyo-night') return 'tokyo';
    return raw;
}

function parseHexColor(input) {
    if (!input) return null;
    const raw = String(input).trim();
    if (!raw) return null;
    const hex = raw.startsWith('#') ? raw.slice(1) : raw;
    if (!/^[0-9a-fA-F]{3}$/.test(hex) && !/^[0-9a-fA-F]{6}$/.test(hex)) return null;
    const full = hex.length === 3
        ? hex.split('').map((c) => c + c).join('')
        : hex;
    const int = parseInt(full, 16);
    return {
        r: (int >> 16) & 255,
        g: (int >> 8) & 255,
        b: int & 255,
        hex: `#${full.toLowerCase()}`,
    };
}

function resolveComputedCssColor(el, fallbackColor) {
    try {
        if (document.body) {
            el.style.display = 'none';
            document.body.appendChild(el);
            const computed = getComputedStyle(el).color || el.style.color;
            document.body.removeChild(el);
            return computed;
        }
    } catch {
        return fallbackColor;
    }
    return fallbackColor;
}

function parseCssColor(input) {
    if (!input || typeof document === 'undefined') return null;
    const raw = String(input).trim();
    if (!raw) return null;

    const el = document.createElement('div');
    el.style.color = '';
    el.style.color = raw;
    if (!el.style.color) return null;

    const computed = resolveComputedCssColor(el, el.style.color);

    const match = computed.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/i);
    if (!match) return null;
    const r = parseInt(match[1], 10);
    const g = parseInt(match[2], 10);
    const b = parseInt(match[3], 10);
    if (![r, g, b].every((v) => Number.isFinite(v))) return null;
    const hex = `#${[r, g, b].map((v) => v.toString(16).padStart(2, '0')).join('')}`;
    return { r, g, b, hex };
}

function parseColor(input) {
    return parseHexColor(input) || parseCssColor(input);
}

function mixColors(base, overlay, ratio) {
    const r = Math.round(base.r + (overlay.r - base.r) * ratio);
    const g = Math.round(base.g + (overlay.g - base.g) * ratio);
    const b = Math.round(base.b + (overlay.b - base.b) * ratio);
    return `rgb(${r} ${g} ${b})`;
}

function rgbaColor(color, alpha) {
    return `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`;
}

/** WCAG relative luminance (0 = black, 1 = white). */
function relativeLuminance(c) {
    const rs = c.r / 255, gs = c.g / 255, bs = c.b / 255;
    const r = rs <= 0.03928 ? rs / 12.92 : Math.pow((rs + 0.055) / 1.055, 2.4);
    const g = gs <= 0.03928 ? gs / 12.92 : Math.pow((gs + 0.055) / 1.055, 2.4);
    const b = bs <= 0.03928 ? bs / 12.92 : Math.pow((bs + 0.055) / 1.055, 2.4);
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** Return white or black text for maximum contrast against the given background. */
function contrastTextColor(bg) {
    return relativeLuminance(bg) > 0.4 ? '#000000' : '#ffffff';
}

function resolveSystemMode() {
    if (typeof window === 'undefined') return 'light';
    try {
        return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
            ? 'dark'
            : 'light';
    } catch {
        return 'light';
    }
}

function resolvePreset(themeName) {
    return THEME_PRESETS[themeName] || THEME_PRESETS.default;
}

function resolveModeForPreset(preset) {
    return preset.mode === 'auto' ? resolveSystemMode() : preset.mode;
}

function resolvePalette(themeName, mode) {
    const preset = resolvePreset(themeName);
    if (mode === 'dark' && preset.dark) return preset.dark;
    if (mode === 'light' && preset.light) return preset.light;
    return preset.dark || preset.light || DEFAULT_LIGHT;
}

function tintPaletteColor(value, tint, ratio) {
    const base = parseColor(value);
    if (!base) return value;
    return mixColors(base, tint, ratio);
}

function buildTintedPalette(basePalette, tintHex, mode) {
    const tint = parseColor(tintHex);
    if (!tint) return basePalette;

    const contrastColor = mode === 'dark' ? '#ffffff' : '#000000';
    const contrast = parseHexColor(contrastColor);

    return {
        ...basePalette,
        bgPrimary: tintPaletteColor(basePalette.bgPrimary, tint, 0.08),
        bgSecondary: tintPaletteColor(basePalette.bgSecondary, tint, 0.12),
        bgHover: tintPaletteColor(basePalette.bgHover, tint, 0.16),
        textPrimary: tintPaletteColor(basePalette.textPrimary, tint, mode === 'dark' ? 0.08 : 0.06),
        textSecondary: tintPaletteColor(basePalette.textSecondary, tint, mode === 'dark' ? 0.12 : 0.1),
        borderColor: tintPaletteColor(basePalette.borderColor, tint, 0.1),
        accent: tint.hex,
        accentHover: contrast ? mixColors(tint, contrast, 0.18) : tint.hex,
        warning: tintPaletteColor(basePalette.warning || DEFAULT_LIGHT.warning, tint, 0.14),
        danger: tintPaletteColor(basePalette.danger, tint, 0.16),
        success: tintPaletteColor(basePalette.success, tint, 0.16),
    };
}

function resolveWarningColor(palette, mode) {
    const explicit = parseColor(palette?.warning);
    if (explicit) return explicit.hex;

    const defaultWarning = parseColor(mode === 'dark' ? DEFAULT_DARK.warning : DEFAULT_LIGHT.warning)
        || parseColor(DEFAULT_LIGHT.warning);
    const accent = parseColor(palette?.accent);

    if (defaultWarning && accent) {
        return mixColors(defaultWarning, accent, mode === 'dark' ? 0.18 : 0.14);
    }

    return mode === 'dark' ? DEFAULT_DARK.warning : DEFAULT_LIGHT.warning;
}

function applyCssVariables(palette, mode) {
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    const accentColor = palette.accent;
    const accentHex = parseColor(accentColor);
    const searchHighlight = accentHex
        ? rgbaColor(accentHex, mode === 'dark' ? 0.35 : 0.2)
        : palette.searchHighlight || palette.searchHighlightColor;
    const accentSoft = accentHex
        ? rgbaColor(accentHex, mode === 'dark' ? 0.16 : 0.12)
        : 'rgba(29, 155, 240, 0.12)';
    const accentSoftStrong = accentHex
        ? rgbaColor(accentHex, mode === 'dark' ? 0.28 : 0.2)
        : 'rgba(29, 155, 240, 0.2)';
    const accentContrastText = accentHex
        ? contrastTextColor(accentHex)
        : (mode === 'dark' ? '#000000' : '#ffffff');
    const accentColorAlpha = accentHex
        ? rgbaColor(accentHex, mode === 'dark' ? 0.35 : 0.25)
        : 'rgba(29, 155, 240, 0.25)';
    const warningColor = resolveWarningColor(palette, mode);

    const vars = {
        '--bg-primary': palette.bgPrimary,
        '--bg-secondary': palette.bgSecondary,
        '--bg-hover': palette.bgHover,
        '--text-primary': palette.textPrimary,
        '--text-secondary': palette.textSecondary,
        '--border-color': palette.borderColor,
        '--accent-color': accentColor,
        '--accent-hover': palette.accentHover || accentColor,
        '--accent-color-alpha': accentColorAlpha,
        '--accent-soft': accentSoft,
        '--accent-soft-strong': accentSoftStrong,
        '--accent-contrast-text': accentContrastText,
        '--warning-color': warningColor,
        '--danger-color': palette.danger || DEFAULT_LIGHT.danger,
        '--success-color': palette.success || DEFAULT_LIGHT.success,
        '--search-highlight-color': searchHighlight || 'rgba(29, 155, 240, 0.2)',
    };

    Object.entries(vars).forEach(([key, value]) => {
        if (value) root.style.setProperty(key, value);
    });
}

function clearCssVariables() {
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    THEME_VAR_KEYS.forEach((key) => root.style.removeProperty(key));
}

function ensureMetaTag(name, options = {}) {
    if (typeof document === 'undefined') return null;
    const id = typeof options.id === 'string' && options.id.trim() ? options.id.trim() : null;
    let tag = id
        ? document.getElementById(id)
        : document.querySelector(`meta[name="${name}"]`);
    if (!tag) {
        tag = document.createElement('meta');
        document.head.appendChild(tag);
    }
    tag.setAttribute('name', name);
    if (id) tag.setAttribute('id', id);
    return tag;
}

function resolveThemeColorForMode(mode) {
    const themeName = normalizeThemeName(currentTheme?.theme || 'default');
    const tint = currentTheme?.tint ? String(currentTheme.tint).trim() : null;
    let palette = resolvePalette(themeName, mode);
    if (themeName === 'default' && tint) {
        palette = buildTintedPalette(palette, tint, mode);
    }
    if (palette?.bgPrimary) return palette.bgPrimary;
    return mode === 'dark' ? DEFAULT_DARK.bgPrimary : DEFAULT_LIGHT.bgPrimary;
}

function updateMetaColor(color, mode) {
    if (typeof document === 'undefined') return;

    const themeMeta = ensureMetaTag('theme-color', { id: 'dynamic-theme-color' });
    if (themeMeta && color) {
        themeMeta.removeAttribute('media');
        themeMeta.setAttribute('content', color);
    }

    const lightThemeMeta = ensureMetaTag('theme-color', { id: 'theme-color-light' });
    if (lightThemeMeta) {
        lightThemeMeta.setAttribute('media', '(prefers-color-scheme: light)');
        lightThemeMeta.setAttribute('content', resolveThemeColorForMode('light'));
    }

    const darkThemeMeta = ensureMetaTag('theme-color', { id: 'theme-color-dark' });
    if (darkThemeMeta) {
        darkThemeMeta.setAttribute('media', '(prefers-color-scheme: dark)');
        darkThemeMeta.setAttribute('content', resolveThemeColorForMode('dark'));
    }

    const tileMeta = ensureMetaTag('msapplication-TileColor');
    if (tileMeta && color) tileMeta.setAttribute('content', color);

    const navMeta = ensureMetaTag('msapplication-navbutton-color');
    if (navMeta && color) navMeta.setAttribute('content', color);

    const statusMeta = ensureMetaTag('apple-mobile-web-app-status-bar-style');
    if (statusMeta) statusMeta.setAttribute('content', mode === 'dark' ? 'black-translucent' : 'default');
}

function emitThemeChange() {
    if (typeof window === 'undefined') return;
    const detail = { ...currentTheme, mode: currentMode };
    window.dispatchEvent(new CustomEvent('piclaw-theme-change', { detail }));
}

function getChatThemeMap() {
    try {
        const raw = getLocalStorageItem(CHAT_THEMES_STORAGE_KEY);
        if (!raw) return {};
        const parsed = JSON.parse(raw);
        return typeof parsed === 'object' && parsed !== null ? parsed : {};
    } catch {
        return {};
    }
}

function setChatTheme(chatJid, theme, tint) {
    const map = getChatThemeMap();
    if (!theme && !tint) {
        delete map[chatJid];
    } else {
        map[chatJid] = { theme: theme || 'default', tint: tint || null };
    }
    setLocalStorageItem(CHAT_THEMES_STORAGE_KEY, JSON.stringify(map));
}

function getChatTheme(chatJid) {
    if (!chatJid) return null;
    const map = getChatThemeMap();
    return map[chatJid] || null;
}

function resolveCurrentChatJid() {
    if (typeof window === 'undefined') return 'web:default';
    try {
        const params = new URL(window.location.href).searchParams;
        const raw = params.get('chat_jid');
        return raw && raw.trim() ? raw.trim() : 'web:default';
    } catch {
        return 'web:default';
    }
}

function applyThemeState(nextTheme, options = {}) {
    if (typeof window === 'undefined' || typeof document === 'undefined') return;
    const themeName = normalizeThemeName(nextTheme?.theme || 'default');
    const tint = nextTheme?.tint ? String(nextTheme.tint).trim() : null;
    const preset = resolvePreset(themeName);
    const mode = resolveModeForPreset(preset);
    const paletteBase = resolvePalette(themeName, mode);

    currentTheme = { theme: themeName, tint };
    currentMode = mode;

    const root = document.documentElement;
    root.dataset.theme = mode;
    root.dataset.colorTheme = themeName;
    root.dataset.tint = tint ? String(tint) : '';
    root.style.colorScheme = mode;

    let palette = paletteBase;
    if (themeName === 'default' && tint) {
        palette = buildTintedPalette(paletteBase, tint, mode);
    }

    if (themeName === 'default' && !tint) {
        clearCssVariables();
    } else {
        applyCssVariables(palette, mode);
    }

    updateMetaColor(palette.bgPrimary, mode);
    emitThemeChange();

    if (options.persist !== false) {
        setLocalStorageItem(THEME_STORAGE_KEY, themeName);
        if (tint) setLocalStorageItem(TINT_STORAGE_KEY, tint);
        else setLocalStorageItem(TINT_STORAGE_KEY, '');
    }
}

function handleSystemThemeChange() {
    const preset = resolvePreset(currentTheme.theme);
    if (preset.mode !== 'auto') return;
    applyThemeState(currentTheme, { persist: false });
}

export function initTheme() {
    if (typeof window === 'undefined') return () => {};

    // Resolve per-chat theme override first, fall back to global
    const chatJid = resolveCurrentChatJid();
    const chatOverride = getChatTheme(chatJid);

    const storedTheme = chatOverride
        ? normalizeThemeName(chatOverride.theme || 'default')
        : normalizeThemeName(getLocalStorageItem(THEME_STORAGE_KEY) || 'default');
    const storedTint = chatOverride
        ? (chatOverride.tint ? String(chatOverride.tint).trim() : null)
        : (() => { const raw = getLocalStorageItem(TINT_STORAGE_KEY); return raw ? raw.trim() : null; })();

    applyThemeState({ theme: storedTheme, tint: storedTint }, { persist: false });

    if (window.matchMedia && !mediaListenerAttached) {
        const media = window.matchMedia('(prefers-color-scheme: dark)');
        if (media.addEventListener) {
            media.addEventListener('change', handleSystemThemeChange);
        } else if (media.addListener) {
            media.addListener(handleSystemThemeChange);
        }
        mediaListenerAttached = true;
        return () => {
            if (media.removeEventListener) {
                media.removeEventListener('change', handleSystemThemeChange);
            } else if (media.removeListener) {
                media.removeListener(handleSystemThemeChange);
            }
            mediaListenerAttached = false;
        };
    }

    return () => {};
}

export function applyThemeFromEvent(payload) {
    if (!payload || typeof payload !== 'object') return;
    const chatJid = payload.chat_jid || payload.chatJid || resolveCurrentChatJid();
    const theme = payload.theme ?? payload.name ?? payload.colorTheme;
    const tint = payload.tint ?? null;

    // Store per-chat override
    setChatTheme(chatJid, theme || 'default', tint);

    // Apply immediately
    applyThemeState({ theme: theme || 'default', tint }, { persist: false });

    // Only update global fallback when on the root/default chat,
    // so branch themes don't bleed into other chats.
    if (!chatJid || chatJid === 'web:default') {
        setLocalStorageItem(THEME_STORAGE_KEY, theme || 'default');
        setLocalStorageItem(TINT_STORAGE_KEY, tint || '');
    }
}

export function getThemeMode() {
    if (typeof document === 'undefined') return 'light';
    const attr = document.documentElement?.dataset?.theme;
    if (attr === 'dark' || attr === 'light') return attr;
    return resolveSystemMode();
}
