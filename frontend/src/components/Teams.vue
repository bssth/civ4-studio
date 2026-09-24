<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {GetPlayers, GetTeams, SetTeams} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, describeType, enums, mapVersion, NONE, withCurrent} from "../store";

type RelationField = 'ContactWithTeam' | 'AtWar' | 'OpenBordersWithTeam' | 'DefensivePactWithTeam' | 'PermanentWarPeace';

const relations: { field: RelationField, title: string, icon: string, color: string }[] = [
  {field: 'ContactWithTeam', title: 'Contact', icon: 'mdi-handshake', color: 'blue'},
  {field: 'AtWar', title: 'At war', icon: 'mdi-sword-cross', color: 'red'},
  {field: 'OpenBordersWithTeam', title: 'Open borders', icon: 'mdi-door-open', color: 'green'},
  {field: 'DefensivePactWithTeam', title: 'Defensive pact', icon: 'mdi-shield', color: 'purple'},
  {field: 'PermanentWarPeace', title: 'Permanent war/peace', icon: 'mdi-lock', color: 'orange'},
];

const teams = ref<editor.Team[]>([]);
// Players are used only to name the teams
const players = ref<editor.Player[]>([]);
const showEmpty = ref(false);
const relation = ref<RelationField>('AtWar');

async function load() {
  const [t, p] = await Promise.all([GetTeams(), GetPlayers()]);
  teams.value = (t ?? []).map(x => editor.Team.createFrom(x));
  players.value = p ?? [];
}

onMounted(load);
watch(mapVersion, load);

watch(teams, batched(() => {
  SetTeams(teams.value.map(t => editor.Team.createFrom(t)));
}), {deep: true});

function teamMembers(team: editor.Team): editor.Player[] {
  return players.value.filter(p => p.Team === team.TeamID && p.CivType !== NONE && p.CivType !== '');
}

function teamName(team: editor.Team): string {
  const members = teamMembers(team);
  if (members.length === 0) return `Team ${team.TeamID}`;
  return members.map(p => p.CivShortDesc || p.LeaderName || p.CivType).join(', ');
}

const visibleTeams = computed(() =>
    teams.value.filter(t => showEmpty.value || teamMembers(t).length > 0));

function byId(id: number): editor.Team | undefined {
  return teams.value.find(t => t.TeamID === id);
}

function hasRelation(team: editor.Team, field: RelationField, other: number): boolean {
  return (team[field] ?? []).includes(other);
}

function setOne(team: editor.Team, field: RelationField, other: number, on: boolean) {
  const list = (team[field] ?? []).filter(id => id !== other);
  if (on) list.push(other);
  list.sort((a, b) => a - b);
  team[field] = list;
}

/** Relations are mutual, so both teams are updated. Conflicting relations are removed. */
function setRelation(a: editor.Team, b: editor.Team, field: RelationField, on: boolean) {
  const apply = (f: RelationField, value: boolean) => {
    setOne(a, f, b.TeamID, value);
    setOne(b, f, a.TeamID, value);
  };
  apply(field, on);
  if (!on) {
    if (field === 'ContactWithTeam') {
      // No war, borders or pacts without contact
      for (const f of ['AtWar', 'OpenBordersWithTeam', 'DefensivePactWithTeam'] as RelationField[]) apply(f, false);
    }
    return;
  }
  if (field !== 'PermanentWarPeace') apply('ContactWithTeam', true);
  if (field === 'AtWar') {
    apply('OpenBordersWithTeam', false);
    apply('DefensivePactWithTeam', false);
  } else if (field === 'OpenBordersWithTeam' || field === 'DefensivePactWithTeam') {
    apply('AtWar', false);
  }
}

function toggle(a: editor.Team, b: editor.Team) {
  setRelation(a, b, relation.value, !hasRelation(a, relation.value, b.TeamID));
}

const currentRelation = computed(() => relations.find(r => r.field === relation.value)!);

function techItems(team: editor.Team) {
  return withCurrent(enums.techs, ...(team.Tech ?? []));
}

function projectItems(team: editor.Team) {
  return withCurrent(enums.projects, ...(team.ProjectType ?? []));
}

function relationSummary(team: editor.Team): string {
  const wars = (team.AtWar ?? []).map(id => byId(id)).filter(Boolean).map(t => teamName(t!));
  return wars.length ? 'at war with ' + wars.join(', ') : '';
}
</script>

<template>
  <div v-if="teams.length === 0" class="pa-5 text-grey">
    No teams in the current map.
  </div>

  <div v-else class="pa-2">
    <v-checkbox v-model="showEmpty" label="Show teams without players" hide-details density="compact" class="px-3" />

    <v-card variant="outlined" class="ma-2">
      <v-card-title class="d-flex align-center flex-wrap">
        Diplomacy
        <v-spacer />
        <v-btn-toggle v-model="relation" mandatory density="compact" divided>
          <v-btn v-for="r in relations" :key="r.field" :value="r.field" size="small">
            <v-icon :icon="r.icon" class="me-1" />{{ r.title }}
          </v-btn>
        </v-btn-toggle>
      </v-card-title>
      <v-card-subtitle>
        Click a cell to toggle "{{ currentRelation.title }}" between two teams. Relations are mutual.
      </v-card-subtitle>
      <v-card-text class="matrix-wrap">
        <table class="matrix">
          <thead>
          <tr>
            <th></th>
            <th v-for="col in visibleTeams" :key="col.TeamID" :title="teamName(col)">
              <div class="col-head">{{ col.TeamID }}</div>
            </th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in visibleTeams" :key="row.TeamID">
            <th class="row-head" :title="teamName(row)">{{ row.TeamID }} · {{ teamName(row) }}</th>
            <td v-for="col in visibleTeams" :key="col.TeamID"
                :class="{self: row.TeamID === col.TeamID}"
                :title="`${teamName(row)} — ${teamName(col)}`"
                @click="row.TeamID !== col.TeamID && toggle(row, col)">
              <v-icon v-if="row.TeamID !== col.TeamID && hasRelation(row, relation, col.TeamID)"
                      :icon="currentRelation.icon" :color="currentRelation.color" size="small" />
            </td>
          </tr>
          </tbody>
        </table>
      </v-card-text>
    </v-card>

    <v-expansion-panels variant="accordion" class="mt-2">
      <v-expansion-panel v-for="team in visibleTeams" :key="team.TeamID">
        <v-expansion-panel-title>
          Team #{{ team.TeamID }} — {{ teamName(team) }}
          <span class="ms-2 text-caption text-grey">
            {{ (team.Tech ?? []).length }} techs <template v-if="relationSummary(team)">· {{ relationSummary(team) }}</template>
          </span>
        </v-expansion-panel-title>
        <v-expansion-panel-text>
          <v-checkbox v-model="team.RevealMap" label="Reveal the whole map" hide-details density="compact" />

          <v-autocomplete label="Starting techs" multiple chips closable-chips clearable
                          density="compact" class="mt-2" hide-details
                          :items="techItems(team)" item-value="type" item-title="description"
                          v-model="team.Tech">
            <template v-slot:item="{ props, item }">
              <v-list-item v-bind="props" :subtitle="describeType(enums.eras, item.raw.group)" />
            </template>
          </v-autocomplete>

          <v-autocomplete label="Completed projects" multiple chips closable-chips clearable
                          density="compact" class="mt-3" hide-details
                          :items="projectItems(team)" item-value="type" item-title="description"
                          v-model="team.ProjectType" />
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
  </div>
</template>

<style scoped>
.matrix-wrap {
  overflow: auto;
  max-height: 45vh;
}

.matrix {
  border-collapse: collapse;
  font-size: 12px;
}

.matrix th, .matrix td {
  border: 1px solid rgba(0, 0, 0, 0.12);
  padding: 0;
}

.matrix thead th {
  position: sticky;
  top: 0;
  background: rgb(var(--v-theme-surface));
  z-index: 1;
}

.col-head {
  width: 26px;
  text-align: center;
}

.row-head {
  position: sticky;
  left: 0;
  background: rgb(var(--v-theme-surface));
  text-align: left;
  padding: 0 8px !important;
  white-space: nowrap;
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: normal;
}

.matrix td {
  width: 26px;
  height: 26px;
  text-align: center;
  cursor: pointer;
}

.matrix td:hover {
  background: rgba(0, 0, 0, 0.06);
}

.matrix td.self {
  background: rgba(0, 0, 0, 0.08);
  cursor: default;
}
</style>
