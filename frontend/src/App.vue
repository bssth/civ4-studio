<template>
  <v-app id="inspire">
    <v-system-bar window style="--wails-draggable:drag">
      <v-icon class="me-4 no-drag" icon="mdi-file-plus-outline" :title="$t('toolbar.new')" @click="newMap" />
      <v-icon class="me-4 no-drag" icon="mdi-folder-open" :title="$t('toolbar.open')" @click="openMap" />
      <v-icon class="me-4 no-drag" icon="mdi-content-save" :title="$t('toolbar.save')" @click="saveMap" />
      <v-icon class="me-4 no-drag" icon="mdi-content-save-edit" :title="$t('toolbar.saveAs')" @click="saveMapAs" />
      <v-icon class="me-4 no-drag" icon="mdi-undo" :disabled="!history.undo"
              :title="history.undo ? $t('world.undo', {what: historyLabel(history.undo)}) : $t('world.nothingToUndo')"
              @click="undo(true)" />
      <v-icon class="me-4 no-drag" icon="mdi-redo" :disabled="!history.redo"
              :title="history.redo ? $t('world.redo', {what: historyLabel(history.redo)}) : $t('world.nothingToRedo')"
              @click="undo(false)" />
      <v-icon class="me-4 no-drag" icon="mdi-rocket-launch" :title="$t('toolbar.launch')" @click="launch" />
      <v-icon class="me-4 no-drag" icon="mdi-cog" :title="$t('toolbar.settings')" @click="tab = 'settings'" />
      <v-icon class="me-4 no-drag" icon="mdi-keyboard-outline" :title="$t('toolbar.shortcuts')" @click="shortcutsOpen = true" />

      <span class="text-caption text-medium-emphasis ms-4 text-truncate" style="max-width: 50%;"
            :title="mapInfo?.path ?? ''">
        <template v-if="mapInfo">
          {{ mapInfo.dirty ? '● ' : '' }}{{ mapInfo.path ? fileName(mapInfo.path) : $t('app.newMapNotSaved') }}
        </template>
        <template v-else>{{ $t('app.noMap') }}</template>
      </span>

      <v-spacer></v-spacer>

      <v-btn class="no-drag" icon="mdi-minus" variant="text" @click="minimize" />
      <v-btn class="ms-2 no-drag" icon="mdi-checkbox-blank-outline" variant="text" @click="maximize" />
      <v-btn class="ms-2 no-drag" icon="mdi-close" variant="text" @click="quit" />
    </v-system-bar>

    <v-app-bar
        class="px-3"
        density="compact"
        flat style="--wails-draggable:no-drag"
    >
      <v-spacer></v-spacer>

      <v-tabs
          color="grey-darken-2"
          centered v-model="tab"
      >
        <v-tab value="game">
          <v-icon icon="mdi-tune-vertical-variant" class="me-1"></v-icon>
          {{ $t('tabs.game') }}
        </v-tab>
        <v-tab value="map">
          <v-icon icon="mdi-earth" class="me-1"></v-icon>
          {{ $t('tabs.map') }}
        </v-tab>
        <v-tab value="world">
          <v-icon icon="mdi-map" class="me-1"></v-icon>
          {{ $t('tabs.world') }}
        </v-tab>
        <v-tab value="teams">
          <v-icon icon="mdi-account-group" class="me-1"></v-icon>
          {{ $t('tabs.teams') }}
        </v-tab>
        <v-tab value="players">
          <v-icon icon="mdi-human-edit" class="me-1"></v-icon>
          {{ $t('tabs.players') }}
        </v-tab>
        <v-tab value="objects">
          <v-icon icon="mdi-home-city" class="me-1"></v-icon>
          {{ $t('tabs.objects') }}
        </v-tab>
        <v-tab value="check">
          <v-icon icon="mdi-clipboard-check-outline" class="me-1"></v-icon>
          {{ $t('tabs.check') }}
        </v-tab>
        <v-tab value="settings">
          <v-icon icon="mdi-cog" class="me-1"></v-icon>
          {{ $t('tabs.settings') }}
        </v-tab>
      </v-tabs>
      <v-spacer />
    </v-app-bar>

    <v-progress-linear v-if="loadingMessage" indeterminate color="primary" />
    <div v-if="loadingMessage" class="px-3 py-1 text-caption text-medium-emphasis text-truncate">
      {{ loadingMessage }}
    </div>

    <v-main class="bg-grey-lighten-3" style="--wails-draggable:no-drag">
      <v-container fluid>
        <v-row no-gutters>
          <v-col cols="12" :md="wide ? 12 : 9" :class="wide ? '' : 'pe-2'">
            <v-sheet height="80vh" rounded="lg" class="overflow-y-auto">
              <Settings v-if="tab === 'settings'" />
              <MapSettings v-else-if="tab === 'game'" />
              <MapProperties v-else-if="tab === 'map'" />
              <WorldView v-else-if="tab === 'world'" />
              <CheckView v-else-if="tab === 'check'" />
              <Teams v-else-if="tab === 'teams'" />
              <Players v-else-if="tab === 'players'" />
              <CitiesUnits v-else-if="tab === 'objects'" />
              <div v-else class="pa-5 text-grey">
                {{ $t('app.openMapHint') }}
              </div>
            </v-sheet>
          </v-col>

          <!-- v-show keeps the console mounted, so it does not miss lines while hidden -->
          <v-col v-show="!wide" cols="12" md="3">
            <v-sheet height="80vh" rounded="lg" class="pa-3">
              <Console />
            </v-sheet>
          </v-col>
        </v-row>
      </v-container>
    </v-main>

    <v-dialog v-model="shortcutsOpen" max-width="620">
      <v-card :title="$t('shortcuts.title')">
        <v-card-text>
          <template v-for="group in shortcutGroups" :key="group.title">
            <h4 class="mt-2 mb-1">{{ $t(group.title) }}</h4>
            <v-table density="compact">
              <tbody>
              <tr v-for="s in group.items" :key="s.text">
                <td style="width: 40%"><kbd v-for="(k, i) in s.keys" :key="i" class="me-1">{{ $te(k) ? $t(k) : k }}</kbd></td>
                <td>{{ $t(s.text) }}</td>
              </tr>
              </tbody>
            </v-table>
          </template>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="shortcutsOpen = false">{{ $t('common.close') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="saveCheck.open" max-width="640">
      <v-card>
        <v-card-title>{{ $t('saveCheck.title', {n: saveCheck.errors}) }}</v-card-title>
        <v-card-text>
          {{ $t('saveCheck.text') }}
          <ProblemList :problems="saveCheck.problems" :limit="6" @navigate="closeSaveCheck(false)" />
        </v-card-text>
        <v-card-actions>
          <v-btn variant="text" @click="closeSaveCheck(false); tab = 'check'">{{ $t('saveCheck.showAll') }}</v-btn>
          <v-spacer />
          <v-btn variant="text" @click="closeSaveCheck(false)">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="error" variant="tonal" @click="closeSaveCheck(true)">{{ $t('saveCheck.saveAnyway') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="errorOpen" color="error" timeout="6000">
      {{ errorMessage }}
    </v-snackbar>
  </v-app>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, reactive, ref, watch} from "vue";
import Console from "./components/Console.vue";
import Settings from "./components/Settings.vue";
import MapSettings from "./components/MapSettings.vue";
import MapProperties from "./components/MapProperties.vue";
import WorldView from "./components/WorldView.vue";
import CheckView from "./components/CheckView.vue";
import ProblemList from "./components/ProblemList.vue";
import Teams from "./components/Teams.vue";
import Players from "./components/Players.vue";
import CitiesUnits from "./components/CitiesUnits.vue";
import {EventsOff, EventsOn, Quit, WindowMaximise, WindowMinimise, WindowToggleMaximise} from "../wailsjs/runtime";
import {useI18n} from "vue-i18n";
import {setUILanguage} from "./i18n";
import {
  CheckGameDir,
  GetConfig,
  LaunchGame,
  LoadGameXML,
  NewMap,
  OpenMapDialog,
  SaveMap,
  SaveMapAs,
  ValidateMap,
} from "../wailsjs/go/editor/App";
import {editor} from "../wailsjs/go/models";
import {
  clearEnums,
  history,
  historyLabel,
  isTextField,
  mapInfo,
  refreshEnums,
  refreshMap,
  refreshMapInfo,
  requestedTab,
  stepHistory
} from "./store";

const {t} = useI18n();

const minimize = WindowMinimise;
const maximize = WindowToggleMaximise;
const quit = Quit;

const tab = ref<string>('game');

// The title bar shows only the file name, the full path is in the tooltip
function fileName(path: string): string {
  return path.split(/[\\/]/).pop() ?? path;
}
// The world map needs the whole width, the console is hidden there
const wide = computed(() => tab.value === 'world');

// Other components ask to show a tab, e.g. the problem list
watch(requestedTab, t => {
  if (t) {
    tab.value = t;
    requestedTab.value = null;
  }
});

// Check before saving: the dialog resolves to true when the user wants to save despite errors
const saveCheck = reactive<{ open: boolean, errors: number, problems: editor.Problem[], resolve: ((save: boolean) => void) | null }>({
  open: false, errors: 0, problems: [], resolve: null,
});

function closeSaveCheck(save: boolean) {
  saveCheck.open = false;
  saveCheck.resolve?.(save);
  saveCheck.resolve = null;
}

async function confirmSave(): Promise<boolean> {
  const problems = (await ValidateMap()) ?? [];
  const errors = problems.filter(p => p.severity === 'error');
  if (errors.length === 0) return true;
  saveCheck.errors = errors.length;
  saveCheck.problems = errors;
  saveCheck.open = true;
  return new Promise(resolve => saveCheck.resolve = resolve);
}

// The dialog closed by clicking outside or Escape means cancel
watch(() => saveCheck.open, open => {
  if (!open && saveCheck.resolve) closeSaveCheck(false);
});
const loadingMessage = ref<string>('');
const errorOpen = ref(false);
const errorMessage = ref('');

function showError(msg: string) {
  errorMessage.value = msg;
  errorOpen.value = true;
}

let bootstrapping = false;

async function bootstrap() {
  if (bootstrapping) return;
  bootstrapping = true;
  try {
    const config = await GetConfig();
    setUILanguage(config?.ui_language);

    // A map may be already opened from the command line
    await refreshMap();

    const reason = await CheckGameDir();
    if (reason) {
      showError(t('app.gameDirNotConfigured', {reason}));
      tab.value = 'settings';
      return;
    }

    loadingMessage.value = t('progress.parsingXml');
    await LoadGameXML();
    await refreshEnums();
  } catch (err: any) {
    showError(String(err));
  } finally {
    loadingMessage.value = '';
    bootstrapping = false;
  }
}

async function run(message: string, action: () => Promise<unknown>) {
  loadingMessage.value = message;
  try {
    await action();
  } catch (err: any) {
    showError(String(err));
  } finally {
    loadingMessage.value = '';
  }
}

function newMap() {
  return run(t('progress.creatingMap'), async () => {
    if (await NewMap()) {
      // Do not rely on the map-loaded event only: editors must drop the old map before any edit
      await refreshMap();
      tab.value = 'map';
    }
  });
}

function openMap() {
  return run(t('progress.loadingMap'), async () => {
    if (await OpenMapDialog()) {
      await refreshMap();
      tab.value = 'game';
    }
  });
}

async function saveMap() {
  if (!mapInfo.value) {
    showError(t('app.noMap'));
    return;
  }
  if (!await confirmSave().catch(err => (showError(String(err)), false))) return;
  return run(t('progress.saving'), async () => {
    await SaveMap('');
    await refreshMapInfo();
  });
}

async function saveMapAs() {
  if (!mapInfo.value) {
    showError(t('app.noMap'));
    return;
  }
  if (!await confirmSave().catch(err => (showError(String(err)), false))) return;
  return run(t('progress.saving'), async () => {
    await SaveMapAs();
    await refreshMapInfo();
  });
}

async function launch() {
  try {
    await LaunchGame();
  } catch (err: any) {
    showError(String(err));
  }
}

async function undo(back: boolean) {
  if (!(back ? history.value.undo : history.value.redo)) return;
  try {
    await stepHistory(back);
  } catch (err: any) {
    showError(String(err));
  }
}

// Shortcuts of the whole editor, shown by F1 and the keyboard icon
const shortcutsOpen = ref(false);
const shortcutGroups = [
  {
    title: 'shortcuts.general', items: [
      {keys: ['Ctrl+N'], text: 'shortcuts.new'}, {keys: ['Ctrl+O'], text: 'shortcuts.open'},
      {keys: ['Ctrl+S'], text: 'shortcuts.save'}, {keys: ['Ctrl+Shift+S'], text: 'shortcuts.saveAs'},
      {keys: ['Ctrl+Z'], text: 'shortcuts.undo'}, {keys: ['Ctrl+Y', 'Ctrl+Shift+Z'], text: 'shortcuts.redo'},
      {keys: ['F1'], text: 'shortcuts.help'},
    ],
  },
  {
    title: 'shortcuts.world', items: [
      {keys: ['Ctrl+F'], text: 'shortcuts.search'}, {keys: ['Ctrl+C'], text: 'shortcuts.copy'},
      {keys: ['Ctrl+V'], text: 'shortcuts.paste'}, {keys: ['Esc'], text: 'shortcuts.escape'},
      {keys: ['shortcuts.keyWheel'], text: 'shortcuts.zoom'}, {keys: ['shortcuts.keyDrag'], text: 'shortcuts.dragStart'},
    ],
  },
];

function onKeyDown(e: KeyboardEvent) {
  if (e.key === 'F1') {
    e.preventDefault();
    shortcutsOpen.value = true;
    return;
  }
  if (!(e.ctrlKey || e.metaKey)) return;
  const key = e.key.toLowerCase();
  if (key === 'z' || key === 'y') {
    // Text fields keep their own undo
    if (isTextField(e.target)) return;
    e.preventDefault();
    undo(key === 'z' && !e.shiftKey);
    return;
  }
  const actions: Record<string, () => unknown> = {
    n: newMap,
    o: openMap,
    s: e.shiftKey ? saveMapAs : saveMap,
  };
  if (actions[key]) {
    e.preventDefault();
    actions[key]();
  }
}

onMounted(() => {
  WindowMaximise();
  EventsOn('xml-progress', (msg: string) => {
    loadingMessage.value = msg;
  });
  EventsOn('xml-done', () => {
    loadingMessage.value = '';
  });
  EventsOn('xml-reset', () => {
    clearEnums();
    bootstrap();
  });
  EventsOn('game-language', () => refreshEnums());
  EventsOn('map-loaded', async () => {
    await refreshMap();
  });
  EventsOn('map-state', async () => {
    await refreshMapInfo();
  });
  EventsOn('history', (state: editor.HistoryState) => {
    history.value = state;
  });
  window.addEventListener('keydown', onKeyDown);
  bootstrap();
});

onUnmounted(() => {
  EventsOff('xml-progress');
  EventsOff('xml-done');
  EventsOff('xml-reset');
  EventsOff('game-language');
  EventsOff('map-loaded');
  EventsOff('map-state');
  EventsOff('history');
  window.removeEventListener('keydown', onKeyDown);
});
</script>

<style>
kbd {
  font-family: monospace;
  font-size: 0.85em;
  padding: 1px 6px;
  border: 1px solid rgba(0, 0, 0, 0.25);
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.05);
  white-space: nowrap;
}

.no-drag {
  --wails-draggable: no-drag;
}

body {
  overflow: hidden;
}
</style>
