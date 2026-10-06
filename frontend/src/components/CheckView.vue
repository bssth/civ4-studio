<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {FixProblems, ReplaceType, ValidateMap} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {civilizations, enums, mapInfo, mapRevision, mapVersion, OptionKey, reloadEditors, xmlReady} from "../store";
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
watch([mapVersion, mapRevision, xmlReady], check);

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

// --- Fixes -------------------------------------------------------------------

const message = ref('');
const error = ref('');
const autoFixable = computed(() => problems.value.filter(p => p.fix && p.fix !== 'replace'));

async function fix(list: editor.Problem[]) {
  try {
    const n = await FixProblems(list);
    message.value = t('check.fixed', {n});
    error.value = '';
    await reloadEditors();
  } catch (err: any) {
    error.value = String(err);
  }
}

// Option lists of the kinds of types ("what" of unknown types)
const whatOptions: Record<string, OptionKey> = {
  era: 'eras', speed: 'speeds', calendar: 'calendars', victory: 'victories', gameOption: 'gameOptions',
  mpOption: 'mpOptions', forceControl: 'forceControls', worldSize: 'worldSizes', climate: 'climates',
  seaLevel: 'seaLevels', tech: 'techs', project: 'projects', leader: 'leaders', handicap: 'handicaps',
  playerColor: 'colors', artStyle: 'artStyles', religion: 'religions', civicOption: 'civicOptions', civic: 'civics',
  terrain: 'terrains', feature: 'features', resource: 'bonuses', improvement: 'improvements', route: 'routes',
  building: 'buildings', unit: 'units', promotion: 'promotions', process: 'processes', unitAI: 'unitAIs',
};
// Kinds that can only be replaced, not removed (the same list is checked by the backend)
const required = new Set(['era', 'speed', 'calendar', 'worldSize', 'climate', 'seaLevel', 'civilization', 'leader',
  'handicap', 'playerColor', 'artStyle', 'civicOption', 'terrain']);
const REMOVE = '\u0000remove';

const replace = ref<{ open: boolean, what: string, from: string, count: string, to: string | null }>(
    {open: false, what: '', from: '', count: '', to: null});

function openReplace(p: editor.Problem) {
  replace.value = {open: true, what: p.args.what, from: p.args.value, count: p.args.count ?? '1', to: null};
}

const replaceItems = computed(() => {
  const what = replace.value.what;
  const options = what === 'civilization'
      ? civilizations.value.map(c => ({value: c.type, title: c.description}))
      : (enums[whatOptions[what]] ?? []).map(o => ({value: o.type, title: `${o.description} (${o.type})`}));
  return required.has(what) ? options : [{value: REMOVE, title: t('check.removeEverywhere')}, ...options];
});

async function doReplace() {
  const r = replace.value;
  if (!r.to) return;
  r.open = false;
  try {
    const n = await ReplaceType(r.what, r.from, r.to === REMOVE ? '' : r.to);
    message.value = t('check.replaced', {n, value: r.from});
    error.value = '';
    await reloadEditors();
  } catch (err: any) {
    error.value = String(err);
  }
}
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
      <v-btn v-if="autoFixable.length" color="primary" variant="tonal" prepend-icon="mdi-auto-fix"
             @click="fix(autoFixable)">{{ $t('check.fixAll', {n: autoFixable.length}) }}</v-btn>
      <v-btn variant="tonal" prepend-icon="mdi-refresh" :loading="loading" @click="check">{{ $t('check.again') }}</v-btn>
    </div>
    <div class="text-caption text-medium-emphasis mt-1">
      <template v-if="xmlReady">{{ $t('check.withData') }}</template>
      <template v-else>{{ $t('check.withoutData') }}</template>
      {{ $t('check.clickHint') }}
    </div>
    <v-alert v-if="message" type="success" variant="tonal" density="compact" class="mt-3" closable
             @click:close="message = ''">{{ message }}</v-alert>
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mt-3" closable
             @click:close="error = ''">{{ error }}</v-alert>
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
      <ProblemList :problems="filtered" fixable @fix="p => fix([p])" @replace="openReplace"/>
    </template>

    <v-dialog v-model="replace.open" max-width="560">
      <v-card :title="$t('check.replaceTitle', {value: replace.from})">
        <v-card-text>
          <div class="text-body-2 mb-3">{{ $t('check.replaceText', {what: $t('what.' + replace.what), n: replace.count}) }}</div>
          <v-autocomplete v-model="replace.to" :items="replaceItems" :label="$t('check.replaceWith')" density="compact"
                          autofocus hide-details/>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="replace.open = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="primary" variant="tonal" :disabled="!replace.to" @click="doReplace">{{ $t('check.replace') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
