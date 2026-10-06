<script setup lang="ts">
import {computed, nextTick, onMounted, onUnmounted, reactive, ref, shallowRef, watch} from "vue";
import {
  GetMapProps,
  GetMapView,
  GetPlayers,
  GetPlot,
  GetSigns,
  PaintPlots,
  SetPlayerStart,
  SetPlot
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
  history,
  historyLabel,
  mapRevision,
  mapVersion,
  NONE,
  playerColor,
  playerName,
  refreshHistory,
  stepHistory
} from "../store";
import {
  brushCells,
  drawMap,
  FLAG_IMPROVEMENT,
  FLAG_ROUTE,
  FLAG_SIGN,
  Layers,
  lineCells,
  mapColumn,
  mod,
  nearestEdge,
  PLOT_LAND,
  PLOT_OCEAN,
  RiverEdge,
  screenColumn,
  StartMarker
} from "../mapRender";
import PlotEditor from "./PlotEditor.vue";
import PlotSigns from "./PlotSigns.vue";
import PaintPanel from "./PaintPanel.vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();

const heightNames = computed(() => [t('height.peak'), t('height.hills'), t('height.flat'), t('height.water')]);

type Mode = 'select' | 'paint' | 'river';

// The view holds big arrays, it is replaced or redrawn explicitly instead of being deeply reactive
const view = shallowRef<editor.MapView | null>(null);
const players = ref<editor.Player[]>([]);
const signs = ref<editor.Sign[]>([]);
const cell = ref(8);
// The map wraps east-west (as almost all maps do): it can be shifted to move the seam out of the way
const wrapX = ref(false);
// Map column shown at the left edge
const offset = ref(0);
const mode = ref<Mode>('select');
const layers = reactive<Layers>({rivers: true, resources: true, cities: true, units: true, starts: true, signs: true, grid: false});
const layerNames: { key: keyof Layers, title: string, icon: string }[] = [
  {key: 'rivers', title: 'layers.rivers', icon: 'mdi-waves'},
  {key: 'resources', title: 'layers.resources', icon: 'mdi-diamond-stone'},
  {key: 'cities', title: 'layers.cities', icon: 'mdi-home-city'},
  {key: 'units', title: 'layers.units', icon: 'mdi-chess-pawn'},
  {key: 'starts', title: 'layers.starts', icon: 'mdi-flag'},
  {key: 'signs', title: 'layers.signs', icon: 'mdi-sign-text'},
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

const canvas = ref<HTMLCanvasElement | null>(null);
const scroller = ref<HTMLElement | null>(null);
const dpr = window.devicePixelRatio || 1;

function setView(v: editor.MapView | null) {
  view.value = v && v.width > 0 && v.height > 0 ? v : null;
}

async function load(fit = false) {
  const [v, p, props, s] = await Promise.all([GetMapView(), GetPlayers(), GetMapProps(), GetSigns()]);
  setView(v);
  players.value = p ?? [];
  signs.value = s ?? [];
  wrapX.value = !!props && props.WrapX !== 0;
  if (!wrapX.value || !view.value) offset.value = 0;
  else offset.value = mod(offset.value, view.value.width);
  await refreshHistory();
  if (fit) {
    await nextTick();
    fitToScreen();
  }
  if (selected.value) await select(selected.value.x, selected.value.y);
}

const refreshView = batched(async () => {
  const [v, s] = await Promise.all([GetMapView(), GetSigns()]);
  setView(v);
  signs.value = s ?? [];
  await refreshHistory();
});

onMounted(async () => {
  await load(true);
  await consumeFocus();
});

onUnmounted(() => {
  cancelAnimationFrame(frame);
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
    el.scrollLeft = (column(target.x) + 0.5) * cell.value - el.clientWidth / 2;
    el.scrollTop = (view.value.height - target.y - 0.5) * cell.value - el.clientHeight / 2;
  }
}

watch(focusPlot, consumeFocus);
watch(mapVersion, () => {
  selected.value = null;
  selectedPlot.value = null;
  offset.value = 0;
  load(true);
});
// Undo and redo may change plots and start positions
watch(mapRevision, () => load());

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
    offset: offset.value,
    wrapX: wrapX.value,
  });
}

watch([view, cell, selected, hover, hoverEdge, mode, offset, () => brush.size], scheduleDraw);
watch([layers, starts, () => enums.colors], scheduleDraw, {deep: true});

/** Screen column of a map column, see offset */
function column(x: number): number {
  return view.value ? screenColumn(x, view.value.width, offset.value) : x;
}

function cellAt(e: MouseEvent): { x: number, y: number } | null {
  const v = view.value;
  if (!v) return null;
  const col = Math.floor(e.offsetX / cell.value);
  const row = Math.floor(e.offsetY / cell.value);
  if (col < 0 || col >= v.width || row < 0 || row >= v.height) return null;
  return {x: mapColumn(col, v.width, offset.value), y: v.height - 1 - row};
}

/** Moves the seam of a wrapping map by the given number of columns */
function shift(columns: number) {
  if (!view.value) return;
  offset.value = mod(offset.value + columns, view.value.width);
}

/** Shifts the map so that the selected plot (or the seam, if nothing is selected) is in the middle of the screen */
function centerSelected() {
  const v = view.value;
  if (!v) return;
  const x = selected.value?.x ?? 0;
  offset.value = mod(x - Math.floor(v.width / 2), v.width);
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
  // The line is drawn in screen columns: on a shifted map neighbour columns may be the two ends of the map
  for (const point of lineCells(column(from.x), from.y, column(to.x), to.y)) {
    for (const c of brushCells(v, mapColumn(point.x, v.width, offset.value), point.y, brush.size, wrapX.value)) {
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
  const fx = e.offsetX / cell.value - column(at.x);
  const fy = e.offsetY / cell.value - (v.height - 1 - at.y);
  return nearestEdge(v, at.x, at.y, fx, fy, wrapX.value);
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
    await stepHistory(undo);
    error.value = '';
  } catch (err: any) {
    error.value = String(err);
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
  if (v.flags[i] & FLAG_SIGN) {
    for (const s of signs.value.filter(s => s.PlotX === h.x && s.PlotY === h.y)) parts.push(`«${s.Caption}»`);
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
        <template v-if="wrapX && view">
          <v-btn icon="mdi-arrow-left-bold" size="small" variant="text" :title="$t('world.seamLeft')"
                 @click="shift(Math.max(1, Math.round(view.width / 8)))"/>
          <v-btn icon="mdi-image-filter-center-focus" size="small" variant="text" :title="$t('world.seamCenter')"
                 @click="centerSelected"/>
          <v-btn icon="mdi-arrow-right-bold" size="small" variant="text" :title="$t('world.seamRight')"
                 @click="shift(-Math.max(1, Math.round(view.width / 8)))"/>
          <v-btn v-if="offset !== 0" icon="mdi-restore" size="small" variant="text" :title="$t('world.seamReset')"
                 @click="offset = 0"/>
        </template>
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
      <template v-else-if="selectedPlot">
        <PlotEditor :plot="selectedPlot" :players="players" @changed="refreshView"/>
        <v-divider/>
        <PlotSigns class="pt-2" :x="selectedPlot.X" :y="selectedPlot.Y" :signs="signs" :players="players"
                   @changed="refreshView"/>
      </template>
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
