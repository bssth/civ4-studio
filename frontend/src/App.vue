<template>
  <v-app id="inspire">
    <v-system-bar window style="--wails-draggable:drag">
      <v-icon class="me-4 no-drag" icon="mdi-file-plus-outline" title="New map (Ctrl+N)" @click="newMap" />
      <v-icon class="me-4 no-drag" icon="mdi-folder-open" title="Open (Ctrl+O)" @click="openMap" />
      <v-icon class="me-4 no-drag" icon="mdi-content-save" title="Save (Ctrl+S)" @click="saveMap" />
      <v-icon class="me-4 no-drag" icon="mdi-content-save-edit" title="Save as (Ctrl+Shift+S)" @click="saveMapAs" />
      <v-icon class="me-4 no-drag" icon="mdi-rocket-launch" title="Launch the game" @click="launch" />
      <v-icon class="me-4 no-drag" icon="mdi-cog" title="Settings" @click="tab = 'settings'" />

      <span class="text-caption text-medium-emphasis ms-4 text-truncate" style="max-width: 50%;">
        <template v-if="mapInfo">
          {{ mapInfo.dirty ? '● ' : '' }}{{ mapInfo.path || 'New map (not saved)' }}
        </template>
        <template v-else>No map loaded</template>
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
        <v-tab value="map">
          <v-icon icon="mdi-tune-vertical-variant" class="me-1"></v-icon>
          Map Settings
        </v-tab>
        <v-tab value="teams">
          <v-icon icon="mdi-account-group" class="me-1"></v-icon>
          Teams
        </v-tab>
        <v-tab value="players">
          <v-icon icon="mdi-human-edit" class="me-1"></v-icon>
          Players
        </v-tab>
        <v-tab value="settings">
          <v-icon icon="mdi-cog" class="me-1"></v-icon>
          Settings
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
          <v-col cols="12" md="9" class="pe-2">
            <v-sheet height="80vh" rounded="lg" class="overflow-y-auto">
              <Settings v-if="tab === 'settings'" />
              <MapSettings v-else-if="tab === 'map'" />
              <Teams v-else-if="tab === 'teams'" />
              <Players v-else-if="tab === 'players'" />
              <div v-else class="pa-5 text-grey">
                To start editing, open a map from the toolbar.
              </div>
            </v-sheet>
          </v-col>

          <v-col cols="12" md="3">
            <v-sheet height="80vh" rounded="lg" class="pa-3">
              <Console />
            </v-sheet>
          </v-col>
        </v-row>
      </v-container>
    </v-main>

    <v-snackbar v-model="errorOpen" color="error" timeout="6000">
      {{ errorMessage }}
    </v-snackbar>
  </v-app>
</template>

<script setup lang="ts">
import {onMounted, onUnmounted, ref} from "vue";
import Console from "./components/Console.vue";
import Settings from "./components/Settings.vue";
import MapSettings from "./components/MapSettings.vue";
import Teams from "./components/Teams.vue";
import Players from "./components/Players.vue";
import {EventsOff, EventsOn, Quit, WindowMaximise, WindowMinimise, WindowToggleMaximise} from "../wailsjs/runtime";
import {
  CheckGameDir,
  LaunchGame,
  LoadGameXML,
  NewMap,
  OpenMapDialog,
  SaveMap,
  SaveMapAs,
} from "../wailsjs/go/editor/App";
import {clearEnums, mapInfo, refreshEnums, refreshMap, refreshMapInfo} from "./store";

const minimize = WindowMinimise;
const maximize = WindowToggleMaximise;
const quit = Quit;

const tab = ref<string>('map');
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
    // A map may be already opened from the command line
    await refreshMap();

    const reason = await CheckGameDir();
    if (reason) {
      showError(`Game directory not configured: ${reason}. Open Settings to fix.`);
      tab.value = 'settings';
      return;
    }

    loadingMessage.value = 'Parsing game XML files...';
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
  return run('Creating map...', async () => {
    if (await NewMap()) {
      tab.value = 'map';
    }
  });
}

function openMap() {
  return run('Loading and parsing map...', async () => {
    if (await OpenMapDialog()) {
      tab.value = 'map';
    }
  });
}

function saveMap() {
  if (!mapInfo.value) {
    showError('No map loaded');
    return;
  }
  return run('Saving...', async () => {
    await SaveMap('');
    await refreshMapInfo();
  });
}

function saveMapAs() {
  if (!mapInfo.value) {
    showError('No map loaded');
    return;
  }
  return run('Saving...', async () => {
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

function onKeyDown(e: KeyboardEvent) {
  if (!(e.ctrlKey || e.metaKey)) return;
  const key = e.key.toLowerCase();
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
  EventsOn('map-loaded', async () => {
    await refreshMap();
  });
  EventsOn('map-state', async () => {
    await refreshMapInfo();
  });
  window.addEventListener('keydown', onKeyDown);
  bootstrap();
});

onUnmounted(() => {
  EventsOff('xml-progress');
  EventsOff('xml-done');
  EventsOff('xml-reset');
  EventsOff('map-loaded');
  EventsOff('map-state');
  window.removeEventListener('keydown', onKeyDown);
});
</script>

<style>
.no-drag {
  --wails-draggable: no-drag;
}

body {
  overflow: hidden;
}
</style>
