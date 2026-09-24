import {reactive, ref} from "vue";
import {GetCivilizations, GetGame, GetMapInfo, GetOptions} from "../wailsjs/go/editor/App";
import {editor} from "../wailsjs/go/models";

export const mapInfo = ref<editor.MapInfo | null>(null);
export const game = ref<editor.Game | null>(null);
// Incremented every time a map is opened or created, so editors reload their local copies
export const mapVersion = ref(0);

export type OptionKey =
    'eras' | 'speeds' | 'calendars' | 'victories' | 'gameOptions' | 'mpOptions' | 'forceControls' |
    'leaders' | 'handicaps' | 'colors' | 'artStyles' | 'techs' | 'religions' | 'civics' | 'civicOptions' |
    'projects' | 'worldSizes' | 'climates' | 'seaLevels';

const optionKeys: OptionKey[] = [
    'eras', 'speeds', 'calendars', 'victories', 'gameOptions', 'mpOptions', 'forceControls',
    'leaders', 'handicaps', 'colors', 'artStyles', 'techs', 'religions', 'civics', 'civicOptions',
    'projects', 'worldSizes', 'climates', 'seaLevels',
];

function emptyEnums(): Record<OptionKey, editor.EnumOption[]> {
    return Object.fromEntries(optionKeys.map(k => [k, []])) as any;
}

export const enums = reactive<Record<OptionKey, editor.EnumOption[]>>(emptyEnums());
export const civilizations = ref<editor.CivilizationOption[]>([]);
export const xmlReady = ref(false);

export async function refreshEnums() {
    const [options, civs] = await Promise.all([GetOptions(), GetCivilizations()]);
    for (const key of optionKeys) {
        enums[key] = options?.[key] ?? [];
    }
    civilizations.value = civs ?? [];
    xmlReady.value = true;
}

export function clearEnums() {
    Object.assign(enums, emptyEnums());
    civilizations.value = [];
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
export function withNone(options: editor.EnumOption[], label = '(none)'): editor.EnumOption[] {
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
