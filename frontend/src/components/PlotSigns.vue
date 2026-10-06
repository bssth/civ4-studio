<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {SetPlotSigns} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {batched, mapRevision, NONE, playerName} from "../store";
import {useI18n} from "vue-i18n";

const props = defineProps<{
  x: number,
  y: number,
  // All signs of the map, the ones of this plot are edited
  signs: editor.Sign[],
  players: editor.Player[],
}>();
const emit = defineEmits<{ (e: 'changed'): void }>();
const {t} = useI18n();

const local = ref<editor.Sign[]>([]);
const error = ref('');
// Last state known to the backend, so selecting a plot does not send it back
let baseline = '';

function plotSigns(): editor.Sign[] {
  return props.signs.filter(s => s.PlotX === props.x && s.PlotY === props.y)
      .map(s => editor.Sign.createFrom(JSON.parse(JSON.stringify(s))));
}

function reset() {
  local.value = plotSigns();
  baseline = JSON.stringify(local.value);
  error.value = '';
}

watch(() => [props.x, props.y], reset, {immediate: true});
// Signs reloaded after a save may be older than the local copy (typing goes on while they load), so only
// undo and redo, which change the map behind the editor, replace it
let revision = mapRevision.value;
watch(() => props.signs, () => {
  if (revision !== mapRevision.value) {
    revision = mapRevision.value;
    reset();
  }
});

const save = batched(async () => {
  const json = JSON.stringify(local.value);
  if (json === baseline) return;
  baseline = json;
  try {
    await SetPlotSigns(props.x, props.y, JSON.parse(json));
    error.value = '';
    emit('changed');
  } catch (err: any) {
    error.value = String(err);
  }
});
watch(local, save, {deep: true});

// -1 shows the sign to everyone
const playerItems = computed(() => [
  {value: -1, title: t('signs.everyone')},
  ...props.players
      .map((p, i) => ({p, i}))
      .filter(({p}) => p.CivType && p.CivType !== NONE)
      .map(({i}) => ({value: i, title: `${i}: ${playerName(props.players, i)}`})),
]);

function items(sign: editor.Sign) {
  const list = playerItems.value;
  // Keep a player that is not in the list (e.g. an empty slot) selectable
  if (!list.some(i => i.value === sign.PlayerType)) {
    return [...list, {value: sign.PlayerType, title: t('common.player', {n: sign.PlayerType})}];
  }
  return list;
}

function add() {
  local.value.push(editor.Sign.createFrom({PlotX: props.x, PlotY: props.y, PlayerType: -1, Caption: t('signs.newCaption')}));
}

function remove(index: number) {
  local.value.splice(index, 1);
}
</script>

<template>
  <div class="px-3 pb-3">
    <div class="d-flex align-center mb-1">
      <h4>{{ $t('signs.title') }}</h4>
      <v-spacer/>
      <v-btn size="small" variant="text" prepend-icon="mdi-sign-text" @click="add">{{ $t('signs.add') }}</v-btn>
    </div>
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-2">{{ error }}</v-alert>
    <div v-if="local.length === 0" class="text-caption text-medium-emphasis">{{ $t('signs.none') }}</div>
    <v-row v-for="(sign, i) in local" :key="i" dense class="align-center">
      <v-col cols="6">
        <v-text-field :label="$t('signs.caption')" density="compact" hide-details v-model="sign.Caption"/>
      </v-col>
      <v-col cols="5">
        <v-select :label="$t('signs.visibleTo')" density="compact" hide-details :items="items(sign)"
                  v-model="sign.PlayerType"/>
      </v-col>
      <v-col cols="1">
        <v-btn icon="mdi-delete" size="small" variant="text" :title="$t('signs.remove')" @click="remove(i)"/>
      </v-col>
    </v-row>
  </div>
</template>
