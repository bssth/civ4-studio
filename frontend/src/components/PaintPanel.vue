<script setup lang="ts">
import {applyBrushPreset, brush, brushIsEmpty, brushPresets, emptyIfNone, enums, noneIfEmpty, withCurrent, withNone} from "../store";
import {PLOT_HILLS, PLOT_LAND, PLOT_OCEAN, PLOT_PEAK} from "../mapRender";

const heights = [
  {value: PLOT_PEAK, title: 'Peak'},
  {value: PLOT_HILLS, title: 'Hills'},
  {value: PLOT_LAND, title: 'Flat'},
  {value: PLOT_OCEAN, title: 'Water'},
];
const sizes = [1, 3, 5, 7, 9];

function optional(options: any[], value: string) {
  return withNone(withCurrent(options, emptyIfNone(value)));
}
</script>

<template>
  <div class="pa-3">
    <h3 class="mb-1">Brush</h3>
    <div class="text-caption text-medium-emphasis mb-2">
      Drag over the map to paint. Only the properties switched on are changed; Ctrl+Z undoes a stroke.
    </div>

    <div class="d-flex flex-wrap" style="gap: 4px">
      <v-btn v-for="p in brushPresets" :key="p.title" size="small" variant="tonal" :prepend-icon="p.icon"
             @click="applyBrushPreset(p)">{{ p.title }}
      </v-btn>
    </div>

    <div class="d-flex align-center mt-3">
      <span class="text-body-2 me-3">Size</span>
      <v-btn-toggle v-model="brush.size" mandatory density="compact" divided variant="outlined">
        <v-btn v-for="s in sizes" :key="s" :value="s" size="small">{{ s }}</v-btn>
      </v-btn-toggle>
    </div>

    <v-divider class="my-3"/>

    <div class="brush-row">
      <v-checkbox-btn v-model="brush.terrain.on" density="compact"/>
      <v-select label="Terrain" density="compact" hide-details :disabled="!brush.terrain.on"
                :items="withCurrent(enums.terrains, brush.terrain.value)" item-value="type" item-title="description"
                v-model="brush.terrain.value"/>
    </div>
    <div class="text-caption text-medium-emphasis mb-2 ms-10">
      Water terrain makes plots water and land terrain makes water plots flat, unless the height is painted too.
    </div>
    <div class="brush-row">
      <v-checkbox-btn v-model="brush.height.on" density="compact"/>
      <v-select label="Height" density="compact" hide-details :disabled="!brush.height.on"
                :items="heights" v-model="brush.height.value"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn v-model="brush.feature.on" density="compact"/>
      <v-select label="Feature" density="compact" hide-details :disabled="!brush.feature.on"
                :items="optional(enums.features, brush.feature.value)" item-value="type" item-title="description"
                :model-value="noneIfEmpty(brush.feature.value)"
                @update:model-value="(v: string) => brush.feature.value = emptyIfNone(v)"/>
      <v-text-field label="Variety" type="number" min="0" density="compact" hide-details style="max-width: 90px"
                    :disabled="!brush.feature.on || !brush.feature.value" v-model.number="brush.variety"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn v-model="brush.bonus.on" density="compact"/>
      <v-autocomplete label="Resource" density="compact" hide-details :disabled="!brush.bonus.on"
                      :items="optional(enums.bonuses, brush.bonus.value)" item-value="type" item-title="description"
                      :model-value="noneIfEmpty(brush.bonus.value)"
                      @update:model-value="(v: string) => brush.bonus.value = emptyIfNone(v)"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn v-model="brush.improvement.on" density="compact"/>
      <v-autocomplete label="Improvement" density="compact" hide-details :disabled="!brush.improvement.on"
                      :items="optional(enums.improvements, brush.improvement.value)" item-value="type"
                      item-title="description"
                      :model-value="noneIfEmpty(brush.improvement.value)"
                      @update:model-value="(v: string) => brush.improvement.value = emptyIfNone(v)"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn v-model="brush.route.on" density="compact"/>
      <v-select label="Route" density="compact" hide-details :disabled="!brush.route.on"
                :items="optional(enums.routes, brush.route.value)" item-value="type" item-title="description"
                :model-value="noneIfEmpty(brush.route.value)"
                @update:model-value="(v: string) => brush.route.value = emptyIfNone(v)"/>
    </div>

    <v-alert v-if="brushIsEmpty()" type="info" variant="tonal" density="compact" class="mt-3">
      Switch on at least one property or pick a preset.
    </v-alert>
  </div>
</template>

<style scoped>
.brush-row {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 8px;
}
</style>
