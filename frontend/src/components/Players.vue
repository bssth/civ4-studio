<script setup lang="ts">
import {computed, onMounted, reactive, ref, watch} from "vue";
import {ClearPlayer, GetPlayers, SetPlayers, SwapPlayers} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {
  batched,
  civilizations,
  describeType,
  enums,
  mapInfo,
  mapRevision,
  mapVersion,
  NONE,
  playerName,
  refreshMap,
  withCurrent,
  withNone
} from "../store";
import {useI18n} from "vue-i18n";

const {t} = useI18n();

const players = ref<editor.Player[]>([]);
const showEmpty = ref(false);
// Allow picking any leader, not only the ones available for the civilization
const anyLeader = ref(false);

async function load() {
  const p = await GetPlayers();
  players.value = (p ?? []).map(x => editor.Player.createFrom(x));
  // A new map has only empty slots, show them so there is something to fill in
  if (players.value.length > 0 && players.value.every(isEmpty)) {
    showEmpty.value = true;
  }
}

onMounted(load);
watch([mapVersion, mapRevision], () => load());

watch(players, batched(() => {
  SetPlayers(players.value.map(p => editor.Player.createFrom(p)));
}), {deep: true});

function isEmpty(p: editor.Player): boolean {
  return (!p.CivType || p.CivType === NONE) && (!p.LeaderType || p.LeaderType === NONE);
}

const civOptions = computed(() => civilizations.value.map(c =>
    editor.EnumOption.createFrom({type: c.type, description: c.description})));

const teamOptions = computed(() => {
  const count = Math.max(mapInfo.value?.teams_count ?? 0, ...players.value.map(p => p.Team + 1));
  return Array.from({length: count}, (_, i) => ({value: i, title: t('teams.team', {n: i})}));
});

function civItems(p: editor.Player) {
  return withNone(withCurrent(civOptions.value, p.CivType === NONE ? '' : p.CivType), t('players.emptySlot'));
}

function leaderItems(p: editor.Player) {
  const civ = civilizations.value.find(c => c.type === p.CivType);
  let leaders = enums.leaders;
  if (civ && !anyLeader.value && civ.leaders.length > 0) {
    leaders = leaders.filter(l => civ.leaders.includes(l.type));
  }
  return withNone(withCurrent(leaders, p.LeaderType === NONE ? '' : p.LeaderType));
}

function optionItems(options: editor.EnumOption[], value: string, allowNone = true) {
  const items = withCurrent(options, value === NONE ? '' : value);
  return allowNone ? withNone(items) : items;
}

function clearSlot(p: editor.Player) {
  Object.assign(p, {
    CivType: NONE, LeaderType: NONE, Color: NONE, ArtStyle: NONE,
    CivDesc: '', CivShortDesc: '', CivAdjective: '', LeaderName: '', FlagDecal: '',
    PlayableCiv: false, StateReligion: '', StartingEra: '',
    CityList: [], CivicOption: [], Civic: [],
  });
}

// --- Attitudes: AttitudePlayer and AttitudeExtra are parallel lists ------------

function attitudeRows(p: editor.Player) {
  const others = p.AttitudePlayer ?? [], values = p.AttitudeExtra ?? [];
  return others.map((other, i) => ({other, value: values[i] ?? 0}));
}

function setAttitude(p: editor.Player, other: number, value: number | null) {
  const rows = attitudeRows(p).filter(r => r.other !== other);
  if (value !== null) rows.push({other, value: Math.round(Number(value) || 0)});
  rows.sort((a, b) => a.other - b.other);
  p.AttitudePlayer = rows.map(r => r.other);
  p.AttitudeExtra = rows.map(r => r.value);
}

function attitudeCandidates(p: editor.Player, index: number) {
  const used = new Set(attitudeRows(p).map(r => r.other));
  return players.value
      .map((q, i) => ({q, i}))
      .filter(({q, i}) => i !== index && !isEmpty(q) && !used.has(i))
      .map(({i}) => ({value: i, title: `#${i} ${playerName(players.value, i)}`}));
}

function attitudeColor(value: number): string {
  return value > 0 ? 'success' : value < 0 ? 'error' : 'grey';
}

function onCivChange(p: editor.Player, civType: string) {
  const wasEmpty = isEmpty(p);
  p.CivType = civType;
  const civ = civilizations.value.find(c => c.type === civType);
  if (!civ) {
    if (civType === NONE) clearSlot(p);
    return;
  }

  p.CivDesc = civ.description;
  p.CivShortDesc = civ.short_description;
  p.CivAdjective = civ.adjective;
  if (civ.color) p.Color = civ.color;
  if (civ.art_style) p.ArtStyle = civ.art_style;
  if (wasEmpty) {
    p.PlayableCiv = civ.playable;
    if (!p.Handicap || p.Handicap === NONE) {
      p.Handicap = enums.handicaps.find(h => h.type === 'HANDICAP_NOBLE')?.type ?? enums.handicaps[0]?.type ?? '';
    }
  }
  if (civ.leaders.length > 0 && !civ.leaders.includes(p.LeaderType)) {
    onLeaderChange(p, civ.leaders[0]);
  }
}

function onLeaderChange(p: editor.Player, leaderType: string) {
  p.LeaderType = leaderType;
  if (leaderType && leaderType !== NONE) {
    p.LeaderName = describeType(enums.leaders, leaderType);
  }
}

/** Fills names, color and art style from game data again */
function resetFromGameData(p: editor.Player) {
  const leader = p.LeaderType;
  onCivChange(p, p.CivType);
  if (leader && leader !== NONE) onLeaderChange(p, leader);
}

// Civics are stored as pairs: CivicOption[i] / Civic[i]
function getCivic(p: editor.Player, option: string): string {
  const idx = (p.CivicOption ?? []).indexOf(option);
  return idx >= 0 ? (p.Civic ?? [])[idx] ?? '' : '';
}

function setCivic(p: editor.Player, option: string, civic: string | null) {
  const options = [...(p.CivicOption ?? [])];
  const civics = [...(p.Civic ?? [])];
  const idx = options.indexOf(option);
  if (idx >= 0) {
    options.splice(idx, 1);
    civics.splice(idx, 1);
  }
  if (civic && civic !== NONE) {
    options.push(option);
    civics.push(civic);
  }
  p.CivicOption = options;
  p.Civic = civics;
}

function civicItems(option: string, current: string) {
  return withNone(withCurrent(enums.civics.filter(c => c.group === option), current), t('players.defaultCivic'));
}

// Names may be text keys (TXT_KEY_...) which the game translates, show game data names for them
function isTextKey(value: string | null | undefined): boolean {
  return !value || value.startsWith('TXT_KEY_');
}

// Slot operations change references in the whole map (units, cities, culture), so the backend does them
const swap = reactive({open: false, from: 0, to: 0});
const clear = reactive({open: false, index: 0, removeAssets: true});
const actionError = ref('');

const slotItems = computed(() => players.value.map((p, i) => ({
  value: i, title: `#${i} ${isEmpty(p) ? t('players.emptySlot') : playerName(players.value, i)}`,
})));

function openSwap(index: number) {
  Object.assign(swap, {open: true, from: index, to: index === 0 ? 1 : 0});
}

function openClear(index: number) {
  Object.assign(clear, {open: true, index, removeAssets: true});
}

async function runSlotAction(action: () => Promise<void>) {
  try {
    await action();
    actionError.value = '';
    // Reloads every editor, including this one
    await refreshMap();
  } catch (err: any) {
    actionError.value = String(err);
  }
}

function doSwap() {
  swap.open = false;
  return runSlotAction(() => SwapPlayers(swap.from, swap.to));
}

function doClear() {
  clear.open = false;
  return runSlotAction(() => ClearPlayer(clear.index, clear.removeAssets));
}

function playerTitle(p: editor.Player): string {
  if (isEmpty(p)) return t('players.emptySlot');
  return isTextKey(p.LeaderName) ? describeType(enums.leaders, p.LeaderType) : p.LeaderName;
}

function civTitle(p: editor.Player): string {
  return isTextKey(p.CivShortDesc) ? describeType(civOptions.value, p.CivType) : p.CivShortDesc;
}
</script>

<template>
  <div v-if="players.length === 0" class="pa-5 text-grey">
    {{ $t('players.none') }}
  </div>

  <div v-else class="pa-2">
    <div class="d-flex flex-wrap px-3">
      <v-checkbox v-model="showEmpty" :label="$t('players.showEmpty')" hide-details density="compact" class="me-6" />
      <v-checkbox v-model="anyLeader" :label="$t('players.anyLeader')" hide-details density="compact" />
    </div>

    <v-alert v-if="actionError" type="error" variant="tonal" density="compact" class="mx-3 mb-2" closable
             @click:close="actionError = ''">{{ actionError }}
    </v-alert>

    <v-dialog v-model="swap.open" max-width="520">
      <v-card :title="$t('players.swapTitle')">
        <v-card-text>
          {{ $t('players.swapText', {n: swap.from}) }}
          <v-select class="mt-3" :label="$t('players.swapWith')" :items="slotItems.filter(s => s.value !== swap.from)"
                    v-model="swap.to" density="compact" hide-details/>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="swap.open = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="primary" variant="tonal" @click="doSwap">{{ $t('players.swap') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="clear.open" max-width="520">
      <v-card :title="$t('players.clearTitle', {n: clear.index})">
        <v-card-text>
          {{ $t('players.clearText', {player: playerName(players, clear.index)}) }}
          <v-checkbox v-model="clear.removeAssets" density="compact" hide-details class="mt-2"
                      :label="$t('players.clearAssets')"/>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="clear.open = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="error" variant="tonal" @click="doClear">{{ $t('players.clear') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-expansion-panels variant="accordion">
      <template v-for="(player, idx) in players" :key="idx">
        <v-expansion-panel v-if="showEmpty || !isEmpty(player)">
          <v-expansion-panel-title>
            #{{ idx }} — {{ playerTitle(player) }}
            <span class="ms-2 text-caption text-grey">
              <template v-if="!isEmpty(player)">
                {{ civTitle(player) }} · {{ $t('players.team', {n: player.Team}) }}
                <template v-if="!player.PlayableCiv"> · {{ $t('players.aiOnly') }}</template>
              </template>
            </span>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <div class="d-flex justify-end mb-1">
              <v-btn size="small" variant="text" prepend-icon="mdi-swap-horizontal" @click="openSwap(idx)">
                {{ $t('players.swapSlot') }}
              </v-btn>
              <v-btn v-if="!isEmpty(player)" size="small" variant="text" color="error" prepend-icon="mdi-account-remove"
                     @click="openClear(idx)">{{ $t('players.clearSlot') }}
              </v-btn>
            </div>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-autocomplete :label="$t('players.civilization')" density="compact" hide-details
                                :items="civItems(player)" item-value="type" item-title="description"
                                :model-value="player.CivType"
                                @update:model-value="(v: string) => onCivChange(player, v ?? NONE)" />
              </v-col>
              <v-col cols="12" md="6">
                <v-autocomplete :label="$t('players.leader')" density="compact" hide-details
                                :items="leaderItems(player)" item-value="type" item-title="description"
                                :model-value="player.LeaderType"
                                @update:model-value="(v: string) => onLeaderChange(player, v ?? NONE)" />
              </v-col>
            </v-row>

            <template v-if="!isEmpty(player)">
              <v-row dense class="mt-1">
                <v-col cols="12" md="6">
                  <v-text-field :label="$t('players.civDesc')" density="compact" hide-details v-model="player.CivDesc" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field :label="$t('players.civShortDesc')" density="compact" hide-details v-model="player.CivShortDesc" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field :label="$t('players.leaderName')" density="compact" hide-details v-model="player.LeaderName" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field :label="$t('players.civAdjective')" density="compact" hide-details v-model="player.CivAdjective">
                    <template v-slot:append>
                      <v-btn icon="mdi-restore" variant="text" size="small"
                             :title="$t('players.resetNames')"
                             @click="resetFromGameData(player)" />
                    </template>
                  </v-text-field>
                </v-col>
              </v-row>

              <v-row dense class="mt-1">
                <v-col cols="6" md="3">
                  <v-select :label="$t('players.teamField')" density="compact" hide-details
                            :items="teamOptions" v-model="player.Team" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select :label="$t('players.handicap')" density="compact" hide-details
                            :items="optionItems(enums.handicaps, player.Handicap, false)"
                            item-value="type" item-title="description" v-model="player.Handicap" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-autocomplete :label="$t('players.color')" density="compact" hide-details
                                  :items="optionItems(enums.colors, player.Color)"
                                  item-value="type" item-title="description" v-model="player.Color" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-autocomplete :label="$t('players.artStyle')" density="compact" hide-details
                                  :items="optionItems(enums.artStyles, player.ArtStyle)"
                                  item-value="type" item-title="description" v-model="player.ArtStyle" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select :label="$t('players.stateReligion')" density="compact" hide-details
                            :items="optionItems(enums.religions, player.StateReligion)"
                            item-value="type" item-title="description"
                            :model-value="player.StateReligion || NONE"
                            @update:model-value="(v: string) => player.StateReligion = v === NONE ? '' : v" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select :label="$t('players.startingEra')" density="compact" hide-details
                            :items="optionItems(enums.eras, player.StartingEra)"
                            item-value="type" item-title="description"
                            :model-value="player.StartingEra || NONE"
                            @update:model-value="(v: string) => player.StartingEra = v === NONE ? '' : v" />
                </v-col>
                <v-col cols="6" md="2">
                  <v-text-field :label="$t('players.startingGold')" type="number" density="compact" hide-details
                                v-model.number="player.StartingGold" />
                </v-col>
                <v-col cols="3" md="2">
                  <v-text-field :label="$t('players.startX')" type="number" density="compact" hide-details
                                v-model.number="player.StartingX" />
                </v-col>
                <v-col cols="3" md="2">
                  <v-text-field :label="$t('players.startY')" type="number" density="compact" hide-details
                                v-model.number="player.StartingY" />
                </v-col>
                <v-col cols="12">
                  <v-text-field :label="$t('players.flagDecal')" density="compact" hide-details v-model="player.FlagDecal" />
                </v-col>
              </v-row>

              <div class="d-flex flex-wrap mt-2">
                <v-checkbox v-model="player.PlayableCiv" :label="$t('players.playable')" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.MinorNationStatus" :label="$t('players.minor')" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.RandomStartLocation" :label="$t('players.randomStart')" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.WhiteFlag" :label="$t('players.whiteFlag')" hide-details density="compact" class="me-4" />
              </div>

              <template v-if="enums.civicOptions.length > 0">
                <h4 class="mt-3 mb-1">{{ $t('players.startingCivics') }}</h4>
                <v-row dense>
                  <v-col v-for="option in enums.civicOptions" :key="option.type" cols="6" md="4">
                    <v-select :label="option.description" density="compact" hide-details
                              :items="civicItems(option.type, getCivic(player, option.type))"
                              item-value="type" item-title="description"
                              :model-value="getCivic(player, option.type) || NONE"
                              @update:model-value="(v: string) => setCivic(player, option.type, v)" />
                  </v-col>
                </v-row>
              </template>

              <v-textarea :label="$t('players.cityList')" rows="3" auto-grow
                          density="compact" hide-details class="mt-3"
                          :model-value="(player.CityList ?? []).join('\n')"
                          @update:model-value="(v: string) => player.CityList = (v ?? '').split('\n').map(s => s.trim()).filter(Boolean)" />

              <h4 class="mt-4 mb-1">{{ $t('players.attitudes') }}</h4>
              <div class="text-caption text-medium-emphasis mb-2">{{ $t('players.attitudesHint') }}</div>
              <div v-for="row in attitudeRows(player)" :key="row.other" class="d-flex align-center" style="gap: 8px">
                <span class="text-body-2" style="width: 220px">#{{ row.other }} {{ playerName(players, row.other) }}</span>
                <v-slider :model-value="row.value" :min="-10" :max="10" :step="1" hide-details density="compact"
                          :color="attitudeColor(row.value)" class="flex-grow-1"
                          @update:model-value="(v: number) => setAttitude(player, row.other, v)"/>
                <span class="text-body-2 text-end" style="width: 32px">{{ row.value > 0 ? '+' : '' }}{{ row.value }}</span>
                <v-btn icon="mdi-close" size="x-small" variant="text" :title="$t('players.removeAttitude')"
                       @click="setAttitude(player, row.other, null)"/>
              </div>
              <v-select v-if="attitudeCandidates(player, idx).length" :label="$t('players.addAttitude')"
                        density="compact" hide-details :items="attitudeCandidates(player, idx)" :model-value="null"
                        style="max-width: 360px" @update:model-value="(o: number) => setAttitude(player, o, 0)"/>
            </template>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </template>
    </v-expansion-panels>
  </div>
</template>
