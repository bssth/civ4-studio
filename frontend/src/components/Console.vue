<script setup lang="ts">
import {onMounted, onUnmounted, ref} from "vue";
import {EventsOff, EventsOn} from "../../wailsjs/runtime";

import { VNumberInput } from 'vuetify/labs/VNumberInput'

const lines = ref<string[]>([]);
const maxLines = ref(25);
const showConfig = ref(false);

onMounted(() => {
  EventsOn('console', (data: any) => {
    alert(JSON.stringify(data));

    if (lines.value.length > maxLines.value) {
      lines.value = lines.value.slice(lines.value.length - maxLines.value);
    }

    const date = new Date();
    data = '[' + date.getHours() + ':' + date.getMinutes() + '] ' + data;
    lines.value.push(data);
  });
});

onUnmounted(() => {
  EventsOff('console');
});
</script>

<template>
  <div>
    <div class="float-right" style="width: 65%;">
      <v-number-input v-if="showConfig"
          reverse v-model="maxLines" density="compact"
          controlVariant="split" variant="solo-filled"
          label="Max. lines" />

      <v-btn icon="mdi-cog" v-else variant="text" @click="showConfig = true" />
    </div>
    <h3 class="mb-6">Console</h3>
    <v-divider class="mb-2" />
    <samp v-html="lines.join('<br>')" />
  </div>
</template>
