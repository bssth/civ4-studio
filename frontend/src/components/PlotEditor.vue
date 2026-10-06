<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {SetPlot} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, emptyIfNone, enums, NONE, noneIfEmpty, playerName, withCurrent, withNone} from "../store";
import {PLOT_HILLS, PLOT_LAND, PLOT_OCEAN, PLOT_PEAK} from "../mapRender";
import {useI18n} from "vue-i18n";

const props = defineProps<{
  plot: editor.Plot,
  players: editor.Player[],
}>();
const emit = defineEmits<{ (e: 'changed'): void }>();
const {t} = useI18n();

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

// Flow directions of the game: 0 north, 1 east, 2 south, 3 west. Without a river the direction is not
// written to the file, so a new river always gets the default direction.
const westEast = computed(() => [{value: 1, title: t('dir.east')}, {value: 3, title: t('dir.west')}]);
const northSouth = computed(() => [{value: 0, title: t('dir.north')}, {value: 2, title: t('dir.south')}]);

const plotTypes = computed(() => [
  {value: PLOT_PEAK, title: t('height.peak')},
  {value: PLOT_HILLS, title: t('height.hills')},
  {value: PLOT_LAND, title: t('height.flat')},
  {value: PLOT_OCEAN, title: t('height.water')},
]);

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
    CityOwner: defaultOwner(), CityName: t('plotEditor.newCity'), CityPopulation: 1,
    ProductionUnit: '', ProductionBuilding: '', ProductionProject: '', ProductionProcess: '',
    BuildingType: [], ReligionType: [], HolyCityReligionType: [], ScriptData: '', PlayerCulture: {},
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
  const south = plot.value.IsNOfRiver, east = plot.value.IsWOfRiver;
  if (south && east) return t('plotEditor.riverBoth');
  if (south) return t('plotEditor.riverSouth');
  if (east) return t('plotEditor.riverEast');
  return '';
});
</script>

<template>
  <div class="pa-3">
    <div class="d-flex align-center mb-2">
      <h3>{{ $t('plotEditor.title', {x: plot.X, y: plot.Y}) }}</h3>
      <v-spacer/>
      <span v-if="riverText" class="text-caption text-medium-emphasis">{{ riverText }}</span>
    </div>

    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-2">{{ error }}</v-alert>

    <v-row dense>
      <v-col cols="7">
        <v-select :label="$t('field.terrain')" density="compact" hide-details
                  :items="withCurrent(enums.terrains, plot.TerrainType)" item-value="type" item-title="description"
                  v-model="plot.TerrainType"/>
      </v-col>
      <v-col cols="5">
        <v-select :label="$t('field.height')" density="compact" hide-details :items="plotTypes" v-model="plot.PlotType"/>
      </v-col>
      <v-col cols="8">
        <v-select :label="$t('field.feature')" density="compact" hide-details
                  :items="optional(enums.features, plot.FeatureType?.[0])" item-value="type" item-title="description"
                  v-model="feature"/>
      </v-col>
      <v-col cols="4">
        <v-text-field :label="$t('field.variety')" type="number" density="compact" hide-details min="0"
                      :disabled="!plot.FeatureType?.length" v-model.number="variety"/>
      </v-col>
      <v-col cols="12">
        <v-autocomplete :label="$t('field.resource')" density="compact" hide-details
                        :items="optional(enums.bonuses, plot.BonusType)" item-value="type" item-title="description"
                        :model-value="noneIfEmpty(plot.BonusType)"
                        @update:model-value="(v: string) => plot.BonusType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="7">
        <v-autocomplete :label="$t('field.improvement')" density="compact" hide-details
                        :items="optional(enums.improvements, plot.ImprovementType)" item-value="type"
                        item-title="description"
                        :model-value="noneIfEmpty(plot.ImprovementType)"
                        @update:model-value="(v: string) => plot.ImprovementType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="5">
        <v-select :label="$t('field.route')" density="compact" hide-details
                  :items="optional(enums.routes, plot.RouteType)" item-value="type" item-title="description"
                  :model-value="noneIfEmpty(plot.RouteType)"
                  @update:model-value="(v: string) => plot.RouteType = emptyIfNone(v)"/>
      </v-col>
      <v-col cols="12">
        <v-text-field :label="$t('plotEditor.landmark')" density="compact" hide-details v-model="plot.Landmark"/>
      </v-col>
    </v-row>
    <v-checkbox v-model="plot.StartingPlot" density="compact" hide-details
                :label="$t('plotEditor.randomStart')"/>

    <v-divider class="my-2"/>
    <h4 class="mb-1">{{ $t('plotEditor.rivers') }}</h4>
    <div class="d-flex align-center" style="gap: 8px">
      <v-checkbox v-model="plot.IsNOfRiver" density="compact" hide-details :label="$t('plotEditor.southEdge')"
                  @update:model-value="(on: boolean | null) => on && (plot.RiverWEDirection = 1)"/>
      <v-select v-if="plot.IsNOfRiver" :label="$t('plotEditor.flows')" density="compact" hide-details style="max-width: 140px"
                :items="westEast" v-model="plot.RiverWEDirection"/>
    </div>
    <div class="d-flex align-center" style="gap: 8px">
      <v-checkbox v-model="plot.IsWOfRiver" density="compact" hide-details :label="$t('plotEditor.eastEdge')"
                  @update:model-value="(on: boolean | null) => on && (plot.RiverNSDirection = 2)"/>
      <v-select v-if="plot.IsWOfRiver" :label="$t('plotEditor.flows')" density="compact" hide-details style="max-width: 140px"
                :items="northSouth" v-model="plot.RiverNSDirection"/>
    </div>

    <v-divider class="my-2"/>
    <div class="d-flex align-center">
      <h4>{{ $t('plotEditor.cities') }}</h4>
      <v-spacer/>
      <v-btn size="small" variant="text" prepend-icon="mdi-plus" @click="addCity">{{ $t('plotEditor.addCity') }}</v-btn>
    </div>
    <v-card v-for="(city, idx) in plot.Cities ?? []" :key="'c' + idx" variant="outlined" class="pa-2 mb-2">
      <v-row dense>
        <v-col cols="12" class="d-flex align-center">
          <v-text-field :label="$t('plotEditor.name')" density="compact" hide-details v-model="city.CityName"/>
          <v-btn icon="mdi-delete" variant="text" size="small" :title="$t('plotEditor.removeCity')"
                 @click="plot.Cities = removeAt(plot.Cities, idx)"/>
        </v-col>
        <v-col cols="8">
          <v-select :label="$t('plotEditor.owner')" density="compact" hide-details :items="ownerItems" v-model="city.CityOwner"/>
        </v-col>
        <v-col cols="4">
          <v-text-field :label="$t('plotEditor.population')" type="number" min="1" density="compact" hide-details
                        v-model.number="city.CityPopulation"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete :label="$t('plotEditor.buildings')" multiple chips closable-chips density="compact" hide-details
                          :items="withCurrent(enums.buildings, ...(city.BuildingType ?? []))"
                          item-value="type" item-title="description" v-model="city.BuildingType"/>
        </v-col>
        <v-col cols="6">
          <v-select :label="$t('plotEditor.religions')" multiple chips density="compact" hide-details
                    :items="withCurrent(enums.religions, ...(city.ReligionType ?? []))"
                    item-value="type" item-title="description" v-model="city.ReligionType"/>
        </v-col>
        <v-col cols="6">
          <v-select :label="$t('plotEditor.holyCity')" multiple chips density="compact" hide-details
                    :items="withCurrent(enums.religions, ...(city.HolyCityReligionType ?? []))"
                    item-value="type" item-title="description" v-model="city.HolyCityReligionType"/>
        </v-col>
      </v-row>
    </v-card>

    <v-divider class="my-2"/>
    <div class="d-flex align-center">
      <h4>{{ $t('plotEditor.units') }}</h4>
      <v-spacer/>
      <v-btn size="small" variant="text" prepend-icon="mdi-plus" @click="addUnit">{{ $t('plotEditor.addUnit') }}</v-btn>
    </div>
    <v-card v-for="(unit, idx) in plot.Units ?? []" :key="'u' + idx" variant="outlined" class="pa-2 mb-2">
      <v-row dense>
        <v-col cols="12" class="d-flex align-center">
          <v-autocomplete :label="$t('plotEditor.unit')" density="compact" hide-details
                          :items="withCurrent(enums.units, unit.UnitType)" item-value="type" item-title="description"
                          v-model="unit.UnitType"/>
          <v-btn icon="mdi-delete" variant="text" size="small" :title="$t('plotEditor.removeUnit')"
                 @click="plot.Units = removeAt(plot.Units, idx)"/>
        </v-col>
        <v-col cols="6">
          <v-select :label="$t('plotEditor.owner')" density="compact" hide-details :items="ownerItems" v-model="unit.UnitOwner"/>
        </v-col>
        <v-col cols="3">
          <v-text-field :label="$t('plotEditor.level')" type="number" min="0" density="compact" hide-details
                        v-model.number="unit.Level"/>
        </v-col>
        <v-col cols="3">
          <v-text-field :label="$t('plotEditor.xp')" type="number" min="0" density="compact" hide-details
                        v-model.number="unit.Experience"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete :label="$t('plotEditor.promotions')" multiple chips closable-chips density="compact" hide-details
                          :items="withCurrent(enums.promotions, ...(unit.PromotionType ?? []))"
                          item-value="type" item-title="description" v-model="unit.PromotionType"/>
        </v-col>
        <v-col cols="12">
          <v-autocomplete :label="$t('plotEditor.aiRole')" density="compact" hide-details
                          :items="optional(enums.unitAIs, unit.UnitAIType)" item-value="type" item-title="description"
                          :model-value="noneIfEmpty(unit.UnitAIType)"
                          @update:model-value="(v: string) => unit.UnitAIType = emptyIfNone(v)"/>
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>
