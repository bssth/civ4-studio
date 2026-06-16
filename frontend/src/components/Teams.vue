<script setup lang="ts">
import {onMounted, ref, watch} from "vue";
import {GetTeams, SetTeams} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {mapInfo} from "../store";

const teams = ref<editor.Team[]>([]);

async function load() {
  const t = await GetTeams();
  teams.value = (t ?? []).map(x => editor.Team.createFrom(x));
}

onMounted(load);
watch(() => mapInfo.value?.path, load);

let pending: any = null;
function persist() {
  if (pending) return;
  pending = Promise.resolve().then(() => {
    pending = null;
    SetTeams(teams.value.map(t => editor.Team.createFrom(t)));
  });
}

watch(teams, persist, {deep: true});
</script>

<template>
  <div v-if="teams.length === 0" class="pa-5 text-grey">
    No teams in the current map.
  </div>

  <div v-else class="pa-2">
    <v-expansion-panels variant="accordion">
      <v-expansion-panel v-for="(team, idx) in teams" :key="idx">
        <v-expansion-panel-title>
          Team #{{ team.TeamID }}
          <span class="ms-2 text-caption text-grey">
            {{ (team.Tech ?? []).length }} techs · {{ (team.AtWar ?? []).length }} wars
          </span>
        </v-expansion-panel-title>
        <v-expansion-panel-text>
          <v-row dense>
            <v-col cols="12" md="3">
              <v-text-field label="Team ID" type="number" density="compact" hide-details
                            v-model.number="team.TeamID" />
            </v-col>
            <v-col cols="12" md="9">
              <v-checkbox v-model="team.RevealMap" label="Reveal map" hide-details density="compact" />
            </v-col>
          </v-row>

          <v-row dense class="mt-2">
            <v-col cols="12" md="6">
              <v-textarea label="Techs (one per line)" rows="3" density="compact" hide-details
                          :model-value="(team.Tech ?? []).join('\n')"
                          @update:model-value="(v) => team.Tech = (v ?? '').split('\n').map(s => s.trim()).filter(Boolean)" />
            </v-col>
            <v-col cols="12" md="6">
              <v-textarea label="Project types (one per line)" rows="3" density="compact" hide-details
                          :model-value="(team.ProjectType ?? []).join('\n')"
                          @update:model-value="(v) => team.ProjectType = (v ?? '').split('\n').map(s => s.trim()).filter(Boolean)" />
            </v-col>
          </v-row>

          <v-row dense class="mt-2">
            <v-col cols="12" md="4">
              <v-text-field label="At war with (comma-separated team IDs)" density="compact" hide-details
                            :model-value="(team.AtWar ?? []).join(', ')"
                            @update:model-value="(v) => team.AtWar = (v ?? '').split(',').map(s => parseInt(s, 10)).filter(n => !isNaN(n))" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field label="Open borders" density="compact" hide-details
                            :model-value="(team.OpenBordersWithTeam ?? []).join(', ')"
                            @update:model-value="(v) => team.OpenBordersWithTeam = (v ?? '').split(',').map(s => parseInt(s, 10)).filter(n => !isNaN(n))" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field label="Defensive pact" density="compact" hide-details
                            :model-value="(team.DefensivePactWithTeam ?? []).join(', ')"
                            @update:model-value="(v) => team.DefensivePactWithTeam = (v ?? '').split(',').map(s => parseInt(s, 10)).filter(n => !isNaN(n))" />
            </v-col>
          </v-row>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
  </div>
</template>
