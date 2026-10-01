<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {CreatePlots, GetMapProps, GetMapStats, SetMapProps} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, describeType, enums, mapVersion, withCurrent, withNone, worldSizes} from "../store";

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
watch(mapVersion, load);

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

const wrapOptions = [
  {value: '00', title: 'Flat (no wrapping)'},
  {value: '10', title: 'Cylinder (wraps east-west)'},
  {value: '01', title: 'Wraps north-south only'},
  {value: '11', title: 'Torus (wraps both ways)'},
];

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
    {title: 'Plots', value: `${s.plots} (${props.value?.GridWidth ?? 0}×${props.value?.GridHeight ?? 0})`},
    {title: 'Land', value: `${land} · ${percent(land)}`},
    {title: 'Water', value: `${s.water} · ${percent(s.water)}`},
    {title: 'Hills / peaks', value: `${s.hills} / ${s.peaks}`},
    {title: 'Resources', value: s.bonuses},
    {title: 'Cities', value: s.cities},
    {title: 'Units', value: s.units},
    {title: 'Starting plots', value: s.starting_plots},
    {title: 'Signs', value: s.signs},
  ];
});
</script>

<template>
  <div v-if="!props" class="pa-5 text-grey">
    No map loaded. Open one from the toolbar to start editing.
  </div>

  <div v-else class="pa-4">
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3" closable
             @click:close="error = ''">{{ error }}
    </v-alert>
    <v-alert v-for="problem in stats?.problems ?? []" :key="problem" type="warning" variant="tonal"
             density="compact" class="mb-3">{{ problem }}
    </v-alert>

    <h3 class="mb-3">Size</h3>
    <template v-if="hasPlots">
      <div class="text-body-2 mb-2">
        The map is <b>{{ props.GridWidth }}×{{ props.GridHeight }}</b> plots.
        <span class="text-medium-emphasis">The size of a map with plots can not be changed.</span>
      </div>
    </template>
    <template v-else>
      <div class="text-body-2 text-medium-emphasis mb-2">
        The map has no plots yet. Choose the size and create an ocean map, then edit it in the game's WorldBuilder.
      </div>
      <v-row dense>
        <v-col cols="12" md="4">
          <v-select label="World size" density="compact" hide-details
                    :items="sizeOptions" item-value="type" item-title="description"
                    :model-value="props.WorldSize" @update:model-value="onWorldSizeChange" />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field label="Width" type="number" density="compact" hide-details min="1" max="512"
                        v-model.number="newWidth" />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field label="Height" type="number" density="compact" hide-details min="1" max="512"
                        v-model.number="newHeight" />
        </v-col>
        <v-col cols="12" md="4">
          <v-btn color="primary" prepend-icon="mdi-waves" block :loading="creating" @click="createPlots">
            Create ocean plots
          </v-btn>
        </v-col>
      </v-row>
    </template>

    <v-divider class="my-4" />
    <h3 class="mb-3">World</h3>
    <v-row dense>
      <v-col v-if="hasPlots" cols="12" md="4">
        <v-select label="World size" density="compact" hide-details
                  :items="sizeOptions" item-value="type" item-title="description"
                  v-model="props.WorldSize" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select label="Climate" density="compact" hide-details
                  :items="withCurrent(enums.climates, props.Climate)" item-value="type" item-title="description"
                  v-model="props.Climate" />
      </v-col>
      <v-col cols="12" md="4">
        <v-select label="Sea level" density="compact" hide-details
                  :items="withCurrent(enums.seaLevels, props.SeaLevel)" item-value="type" item-title="description"
                  v-model="props.SeaLevel" />
      </v-col>
      <v-col cols="12" md="6">
        <v-select label="Wrapping" density="compact" hide-details
                  :items="wrapOptions" v-model="wrapping" />
      </v-col>
      <v-col cols="12" md="6" class="d-flex align-center">
        <v-checkbox v-model="props.RandomizeResources" label="Randomize resources at game start"
                    hide-details density="compact" />
      </v-col>
    </v-row>

    <div class="mt-4">
      <div class="text-body-2 mb-1">
        Latitudes: from <b>{{ props.BottomLatitude }}°</b> (bottom) to <b>{{ props.TopLatitude }}°</b> (top)
      </div>
      <v-range-slider v-model="latitudeRange" :min="-90" :max="90" :step="1" strict
                      thumb-label hide-details class="px-2 mt-6" />
      <div class="text-caption text-medium-emphasis">
        Lower values than ±90 remove polar ice caps, e.g. a map of Europe could use 30..70.
      </div>
    </div>

    <v-divider class="my-4" />
    <h3 class="mb-3">Contents</h3>
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
        <div class="text-body-2 mb-2">Terrain</div>
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
