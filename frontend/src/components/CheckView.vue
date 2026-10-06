<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {ValidateMap} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {mapInfo, mapVersion, xmlReady} from "../store";
import ProblemList from "./ProblemList.vue";
import {useI18n} from "vue-i18n";

const {t, te} = useI18n();

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

const hasUnknownTypes = computed(() => problems.value.some(p => p.code === 'unknownType'));

const errors = computed(() => problems.value.filter(p => p.severity === 'error').length);
const warnings = computed(() => problems.value.length - errors.value);

const sections = computed(() => {
  const counts = new Map<string, number>();
  for (const p of problems.value) counts.set(p.section, (counts.get(p.section) ?? 0) + 1);
  return [{value: 'all', title: t('check.all', {n: problems.value.length})},
    ...[...counts].map(([s, n]) => ({value: s, title: `${te('sections.' + s) ? t('sections.' + s) : s} (${n})`}))];
});

const filtered = computed(() => section.value === 'all'
    ? problems.value : problems.value.filter(p => p.section === section.value));
</script>

<template>
  <div v-if="!mapInfo" class="pa-5 text-grey">
    {{ $t('common.noMapHint') }}
  </div>

  <div v-else class="pa-4">
    <div class="d-flex align-center flex-wrap" style="gap: 8px">
      <h3>{{ $t('check.title') }}</h3>
      <v-chip :color="errors ? 'error' : 'success'" size="small">{{ $t('check.errors', {n: errors}) }}</v-chip>
      <v-chip :color="warnings ? 'warning' : 'success'" size="small">{{ $t('check.warnings', {n: warnings}) }}</v-chip>
      <v-spacer/>
      <v-btn variant="tonal" prepend-icon="mdi-refresh" :loading="loading" @click="check">{{ $t('check.again') }}</v-btn>
    </div>
    <div class="text-caption text-medium-emphasis mt-1">
      <template v-if="xmlReady">{{ $t('check.withData') }}</template>
      <template v-else>{{ $t('check.withoutData') }}</template>
      {{ $t('check.clickHint') }}
    </div>
    <v-alert v-if="hasUnknownTypes" type="info" variant="tonal" density="compact" class="mt-3">
      {{ $t('check.modHint') }}
    </v-alert>

    <v-alert v-if="!loading && problems.length === 0" type="success" variant="tonal" class="mt-4">
      {{ $t('check.noProblems') }}
    </v-alert>

    <template v-else>
      <v-chip-group v-model="section" mandatory class="mt-2">
        <v-chip v-for="s in sections" :key="s.value" :value="s.value" filter size="small">{{ s.title }}</v-chip>
      </v-chip-group>
      <ProblemList :problems="filtered"/>
    </template>
  </div>
</template>
