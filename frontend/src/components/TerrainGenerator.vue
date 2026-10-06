<script setup lang="ts">
import {nextTick, onMounted, reactive, ref, watch} from "vue";
import {GenerateTerrain, GetMapView, GetPlayers, PlaceStarts} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {mapRevision, mapVersion, NONE, playerColor} from "../store";
import {drawMap, StartMarker} from "../mapRender";
import {useI18n} from "vue-i18n";

const emit = defineEmits<{ (e: 'changed'): void }>();
const {t} = useI18n();

function randomSeed(): number {
  return Math.floor(Math.random() * 1_000_000);
}

const options = reactive({
  seed: randomSeed(), land: 35, continents: 3, hills: 15, peaks: 5, forests: 40, rivers: 15, resources: true,
});
const busy = ref(false);
const error = ref('');
const message = ref('');
const confirmOpen = ref(false);
const canvas = ref<HTMLCanvasElement | null>(null);

const sliders: { key: 'land' | 'continents' | 'hills' | 'peaks' | 'forests' | 'rivers', min: number, max: number, unit: string }[] = [
  {key: 'land', min: 5, max: 95, unit: '%'},
  {key: 'continents', min: 1, max: 12, unit: ''},
  {key: 'hills', min: 0, max: 50, unit: '%'},
  {key: 'peaks', min: 0, max: 30, unit: '%'},
  {key: 'forests', min: 0, max: 100, unit: '%'},
  {key: 'rivers', min: 0, max: 100, unit: ''},
];

/** A small picture of the whole map, so the result is seen without going to the World tab */
async function drawPreview() {
  const [view, players] = await Promise.all([GetMapView(), GetPlayers()]);
  await nextTick();
  const c = canvas.value;
  if (!c || !view || view.width === 0) return;
  const cell = Math.max(1, Math.min(6, Math.floor(420 / view.width)));
  c.width = view.width * cell;
  c.height = view.height * cell;
  const ctx = c.getContext('2d');
  if (!ctx) return;
  const list = players ?? [];
  const starts: StartMarker[] = list
      .map((p, i) => ({p, i}))
      .filter(({p}) => p.CivType && p.CivType !== NONE)
      .map(({p, i}) => ({player: i, x: p.StartingX, y: p.StartingY, color: playerColor(list, i), random: p.RandomStartLocation}));
  drawMap(ctx, view, {
    cell,
    layers: {rivers: cell >= 3, resources: false, cities: true, units: false, starts: true, signs: false, grid: false},
    ownerColor: owner => playerColor(list, owner),
    starts,
    selected: null,
  });
}

onMounted(drawPreview);
watch([mapVersion, mapRevision], drawPreview);

async function generate() {
  confirmOpen.value = false;
  busy.value = true;
  try {
    const r = await GenerateTerrain(editor.TerrainOptions.createFrom({...options}));
    message.value = t('generator.done', {land: r.land, rivers: r.rivers, resources: r.resources});
    error.value = '';
    emit('changed');
    await drawPreview();
  } catch (err: any) {
    error.value = String(err);
  } finally {
    busy.value = false;
  }
}

async function placeStarts() {
  busy.value = true;
  try {
    const n = await PlaceStarts(options.seed);
    message.value = t('generator.startsPlaced', {n});
    error.value = '';
    emit('changed');
    await drawPreview();
  } catch (err: any) {
    error.value = String(err);
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div>
    <h3 class="mb-1">{{ $t('generator.title') }}</h3>
    <div class="text-caption text-medium-emphasis mb-3">{{ $t('generator.hint') }}</div>
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3" closable
             @click:close="error = ''">{{ error }}
    </v-alert>
    <v-alert v-if="message" type="success" variant="tonal" density="compact" class="mb-3" closable
             @click:close="message = ''">{{ message }}
    </v-alert>
    <v-row dense>
      <v-col cols="12" md="6">
        <div class="d-flex align-center mb-2" style="gap: 8px">
          <v-text-field :label="$t('generator.seed')" type="number" density="compact" hide-details
                        v-model.number="options.seed" style="max-width: 180px"/>
          <v-btn icon="mdi-dice-multiple" size="small" variant="text" :title="$t('generator.newSeed')"
                 @click="options.seed = randomSeed()"/>
          <v-checkbox v-model="options.resources" :label="$t('generator.resources')" density="compact" hide-details/>
        </div>
        <div v-for="s in sliders" :key="s.key" class="d-flex align-center">
          <span class="text-body-2" style="width: 150px">{{ $t('generator.' + s.key) }}</span>
          <v-slider v-model="options[s.key]" :min="s.min" :max="s.max" :step="1" hide-details density="compact"
                    class="flex-grow-1"/>
          <span class="text-body-2 text-end" style="width: 48px">{{ options[s.key] }}{{ s.unit }}</span>
        </div>
        <div class="d-flex flex-wrap mt-3" style="gap: 8px">
          <v-btn color="primary" prepend-icon="mdi-earth" :loading="busy" @click="confirmOpen = true">
            {{ $t('generator.generate') }}
          </v-btn>
          <v-btn variant="tonal" prepend-icon="mdi-flag-variant" :loading="busy" @click="placeStarts">
            {{ $t('generator.placeStarts') }}
          </v-btn>
        </div>
      </v-col>
      <v-col cols="12" md="6" class="d-flex justify-center align-start">
        <canvas ref="canvas" class="preview" :title="$t('generator.preview')"/>
      </v-col>
    </v-row>

    <v-dialog v-model="confirmOpen" max-width="480">
      <v-card :title="$t('generator.confirmTitle')">
        <v-card-text>{{ $t('generator.confirmText') }}</v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="confirmOpen = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="primary" variant="tonal" @click="generate">{{ $t('generator.generate') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>
.preview {
  max-width: 100%;
  image-rendering: pixelated;
  border: 1px solid rgba(0, 0, 0, 0.15);
}
</style>
