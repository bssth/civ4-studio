<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {GetPlayers, SetPlayers} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, civilizations, describeType, enums, mapInfo, mapVersion, NONE, withCurrent, withNone} from "../store";

const players = ref<editor.Player[]>([]);
const showEmpty = ref(false);
// Allow picking any leader, not only the ones available for the civilization
const anyLeader = ref(false);

async function load() {
  const p = await GetPlayers();
  players.value = (p ?? []).map(x => editor.Player.createFrom(x));
}

onMounted(load);
watch(mapVersion, load);

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
  return Array.from({length: count}, (_, i) => ({value: i, title: `Team ${i}`}));
});

function civItems(p: editor.Player) {
  return withNone(withCurrent(civOptions.value, p.CivType === NONE ? '' : p.CivType), '(empty slot)');
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
  return withNone(withCurrent(enums.civics.filter(c => c.group === option), current), '(default)');
}

// Names may be text keys (TXT_KEY_...) which the game translates, show game data names for them
function isTextKey(value: string | null | undefined): boolean {
  return !value || value.startsWith('TXT_KEY_');
}

function playerTitle(p: editor.Player): string {
  if (isEmpty(p)) return '(empty slot)';
  return isTextKey(p.LeaderName) ? describeType(enums.leaders, p.LeaderType) : p.LeaderName;
}

function civTitle(p: editor.Player): string {
  return isTextKey(p.CivShortDesc) ? describeType(civOptions.value, p.CivType) : p.CivShortDesc;
}
</script>

<template>
  <div v-if="players.length === 0" class="pa-5 text-grey">
    No players in the current map.
  </div>

  <div v-else class="pa-2">
    <div class="d-flex flex-wrap px-3">
      <v-checkbox v-model="showEmpty" label="Show empty player slots" hide-details density="compact" class="me-6" />
      <v-checkbox v-model="anyLeader" label="Allow any leader for a civilization" hide-details density="compact" />
    </div>

    <v-expansion-panels variant="accordion">
      <template v-for="(player, idx) in players" :key="idx">
        <v-expansion-panel v-if="showEmpty || !isEmpty(player)">
          <v-expansion-panel-title>
            #{{ idx }} — {{ playerTitle(player) }}
            <span class="ms-2 text-caption text-grey">
              <template v-if="!isEmpty(player)">
                {{ civTitle(player) }} · team {{ player.Team }}
                <template v-if="!player.PlayableCiv"> · AI only</template>
              </template>
            </span>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-autocomplete label="Civilization" density="compact" hide-details
                                :items="civItems(player)" item-value="type" item-title="description"
                                :model-value="player.CivType"
                                @update:model-value="(v: string) => onCivChange(player, v ?? NONE)" />
              </v-col>
              <v-col cols="12" md="6">
                <v-autocomplete label="Leader" density="compact" hide-details
                                :items="leaderItems(player)" item-value="type" item-title="description"
                                :model-value="player.LeaderType"
                                @update:model-value="(v: string) => onLeaderChange(player, v ?? NONE)" />
              </v-col>
            </v-row>

            <template v-if="!isEmpty(player)">
              <v-row dense class="mt-1">
                <v-col cols="12" md="6">
                  <v-text-field label="Civ description" density="compact" hide-details v-model="player.CivDesc" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field label="Civ short description" density="compact" hide-details v-model="player.CivShortDesc" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field label="Leader name" density="compact" hide-details v-model="player.LeaderName" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field label="Civ adjective" density="compact" hide-details v-model="player.CivAdjective">
                    <template v-slot:append>
                      <v-btn icon="mdi-restore" variant="text" size="small"
                             title="Reset names, color and art style from game data"
                             @click="resetFromGameData(player)" />
                    </template>
                  </v-text-field>
                </v-col>
              </v-row>

              <v-row dense class="mt-1">
                <v-col cols="6" md="3">
                  <v-select label="Team" density="compact" hide-details
                            :items="teamOptions" v-model="player.Team" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select label="Handicap (AI)" density="compact" hide-details
                            :items="optionItems(enums.handicaps, player.Handicap, false)"
                            item-value="type" item-title="description" v-model="player.Handicap" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-autocomplete label="Color" density="compact" hide-details
                                  :items="optionItems(enums.colors, player.Color)"
                                  item-value="type" item-title="description" v-model="player.Color" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-autocomplete label="Art style" density="compact" hide-details
                                  :items="optionItems(enums.artStyles, player.ArtStyle)"
                                  item-value="type" item-title="description" v-model="player.ArtStyle" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select label="State religion" density="compact" hide-details
                            :items="optionItems(enums.religions, player.StateReligion)"
                            item-value="type" item-title="description"
                            :model-value="player.StateReligion || NONE"
                            @update:model-value="(v: string) => player.StateReligion = v === NONE ? '' : v" />
                </v-col>
                <v-col cols="6" md="3">
                  <v-select label="Starting era" density="compact" hide-details
                            :items="optionItems(enums.eras, player.StartingEra)"
                            item-value="type" item-title="description"
                            :model-value="player.StartingEra || NONE"
                            @update:model-value="(v: string) => player.StartingEra = v === NONE ? '' : v" />
                </v-col>
                <v-col cols="6" md="2">
                  <v-text-field label="Starting gold" type="number" density="compact" hide-details
                                v-model.number="player.StartingGold" />
                </v-col>
                <v-col cols="3" md="2">
                  <v-text-field label="Start X" type="number" density="compact" hide-details
                                v-model.number="player.StartingX" />
                </v-col>
                <v-col cols="3" md="2">
                  <v-text-field label="Start Y" type="number" density="compact" hide-details
                                v-model.number="player.StartingY" />
                </v-col>
                <v-col cols="12">
                  <v-text-field label="Flag decal" density="compact" hide-details v-model="player.FlagDecal" />
                </v-col>
              </v-row>

              <div class="d-flex flex-wrap mt-2">
                <v-checkbox v-model="player.PlayableCiv" label="Playable by human" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.MinorNationStatus" label="Minor nation" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.RandomStartLocation" label="Random start location" hide-details density="compact" class="me-4" />
                <v-checkbox v-model="player.WhiteFlag" label="White flag background" hide-details density="compact" class="me-4" />
              </div>

              <template v-if="enums.civicOptions.length > 0">
                <h4 class="mt-3 mb-1">Starting civics</h4>
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

              <v-textarea label="City names (one per line, used for new cities)" rows="3" auto-grow
                          density="compact" hide-details class="mt-3"
                          :model-value="(player.CityList ?? []).join('\n')"
                          @update:model-value="(v: string) => player.CityList = (v ?? '').split('\n').map(s => s.trim()).filter(Boolean)" />
            </template>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </template>
    </v-expansion-panels>
  </div>
</template>
