<script setup lang="ts">
import {computed, watch} from "vue";
import {SetGame} from "../../wailsjs/go/editor/App";
import {batched, enums, game, withCurrent} from "../store";
import {editor} from "../../wailsjs/go/models";

// Persist Game changes back to the Go side, a flurry of synchronous edits collapses into a single call
watch(() => game.value && JSON.stringify(game.value), batched(() => {
  if (game.value) {
    SetGame(editor.Game.createFrom(game.value));
  }
}));

function toggleInArray(arr: string[] | null | undefined, value: string, on: boolean): string[] {
  const list = (arr ?? []).filter(v => v !== value);
  if (on) list.push(value);
  return list;
}

const victorySet = computed({
  get: () => game.value?.Victory ?? [],
  set: v => { if (game.value) game.value.Victory = v; },
});
const optionSet = computed({
  get: () => game.value?.Option ?? [],
  set: v => { if (game.value) game.value.Option = v; },
});
const mpOptionSet = computed({
  get: () => game.value?.MPOption ?? [],
  set: v => { if (game.value) game.value.MPOption = v; },
});
const forceControlSet = computed({
  get: () => game.value?.ForceControl ?? [],
  set: v => { if (game.value) game.value.ForceControl = v; },
});
</script>

<template>
  <div v-if="!game" class="pa-5 text-grey">
    No map loaded. Open one from the toolbar to start editing.
  </div>

  <div v-else class="pa-4">
    <h3 class="mb-3">Game</h3>

    <v-row dense>
      <v-col cols="12" md="4">
        <v-select label="Era" density="compact" hide-details
                  v-model="game.Era"
                  :items="withCurrent(enums.eras, game.Era)" item-value="type" item-title="description" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select label="Speed" density="compact" hide-details
                  v-model="game.Speed"
                  :items="withCurrent(enums.speeds, game.Speed)" item-value="type" item-title="description" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select label="Calendar" density="compact" hide-details
                  v-model="game.Calendar"
                  :items="withCurrent(enums.calendars, game.Calendar)" item-value="type" item-title="description" />
      </v-col>
    </v-row>

    <v-row dense class="mt-2">
      <v-col cols="12" md="6">
        <v-text-field label="Starting turn" type="number" density="compact" hide-details
                      v-model.number="game.GameTurn" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field label="Starting year" type="number" density="compact" hide-details
                      v-model.number="game.StartYear" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field label="Max turns" type="number" density="compact" hide-details
                      v-model.number="game.MaxTurns" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field label="Target score" type="number" density="compact" hide-details
                      v-model.number="game.TargetScore" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field label="Max city elimination" type="number" density="compact" hide-details
                      v-model.number="game.MaxCityElimination" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field label="Advanced start points" type="number" density="compact" hide-details
                      v-model.number="game.NumAdvancedStartPoints" />
      </v-col>
    </v-row>

    <v-row dense class="mt-2">
      <v-col cols="12">
        <v-text-field label="Description" density="compact" hide-details v-model="game.Description" />
      </v-col>
      <v-col cols="12">
        <v-text-field label="Mod path" density="compact" hide-details v-model="game.ModPath" />
      </v-col>
    </v-row>

    <v-checkbox v-model="game.Tutorial" hide-details density="compact" label="Tutorial enabled" />

    <v-divider class="my-3" />
    <h4>Victory conditions</h4>
    <div class="d-flex flex-wrap">
      <v-checkbox v-for="opt in withCurrent(enums.victories, ...victorySet)" :key="opt.type"
                  :label="opt.description"
                  :model-value="victorySet.includes(opt.type)"
                  @update:model-value="(v) => victorySet = toggleInArray(victorySet, opt.type, !!v)"
                  hide-details density="compact" class="me-4" />
    </div>

    <v-divider class="my-3" />
    <h4>Game options</h4>
    <div class="d-flex flex-wrap">
      <v-checkbox v-for="opt in withCurrent(enums.gameOptions, ...optionSet)" :key="opt.type"
                  :label="opt.description"
                  :model-value="optionSet.includes(opt.type)"
                  @update:model-value="(v) => optionSet = toggleInArray(optionSet, opt.type, !!v)"
                  hide-details density="compact" class="me-4" />
    </div>

    <v-divider class="my-3" />
    <h4>Multiplayer options</h4>
    <div class="d-flex flex-wrap">
      <v-checkbox v-for="opt in withCurrent(enums.mpOptions, ...mpOptionSet)" :key="opt.type"
                  :label="opt.description"
                  :model-value="mpOptionSet.includes(opt.type)"
                  @update:model-value="(v) => mpOptionSet = toggleInArray(mpOptionSet, opt.type, !!v)"
                  hide-details density="compact" class="me-4" />
    </div>

    <v-divider class="my-3" />
    <h4>Locked (force control)</h4>
    <div class="d-flex flex-wrap">
      <v-checkbox v-for="opt in withCurrent(enums.forceControls, ...forceControlSet)" :key="opt.type"
                  :label="opt.description"
                  :model-value="forceControlSet.includes(opt.type)"
                  @update:model-value="(v) => forceControlSet = toggleInArray(forceControlSet, opt.type, !!v)"
                  hide-details density="compact" class="me-4" />
    </div>
  </div>
</template>
