<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {SetPlot} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, emptyIfNone, enums, NONE, noneIfEmpty, playerName, withCurrent, withNone} from "../store";
import {PLOT_HILLS, PLOT_LAND, PLOT_OCEAN, PLOT_PEAK} from "../mapRender";

const props = defineProps<{
  plot: editor.Plot,
  players: editor.Player[],
}>();
const emit = defineEmits<{ (e: 'changed'): void }>();

// Local editable copy, sent to the backend on every change
const plot = ref<editor.Plot>(JSON.parse(JSON.stringify(props.plot)));
const error = ref('');
// Last state known to the backend, so selecting a plot does not send it back
let baseline = JSON.stringify(plot.value);

watch(() => props.plot, p => {
  baseline = JSON.stringify(p);
  plot.value = JSON.parse(baseline);
  error.value = '';
});

const save = batched(async () => {
  const json = JSON.stringify(plot.value);
  if (json === baseline) return;
  baseline = json;
  try {
    await SetPlot(JSON.parse(json));
    error.value = '';
    emit('changed');
  } catch (err: any) {
    error.value = String(err);
  }
});
watch(plot, save, {deep: true});

const plotTypes = [
  {value: PLOT_PEAK, title: 'Peak'},
  {value: PLOT_HILLS, title: 'Hills'},
  {value: PLOT_LAND, title: 'Flat'},
  {value: PLOT_OCEAN, title: 'Water'},
];

function optional(options: editor.EnumOption[], value: string | null | undefined) {
  return withNone(withCurrent(options, emptyIfNone(value)));
}

const feature = computed({
  get: () => noneIfEmpty(plot.value.FeatureType?.[0]),
  set: (v: string) => {
    if (v === NONE) {
      plot.value.FeatureType = [];
      plot.value.FeatureVariety = [];
    } else {
      plot.value.FeatureType = [v];
      plot.value.FeatureVariety = [plot.value.FeatureVariety?.[0] ?? '0'];
    }
  },
});

const variety = computed({
  get: () => Number(plot.value.FeatureVariety?.[0] ?? 0),
  set: (v: number) => {
    if (plot.value.FeatureType?.length) plot.value.FeatureVariety = [String(v || 0)];
  },
});

const ownerItems = computed(() => {
  const items = props.players
      .map((p, i) => ({value: i, title: `#${i} ${playerName(props.players, i)}`, empty: !p.CivType || p.CivType === NONE}))
      .filter(i => !i.empty);
  // Owners that are not regular players (e.g. barbarians) must still be shown
  const used = new Set<number>([...(plot.value.Cities ?? []).map(c => c.CityOwner), ...(plot.value.Units ?? []).map(u => u.UnitOwner)]);
  for (const owner of used) {
    if (!items.some(i => i.value === owner)) items.push({value: owner, title: `#${owner}`, empty: false});
  }
  return items;
});

function defaultOwner(): number {
  return ownerItems.value[0]?.value ?? 0;
}

function addCity() {
  plot.value.Cities = [...(plot.value.Cities ?? []), {
    CityOwner: defaultOwner(), CityName: 'New city', CityPopulation: 1,
    ProductionUnit: '', ProductionBuilding: '', ProductionProject: '', ProductionProcess: '',
    BuildingType: [], ReligionType: [], HolyCityReligionType: [], ScriptData: '', PlayerCulture: null,
  }];
}

function addUnit() {
  const warrior = enums.units.find(u => u.type === 'UNIT_WARRIOR')?.type ?? enums.units[0]?.type ?? 'UNIT_WARRIOR';
  plot.value.Units = [...(plot.value.Units ?? []), {
    UnitType: warrior, UnitOwner: defaultOwner(), Level: 1, Experience: 0,
    PromotionType: [], UnitAIType: '', Damage: 0, FacingDirection: 4,
  }];
}

function removeAt<T>(list: T[] | null, index: number): T[] {
  const copy = [...(list ?? [])];
  copy.splice(index, 1);
  return copy;
}

const riverText = computed(() => {
  const parts = [];
  if (plot.value.IsNOfRiver) parts.push('south edge');
  if (plot.value.IsWOfRiver) parts.push('east edge');
  return parts.length ? 'River along the ' + parts.join(' and ') : '';
});
</script>

<template>
  <div class="pa-3">
    <div class="d-flex align-center mb-2">
      <h3>Plot {{ plot.X }}, {{ plot.Y }}</h3>
      <v-spacer/>
      <span v-if="riverText" class="text-caption text-medium-emphasis">{{ riverText }}</span>
    </div>

    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-2">{{ error }}</v-alert>

    <v-row dense>
      <v-col cols="7">
        <v-select label="Terrain" density="compact" hide-details
                  :items="withCurrent(enums.terrains, plot.TerrainType)" item-value="type" item-title="description"
                  v-model="plot.TerrainType"/>
      </v-col>
      <v-col cols="5">
        <v-select label="Height" density="compact" hide-details :items="plotTypes" v-model="plot.PlotType"/>
      </v-col>
      <v-col cols="8">
        <v-select label="Feature" density="compact" hide-details
                  :items="optional(enums.features, plot.FeatureType?.[0])" item-value="type" item-title="description"
                  v-model="feature"/>
      </v-col>
      <v-col cols="4">
        <v-text-field label="Variety" type="number" density="compact" hide-details min="0"
                      :disabled="!plot.FeatureType?.length" v-model.number="variety"/>
      </v-col>
      <v-col cols="12">
        <v-autocomplete label="Resource" density="compact" hide-details
                        :items="optional(enums.bonuses, plot.BonusType)" item-value="type" item-title="description"
                        :model-value="noneIfEmpty(plot.BonusType)"
                        @update:model-value="(v: string) => plot.BonusType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="7">
        <v-autocomplete label="Improvement" density="compact" hide-details
                        :items="optional(enums.improvements, plot.ImprovementType)" item-value="type"
                        item-title="description"
                        :model-value="noneIfEmpty(plot.ImprovementType)"
                        @update:model-value="(v: string) => plot.ImprovementType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="5">
        <v-select label="Route" density="compact" hide-details
                  :items="optional(enums.routes, plot.RouteType)" item-value="type" item-title="description"
                  :model-value="noneIfEmpty(plot.RouteType)"
                  @update:model-value="(v: string) => plot.RouteType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="12">
        <v-text-field label="Landmark text" density="compact" hide-details v-model="plot.Landmark"/>
      </v-col>
    </v-row>
    <v-checkbox v-model="plot.StartingPlot" density="compact" hide-details
                label="Starting plot for a random civilization"/>

    <v-divider class="my-2"/>
    <div class="d-flex align-center">
      <h4>Cities</h4>
      <v-spacer/>
      <v-btn size="small" variant="text" prepend-icon="mdi-plus" @click="addCity">Add city</v-btn>
    </div>
    <v-card v-for="(city, idx) in plot.Cities ?? []" :key="'c' + idx" variant="outlined" class="pa-2 mb-2">
      <v-row dense>
        <v-col cols="12" class="d-flex align-center">
          <v-text-field label="Name" density="compact" hide-details v-model="city.CityName"/>
          <v-btn icon="mdi-delete" variant="text" size="small" title="Remove city"
                 @click="plot.Cities = removeAt(plot.Cities, idx)"/>
        </v-col>
        <v-col cols="8">
          <v-select label="Owner" density="compact" hide-details :items="ownerItems" v-model="city.CityOwner"/>
        </v-col>
        <v-col cols="4">
          <v-text-field label="Population" type="number" min="1" density="compact" hide-details
                        v-model.number="city.CityPopulation"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete label="Buildings" multiple chips closable-chips density="compact" hide-details
                          :items="withCurrent(enums.buildings, ...(city.BuildingType ?? []))"
                          item-value="type" item-title="description" v-model="city.BuildingType"/>
        </v-col>
        <v-col cols="6">
          <v-select label="Religions" multiple chips density="compact" hide-details
                    :items="withCurrent(enums.religions, ...(city.ReligionType ?? []))"
                    item-value="type" item-title="description" v-model="city.ReligionType"/>
        </v-col>
        <v-col cols="6">
          <v-select label="Holy city of" multiple chips density="compact" hide-details
                    :items="withCurrent(enums.religions, ...(city.HolyCityReligionType ?? []))"
                    item-value="type" item-title="description" v-model="city.HolyCityReligionType"/>
        </v-col>
      </v-row>
    </v-card>

    <v-divider class="my-2"/>
    <div class="d-flex align-center">
      <h4>Units</h4>
      <v-spacer/>
      <v-btn size="small" variant="text" prepend-icon="mdi-plus" @click="addUnit">Add unit</v-btn>
    </div>
    <v-card v-for="(unit, idx) in plot.Units ?? []" :key="'u' + idx" variant="outlined" class="pa-2 mb-2">
      <v-row dense>
        <v-col cols="12" class="d-flex align-center">
          <v-autocomplete label="Unit" density="compact" hide-details
                          :items="withCurrent(enums.units, unit.UnitType)" item-value="type" item-title="description"
                          v-model="unit.UnitType"/>
          <v-btn icon="mdi-delete" variant="text" size="small" title="Remove unit"
                 @click="plot.Units = removeAt(plot.Units, idx)"/>
        </v-col>
        <v-col cols="6">
          <v-select label="Owner" density="compact" hide-details :items="ownerItems" v-model="unit.UnitOwner"/>
        </v-col>
        <v-col cols="3">
          <v-text-field label="Level" type="number" min="0" density="compact" hide-details
                        v-model.number="unit.Level"/>
        </v-col>
        <v-col cols="3">
          <v-text-field label="XP" type="number" min="0" density="compact" hide-details
                        v-model.number="unit.Experience"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete label="Promotions" multiple chips closable-chips density="compact" hide-details
                          :items="withCurrent(enums.promotions, ...(unit.PromotionType ?? []))"
                          item-value="type" item-title="description" v-model="unit.PromotionType"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete label="AI role" density="compact" hide-details
                          :items="optional(enums.unitAIs, unit.UnitAIType)" item-value="type" item-title="description"
                          :model-value="noneIfEmpty(unit.UnitAIType)"
                          @update:model-value="(v: string) => unit.UnitAIType = emptyIfNone(v)"/>
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>
