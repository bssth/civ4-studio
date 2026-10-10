<script setup lang="ts">
import {reactive, ref} from "vue";
import {ImportImage} from "../../wailsjs/go/editor/App";
import {editor} from "../../wailsjs/go/models";
import {reloadEditors} from "../store";
import {useI18n} from "vue-i18n";

const emit = defineEmits<{ (e: 'changed'): void }>();
const {t} = useI18n();

const options = reactive({
  mode: 'height', invert: false, seed: 1, land: 35, hills: 15, peaks: 5, forests: 40, rivers: 15, resources: true,
});
const image = ref<{ name: string, url: string, width: number, height: number } | null>(null);
const file = ref<HTMLInputElement | null>(null);
const busy = ref(false);
const error = ref('');
const message = ref('');
const confirmOpen = ref(false);

const sliders: { key: 'land' | 'hills' | 'peaks' | 'forests' | 'rivers', min: number, max: number, unit: string, height?: boolean }[] = [
  {key: 'land', min: 5, max: 95, unit: '%', height: true},
  {key: 'hills', min: 0, max: 50, unit: '%', height: true},
  {key: 'peaks', min: 0, max: 30, unit: '%', height: true},
  {key: 'forests', min: 0, max: 100, unit: '%', height: true},
  {key: 'rivers', min: 0, max: 100, unit: ''},
];

function choose() {
  file.value?.click();
}

function onFile(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0];
  if (!f) return;
  const reader = new FileReader();
  reader.onload = () => {
    const url = String(reader.result);
    const img = new Image();
    img.onload = () => {
      image.value = {name: f.name, url, width: img.naturalWidth, height: img.naturalHeight};
      error.value = '';
    };
    img.onerror = () => error.value = t('imageImport.notImage');
    img.src = url;
  };
  reader.readAsDataURL(f);
  // The same file can be chosen again after changing it
  (e.target as HTMLInputElement).value = '';
}

async function importImage() {
  confirmOpen.value = false;
  if (!image.value) return;
  busy.value = true;
  try {
    const r = await ImportImage(image.value.url, editor.ImageImportOptions.createFrom({...options}));
    message.value = t('imageImport.done', {land: r.land, rivers: r.rivers, resources: r.resources});
    error.value = '';
    await reloadEditors();
    emit('changed');
  } catch (err: any) {
    error.value = String(err);
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div>
    <h3 class="mb-1">{{ $t('imageImport.title') }}</h3>
    <div class="text-caption text-medium-emphasis mb-3">{{ $t('imageImport.hint') }}</div>
    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3" closable
             @click:close="error = ''">{{ error }}
    </v-alert>
    <v-alert v-if="message" type="success" variant="tonal" density="compact" class="mb-3" closable
             @click:close="message = ''">{{ message }}
    </v-alert>
    <input ref="file" type="file" accept="image/png,image/jpeg,image/gif" class="d-none" @change="onFile"/>
    <v-row dense>
      <v-col cols="12" md="6">
        <div class="d-flex align-center flex-wrap mb-2" style="gap: 8px">
          <v-btn variant="tonal" prepend-icon="mdi-image-outline" @click="choose">{{ $t('imageImport.choose') }}</v-btn>
          <span v-if="image" class="text-body-2 text-truncate" style="max-width: 260px">
            {{ image.name }} ({{ image.width }}×{{ image.height }})
          </span>
        </div>
        <v-btn-toggle v-model="options.mode" mandatory density="compact" divided variant="outlined" color="primary"
                      class="mb-2">
          <v-btn value="height" size="small" prepend-icon="mdi-image-filter-hdr">{{ $t('imageImport.height') }}</v-btn>
          <v-btn value="colors" size="small" prepend-icon="mdi-palette">{{ $t('imageImport.colors') }}</v-btn>
        </v-btn-toggle>
        <div class="text-caption text-medium-emphasis mb-2">
          {{ options.mode === 'height' ? $t('imageImport.heightHint') : $t('imageImport.colorsHint') }}
        </div>
        <v-checkbox v-if="options.mode === 'height'" v-model="options.invert" :label="$t('imageImport.invert')"
                    density="compact" hide-details/>
        <template v-for="s in sliders" :key="s.key">
          <div v-if="!s.height || options.mode === 'height'" class="d-flex align-center">
            <span class="text-body-2" style="width: 150px">{{ $t('generator.' + s.key) }}</span>
            <v-slider v-model="options[s.key]" :min="s.min" :max="s.max" :step="1" hide-details density="compact"
                      class="flex-grow-1"/>
            <span class="text-body-2 text-end" style="width: 48px">{{ options[s.key] }}{{ s.unit }}</span>
          </div>
        </template>
        <v-checkbox v-model="options.resources" :label="$t('generator.resources')" density="compact" hide-details/>
        <v-btn color="primary" prepend-icon="mdi-import" class="mt-2" :disabled="!image" :loading="busy"
               @click="confirmOpen = true">{{ $t('imageImport.import') }}
        </v-btn>
      </v-col>
      <v-col cols="12" md="6" class="d-flex justify-center align-start">
        <img v-if="image" :src="image.url" :alt="image.name" class="picture"/>
      </v-col>
    </v-row>

    <v-dialog v-model="confirmOpen" max-width="480">
      <v-card :title="$t('imageImport.confirmTitle')">
        <v-card-text>{{ $t('generator.confirmText') }}</v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn variant="text" @click="confirmOpen = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn color="primary" variant="tonal" @click="importImage">{{ $t('imageImport.import') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>
.picture {
  max-width: 100%;
  max-height: 220px;
  border: 1px solid rgba(0, 0, 0, 0.15);
}
</style>
