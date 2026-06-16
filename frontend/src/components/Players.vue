<script setup lang="ts">
import {onMounted, ref, watch} from "vue";
import {GetPlayers, SetPlayers} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {mapInfo} from "../store";

const players = ref<editor.Player[]>([]);
const showEmpty = ref(false);

async function load() {
  const p = await GetPlayers();
  players.value = (p ?? []).map(x => editor.Player.createFrom(x));
}

onMounted(load);
watch(() => mapInfo.value?.path, load);

let pending: any = null;
function persist() {
  if (pending) return;
  pending = Promise.resolve().then(() => {
    pending = null;
    SetPlayers(players.value.map(p => editor.Player.createFrom(p)));
  });
}

watch(players, persist, {deep: true});

function isEmpty(p: editor.Player): boolean {
  return p.CivType === "NONE" && p.LeaderType === "NONE";
}
</script>

<template>
  <div v-if="players.length === 0" class="pa-5 text-grey">
    No players in the current map.
  </div>

  <div v-else class="pa-2">
    <v-checkbox v-model="showEmpty" label="Show empty player slots" hide-details density="compact" class="px-3" />

    <v-expansion-panels variant="accordion">
      <template v-for="(player, idx) in players" :key="idx">
        <v-expansion-panel v-if="showEmpty || !isEmpty(player)">
          <v-expansion-panel-title>
            #{{ idx }} — {{ player.LeaderName || player.LeaderType || "(empty)" }}
            <span class="ms-2 text-caption text-grey">
              {{ player.CivShortDesc || player.CivType }}
              <template v-if="!isEmpty(player)">· team {{ player.Team }}</template>
            </span>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <v-row dense>
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
                <v-text-field label="Civ adjective" density="compact" hide-details v-model="player.CivAdjective" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field label="Leader type" density="compact" hide-details v-model="player.LeaderType" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field label="Civ type" density="compact" hide-details v-model="player.CivType" />
              </v-col>
              <v-col cols="6" md="3">
                <v-text-field label="Team" type="number" density="compact" hide-details v-model.number="player.Team" />
              </v-col>
              <v-col cols="6" md="3">
                <v-text-field label="Starting gold" type="number" density="compact" hide-details v-model.number="player.StartingGold" />
              </v-col>
              <v-col cols="6" md="3">
                <v-text-field label="Starting X" type="number" density="compact" hide-details v-model.number="player.StartingX" />
              </v-col>
              <v-col cols="6" md="3">
                <v-text-field label="Starting Y" type="number" density="compact" hide-details v-model.number="player.StartingY" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="Handicap" density="compact" hide-details v-model="player.Handicap" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="Color" density="compact" hide-details v-model="player.Color" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="Art style" density="compact" hide-details v-model="player.ArtStyle" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="State religion" density="compact" hide-details v-model="player.StateReligion" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="Starting era" density="compact" hide-details v-model="player.StartingEra" />
              </v-col>
              <v-col cols="6" md="4">
                <v-text-field label="Flag decal" density="compact" hide-details v-model="player.FlagDecal" />
              </v-col>
            </v-row>

            <div class="d-flex flex-wrap mt-2">
              <v-checkbox v-model="player.PlayableCiv" label="Playable" hide-details density="compact" class="me-4" />
              <v-checkbox v-model="player.MinorNationStatus" label="Minor nation" hide-details density="compact" class="me-4" />
              <v-checkbox v-model="player.RandomStartLocation" label="Random start location" hide-details density="compact" class="me-4" />
              <v-checkbox v-model="player.WhiteFlag" label="White flag bg" hide-details density="compact" class="me-4" />
            </div>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </template>
    </v-expansion-panels>
  </div>
</template>
