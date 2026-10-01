<script setup lang="ts">
import {computed} from "vue";
import {editor} from "../../wailsjs/go/models";
import {requestedTab, showPlot} from "../store";

const props = defineProps<{ problems: editor.Problem[], limit?: number }>();
const emit = defineEmits<{ (e: 'navigate'): void }>();

const shown = computed(() => props.limit ? props.problems.slice(0, props.limit) : props.problems);

const sectionTabs: Record<string, string> = {game: 'game', map: 'map', teams: 'teams', players: 'players', plots: 'world'};

function canNavigate(p: editor.Problem): boolean {
  return p.x >= 0 || !!sectionTabs[p.section];
}

function navigate(p: editor.Problem) {
  if (p.x >= 0 && p.y >= 0) {
    showPlot(p.x, p.y);
  } else {
    requestedTab.value = sectionTabs[p.section] ?? null;
  }
  emit('navigate');
}
</script>

<template>
  <v-list density="compact">
    <v-list-item v-for="(p, idx) in shown" :key="idx" @click="canNavigate(p) && navigate(p)">
      <v-list-item-title class="problem-text">{{ p.message }}</v-list-item-title>
      <v-list-item-subtitle>{{ p.x >= 0 ? `${p.section} · plot ${p.x}, ${p.y}` : p.section }}</v-list-item-subtitle>
      <template v-slot:prepend>
        <v-icon :icon="p.severity === 'error' ? 'mdi-alert-circle' : 'mdi-alert'"
                :color="p.severity === 'error' ? 'error' : 'warning'"/>
      </template>
      <template v-slot:append v-if="canNavigate(p)">
        <v-icon icon="mdi-chevron-right" size="small"/>
      </template>
    </v-list-item>
    <v-list-item v-if="limit && problems.length > limit" class="text-medium-emphasis"
                 :title="`… and ${problems.length - limit} more`"/>
  </v-list>
</template>

<style scoped>
.problem-text {
  white-space: normal;
  word-break: break-word;
}
</style>
