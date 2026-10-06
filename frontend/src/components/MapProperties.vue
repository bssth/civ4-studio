<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {CreatePlots, GetMapProps, GetMapStats, ResizeMap, SetMapProps} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {useI18n} from "vue-i18n";
import {problemText} from "../problems";
import {batched, describeType, enums, mapRevision, mapVersion, withCurrent, withNone, worldSizes} from "../store";

const {t} = useI18n();

const props = ref<editor.MapProps | null>(null);
const stats = ref<editor.MapStats | null>(null);
const error = ref('');
// Size of a new map, used only while the map has no plots
const newWidth = ref(84);
const newHeight = ref(52);
const creating = ref(false);

async function load() {
  const p = await GetMapProps();
  props.value = p ? editor.MapProps.createFrom(p) : null;
  if (p && p.GridWidth > 0 && p.GridHeight > 0) {
    newWidth.value = p.GridWidth;
    newHeight.value = p.GridHeight;
  }
  await refreshStats();
}

async function refreshStats() {
  stats.value = await GetMapStats();
}

onMounted(load);
watch([mapVersion, mapRevision], () => load());

watch(() => props.value && JSON.stringify(props.value), batched(async () => {
  if (!props.value) return;
  try {
    await SetMapProps(editor.MapProps.createFrom(props.value));
    error.value = '';
    await refreshStats();
  } catch (err: any) {
    error.value = String(err);
  }
}));

const hasPlots = computed(() => (stats.value?.plots ?? 0) > 0);

const latitudeRange = computed({
  get: () => props.value ? [props.value.BottomLatitude, props.value.TopLatitude] : [-90, 90],
  set: ([bottom, top]) => {
    if (!props.value) return;
    props.value.BottomLatitude = Math.round(bottom);
    props.value.TopLatitude = Math.round(top);
  },
});

const wrapping = computed({
  get: () => props.value ? `${props.value.WrapX ? 1 : 0}${props.value.WrapY ? 1 : 0}` : '10',
  set: (v: string) => {
    if (!props.value) return;
    props.value.WrapX = v[0] === '1' ? 1 : 0;
    props.value.WrapY = v[1] === '1' ? 1 : 0;
  },
});

const wrapOptions = computed(() => [
  {value: '00', title: t('mapProps.wrapFlat')},
  {value: '10', title: t('mapProps.wrapCylinder')},
  {value: '01', title: t('mapProps.wrapNS')},
  {value: '11', title: t('mapProps.wrapTorus')},
]);

const sizeOptions = computed(() => withCurrent(
    worldSizes.value.map(s => editor.EnumOption.createFrom({type: s.type, description: s.description})),
    props.value?.WorldSize));

function onWorldSizeChange(type: string) {
  if (!props.value) return;
  props.value.WorldSize = type;
  const size = worldSizes.value.find(s => s.type === type);
  // Suggest the default size of the world size for a map that is not created yet
  if (!hasPlots.value && size && size.width > 0 && size.height > 0) {
    newWidth.value = size.width;
    newHeight.value = size.height;
  }
}

// --- Resize -------------------------------------------------------------------

type Side = 'west' | 'east' | 'north' | 'south';
const sides: Side[] = ['west', 'east', 'north', 'south'];
// Columns or rows to add (positive) or remove (negative) on each side
const resize = ref<Record<Side, number>>({west: 0, east: 0, north: 0, south: 0});
const resizing = ref(false);
const resizeResult = ref<editor.ResizeResult | null>(null);

function sideValue(side: Side): number {
  const v = Number(resize.value[side]);
  return Number.isFinite(v) ? Math.trunc(v) : 0;
}

const newSize = computed(() => ({
  width: (props.value?.GridWidth ?? 0) + sideValue('west') + sideValue('east'),
  height: (props.value?.GridHeight ?? 0) + sideValue('north') + sideValue('south'),
}));
const newSizeValid = computed(() => [newSize.value.width, newSize.value.height].every(n => n >= 4 && n <= 512));
const resizeChanged = computed(() => sides.some(s => sideValue(s) !== 0));

// Default size of the chosen world size, if the map differs from it
const worldSizeTarget = computed(() => {
  const p = props.value;
  const size = worldSizes.value.find(s => s.type === p?.WorldSize);
  if (!p || !size || size.width <= 0 || size.height <= 0) return null;
  return size.width === p.GridWidth && size.height === p.GridHeight ? null : size;
});

function fitWorldSize() {
  const p = props.value, size = worldSizeTarget.value;
  if (!p || !size) return;
  // Split the difference between both sides, so the map stays in the middle
  const dw = size.width - p.GridWidth, dh = size.height - p.GridHeight;
  resize.value = {west: Math.trunc(dw / 2), east: dw - Math.trunc(dw / 2), north: dh - Math.trunc(dh / 2), south: Math.trunc(dh / 2)};
}

function clearResize() {
  resize.value = {west: 0, east: 0, north: 0, south: 0};
}

watch(mapVersion, () => {
  clearResize();
  resizeResult.value = null;
});

async function applyResize() {
  resizing.value = true;
  try {
    resizeResult.value = await ResizeMap(sideValue('west'), sideValue('east'), sideValue('north'), sideValue('south'));
    error.value = '';
    clearResize();
    await load();
  } catch (err: any) {
    error.value = String(err);
  } finally {
    resizing.value = false;
  }
}

const resizeResultText = computed(() => {
  const r = resizeResult.value;
  if (!r) return '';
  const parts = [t('resize.done', {width: props.value?.GridWidth ?? 0, height: props.value?.GridHeight ?? 0})];
  if (r.units || r.cities || r.signs) parts.push(t('resize.lost', {units: r.units, cities: r.cities, signs: r.signs}));
  if (r.starts?.length) parts.push(t('resize.startsOutside', {players: r.starts.join(', ')}));
  parts.push(t('resize.undoHint'));
  return parts.join(' ');
});

async function createPlots() {
  creating.value = true;
  try {
    await CreatePlots(newWidth.value, newHeight.value);
    error.value = '';
    await load();
  } catch (err: any) {
    error.value = String(err);
  } finally {
    creating.value = false;
  }
}

function percent(n: number): string {
  const total = stats.value?.plots ?? 0;
  return total ? Math.round(n * 100 / total) + '%' : '—';
}

const statRows = computed(() => {
  const s = stats.value;
  if (!s) return [];
  const land = s.peaks + s.hills + s.flat;
  return [
    {title: t('mapProps.statPlots'), value: `${s.plots} (${props.value?.GridWidth ?? 0}×${props.value?.GridHeight ?? 0})`},
    {title: t('mapProps.statLand'), value: `${land} · ${percent(land)}`},
    {title: t('mapProps.statWater'), value: `${s.water} · ${percent(s.water)}`},
    {title: t('mapProps.statHills'), value: `${s.hills} / ${s.peaks}`},
    {title: t('mapProps.statResources'), value: s.bonuses},
    {title: t('mapProps.statCities'), value: s.cities},
    {title: t('mapProps.statUnits'), value: s.units},
    {title: t('mapProps.statStarts'), value: s.starting_plots},
    {title: t('mapProps.statSigns'), value: s.signs},
  ];
});
</script>

<template>
  <div v-if="!props" class="pa-5 text-grey">
    {{ $t('common.noMapHint') }}
  </div>

  <div v-else class="pa-4">
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3" closable
             @click:close="error = ''">{{ error }}
    </v-alert>
    <v-alert v-for="problem in stats?.problems ?? []" :key="problem.code" type="warning" variant="tonal"
             density="compact" class="mb-3">{{ problemText(problem) }}
    </v-alert>

    <h3 class="mb-3">{{ $t('mapProps.size') }}</h3>
    <template v-if="hasPlots">
      <div class="text-body-2 mb-2">
        {{ $t('mapProps.sizeIs') }} <b>{{ props.GridWidth }}×{{ props.GridHeight }}</b> {{ $t('mapProps.plots') }}
      </div>
      <div class="text-caption text-medium-emphasis mb-2">{{ $t('resize.hint') }}</div>
      <v-row dense class="align-center">
        <v-col v-for="side in sides" :key="side" cols="6" md="2">
          <v-text-field :label="$t(`resize.${side}`)" type="number" density="compact" hide-details
                        v-model.number="resize[side]" />
        </v-col>
        <v-col cols="12" md="4" class="d-flex align-center ga-2">
          <span class="text-body-2 text-no-wrap">→ <b>{{ newSize.width }}×{{ newSize.height }}</b></span>
          <v-btn color="primary" prepend-icon="mdi-resize" :disabled="!resizeChanged || !newSizeValid"
                 :loading="resizing" @click="applyResize">{{ $t('resize.apply') }}</v-btn>
        </v-col>
      </v-row>
      <div class="d-flex align-center flex-wrap ga-2 mt-1">
        <v-btn v-if="worldSizeTarget" size="small" variant="text" prepend-icon="mdi-arrow-expand"
               @click="fitWorldSize">{{ $t('resize.fitWorldSize', {size: worldSizeTarget.description, width: worldSizeTarget.width, height: worldSizeTarget.height}) }}</v-btn>
        <v-btn v-if="resizeChanged" size="small" variant="text" prepend-icon="mdi-close" @click="clearResize">{{ $t('common.cancel') }}</v-btn>
        <span v-if="resizeChanged && !newSizeValid" class="text-caption text-error">{{ $t('resize.invalid', {min: 4, max: 512}) }}</span>
      </div>
      <v-alert v-if="resizeResult" type="info" variant="tonal" density="compact" class="mt-2" closable
               @click:close="resizeResult = null">{{ resizeResultText }}
      </v-alert>
    </template>
    <template v-else>
      <div class="text-body-2 text-medium-emphasis mb-2">
        {{ $t('mapProps.noPlots') }}
      </div>
      <v-row dense>
        <v-col cols="12" md="4">
          <v-select :label="$t('mapProps.worldSize')" density="compact" hide-details
                    :items="sizeOptions" item-value="type" item-title="description"
                    :model-value="props.WorldSize" @update:model-value="onWorldSizeChange" />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field :label="$t('mapProps.width')" type="number" density="compact" hide-details min="1" max="512"
                        v-model.number="newWidth" />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field :label="$t('mapProps.height')" type="number" density="compact" hide-details min="1" max="512"
                        v-model.number="newHeight" />
        </v-col>
        <v-col cols="12" md="4">
          <v-btn color="primary" prepend-icon="mdi-waves" block :loading="creating" @click="createPlots">
            {{ $t('mapProps.createPlots') }}
          </v-btn>
        </v-col>
      </v-row>
    </template>

    <v-divider class="my-4" />
    <h3 class="mb-3">{{ $t('mapProps.world') }}</h3>
    <v-row dense>
      <v-col v-if="hasPlots" cols="12" md="4">
        <v-select :label="$t('mapProps.worldSize')" density="compact" hide-details
                  :items="sizeOptions" item-value="type" item-title="description"
                  v-model="props.WorldSize" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select :label="$t('mapProps.climate')" density="compact" hide-details
                  :items="withCurrent(enums.climates, props.Climate)" item-value="type" item-title="description"
                  v-model="props.Climate" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select :label="$t('mapProps.seaLevel')" density="compact" hide-details
                  :items="withCurrent(enums.seaLevels, props.SeaLevel)" item-value="type" item-title="description"
                  v-model="props.SeaLevel" />
      </v-col>
      <v-col cols="12" md="6">
        <v-select :label="$t('mapProps.wrapping')" density="compact" hide-details
                  :items="wrapOptions" v-model="wrapping" />
      </v-col>
      <v-col cols="12" md="6" class="d-flex align-center">
        <v-checkbox v-model="props.RandomizeResources" :label="$t('mapProps.randomizeResources')"
                    hide-details density="compact" />
      </v-col>
    </v-row>

    <div class="mt-4">
      <div class="text-body-2 mb-1">
        {{ $t('mapProps.latFrom') }} <b>{{ props.BottomLatitude }}°</b> {{ $t('mapProps.latBottom') }} <b>{{ props.TopLatitude }}°</b> {{ $t('mapProps.latTop') }}
      </div>
      <v-range-slider v-model="latitudeRange" :min="-90" :max="90" :step="1" strict
                      thumb-label hide-details class="px-2 mt-6" />
      <div class="text-caption text-medium-emphasis">
        {{ $t('mapProps.latHint') }}
      </div>
    </div>

    <v-divider class="my-4" />
    <h3 class="mb-3">{{ $t('mapProps.contents') }}</h3>
    <v-row dense>
      <v-col cols="12" md="5">
        <v-table density="compact">
          <tbody>
          <tr v-for="row in statRows" :key="row.title">
            <td class="text-medium-emphasis">{{ row.title }}</td>
            <td class="text-right">{{ row.value }}</td>
          </tr>
          </tbody>
        </v-table>
      </v-col>
      <v-col cols="12" md="7">
        <div class="text-body-2 mb-2">{{ $t('mapProps.terrain') }}</div>
        <div v-for="t in stats?.terrains ?? []" :key="t.type" class="d-flex align-center mb-1">
          <div class="terrain-name text-caption text-truncate" :title="t.type">
            {{ describeType(withNone(enums.terrains), t.type) }}
          </div>
          <v-progress-linear :model-value="t.count * 100 / (stats?.plots || 1)" height="8" rounded
                             color="primary" class="mx-2" />
          <div class="text-caption terrain-count">{{ t.count }}</div>
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<style scoped>
.terrain-name {
  width: 140px;
  flex-shrink: 0;
}

.terrain-count {
  width: 50px;
  text-align: right;
  flex-shrink: 0;
}
</style>
