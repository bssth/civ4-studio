import { createApp } from 'vue'

import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import { en, ru } from 'vuetify/locale'
import '@mdi/font/css/materialdesignicons.css'

import App from './App.vue'
import { bindVuetifyLocale, i18n } from './i18n'

const vuetify = createVuetify({
    components,
    directives,
    icons: {
        defaultSet: 'mdi',
        aliases,
        sets: {
            mdi,
        },
    },
    locale: {
        locale: 'en',
        fallback: 'en',
        messages: { en, ru },
    },
})

bindVuetifyLocale(vuetify.locale)

createApp(App).use(vuetify).use(i18n).mount('#app')
