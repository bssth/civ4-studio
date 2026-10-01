<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {ValidateMap} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {mapInfo, mapVersion, xmlReady} from "../store";
import ProblemList from "./ProblemList.vue";

const problems = ref<editor.Problem[]>([]);
const loading = ref(false);
const section = ref<string>('all');

async function check() {
  loading.value = true;
  try {
    problems.value = (await ValidateMap()) ?? [];
  } finally {
    loading.value = false;
  }
}

onMounted(check);
watch([mapVersion, xmlReady], check);

const hasUnknownTypes = computed(() => problems.value.some(p => p.message.startsWith('unknown ')));

const errors = computed(() => problems.value.filter(p => p.severity === 'error').length);
const warnings = computed(() => problems.value.length - errors.value);

const sections = computed(() => {
  const counts = new Map<string, number>();
  for (const p of problems.value) counts.set(p.section, (counts.get(p.section) ?? 0) + 1);
  return [{value: 'all', title: `All (${problems.value.length})`},
    ...[...counts].map(([s, n]) => ({value: s, title: `${s[0].toUpperCase()}${s.slice(1)} (${n})`}))];
});

const filtered = computed(() => section.value === 'all'
    ? problems.value : problems.value.filter(p => p.section === section.value));
</script>

<template>
  <div v-if="!mapInfo" class="pa-5 text-grey">
    No map loaded. Open one from the toolbar to start editing.
  </div>

  <div v-else class="pa-4">
    <div class="d-flex align-center flex-wrap" style="gap: 8px">
      <h3>Scenario check</h3>
      <v-chip :color="errors ? 'error' : 'success'" size="small">{{ errors }} errors</v-chip>
      <v-chip :color="warnings ? 'warning' : 'success'" size="small">{{ warnings }} warnings</v-chip>
      <v-spacer/>
      <v-btn variant="tonal" prepend-icon="mdi-refresh" :loading="loading" @click="check">Check again</v-btn>
    </div>
    <div class="text-caption text-medium-emphasis mt-1">
      <template v-if="xmlReady">Types are checked against the loaded game and mod files.</template>
      <template v-else>Game data is not loaded, so types (civilizations, techs, units...) are not checked.</template>
      Click a problem to go to its place.
    </div>
    <v-alert v-if="hasUnknownTypes" type="info" variant="tonal" density="compact" class="mt-3">
      Unknown types usually mean that the map is made for a mod that is not selected in Settings.
    </v-alert>

    <v-alert v-if="!loading && problems.length === 0" type="success" variant="tonal" class="mt-4">
      No problems found.
    </v-alert>

    <template v-else>
      <v-chip-group v-model="section" mandatory class="mt-2">
        <v-chip v-for="s in sections" :key="s.value" :value="s.value" filter size="small">{{ s.title }}</v-chip>
      </v-chip-group>
      <ProblemList :problems="filtered"/>
    </template>
  </div>
</template>
