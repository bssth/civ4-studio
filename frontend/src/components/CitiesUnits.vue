<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue";
import {GetCities, GetPlayers, GetUnits} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {describeType, enums, mapRevision, mapVersion, NONE, playerColor, playerName, showPlot} from "../store";
import {useI18n} from "vue-i18n";

const {t} = useI18n();

type CityAt = { x: number, y: number, index: number, city: editor.City };
type UnitAt = { x: number, y: number, index: number, unit: editor.Unit };

const cities = ref<CityAt[]>([]);
const units = ref<UnitAt[]>([]);
const players = ref<editor.Player[]>([]);
const view = ref<'cities' | 'units'>('cities');
const search = ref('');
// -1 shows everyone
const owner = ref(-1);

async function load() {
  const [c, u, p] = await Promise.all([GetCities(), GetUnits(), GetPlayers()]);
  cities.value = (c ?? []).flatMap(e => e.city ? [{x: e.x, y: e.y, index: e.index, city: e.city as editor.City}] : []);
  units.value = (u ?? []).flatMap(e => e.unit ? [{x: e.x, y: e.y, index: e.index, unit: e.unit as editor.Unit}] : []);
  players.value = p ?? [];
}

onMounted(load);
watch([mapVersion, mapRevision], () => load());

const ownerItems = computed(() => {
  const owners = new Set<number>([...cities.value.map(c => c.city.CityOwner), ...units.value.map(u => u.unit.UnitOwner)]);
  return [
    {value: -1, title: t('objects.everyone')},
    ...[...owners].sort((a, b) => a - b).map(o => ({value: o, title: ownerName(o)})),
  ];
});

function ownerName(index: number): string {
  const p = players.value[index];
  // The slot after the last player is the barbarians
  if (!p && index === players.value.length) return t('objects.barbarians');
  if (!p || !p.CivType || p.CivType === NONE) return `#${index} ${t('objects.emptySlot')}`;
  return `#${index} ${playerName(players.value, index)}`;
}

function production(c: editor.City): string {
  const pairs: [string | undefined, editor.EnumOption[]][] = [
    [c.ProductionUnit, enums.units], [c.ProductionBuilding, enums.buildings],
    [c.ProductionProject, enums.projects], [c.ProductionProcess, enums.processes],
  ];
  for (const [value, options] of pairs) {
    if (value && value !== NONE) return describeType(options, value);
  }
  return '';
}

function culture(c: editor.City): number {
  return Object.values(c.PlayerCulture ?? {}).reduce((sum, v) => sum + Number(v), 0);
}

const cityRows = computed(() => cities.value
    .filter(e => owner.value === -1 || e.city.CityOwner === owner.value)
    .map(e => ({
      key: `${e.x},${e.y},${e.index}`,
      name: e.city.CityName,
      owner: e.city.CityOwner,
      ownerName: ownerName(e.city.CityOwner),
      population: e.city.CityPopulation,
      buildings: e.city.BuildingType?.length ?? 0,
      religions: [...(e.city.ReligionType ?? [])].map(r => describeType(enums.religions, r)).join(', '),
      holy: (e.city.HolyCityReligionType?.length ?? 0) > 0,
      production: production(e.city),
      culture: culture(e.city),
      x: e.x,
      y: e.y,
      place: `${e.x}, ${e.y}`,
    })));

const unitRows = computed(() => units.value
    .filter(e => owner.value === -1 || e.unit.UnitOwner === owner.value)
    .map(e => ({
      key: `${e.x},${e.y},${e.index}`,
      type: describeType(enums.units, e.unit.UnitType),
      owner: e.unit.UnitOwner,
      ownerName: ownerName(e.unit.UnitOwner),
      level: e.unit.Level,
      experience: e.unit.Experience,
      promotions: e.unit.PromotionType?.length ?? 0,
      ai: describeType(enums.unitAIs, e.unit.UnitAIType),
      damage: e.unit.Damage,
      x: e.x,
      y: e.y,
      place: `${e.x}, ${e.y}`,
    })));

// Units of each type, for the summary above the table
const unitSummary = computed(() => {
  const counts = new Map<string, number>();
  for (const r of unitRows.value) counts.set(r.type, (counts.get(r.type) ?? 0) + 1);
  return [...counts.entries()].sort((a, b) => b[1] - a[1]);
});

const cityHeaders = computed(() => [
  {title: t('objects.name'), key: 'name'},
  {title: t('objects.owner'), key: 'ownerName'},
  {title: t('objects.population'), key: 'population', align: 'end' as const},
  {title: t('objects.buildings'), key: 'buildings', align: 'end' as const},
  {title: t('objects.religions'), key: 'religions'},
  {title: t('objects.production'), key: 'production'},
  {title: t('objects.culture'), key: 'culture', align: 'end' as const},
  {title: t('objects.place'), key: 'place', sortable: false},
]);

const unitHeaders = computed(() => [
  {title: t('objects.unit'), key: 'type'},
  {title: t('objects.owner'), key: 'ownerName'},
  {title: t('objects.level'), key: 'level', align: 'end' as const},
  {title: t('objects.experience'), key: 'experience', align: 'end' as const},
  {title: t('objects.promotions'), key: 'promotions', align: 'end' as const},
  {title: t('objects.ai'), key: 'ai'},
  {title: t('objects.damage'), key: 'damage', align: 'end' as const},
  {title: t('objects.place'), key: 'place', sortable: false},
]);

function open(_: Event, row: { item: { x: number, y: number } }) {
  showPlot(row.item.x, row.item.y);
}
</script>

<template>
  <div class="pa-4">
    <div class="d-flex align-center flex-wrap mb-3" style="gap: 12px">
      <v-btn-toggle v-model="view" mandatory density="compact" divided variant="outlined" color="primary">
        <v-btn value="cities" prepend-icon="mdi-home-city">{{ $t('objects.cities', {n: cities.length}) }}</v-btn>
        <v-btn value="units" prepend-icon="mdi-chess-pawn">{{ $t('objects.units', {n: units.length}) }}</v-btn>
      </v-btn-toggle>
      <v-select :label="$t('objects.owner')" density="compact" hide-details :items="ownerItems" v-model="owner"
                style="max-width: 280px"/>
      <v-text-field v-model="search" :label="$t('objects.search')" prepend-inner-icon="mdi-magnify" clearable
                    density="compact" hide-details style="max-width: 280px"/>
      <v-spacer/>
      <span class="text-caption text-medium-emphasis">{{ $t('objects.hint') }}</span>
    </div>

    <v-data-table v-if="view === 'cities'" :headers="cityHeaders" :items="cityRows" :search="search ?? ''"
                  item-value="key" density="compact" hover :items-per-page="25" class="objects-table"
                  :no-data-text="$t('objects.noCities')" @click:row="open">
      <template v-slot:item.name="{ item }">
        {{ item.name }}
        <v-icon v-if="item.holy" icon="mdi-star-circle" size="x-small" color="amber" :title="$t('objects.holyCity')"/>
      </template>
      <template v-slot:item.ownerName="{ item }">
        <v-icon icon="mdi-circle" size="x-small" :color="playerColor(players, item.owner)" class="me-1"/>{{ item.ownerName }}
      </template>
    </v-data-table>

    <template v-else>
      <div v-if="unitSummary.length" class="d-flex flex-wrap mb-2" style="gap: 6px">
        <v-chip v-for="[type, n] in unitSummary" :key="type" size="small" variant="tonal">{{ type }} × {{ n }}</v-chip>
      </div>
      <v-data-table :headers="unitHeaders" :items="unitRows" :search="search ?? ''"
                    item-value="key" density="compact" hover :items-per-page="25" class="objects-table"
                    :no-data-text="$t('objects.noUnits')" @click:row="open">
        <template v-slot:item.ownerName="{ item }">
          <v-icon icon="mdi-circle" size="x-small" :color="playerColor(players, item.owner)" class="me-1"/>{{ item.ownerName }}
        </template>
        <template v-slot:item.damage="{ item }">{{ item.damage ? item.damage + '%' : '' }}</template>
      </v-data-table>
    </template>
  </div>
</template>

<style scoped>
.objects-table :deep(tbody tr) {
  cursor: pointer;
}
</style>
