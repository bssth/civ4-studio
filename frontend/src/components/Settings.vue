<script setup lang="ts">
import {
  CheckGameDir,
  ChooseGameDir,
  GetConfig,
  GetLanguages,
  GetModsList,
  GetVersion,
  ResetGameXML,
  SetConfig,
  WriteConsole
} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {computed, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {BrowserOpenURL} from "../../wailsjs/runtime";
import {refreshEnums, xmlReady} from "../store";
import {setUILanguage} from "../i18n";

const {t, te} = useI18n();

const config = ref<editor.Config | null>(null);
const mods = ref<string[]>([]);
const languages = ref<editor.LanguageOption[]>([]);
const dirError = ref('');
const version = ref('');
const releasesUrl = 'https://github.com/bssth/civ4-studio/releases';

async function refreshDirState() {
  dirError.value = await CheckGameDir();
  mods.value = (await GetModsList()) ?? [];
}

async function refreshLanguages() {
  languages.value = (await GetLanguages()) ?? [];
}

onMounted(async () => {
  config.value = await GetConfig();
  version.value = await GetVersion();
  await Promise.all([refreshDirState(), refreshLanguages()]);
});

// Languages appear once the game data is loaded
watch(xmlReady, refreshLanguages);

async function saveConfig() {
  if (!config.value) {
    return;
  }

  await SetConfig(editor.Config.createFrom(config.value));
  WriteConsole(t('settings.saved'));
  await refreshDirState();
}

async function browse() {
  const dir = await ChooseGameDir();
  if (dir && config.value) {
    config.value.game_dir = dir;
    config.value.mod = '';
    await saveConfig();
  }
}

const uiLanguageItems = computed(() => [
  {value: '', title: t('settings.uiLanguageAuto')},
  {value: 'en', title: 'English'},
  {value: 'ru', title: 'Русский'},
]);

async function changeUILanguage(language: string | null) {
  if (!config.value) return;
  config.value.ui_language = language ?? '';
  setUILanguage(config.value.ui_language);
  await saveConfig();
}

function languageTitle(name: string): string {
  return te(`languages.${name}`) ? t(`languages.${name}`) : name;
}

const gameLanguageItems = computed(() => languages.value.map(l => ({
  value: l.name,
  title: languageTitle(l.name),
  subtitle: t('settings.texts', {n: l.texts}),
})));

async function changeGameLanguage(language: string | null) {
  if (!config.value) return;
  config.value.language = language ?? '';
  await saveConfig();
  // Names of the game data (civilizations, techs...) are translated by the backend
  await refreshEnums();
}
</script>

<template>
  <div v-if="!config" class="pa-5 text-grey">
    ...
  </div>

  <div v-else class="pa-5">
    <v-text-field :label="$t('settings.gameDir')" @change="saveConfig"
                  prepend-icon="mdi-controller-classic" v-model="config.game_dir"
                  :error-messages="dirError ? [dirError] : []"
                  :hint="dirError ? '' : $t('settings.gameDirOk')" persistent-hint>
      <template v-slot:append>
        <v-btn variant="tonal" @click="browse">{{ $t('settings.browse') }}</v-btn>
      </template>
    </v-text-field>

    <v-select :label="$t('settings.mod')" class="mt-4" @update:modelValue="saveConfig" clearable
              prepend-icon="mdi-puzzle" :items="mods" v-model="config.mod"
              :hint="$t('settings.modHint')" persistent-hint/>

    <v-checkbox v-model="config.auto_save" class="mt-4" @change="saveConfig"
                :hint="$t('settings.autosaveHint')" persistent-hint>
      <template v-slot:label>
        {{ $t('settings.autosave') }}
      </template>
    </v-checkbox>

    <div class="mt-6 d-flex align-center">
      <v-chip :color="xmlReady ? 'success' : 'grey'" class="me-3">
        {{ xmlReady ? $t('settings.dataLoaded') : $t('settings.dataNotLoaded') }}
      </v-chip>
      <v-btn variant="tonal" prepend-icon="mdi-refresh" :disabled="!!dirError" @click="ResetGameXML">
        {{ $t('settings.reload') }}
      </v-btn>
    </div>

    <v-row class="mt-6">
      <v-col cols="12" md="6">
        <v-select :label="$t('settings.uiLanguage')" prepend-icon="mdi-translate" :items="uiLanguageItems"
                  :model-value="config.ui_language ?? ''" @update:model-value="changeUILanguage" hide-details/>
      </v-col>
      <v-col cols="12" md="6">
        <v-select :label="$t('settings.gameLanguage')" prepend-icon="mdi-book-alphabet"
                  :items="gameLanguageItems" :disabled="languages.length < 2"
                  :model-value="config.language || 'English'" @update:model-value="changeGameLanguage"
                  :hint="$t('settings.gameLanguageHint')" persistent-hint>
          <template v-slot:item="{ props, item }">
            <v-list-item v-bind="props" :subtitle="item.raw.subtitle"/>
          </template>
        </v-select>
      </v-col>
    </v-row>

    <v-divider class="my-6"/>
    <div class="text-caption text-medium-emphasis">
      {{ $t('settings.version', {version: version === 'dev' ? $t('settings.devBuild') : version}) }} ·
      <a href="#" @click.prevent="BrowserOpenURL(releasesUrl)">{{ $t('settings.releases') }}</a>
    </div>
  </div>
</template>
