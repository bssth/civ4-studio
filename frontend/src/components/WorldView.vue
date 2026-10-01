<script setup lang="ts">
import {computed, nextTick, onMounted, onUnmounted, reactive, ref, watch} from "vue";
import {GetMapView, GetPlayers, GetPlot, SetPlayerStart} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, describeType, enums, mapVersion, NONE, playerColor, playerName} from "../store";
import {drawMap, Layers, StartMarker} from "../mapRender";
import PlotEditor from "./PlotEditor.vue";

const view = ref<editor.MapView | null>(null);
const players = ref<editor.Player[]>([]);
const cell = ref(8);
const layers = reactive<Layers>({rivers: true, resources: true, cities: true, units: true, starts: true, grid: false});
const layerNames: { key: keyof Layers, title: string, icon: string }[] = [
  {key: 'rivers', title: 'Rivers', icon: 'mdi-waves'},
  {key: 'resources', title: 'Resources', icon: 'mdi-diamond-stone'},
  {key: 'cities', title: 'Cities', icon: 'mdi-home-city'},
  {key: 'units', title: 'Units', icon: 'mdi-chess-pawn'},
  {key: 'starts', title: 'Start positions', icon: 'mdi-flag'},
  {key: 'grid', title: 'Grid', icon: 'mdi-grid'},
];
const activeLayers = computed({
  get: () => layerNames.filter(l => layers[l.key]).map(l => l.key),
  set: (keys: (keyof Layers)[]) => layerNames.forEach(l => layers[l.key] = keys.includes(l.key)),
});

const selected = ref<{ x: number, y: number } | null>(null);
const selectedPlot = ref<editor.Plot | null>(null);
const hover = ref<{ x: number, y: number } | null>(null);
const error = ref('');

const canvas = ref<HTMLCanvasElement | null>(null);
const scroller = ref<HTMLElement | null>(null);
const dpr = window.devicePixelRatio || 1;

async function load(fit = false) {
  const [v, p] = await Promise.all([GetMapView(), GetPlayers()]);
  view.value = v && v.width > 0 && v.height > 0 ? v : null;
  players.value = p ?? [];
  if (fit) {
    await nextTick();
    fitToScreen();
  }
  if (selected.value) await select(selected.value.x, selected.value.y);
}

const refreshView = batched(async () => {
  const v = await GetMapView();
  view.value = v && v.width > 0 && v.height > 0 ? v : null;
});

onMounted(() => load(true));
watch(mapVersion, () => {
  selected.value = null;
  selectedPlot.value = null;
  load(true);
});

const maxCell = computed(() => {
  if (!view.value) return 32;
  // Browsers limit canvas size, keep it well below the limit
  return Math.max(2, Math.min(32, Math.floor(8000 / dpr / Math.max(view.value.width, view.value.height))));
});

function fitToScreen() {
  if (!view.value || !scroller.value) return;
  const byWidth = Math.floor((scroller.value.clientWidth - 4) / view.value.width);
  const byHeight = Math.floor((scroller.value.clientHeight - 4) / view.value.height);
  cell.value = Math.max(2, Math.min(maxCell.value, byWidth, byHeight));
}

// Start position being dragged
const drag = ref<{ player: number, x: number, y: number, moved: boolean } | null>(null);

const starts = computed<StartMarker[]>(() => players.value
    .map((p, i) => ({p, i}))
    .filter(({p}) => p.CivType && p.CivType !== NONE)
    .map(({p, i}) => {
      const dragged = drag.value?.player === i;
      return {
        player: i,
        x: dragged ? drag.value!.x : p.StartingX,
        y: dragged ? drag.value!.y : p.StartingY,
        color: playerColor(players.value, i),
        random: p.RandomStartLocation,
      };
    }));

let frame = 0;

function scheduleDraw() {
  cancelAnimationFrame(frame);
  frame = requestAnimationFrame(draw);
}

function draw() {
  const c = canvas.value, v = view.value;
  if (!c || !v) return;
  const size = cell.value;
  c.width = v.width * size * dpr;
  c.height = v.height * size * dpr;
  c.style.width = `${v.width * size}px`;
  c.style.height = `${v.height * size}px`;
  const ctx = c.getContext('2d');
  if (!ctx) return;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  drawMap(ctx, v, {
    cell: size,
    layers,
    ownerColor: owner => playerColor(players.value, owner),
    starts: starts.value,
    selected: selected.value,
  });
}

watch([view, cell, layers, selected, starts, () => enums.colors], scheduleDraw, {deep: true});
onUnmounted(() => cancelAnimationFrame(frame));

function cellAt(e: MouseEvent): { x: number, y: number } | null {
  const v = view.value;
  if (!v) return null;
  const x = Math.floor(e.offsetX / cell.value);
  const row = Math.floor(e.offsetY / cell.value);
  if (x < 0 || x >= v.width || row < 0 || row >= v.height) return null;
  return {x, y: v.height - 1 - row};
}

function onMouseDown(e: MouseEvent) {
  const at = cellAt(e);
  if (!at || e.button !== 0) return;
  const marker = layers.starts ? starts.value.find(s => s.x === at.x && s.y === at.y) : undefined;
  drag.value = marker ? {player: marker.player, x: at.x, y: at.y, moved: false} : null;
}

function onMouseMove(e: MouseEvent) {
  const at = cellAt(e);
  hover.value = at;
  if (drag.value && at && (drag.value.x !== at.x || drag.value.y !== at.y)) {
    drag.value = {...drag.value, x: at.x, y: at.y, moved: true};
  }
}

async function onMouseUp(e: MouseEvent) {
  const at = cellAt(e);
  const d = drag.value;
  drag.value = null;
  if (d && d.moved && at) {
    try {
      await SetPlayerStart(d.player, at.x, at.y);
      const p = players.value[d.player];
      p.StartingX = at.x;
      p.StartingY = at.y;
      p.RandomStartLocation = false;
      error.value = '';
    } catch (err: any) {
      error.value = String(err);
    }
    return;
  }
  if (at) await select(at.x, at.y);
}

function onMouseLeave() {
  hover.value = null;
  drag.value = null;
}

async function select(x: number, y: number) {
  selected.value = {x, y};
  selectedPlot.value = await GetPlot(x, y);
}

function onWheel(e: WheelEvent) {
  if (!e.ctrlKey) return;
  e.preventDefault();
  cell.value = Math.max(2, Math.min(maxCell.value, cell.value + (e.deltaY < 0 ? 1 : -1)));
}

const hoverText = computed(() => {
  const v = view.value, h = hover.value;
  if (!v || !h) return '';
  const i = h.y * v.width + h.x;
  const parts = [`${h.x}, ${h.y}`];
  const t = v.terrain[i];
  if (t >= 0) parts.push(describeType(enums.terrains, v.terrains[t]));
  parts.push(['Peak', 'Hills', 'Flat', 'Water'][v.plot_type[i]] ?? '');
  if (v.feature[i] >= 0) parts.push(describeType(enums.features, v.features[v.feature[i]]));
  if (v.bonus[i] >= 0) parts.push(describeType(enums.bonuses, v.bonuses[v.bonus[i]]));
  if (v.city_owner[i] >= 0) parts.push(`city of ${playerName(players.value, v.city_owner[i])}`);
  if (v.unit_count[i] > 0) parts.push(`${v.unit_count[i]} unit(s) of ${playerName(players.value, v.unit_owner[i])}`);
  const here = starts.value.filter(s => s.x === h.x && s.y === h.y);
  for (const s of here) parts.push(`start of #${s.player} ${playerName(players.value, s.player)}${s.random ? ' (random start)' : ''}`);
  return parts.filter(Boolean).join(' · ');
});

const canvasCursor = computed(() => {
  if (drag.value) return 'grabbing';
  const h = hover.value;
  if (h && layers.starts && starts.value.some(s => s.x === h.x && s.y === h.y)) return 'grab';
  return 'crosshair';
});
</script>

<template>
  <div v-if="!view" class="pa-5 text-grey">
    The map has no plots. Open a map, or create plots for a new map on the Map tab.
  </div>

  <div v-else class="world d-flex fill-height">
    <div class="d-flex flex-column flex-grow-1" style="min-width: 0">
      <div class="d-flex align-center flex-wrap px-2 pt-2" style="gap: 8px">
        <v-btn-toggle v-model="activeLayers" multiple density="compact" divided variant="outlined">
          <v-btn v-for="l in layerNames" :key="l.key" :value="l.key" size="small" :title="l.title">
            <v-icon :icon="l.icon"/>
          </v-btn>
        </v-btn-toggle>
        <v-spacer/>
        <v-icon icon="mdi-magnify-minus-outline" size="small"/>
        <v-slider v-model="cell" :min="2" :max="maxCell" :step="1" hide-details density="compact"
                  style="max-width: 180px; min-width: 120px"/>
        <v-icon icon="mdi-magnify-plus-outline" size="small"/>
        <v-btn size="small" variant="text" prepend-icon="mdi-fit-to-screen-outline" @click="fitToScreen">Fit</v-btn>
      </div>
      <div class="px-3 py-1 text-caption text-medium-emphasis text-truncate hover-line">
        {{ hoverText || 'Click a plot to edit it. Drag a flag to move a start position. Ctrl+wheel zooms.' }}
      </div>
      <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mx-2 mb-1" closable
               @click:close="error = ''">{{ error }}
      </v-alert>
      <div ref="scroller" class="scroller flex-grow-1" @wheel="onWheel">
        <canvas ref="canvas" :style="{cursor: canvasCursor}"
                @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp"
                @mouseleave="onMouseLeave"/>
      </div>
    </div>

    <div class="editor-panel">
      <PlotEditor v-if="selectedPlot" :plot="selectedPlot" :players="players" @changed="refreshView"/>
      <div v-else class="pa-4 text-grey text-body-2">
        <template v-if="selected">There is no plot at {{ selected.x }}, {{ selected.y }}.</template>
        <template v-else>Select a plot on the map to see and edit its terrain, resources, cities and units.</template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.world {
  height: 100%;
}

.scroller {
  overflow: auto;
  min-height: 0;
  background: #0b1e36;
  margin: 0 8px 8px;
  border-radius: 4px;
}

.scroller canvas {
  display: block;
  image-rendering: pixelated;
}

.hover-line {
  min-height: 22px;
}

.editor-panel {
  width: 380px;
  flex-shrink: 0;
  overflow-y: auto;
  border-left: 1px solid rgba(0, 0, 0, 0.12);
}
</style>
