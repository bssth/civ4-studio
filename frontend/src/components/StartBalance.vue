<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {GetPlayers, StartBalance} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {describeType, enums, mapRevision, mapVersion, playerColor, playerName, showPlot} from "../store";

const starts = ref<editor.StartInfo[]>([]);
const players = ref<editor.Player[]>([]);
// Index of the open panel, the analysis is loaded only while it is open
const open = ref<number | undefined>(undefined);

async function load() {
  if (open.value === undefined) return;
  const [s, p] = await Promise.all([StartBalance(), GetPlayers()]);
  starts.value = s ?? [];
  players.value = p ?? [];
}

watch([open, mapVersion, mapRevision], load);

type Column = 'land' | 'good' | 'hills' | 'forests' | 'water' | 'food' | 'strategic' | 'luxury' | 'nearest';
// Columns where more is better; their best and worst values are highlighted
const compared: Column[] = ['land', 'good', 'hills', 'forests', 'food', 'strategic', 'luxury', 'nearest'];

const ranges = computed(() => {
  const r = {} as Record<Column, { min: number, max: number }>;
  for (const c of compared) {
    const values = starts.value.map(s => s[c]);
    r[c] = {min: Math.min(...values), max: Math.max(...values)};
  }
  return r;
});

function highlight(c: Column, value: number): string {
  const r = ranges.value[c];
  if (!r || starts.value.length < 2 || r.min === r.max) return '';
  if (value === r.max) return 'text-success font-weight-bold';
  if (value === r.min) return 'text-error font-weight-bold';
  return '';
}

function resourceList(s: editor.StartInfo): string {
  return (s.resources ?? []).map(r => describeType(enums.bonuses, r)).join(', ');
}

const columns: { key: Column, title: string }[] = [
  {key: 'land', title: 'balance.land'}, {key: 'good', title: 'balance.good'}, {key: 'hills', title: 'balance.hills'},
  {key: 'forests', title: 'balance.forests'}, {key: 'water', title: 'balance.water'}, {key: 'food', title: 'balance.food'},
  {key: 'strategic', title: 'balance.strategic'}, {key: 'luxury', title: 'balance.luxury'},
];
</script>

<template>
  <v-expansion-panels v-model="open" class="px-3 mb-2">
    <v-expansion-panel>
      <v-expansion-panel-title>
        <v-icon icon="mdi-scale-balance" class="me-2"/>{{ $t('balance.title') }}
      </v-expansion-panel-title>
      <v-expansion-panel-text>
        <div class="d-flex align-center mb-2">
          <span class="text-caption text-medium-emphasis">{{ $t('balance.hint') }}</span>
          <v-spacer/>
          <v-btn size="small" variant="text" prepend-icon="mdi-refresh" @click="load">{{ $t('check.again') }}</v-btn>
        </div>
        <div v-if="starts.length === 0" class="text-body-2 text-grey">{{ $t('balance.none') }}</div>
        <v-table v-else density="compact" class="balance-table" hover>
          <thead>
          <tr>
            <th>{{ $t('balance.player') }}</th>
            <th>{{ $t('balance.plot') }}</th>
            <th v-for="c in columns" :key="c.key" class="text-end">{{ $t(c.title) }}</th>
            <th class="text-center">{{ $t('balance.coastal') }}</th>
            <th class="text-center">{{ $t('balance.river') }}</th>
            <th class="text-end">{{ $t('balance.nearest') }}</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="s in starts" :key="s.player" @click="showPlot(s.x, s.y)">
            <td class="text-no-wrap">
              <v-icon icon="mdi-circle" size="x-small" :color="playerColor(players, s.player)" class="me-1"/>
              #{{ s.player }} {{ playerName(players, s.player) }}
            </td>
            <td class="text-no-wrap">{{ s.x }}, {{ s.y }}</td>
            <td v-for="c in columns" :key="c.key" class="text-end" :class="highlight(c.key, s[c.key])"
                :title="['food', 'strategic', 'luxury'].includes(c.key) ? resourceList(s) : ''">{{ s[c.key] }}</td>
            <td class="text-center"><v-icon v-if="s.coastal" icon="mdi-anchor" size="small" color="primary"/></td>
            <td class="text-center"><v-icon v-if="s.river" icon="mdi-waves" size="small" color="primary"/></td>
            <td class="text-end" :class="highlight('nearest', s.nearest)">{{ s.nearest || '—' }}</td>
          </tr>
          </tbody>
        </v-table>
      </v-expansion-panel-text>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<style scoped>
.balance-table tbody tr {
  cursor: pointer;
}
</style>
