<template>
  <v-app id="inspire">
    <v-system-bar window>
      <v-icon class="me-4" icon="mdi-content-save-all" />
      <v-icon class="me-4" icon="mdi-folder-open" />
      <v-icon class="me-4" icon="mdi-rocket-launch" />
      <v-icon class="me-4" icon="mdi-cog" @click="tab = 'settings'" />

      <v-spacer></v-spacer>

      <v-btn icon="mdi-minus" variant="text" @click="minimize" />
      <v-btn class="ms-2" icon="mdi-checkbox-blank-outline" variant="text" @click="maximize" />
      <v-btn class="ms-2" icon="mdi-close" variant="text" @click="quit" />
    </v-system-bar>

    <v-app-bar
        class="px-3"
        density="compact"
        flat style="--wails-draggable:no-drag"
    >
      <v-avatar
          class="hidden-md-and-up"
          color="grey-darken-1"
          size="32"
      ></v-avatar>

      <v-spacer></v-spacer>

      <v-tabs
          color="grey-darken-2"
          centered v-model="tab"
      >
        <v-tab key="link" value="wb">
          <v-icon icon="mdi-map"></v-icon>
          Map Edit
        </v-tab>
        <v-tab key="link" value="map">
          <v-icon icon="mdi-tune-vertical-variant"></v-icon>
          Map Setting
        </v-tab>
        <v-tab key="link" value="teams">
          <v-icon icon="mdi-account-group"></v-icon>
          Teams
        </v-tab>
        <v-tab key="link" value="players">
          <v-icon icon="mdi-human-edit"></v-icon>
          Players
        </v-tab>
      </v-tabs>
      <v-spacer />
    </v-app-bar>

    <v-main class="bg-grey-lighten-3" style="--wails-draggable:no-drag">
      <v-container>
        <v-row>
          <v-col
              cols="12"
              md="9"
          >
            <v-sheet
                height="80vh"
                rounded="lg"
            >
              <!-- wb map teams players -->
              <Settings v-if="tab == 'settings'" />
              <div v-else>
                To start editing, open a map or create a new one
              </div>
            </v-sheet>
          </v-col>

          <v-col
              cols="12"
              md="3"
          >
            <v-sheet
                height="80vh"
                rounded="lg" class="pa-3"
            >
              <Console />
            </v-sheet>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import {ref} from "vue";
import Console from "./components/Console.vue";
import {Quit, WindowMaximise, WindowMinimise, WindowToggleMaximise} from "../wailsjs/runtime";
import Settings from "./components/Settings.vue";

const minimize = WindowMinimise;
const maximize = WindowToggleMaximise;
const quit = Quit;

const tab = ref('');

WindowMaximise();
</script>

<style>
body {
  overflow: hidden;
}
</style>
