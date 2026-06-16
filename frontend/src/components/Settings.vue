<script setup lang="ts">
import {GetConfig, GetModsList, SetConfig, WriteConsole} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {ref} from "vue";

const config = ref<editor.Config | null>(null);
const mods = ref<string[]>([]);

GetConfig().then(c => {
  config.value = c;
});

GetModsList().then(m => {
  mods.value = m ?? [];
});

const saveConfig = () => {
  if (!config.value) {
    return;
  }

  SetConfig(editor.Config.createFrom(config.value));
  WriteConsole('Config saved');
}
</script>

<template>
  <div v-if="!config" class="pa-5 text-grey">
    Loading...
  </div>

  <div v-else class="pa-5">
    <v-text-field label="Game directory" @change="saveConfig"
                  prepend-icon="mdi-controller-classic" v-model="config.game_dir" />

    <v-select label="Use mod" @update:modelValue="saveConfig"
              :items="mods" v-model="config.mod" />

    <v-checkbox v-model="config.auto_save" @change="saveConfig">
      <template v-slot:label>
        Autosave map every 5 minutes
      </template>
    </v-checkbox>
  </div>
</template>
