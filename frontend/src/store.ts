import {reactive, ref} from "vue";
import {
    GetGame,
    GetMapInfo,
    GetEraOptions,
    GetSpeedOptions,
    GetCalendarOptions,
    GetVictoryOptions,
    GetGameOptionOptions,
    GetMPOptionOptions,
    GetForceControlOptions,
} from "../wailsjs/go/editor/App";
import {editor} from "../wailsjs/go/models";

export const mapInfo = ref<editor.MapInfo | null>(null);
export const game = ref<editor.Game | null>(null);

export const enums = reactive<{
    eras: editor.EnumOption[];
    speeds: editor.EnumOption[];
    calendars: editor.EnumOption[];
    victories: editor.EnumOption[];
    gameOptions: editor.EnumOption[];
    mpOptions: editor.EnumOption[];
    forceControls: editor.EnumOption[];
}>({
    eras: [],
    speeds: [],
    calendars: [],
    victories: [],
    gameOptions: [],
    mpOptions: [],
    forceControls: [],
});

export const xmlReady = ref(false);

export async function refreshEnums() {
    const [eras, speeds, calendars, victories, gameOptions, mpOptions, forceControls] = await Promise.all([
        GetEraOptions(),
        GetSpeedOptions(),
        GetCalendarOptions(),
        GetVictoryOptions(),
        GetGameOptionOptions(),
        GetMPOptionOptions(),
        GetForceControlOptions(),
    ]);
    enums.eras = eras ?? [];
    enums.speeds = speeds ?? [];
    enums.calendars = calendars ?? [];
    enums.victories = victories ?? [];
    enums.gameOptions = gameOptions ?? [];
    enums.mpOptions = mpOptions ?? [];
    enums.forceControls = forceControls ?? [];
    xmlReady.value = true;
}

export async function refreshMap() {
    const [info, g] = await Promise.all([GetMapInfo(), GetGame()]);
    mapInfo.value = info;
    game.value = g;
}
