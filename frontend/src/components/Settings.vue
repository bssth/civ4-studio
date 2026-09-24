<script setup lang="ts">
import {
  CheckGameDir,
  ChooseGameDir,
  GetConfig,
  GetModsList,
  ResetGameXML,
  SetConfig,
  WriteConsole
} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {onMounted, ref} from "vue";
import {xmlReady} from "../store";

const config = ref<editor.Config | null>(null);
const mods = ref<string[]>([]);
const dirError = ref('');

async function refreshDirState() {
  dirError.value = await CheckGameDir();
  mods.value = (await GetModsList()) ?? [];
}

onMounted(async () => {
  config.value = await GetConfig();
  await refreshDirState();
});

async function saveConfig() {
  if (!config.value) {
    return;
  }

  await SetConfig(editor.Config.createFrom(config.value));
  WriteConsole('Config saved');
  await refreshDirState();
}

async function browse() {
  const dir = await ChooseGameDir();
  if (dir && config.value) {
    config.value.game_dir = dir;
    config.value.mod = '';
    await saveConfig();
  }
}
</script>

<template>
  <div v-if="!config" class="pa-5 text-grey">
    Loading...
  </div>

  <div v-else class="pa-5">
    <v-text-field label="Beyond the Sword directory" @change="saveConfig"
                  prepend-icon="mdi-controller-classic" v-model="config.game_dir"
                  :error-messages="dirError ? [dirError] : []"
                  :hint="dirError ? '' : 'Game directory looks fine'" persistent-hint>
      <template v-slot:append>
        <v-btn variant="tonal" @click="browse">Browse...</v-btn>
      </template>
    </v-text-field>

    <v-select label="Use mod" class="mt-4" @update:modelValue="saveConfig" clearable
              prepend-icon="mdi-puzzle" :items="mods" v-model="config.mod"
              hint="Game data (civilizations, techs etc.) is loaded from the mod when it is selected"
              persistent-hint />

    <v-checkbox v-model="config.auto_save" class="mt-4" @change="saveConfig"
                hint="Only maps that were saved at least once. The original file is kept as .bak"
                persistent-hint>
      <template v-slot:label>
        Autosave map every 5 minutes
      </template>
    </v-checkbox>

    <div class="mt-6 d-flex align-center">
      <v-chip :color="xmlReady ? 'success' : 'grey'" class="me-3">
        {{ xmlReady ? 'Game data loaded' : 'Game data not loaded' }}
      </v-chip>
      <v-btn variant="tonal" prepend-icon="mdi-refresh" :disabled="!!dirError" @click="ResetGameXML">
        Reload game data
      </v-btn>
    </div>
  </div>
</template>
