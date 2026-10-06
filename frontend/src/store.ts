import {reactive, ref} from "vue";
import {GetCivilizations, GetGame, GetMapInfo, GetOptions, GetWorldSizes} from "../wailsjs/go/editor/App";
import {editor} from "../wailsjs/go/models";
import {t} from "./i18n";

export const mapInfo = ref<editor.MapInfo | null>(null);
export const game = ref<editor.Game | null>(null);
// Incremented every time a map is opened or created, so editors reload their local copies
export const mapVersion = ref(0);

export type OptionKey =
    'eras' | 'speeds' | 'calendars' | 'victories' | 'gameOptions' | 'mpOptions' | 'forceControls' |
    'leaders' | 'handicaps' | 'colors' | 'artStyles' | 'techs' | 'religions' | 'civics' | 'civicOptions' |
    'projects' | 'worldSizes' | 'climates' | 'seaLevels' | 'terrains' |
    'features' | 'bonuses' | 'improvements' | 'routes' | 'units' | 'unitAIs' | 'buildings' | 'promotions';

const optionKeys: OptionKey[] = [
    'eras', 'speeds', 'calendars', 'victories', 'gameOptions', 'mpOptions', 'forceControls',
    'leaders', 'handicaps', 'colors', 'artStyles', 'techs', 'religions', 'civics', 'civicOptions',
    'projects', 'worldSizes', 'climates', 'seaLevels', 'terrains',
    'features', 'bonuses', 'improvements', 'routes', 'units', 'unitAIs', 'buildings', 'promotions',
];

function emptyEnums(): Record<OptionKey, editor.EnumOption[]> {
    return Object.fromEntries(optionKeys.map(k => [k, []])) as any;
}

export const enums = reactive<Record<OptionKey, editor.EnumOption[]>>(emptyEnums());
export const civilizations = ref<editor.CivilizationOption[]>([]);
export const worldSizes = ref<editor.WorldSizeOption[]>([]);
export const xmlReady = ref(false);

export async function refreshEnums() {
    const [options, civs, sizes] = await Promise.all([GetOptions(), GetCivilizations(), GetWorldSizes()]);
    const byKey = new Map((options ?? []).map(list => [list.key, list.options ?? []]));
    for (const key of optionKeys) {
        enums[key] = byKey.get(key) ?? [];
    }
    civilizations.value = civs ?? [];
    worldSizes.value = sizes ?? [];
    xmlReady.value = true;
}

export function clearEnums() {
    Object.assign(enums, emptyEnums());
    civilizations.value = [];
    worldSizes.value = [];
    xmlReady.value = false;
}

export async function refreshMapInfo() {
    mapInfo.value = await GetMapInfo();
}

export async function refreshMap() {
    const [info, g] = await Promise.all([GetMapInfo(), GetGame()]);
    mapInfo.value = info;
    game.value = g;
    mapVersion.value++;
}

export const NONE = 'NONE';

/**
 * Returns options with the current value appended if it is unknown (e.g. it comes from another mod),
 * so selects never hide values stored in the map.
 */
export function withCurrent(options: editor.EnumOption[], ...values: (string | null | undefined)[]): editor.EnumOption[] {
    const missing = values.filter((v): v is string => !!v && !options.some(o => o.type === v));
    if (missing.length === 0) {
        return options;
    }
    return [...options, ...[...new Set(missing)].map(v => editor.EnumOption.createFrom({type: v, description: v}))];
}

/** Options with a leading NONE entry, used by fields that may be empty */
export function withNone(options: editor.EnumOption[], label = t('common.none')): editor.EnumOption[] {
    return [editor.EnumOption.createFrom({type: NONE, description: label}), ...options];
}

export function describeType(options: editor.EnumOption[], type: string | null | undefined): string {
    if (!type) return '';
    return options.find(o => o.type === type)?.description ?? type;
}

/**
 * Calls fn at most once per tick. Editors keep a local reactive copy and push it
 * to the backend on every change.
 */
export function batched(fn: () => void): () => void {
    let pending = false;
    return () => {
        if (pending) return;
        pending = true;
        Promise.resolve().then(() => {
            pending = false;
            fn();
        });
    };
}

/** Value of an optional type field for selects built with withNone: empty values become NONE */
export function noneIfEmpty(value: string | null | undefined): string {
    return value ? value : NONE;
}

/** Inverse of noneIfEmpty */
export function emptyIfNone(value: string | null | undefined): string {
    return !value || value === NONE ? '' : value;
}

// Colors for owners without a known player color
const fallbackColors = ['#e6194b', '#3cb44b', '#ffe119', '#4363d8', '#f58231', '#911eb4', '#46f0f0', '#f032e6',
    '#bcf60c', '#fabebe', '#008080', '#e6beff', '#9a6324', '#fffac8', '#800000', '#aaffc3', '#808000', '#000075'];

/** Color of a player for markers on the map */
export function playerColor(players: editor.Player[], index: number): string {
    const p = players[index];
    const known = p && enums.colors.find(c => c.type === p.Color)?.color;
    return known || fallbackColors[((index % fallbackColors.length) + fallbackColors.length) % fallbackColors.length];
}

/** Short readable name of a player */
export function playerName(players: editor.Player[], index: number): string {
    const p = players[index];
    if (!p || !p.CivType || p.CivType === NONE) return t('common.player', {n: index});
    const name = p.CivShortDesc && !p.CivShortDesc.startsWith('TXT_KEY_') ? p.CivShortDesc : humanizeCiv(p.CivType);
    return `${name}`;
}

function humanizeCiv(type: string): string {
    return type.replace(/^CIVILIZATION_/, '').toLowerCase().replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}

// Navigation requests between tabs, e.g. from the problem list to a plot on the world map
export const requestedTab = ref<string | null>(null);
export const focusPlot = ref<{ x: number, y: number } | null>(null);

export function showPlot(x: number, y: number) {
    focusPlot.value = {x, y};
    requestedTab.value = 'world';
}

// Brush of the World tab. Each property is painted only when it is switched on;
// an empty value removes a feature, resource, improvement or route.
export interface BrushProperty<T> {
    on: boolean;
    value: T;
}

export const brush = reactive({
    size: 1,
    terrain: {on: true, value: 'TERRAIN_GRASS'} as BrushProperty<string>,
    height: {on: false, value: 2} as BrushProperty<number>,
    feature: {on: false, value: ''} as BrushProperty<string>,
    variety: 0,
    bonus: {on: false, value: ''} as BrushProperty<string>,
    improvement: {on: false, value: ''} as BrushProperty<string>,
    route: {on: false, value: ''} as BrushProperty<string>,
});

export interface BrushPreset {
    // Translation key in brush.presets
    key: string;
    icon: string;
    terrain?: string;
    height?: number;
    feature?: string;
}

// Presets switch on only what they set
export const brushPresets: BrushPreset[] = [
    {key: 'ocean', icon: 'mdi-waves', terrain: 'TERRAIN_OCEAN', height: 3, feature: ''},
    {key: 'coast', icon: 'mdi-wave', terrain: 'TERRAIN_COAST', height: 3, feature: ''},
    {key: 'grassland', icon: 'mdi-grass', terrain: 'TERRAIN_GRASS'},
    {key: 'plains', icon: 'mdi-barley', terrain: 'TERRAIN_PLAINS'},
    {key: 'desert', icon: 'mdi-weather-sunny', terrain: 'TERRAIN_DESERT'},
    {key: 'tundra', icon: 'mdi-snowflake-variant', terrain: 'TERRAIN_TUNDRA'},
    {key: 'snow', icon: 'mdi-snowflake', terrain: 'TERRAIN_SNOW'},
    {key: 'flat', icon: 'mdi-minus', height: 2},
    {key: 'hills', icon: 'mdi-image-filter-hdr', height: 1},
    {key: 'peak', icon: 'mdi-triangle', height: 0},
    {key: 'forest', icon: 'mdi-pine-tree', feature: 'FEATURE_FOREST'},
    {key: 'jungle', icon: 'mdi-palm-tree', feature: 'FEATURE_JUNGLE'},
    {key: 'noFeature', icon: 'mdi-eraser', feature: ''},
];

export function applyBrushPreset(preset: BrushPreset) {
    brush.terrain.on = preset.terrain !== undefined;
    if (preset.terrain !== undefined) brush.terrain.value = preset.terrain;
    brush.height.on = preset.height !== undefined;
    if (preset.height !== undefined) brush.height.value = preset.height;
    brush.feature.on = preset.feature !== undefined;
    if (preset.feature !== undefined) brush.feature.value = preset.feature;
    brush.bonus.on = false;
    brush.improvement.on = false;
    brush.route.on = false;
}

/** The paint operation for the current brush, without cells */
export function brushOperation(): Omit<editor.PaintOp, 'cells' | 'convertValues'> {
    return {
        terrain: brush.terrain.on ? brush.terrain.value : undefined,
        plot_type: brush.height.on ? brush.height.value : undefined,
        feature: brush.feature.on ? brush.feature.value : undefined,
        feature_variety: brush.variety || 0,
        bonus: brush.bonus.on ? brush.bonus.value : undefined,
        improvement: brush.improvement.on ? brush.improvement.value : undefined,
        route: brush.route.on ? brush.route.value : undefined,
    };
}

export function brushIsEmpty(): boolean {
    return !(brush.terrain.on || brush.height.on || brush.feature.on || brush.bonus.on || brush.improvement.on || brush.route.on);
}
