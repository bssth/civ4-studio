<script setup lang="ts">
import {computed, nextTick, onMounted, onUnmounted, reactive, ref, shallowRef, watch} from "vue";
import {
  GetMapView,
  GetPlayers,
  GetPlot,
  HistoryState,
  PaintPlots,
  Redo,
  SetPlayerStart,
  SetPlot,
  Undo
} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {
  batched,
  brush,
  brushIsEmpty,
  brushOperation,
  describeType,
  enums,
  focusPlot,
  mapVersion,
  NONE,
  playerColor,
  playerName
} from "../store";
import {
  brushCells,
  drawMap,
  FLAG_IMPROVEMENT,
  FLAG_ROUTE,
  Layers,
  lineCells,
  nearestEdge,
  PLOT_LAND,
  PLOT_OCEAN,
  RiverEdge,
  StartMarker
} from "../mapRender";
import PlotEditor from "./PlotEditor.vue";
import PaintPanel from "./PaintPanel.vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();

const heightNames = computed(() => [t('height.peak'), t('height.hills'), t('height.flat'), t('height.water')]);

/** Text of a history label from the backend: "paint:<plots>", "plot:<x>,<y>", "start:<player>" */
function historyLabel(label: string): string {
  const [kind, value = ''] = label.split(':');
  switch (kind) {
    case 'paint':
      return t('history.paint', {n: value});
    case 'plot': {
      const [x, y] = value.split(',');
      return t('history.plot', {x, y});
    }
    case 'start':
      return t('history.start', {n: value});
  }
  return label;
}

type Mode = 'select' | 'paint' | 'river';

// The view holds big arrays, it is replaced or redrawn explicitly instead of being deeply reactive
const view = shallowRef<editor.MapView | null>(null);
const players = ref<editor.Player[]>([]);
const cell = ref(8);
const mode = ref<Mode>('select');
const layers = reactive<Layers>({rivers: true, resources: true, cities: true, units: true, starts: true, grid: false});
const layerNames: { key: keyof Layers, title: string, icon: string }[] = [
  {key: 'rivers', title: 'layers.rivers', icon: 'mdi-waves'},
  {key: 'resources', title: 'layers.resources', icon: 'mdi-diamond-stone'},
  {key: 'cities', title: 'layers.cities', icon: 'mdi-home-city'},
  {key: 'units', title: 'layers.units', icon: 'mdi-chess-pawn'},
  {key: 'starts', title: 'layers.starts', icon: 'mdi-flag'},
  {key: 'grid', title: 'layers.grid', icon: 'mdi-grid'},
];
const activeLayers = computed({
  get: () => layerNames.filter(l => layers[l.key]).map(l => l.key),
  set: (keys: (keyof Layers)[]) => layerNames.forEach(l => layers[l.key] = keys.includes(l.key)),
});

const selected = ref<{ x: number, y: number } | null>(null);
const selectedPlot = ref<editor.Plot | null>(null);
const hover = ref<{ x: number, y: number } | null>(null);
const hoverEdge = ref<RiverEdge | null>(null);
const error = ref('');
const history = ref<editor.HistoryState>(editor.HistoryState.createFrom({undo: '', redo: ''}));

const canvas = ref<HTMLCanvasElement | null>(null);
const scroller = ref<HTMLElement | null>(null);
const dpr = window.devicePixelRatio || 1;

function setView(v: editor.MapView | null) {
  view.value = v && v.width > 0 && v.height > 0 ? v : null;
}

async function refreshHistory() {
  history.value = await HistoryState();
}

async function load(fit = false) {
  const [v, p] = await Promise.all([GetMapView(), GetPlayers()]);
  setView(v);
  players.value = p ?? [];
  await refreshHistory();
  if (fit) {
    await nextTick();
    fitToScreen();
  }
  if (selected.value) await select(selected.value.x, selected.value.y);
}

const refreshView = batched(async () => {
  setView(await GetMapView());
  await refreshHistory();
});

onMounted(async () => {
  await load(true);
  await consumeFocus();
  window.addEventListener('keydown', onKeyDown);
});

onUnmounted(() => {
  cancelAnimationFrame(frame);
  window.removeEventListener('keydown', onKeyDown);
});

// A plot requested from another tab (e.g. the problem list): zoom in, center it and open its editor
async function consumeFocus() {
  const target = focusPlot.value;
  if (!target || !view.value) return;
  focusPlot.value = null;
  mode.value = 'select';
  cell.value = Math.max(cell.value, Math.min(12, maxCell.value));
  await select(target.x, target.y);
  await nextTick();
  const el = scroller.value;
  if (el) {
    el.scrollLeft = (target.x + 0.5) * cell.value - el.clientWidth / 2;
    el.scrollTop = (view.value.height - target.y - 0.5) * cell.value - el.clientHeight / 2;
  }
}

watch(focusPlot, consumeFocus);
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
  if (c.width !== v.width * size * dpr || c.height !== v.height * size * dpr) {
    c.width = v.width * size * dpr;
    c.height = v.height * size * dpr;
    c.style.width = `${v.width * size}px`;
    c.style.height = `${v.height * size}px`;
  }
  const ctx = c.getContext('2d');
  if (!ctx) return;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  drawMap(ctx, v, {
    cell: size,
    layers,
    ownerColor: owner => playerColor(players.value, owner),
    starts: starts.value,
    selected: mode.value === 'select' ? selected.value : null,
    brush: mode.value === 'paint' && hover.value ? {...hover.value, size: brush.size} : null,
    edge: mode.value === 'river' ? hoverEdge.value : null,
  });
}

watch([view, cell, selected, hover, hoverEdge, mode, () => brush.size], scheduleDraw);
watch([layers, starts, () => enums.colors], scheduleDraw, {deep: true});

function cellAt(e: MouseEvent): { x: number, y: number } | null {
  const v = view.value;
  if (!v) return null;
  const x = Math.floor(e.offsetX / cell.value);
  const row = Math.floor(e.offsetY / cell.value);
  if (x < 0 || x >= v.width || row < 0 || row >= v.height) return null;
  return {x, y: v.height - 1 - row};
}

// --- Painting -------------------------------------------------------------

// Cells of the current stroke, sent to the backend as one undoable step when the mouse is released
let stroke: { cells: Map<string, { x: number, y: number }>, last: { x: number, y: number } } | null = null;

function isWaterName(terrain: string): boolean {
  return /OCEAN|COAST|LAKE/.test(terrain);
}

/** Shows the stroke immediately; the backend result replaces it after the stroke */
function preview(cells: { x: number, y: number }[]) {
  const v = view.value;
  if (!v) return;
  const op = brushOperation();
  const dictionary = (list: string[], value: string) => {
    let i = list.indexOf(value);
    if (i < 0) {
      list.push(value);
      i = list.length - 1;
    }
    return i;
  };
  for (const c of cells) {
    const i = c.y * v.width + c.x;
    if (op.terrain !== undefined) {
      v.terrain[i] = dictionary(v.terrains, op.terrain);
      if (op.plot_type === undefined) {
        const water = isWaterName(op.terrain);
        if (water) v.plot_type[i] = PLOT_OCEAN;
        else if (v.plot_type[i] === PLOT_OCEAN) v.plot_type[i] = PLOT_LAND;
      }
    }
    if (op.plot_type !== undefined) v.plot_type[i] = op.plot_type;
    if (op.feature !== undefined) v.feature[i] = op.feature ? dictionary(v.features, op.feature) : -1;
    if (op.bonus !== undefined) v.bonus[i] = op.bonus ? dictionary(v.bonuses, op.bonus) : -1;
    if (op.improvement !== undefined) v.flags[i] = op.improvement ? v.flags[i] | FLAG_IMPROVEMENT : v.flags[i] & ~FLAG_IMPROVEMENT;
    if (op.route !== undefined) v.flags[i] = op.route ? v.flags[i] | FLAG_ROUTE : v.flags[i] & ~FLAG_ROUTE;
  }
  scheduleDraw();
}

function paintAt(from: { x: number, y: number }, to: { x: number, y: number }) {
  const v = view.value;
  if (!v || !stroke) return;
  const added = [];
  for (const point of lineCells(from.x, from.y, to.x, to.y)) {
    for (const c of brushCells(v, point.x, point.y, brush.size)) {
      const key = `${c.x},${c.y}`;
      if (!stroke.cells.has(key)) {
        stroke.cells.set(key, c);
        added.push(c);
      }
    }
  }
  stroke.last = to;
  preview(added);
}

async function finishStroke() {
  const s = stroke;
  stroke = null;
  if (!s || s.cells.size === 0) return;
  try {
    await PaintPlots(editor.PaintOp.createFrom({...brushOperation(), cells: [...s.cells.values()]}));
    error.value = '';
  } catch (err: any) {
    error.value = String(err);
  }
  setView(await GetMapView());
  await refreshHistory();
  if (selected.value) selectedPlot.value = await GetPlot(selected.value.x, selected.value.y);
}

// --- Rivers ---------------------------------------------------------------

function edgeAt(e: MouseEvent): RiverEdge | null {
  const v = view.value, at = cellAt(e);
  if (!v || !at) return null;
  const fx = e.offsetX / cell.value - at.x;
  const fy = e.offsetY / cell.value - (v.height - 1 - at.y);
  return nearestEdge(v, at.x, at.y, fx, fy);
}

// Flow directions of the game: 0 north, 1 east, 2 south, 3 west. Without a river the direction is not
// written to the file, so a new river always gets the default direction.
async function toggleRiver(edge: RiverEdge) {
  const plot = await GetPlot(edge.x, edge.y);
  if (!plot) return;
  const p = editor.Plot.createFrom(plot);
  if (edge.side === 'south') {
    p.IsNOfRiver = !p.IsNOfRiver;
    if (p.IsNOfRiver) p.RiverWEDirection = 1;
  } else {
    p.IsWOfRiver = !p.IsWOfRiver;
    if (p.IsWOfRiver) p.RiverNSDirection = 2;
  }
  try {
    await SetPlot(p);
    error.value = '';
  } catch (err: any) {
    error.value = String(err);
  }
  setView(await GetMapView());
  await refreshHistory();
  if (selected.value) selectedPlot.value = await GetPlot(selected.value.x, selected.value.y);
}

// --- Mouse ----------------------------------------------------------------

function onMouseDown(e: MouseEvent) {
  const at = cellAt(e);
  if (!at || e.button !== 0) return;
  if (mode.value === 'paint') {
    if (brushIsEmpty()) {
      error.value = t('world.brushEmpty');
      return;
    }
    stroke = {cells: new Map(), last: at};
    paintAt(at, at);
    return;
  }
  if (mode.value === 'river') {
    const edge = edgeAt(e);
    if (edge) toggleRiver(edge);
    return;
  }
  const marker = layers.starts ? starts.value.find(s => s.x === at.x && s.y === at.y) : undefined;
  drag.value = marker ? {player: marker.player, x: at.x, y: at.y, moved: false} : null;
}

function onMouseMove(e: MouseEvent) {
  const at = cellAt(e);
  hover.value = at;
  if (mode.value === 'river') {
    hoverEdge.value = edgeAt(e);
    return;
  }
  if (stroke && at && (e.buttons & 1) && (at.x !== stroke.last.x || at.y !== stroke.last.y)) {
    paintAt(stroke.last, at);
    return;
  }
  if (drag.value && at && (drag.value.x !== at.x || drag.value.y !== at.y)) {
    drag.value = {...drag.value, x: at.x, y: at.y, moved: true};
  }
}

async function onMouseUp(e: MouseEvent) {
  if (stroke) {
    await finishStroke();
    return;
  }
  if (mode.value !== 'select') return;
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
    await refreshHistory();
    return;
  }
  if (at) await select(at.x, at.y);
}

function onMouseLeave() {
  hover.value = null;
  hoverEdge.value = null;
  drag.value = null;
  if (stroke) finishStroke();
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

// --- Undo -----------------------------------------------------------------

async function step(undo: boolean) {
  try {
    history.value = undo ? await Undo() : await Redo();
    error.value = '';
  } catch (err: any) {
    error.value = String(err);
    return;
  }
  // Undo may change plots and start positions
  const [v, p] = await Promise.all([GetMapView(), GetPlayers()]);
  setView(v);
  players.value = p ?? [];
  if (selected.value) selectedPlot.value = await GetPlot(selected.value.x, selected.value.y);
}

function onKeyDown(e: KeyboardEvent) {
  if (!(e.ctrlKey || e.metaKey)) return;
  // Text fields keep their own undo
  const target = e.target as HTMLElement | null;
  if (target && (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName))) return;
  const key = e.key.toLowerCase();
  if (key === 'z' && !e.shiftKey && history.value.undo) {
    e.preventDefault();
    step(true);
  } else if ((key === 'y' || (key === 'z' && e.shiftKey)) && history.value.redo) {
    e.preventDefault();
    step(false);
  }
}

const hoverText = computed(() => {
  const v = view.value, h = hover.value;
  if (!v || !h) return '';
  const i = h.y * v.width + h.x;
  const parts = [`${h.x}, ${h.y}`];
  const terrain = v.terrain[i];
  if (terrain >= 0) parts.push(describeType(enums.terrains, v.terrains[terrain]));
  parts.push(heightNames.value[v.plot_type[i]] ?? '');
  if (v.feature[i] >= 0) parts.push(describeType(enums.features, v.features[v.feature[i]]));
  if (v.bonus[i] >= 0) parts.push(describeType(enums.bonuses, v.bonuses[v.bonus[i]]));
  if (v.city_owner[i] >= 0) parts.push(t('world.cityOf', {player: playerName(players.value, v.city_owner[i])}));
  if (v.unit_count[i] > 0) parts.push(t('world.unitsOf', {n: v.unit_count[i], player: playerName(players.value, v.unit_owner[i])}));
  const here = starts.value.filter(s => s.x === h.x && s.y === h.y);
  for (const s of here) {
    parts.push(t('world.startOf', {n: s.player, player: playerName(players.value, s.player)}) + (s.random ? ' ' + t('world.randomStart') : ''));
  }
  return parts.filter(Boolean).join(' · ');
});

const hints: Record<Mode, string> = {
  select: 'world.hintSelect',
  paint: 'world.hintPaint',
  river: 'world.hintRiver',
};

const canvasCursor = computed(() => {
  if (mode.value !== 'select') return 'crosshair';
  if (drag.value) return 'grabbing';
  const h = hover.value;
  if (h && layers.starts && starts.value.some(s => s.x === h.x && s.y === h.y)) return 'grab';
  return 'crosshair';
});
</script>

<template>
  <div v-if="!view" class="pa-5 text-grey">
    {{ $t('world.noPlots') }}
  </div>

  <div v-else class="world d-flex fill-height">
    <div class="d-flex flex-column flex-grow-1" style="min-width: 0">
      <div class="d-flex align-center flex-wrap px-2 pt-2" style="gap: 8px">
        <v-btn-toggle v-model="mode" mandatory density="compact" divided variant="outlined" color="primary">
          <v-btn value="select" size="small" :title="$t('world.modeSelect')">
            <v-icon icon="mdi-cursor-default"/>
          </v-btn>
          <v-btn value="paint" size="small" :title="$t('world.modePaint')">
            <v-icon icon="mdi-brush"/>
          </v-btn>
          <v-btn value="river" size="small" :title="$t('world.modeRiver')">
            <v-icon icon="mdi-current-ac"/>
          </v-btn>
        </v-btn-toggle>
        <v-btn icon="mdi-undo" size="small" variant="text" :disabled="!history.undo"
               :title="history.undo ? $t('world.undo', {what: historyLabel(history.undo)}) : $t('world.nothingToUndo')" @click="step(true)"/>
        <v-btn icon="mdi-redo" size="small" variant="text" :disabled="!history.redo"
               :title="history.redo ? $t('world.redo', {what: historyLabel(history.redo)}) : $t('world.nothingToRedo')" @click="step(false)"/>
        <v-btn-toggle v-model="activeLayers" multiple density="compact" divided variant="outlined">
          <v-btn v-for="l in layerNames" :key="l.key" :value="l.key" size="small" :title="$t(l.title)">
            <v-icon :icon="l.icon"/>
          </v-btn>
        </v-btn-toggle>
        <v-spacer/>
        <v-icon icon="mdi-magnify-minus-outline" size="small"/>
        <v-slider v-model="cell" :min="2" :max="maxCell" :step="1" hide-details density="compact"
                  style="max-width: 180px; min-width: 120px"/>
        <v-icon icon="mdi-magnify-plus-outline" size="small"/>
        <v-btn size="small" variant="text" prepend-icon="mdi-fit-to-screen-outline" @click="fitToScreen">{{ $t('world.fit') }}</v-btn>
      </div>
      <div class="px-3 py-1 text-caption text-medium-emphasis text-truncate hover-line">
        {{ hoverText || $t(hints[mode]) }}
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
      <PaintPanel v-if="mode === 'paint'"/>
      <div v-else-if="mode === 'river'" class="pa-3">
        <h3 class="mb-1">{{ $t('world.riversTitle') }}</h3>
        <div class="text-body-2">{{ $t('world.riversHelp') }}</div>
        <div class="text-caption text-medium-emphasis mt-2">{{ $t('world.riversDirection') }}</div>
      </div>
      <PlotEditor v-else-if="selectedPlot" :plot="selectedPlot" :players="players" @changed="refreshView"/>
      <div v-else class="pa-4 text-grey text-body-2">
        <template v-if="selected">{{ $t('world.noPlotAt', {x: selected.x, y: selected.y}) }}</template>
        <template v-else>{{ $t('world.selectHint') }}</template>
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
