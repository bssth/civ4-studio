<script setup lang="ts">
import {
  applyBrushPreset,
  brush,
  BRUSH_FILL,
  brushIsEmpty,
  brushPresets,
  emptyIfNone,
  enums,
  noneIfEmpty,
  withCurrent,
  withNone
} from "../store";
import {PLOT_HILLS, PLOT_LAND, PLOT_OCEAN, PLOT_PEAK} from "../mapRender";
import {computed} from "vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();

const heights = computed(() => [
  {value: PLOT_PEAK, title: t('height.peak')},
  {value: PLOT_HILLS, title: t('height.hills')},
  {value: PLOT_LAND, title: t('height.flat')},
  {value: PLOT_OCEAN, title: t('height.water')},
]);
const sizes = [1, 3, 5, 7, 9];

function optional(options: any[], value: string) {
  return withNone(withCurrent(options, emptyIfNone(value)));
}
</script>

<template>
  <div class="pa-3">
    <h3 class="mb-1">{{ $t('brush.title') }}</h3>
    <div class="text-caption text-medium-emphasis mb-2">
      {{ $t('brush.help') }}
    </div>

    <div class="d-flex flex-wrap" style="gap: 4px">
      <v-btn v-for="p in brushPresets" :key="p.key" size="small" variant="tonal" :prepend-icon="p.icon"
             @click="applyBrushPreset(p)">{{ $t('brush.presets.' + p.key) }}
      </v-btn>
    </div>

    <div class="d-flex align-center mt-3">
      <span class="text-body-2 me-3">{{ $t('brush.size') }}</span>
      <v-btn-toggle v-model="brush.size" mandatory density="compact" divided variant="outlined">
        <v-btn v-for="s in sizes" :key="s" :value="s" size="small">{{ s }}</v-btn>
        <v-btn :value="BRUSH_FILL" size="small" :title="$t('brush.fillTitle')">
          <v-icon icon="mdi-format-color-fill"/>
        </v-btn>
      </v-btn-toggle>
    </div>
    <div v-if="brush.size === BRUSH_FILL" class="text-caption text-medium-emphasis mt-1">{{ $t('brush.fillHint') }}</div>

    <v-divider class="my-3"/>

    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.terrain.on" density="compact"/>
      <v-select :label="$t('field.terrain')" density="compact" hide-details :disabled="!brush.terrain.on"
                :items="withCurrent(enums.terrains, brush.terrain.value)" item-value="type" item-title="description"
                v-model="brush.terrain.value"/>
    </div>
    <div class="text-caption text-medium-emphasis mb-2 ms-10">
      {{ $t('brush.autoHeight') }}
    </div>
    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.height.on" density="compact"/>
      <v-select :label="$t('field.height')" density="compact" hide-details :disabled="!brush.height.on"
                :items="heights" v-model="brush.height.value"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.feature.on" density="compact"/>
      <v-select :label="$t('field.feature')" density="compact" hide-details :disabled="!brush.feature.on"
                :items="optional(enums.features, brush.feature.value)" item-value="type" item-title="description"
                :model-value="noneIfEmpty(brush.feature.value)"
                @update:model-value="(v: string) => brush.feature.value = emptyIfNone(v)"/>
      <v-text-field :label="$t('field.variety')" type="number" min="0" density="compact" hide-details style="max-width: 90px"
                    :disabled="!brush.feature.on || !brush.feature.value" v-model.number="brush.variety"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.bonus.on" density="compact"/>
      <v-autocomplete :label="$t('field.resource')" density="compact" hide-details :disabled="!brush.bonus.on"
                      :items="optional(enums.bonuses, brush.bonus.value)" item-value="type" item-title="description"
                      :model-value="noneIfEmpty(brush.bonus.value)"
                      @update:model-value="(v: string) => brush.bonus.value = emptyIfNone(v)"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.improvement.on" density="compact"/>
      <v-autocomplete :label="$t('field.improvement')" density="compact" hide-details :disabled="!brush.improvement.on"
                      :items="optional(enums.improvements, brush.improvement.value)" item-value="type"
                      item-title="description"
                      :model-value="noneIfEmpty(brush.improvement.value)"
                      @update:model-value="(v: string) => brush.improvement.value = emptyIfNone(v)"/>
    </div>
    <div class="brush-row">
      <v-checkbox-btn class="flex-grow-0" v-model="brush.route.on" density="compact"/>
      <v-select :label="$t('field.route')" density="compact" hide-details :disabled="!brush.route.on"
                :items="optional(enums.routes, brush.route.value)" item-value="type" item-title="description"
                :model-value="noneIfEmpty(brush.route.value)"
                @update:model-value="(v: string) => brush.route.value = emptyIfNone(v)"/>
    </div>

    <v-alert v-if="brushIsEmpty()" type="info" variant="tonal" density="compact" class="mt-3">
      {{ $t('brush.empty') }}
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
