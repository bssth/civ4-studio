<template>
  <v-app id="inspire">
    <v-system-bar window style="--wails-draggable:drag">
      <v-icon class="me-4 no-drag" icon="mdi-folder-open" @click="openMap" />
      <v-icon class="me-4 no-drag" icon="mdi-content-save" @click="saveMap" />
      <v-icon class="me-4 no-drag" icon="mdi-rocket-launch" @click="launch" />
      <v-icon class="me-4 no-drag" icon="mdi-cog" @click="tab = 'settings'" />

      <span class="text-caption text-medium-emphasis ms-4 text-truncate" style="max-width: 50%;">
        {{ mapInfo?.path || 'No map loaded' }}
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
  OpenMapDialog,
  SaveMap,
} from "../wailsjs/go/editor/App";
import {mapInfo, refreshEnums, refreshMap} from "./store";

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

async function bootstrap() {
  const reason = await CheckGameDir();
  if (reason) {
    showError(`Game directory not configured: ${reason}. Open Settings to fix.`);
    tab.value = 'settings';
    return;
  }

  loadingMessage.value = 'Parsing game XML files...';
  try {
    await LoadGameXML();
    await refreshEnums();
  } catch (err: any) {
    showError(String(err));
  } finally {
    loadingMessage.value = '';
  }
}

async function openMap() {
  loadingMessage.value = 'Loading and parsing map...';
  try {
    const path = await OpenMapDialog();
    if (path) {
      await refreshMap();
      tab.value = 'map';
    }
  } catch (err: any) {
    showError(String(err));
  } finally {
    loadingMessage.value = '';
  }
}

async function saveMap() {
  if (!mapInfo.value) {
    showError('No map loaded');
    return;
  }
  loadingMessage.value = 'Saving...';
  try {
    const path = await SaveMap('');
    if (path) {
      await refreshMap();
    }
  } catch (err: any) {
    showError(String(err));
  } finally {
    loadingMessage.value = '';
  }
}

async function launch() {
  try {
    await LaunchGame();
  } catch (err: any) {
    showError(String(err));
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
  EventsOn('map-loaded', async () => {
    await refreshMap();
  });
  bootstrap();
});

onUnmounted(() => {
  EventsOff('xml-progress');
  EventsOff('xml-done');
  EventsOff('map-loaded');
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
