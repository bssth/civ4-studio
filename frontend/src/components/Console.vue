<script setup lang="ts">
import {nextTick, onMounted, onUnmounted, ref} from "vue";
import {EventsOff, EventsOn} from "../../wailsjs/runtime";

const maxLines = 500;
const lines = ref<string[]>([]);
const output = ref<HTMLElement | null>(null);

function stamp(): string {
  const date = new Date();
  return '[' + date.getHours().toString().padStart(2, '0')
      + ':' + date.getMinutes().toString().padStart(2, '0') + '] ';
}

onMounted(() => {
  EventsOn('console', async (data: any) => {
    const el = output.value;
    const atBottom = !el || el.scrollHeight - el.scrollTop - el.clientHeight < 20;

    lines.value.push(stamp() + String(data).trimEnd());
    if (lines.value.length > maxLines) {
      lines.value.splice(0, lines.value.length - maxLines);
    }

    if (atBottom) {
      await nextTick();
      output.value?.scrollTo({top: output.value.scrollHeight});
    }
  });
});

onUnmounted(() => {
  EventsOff('console');
});
</script>

<template>
  <div class="d-flex flex-column fill-height">
    <div class="d-flex align-center">
      <h3>Console</h3>
      <v-spacer />
      <v-btn icon="mdi-delete-sweep" variant="text" size="small" title="Clear" @click="lines = []" />
    </div>
    <v-divider class="mb-2" />
    <div ref="output" class="console-output flex-grow-1">
      <div v-for="(line, idx) in lines" :key="idx">{{ line }}</div>
    </div>
  </div>
</template>

<style scoped>
.console-output {
  overflow-y: auto;
  min-height: 0;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
