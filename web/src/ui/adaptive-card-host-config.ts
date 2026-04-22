/**
 * ui/adaptive-card-host-config.ts – HostConfig for Adaptive Cards mapped from PiClaw CSS vars.
 *
 * Maps PiClaw theme tokens into the Microsoft Adaptive Cards HostConfig format
 * so cards render consistently in both light and dark themes.
 */

type RgbColor = { r: number; g: number; b: number };

function parseHexColor(input: string): RgbColor | null {
  const raw = String(input || "").trim();
  const match = raw.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
  if (!match) return null;
  const hex = match[1].length === 3
    ? match[1].split("").map((part) => `${part}${part}`).join("")
    : match[1];
  return {
    r: parseInt(hex.slice(0, 2), 16),
    g: parseInt(hex.slice(2, 4), 16),
    b: parseInt(hex.slice(4, 6), 16),
  };
}

function parseRgbColor(input: string): RgbColor | null {
  const raw = String(input || "").trim();
  const match = raw.match(/^rgba?\((\d+)[,\s]+(\d+)[,\s]+(\d+)/i);
  if (!match) return null;
  const r = Number(match[1]);
  const g = Number(match[2]);
  const b = Number(match[3]);
  if (![r, g, b].every((value) => Number.isFinite(value))) return null;
  return { r, g, b };
}

function parseColor(input: string): RgbColor | null {
  return parseHexColor(input) || parseRgbColor(input);
}

function relativeLuminance(color: RgbColor): number {
  const toLinear = (channel: number) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  };
  const r = toLinear(color.r);
  const g = toLinear(color.g);
  const b = toLinear(color.b);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrastRatio(a: RgbColor, b: RgbColor): number {
  const lighter = Math.max(relativeLuminance(a), relativeLuminance(b));
  const darker = Math.min(relativeLuminance(a), relativeLuminance(b));
  return (lighter + 0.05) / (darker + 0.05);
}

export function pickHighestContrastColor(background: string, candidates: string[], fallback = "#ffffff"): string {
  const backgroundColor = parseColor(background);
  if (!backgroundColor) return fallback;

  let best = fallback;
  let bestScore = -1;
  for (const candidate of candidates) {
    const parsed = parseColor(candidate);
    if (!parsed) continue;
    const score = contrastRatio(backgroundColor, parsed);
    if (score > bestScore) {
      best = candidate;
      bestScore = score;
    }
  }
  return best;
}

export function getAdaptiveCardThemeValues() {
  const style = getComputedStyle(document.documentElement);
  const getAny = (names: string[], fallback: string) => {
    for (const name of names) {
      const value = style.getPropertyValue(name).trim();
      if (value) return value;
    }
    return fallback;
  };

  const fg = getAny(["--text-primary", "--color-text"], "#0f1419");
  const fgMuted = getAny(["--text-secondary", "--color-text-muted"], "#536471");
  const bgPrimary = getAny(["--bg-primary", "--color-bg-primary"], "#ffffff");
  const bg = getAny(["--bg-secondary", "--color-bg-secondary"], "#f7f9fa");
  const bgEmphasis = getAny(["--bg-hover", "--bg-tertiary", "--color-bg-tertiary"], "#e8ebed");
  const accent = getAny(["--accent-color", "--color-accent"], "#1d9bf0");
  const good = getAny(["--success-color", "--color-success"], "#00ba7c");
  const warning = getAny(["--warning-color", "--color-warning", "--accent-color"], "#f0b429");
  const attention = getAny(["--danger-color", "--color-error"], "#f4212e");
  const border = getAny(["--border-color", "--color-border"], "#eff3f4");
  const fontFamily = getAny(["--font-family"], "system-ui, sans-serif");
  const buttonTextColor = pickHighestContrastColor(accent, [fg, bgPrimary], fg);

  return {
    fg,
    fgMuted,
    bgPrimary,
    bg,
    bgEmphasis,
    accent,
    good,
    warning,
    attention,
    border,
    fontFamily,
    buttonTextColor,
  };
}

/** Build an Adaptive Cards HostConfig from computed CSS variables. */
export function buildHostConfig(): Record<string, unknown> {
  const {
    fg,
    fgMuted,
    bg,
    bgEmphasis,
    accent,
    good,
    warning,
    attention,
    border,
    fontFamily,
  } = getAdaptiveCardThemeValues();

  return {
    fontFamily,
    containerStyles: {
      default: {
        backgroundColor: bg,
        foregroundColors: {
          default: { default: fg, subtle: fgMuted },
          accent: { default: accent, subtle: accent },
          good: { default: good, subtle: good },
          warning: { default: warning, subtle: warning },
          attention: { default: attention, subtle: attention },
        },
      },
      emphasis: {
        backgroundColor: bgEmphasis,
        foregroundColors: {
          default: { default: fg, subtle: fgMuted },
          accent: { default: accent, subtle: accent },
          good: { default: good, subtle: good },
          warning: { default: warning, subtle: warning },
          attention: { default: attention, subtle: attention },
        },
      },
    },
    actions: {
      actionsOrientation: "horizontal",
      actionAlignment: "left",
      buttonSpacing: 8,
      maxActions: 5,
      showCard: { actionMode: "inline" },
      spacing: "default",
    },
    adaptiveCard: {
      allowCustomStyle: false,
    },
    spacing: {
      small: 4,
      default: 8,
      medium: 12,
      large: 16,
      extraLarge: 24,
      padding: 12,
    },
    separator: {
      lineThickness: 1,
      lineColor: border,
    },
    fontSizes: {
      small: 12,
      default: 14,
      medium: 16,
      large: 18,
      extraLarge: 22,
    },
    fontWeights: {
      lighter: 300,
      default: 400,
      bolder: 600,
    },
    imageSizes: {
      small: 40,
      medium: 80,
      large: 120,
    },
    textBlock: {
      headingLevel: 2,
    },
  };
}
