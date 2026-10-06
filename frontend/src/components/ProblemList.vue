<script setup lang="ts">
import {computed} from "vue";
import {editor} from "../../wailsjs/go/models";
import {requestedTab, showPlot} from "../store";
import {problemText} from "../problems";
import {useI18n} from "vue-i18n";

const {t, te} = useI18n();

function sectionName(section: string): string {
  return te(`sections.${section}`) ? t(`sections.${section}`) : section;
}

// fixable shows the buttons of automatic fixes
const props = defineProps<{ problems: editor.Problem[], limit?: number, fixable?: boolean }>();
const emit = defineEmits<{
  (e: 'navigate'): void,
  (e: 'fix', problem: editor.Problem): void,
  (e: 'replace', problem: editor.Problem): void,
}>();

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
      <v-list-item-title class="problem-text">{{ problemText(p) }}</v-list-item-title>
      <v-list-item-subtitle>{{ sectionName(p.section) }}<template v-if="p.x >= 0"> · {{ $t('check.plotAt', {x: p.x, y: p.y}) }}</template></v-list-item-subtitle>
      <template v-slot:prepend>
        <v-icon :icon="p.severity === 'error' ? 'mdi-alert-circle' : 'mdi-alert'"
                :color="p.severity === 'error' ? 'error' : 'warning'"/>
      </template>
      <template v-slot:append>
        <template v-if="fixable && p.fix">
          <v-btn v-if="p.fix === 'replace'" size="small" variant="tonal" prepend-icon="mdi-swap-horizontal" class="me-2"
                 @click.stop="emit('replace', p)">{{ $t('check.replace') }}</v-btn>
          <v-btn v-else size="small" variant="tonal" color="primary" prepend-icon="mdi-auto-fix" class="me-2"
                 :title="$t('fixes.' + p.fix)" @click.stop="emit('fix', p)">{{ $t('check.fix') }}</v-btn>
        </template>
        <v-icon v-if="canNavigate(p)" icon="mdi-chevron-right" size="small"/>
      </template>
    </v-list-item>
    <v-list-item v-if="limit && problems.length > limit" class="text-medium-emphasis"
                 :title="$t('check.more', {n: problems.length - limit})"/>
  </v-list>
</template>

<style scoped>
.problem-text {
  white-space: normal;
  word-break: break-word;
}
</style>
