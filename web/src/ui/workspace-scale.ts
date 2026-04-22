export const WORKSPACE_SCALE_STORAGE_KEY = 'workspaceExplorerScale';
export const WORKSPACE_SCALE_PRESETS = ['compact', 'default', 'comfortable'] as const;

export type WorkspaceScalePreset = (typeof WORKSPACE_SCALE_PRESETS)[number];

export interface WorkspaceScaleEnvironment {
    width?: number | null;
    isTouch?: boolean | null;
}

export interface WorkspaceScaleResolutionOptions extends WorkspaceScaleEnvironment {
    stored?: string | null;
}

export interface WorkspaceScaleMetrics {
    indentPx: number;
}

const WORKSPACE_SCALE_SET = new Set<string>(WORKSPACE_SCALE_PRESETS);

const WORKSPACE_SCALE_METRICS: Record<WorkspaceScalePreset, WorkspaceScaleMetrics> = {
    compact: { indentPx: 14 },
    default: { indentPx: 16 },
    comfortable: { indentPx: 18 },
};

export function normalizeWorkspaceScale(value: string | null | undefined, fallback: WorkspaceScalePreset = 'default'): WorkspaceScalePreset {
    if (typeof value !== 'string') return fallback;
    const normalized = value.trim().toLowerCase();
    return WORKSPACE_SCALE_SET.has(normalized) ? (normalized as WorkspaceScalePreset) : fallback;
}

export function readWorkspaceScaleEnvironment(): WorkspaceScaleEnvironment {
    if (typeof window === 'undefined') {
        return { width: 0, isTouch: false };
    }

    const width = Number(window.innerWidth) || 0;
    const coarsePointer = Boolean(window.matchMedia?.('(pointer: coarse)')?.matches);
    const noHover = Boolean(window.matchMedia?.('(hover: none)')?.matches);
    const touchPoints = Number(globalThis.navigator?.maxTouchPoints || 0) > 0;

    return {
        width,
        isTouch: coarsePointer || (touchPoints && noHover),
    };
}

export function getResponsiveWorkspaceScale(env: WorkspaceScaleEnvironment = {}): WorkspaceScalePreset {
    const width = Math.max(0, Number(env.width) || 0);
    const isTouch = Boolean(env.isTouch);

    if (isTouch) return 'comfortable';
    if (width > 0 && width < 1180) return 'comfortable';
    return 'default';
}

export function clampWorkspaceScale(scale: WorkspaceScalePreset, env: WorkspaceScaleEnvironment = {}): WorkspaceScalePreset {
    if (Boolean(env.isTouch) && scale === 'compact') return 'default';
    return scale;
}

export function resolveWorkspaceScale(options: WorkspaceScaleResolutionOptions = {}): WorkspaceScalePreset {
    const responsive = getResponsiveWorkspaceScale(options);
    const requested = options.stored ? normalizeWorkspaceScale(options.stored, responsive) : responsive;
    return clampWorkspaceScale(requested, options);
}

export function getWorkspaceScaleMetrics(scale: WorkspaceScalePreset): WorkspaceScaleMetrics {
    return WORKSPACE_SCALE_METRICS[normalizeWorkspaceScale(scale)];
}
