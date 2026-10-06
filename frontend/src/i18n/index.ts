import {createI18n} from "vue-i18n";
import en from "./en";
import ru from "./ru";

export const uiLanguages = ['en', 'ru'] as const;
export type UILanguage = typeof uiLanguages[number];

/** The configured language, or the system one when it is not set (English if not supported) */
export function resolveUILanguage(configured?: string | null): UILanguage {
    const wanted = (configured || navigator.language || 'en').toLowerCase().slice(0, 2);
    return (uiLanguages as readonly string[]).includes(wanted) ? wanted as UILanguage : 'en';
}

export const i18n = createI18n({
    legacy: false,
    locale: resolveUILanguage(),
    fallbackLocale: 'en',
    messages: {en, ru},
});

// Vuetify has its own strings (no data, close etc.), it follows the same language
type VuetifyLocale = { current: { value: string } };
let vuetifyLocale: VuetifyLocale | null = null;

export function bindVuetifyLocale(locale: VuetifyLocale) {
    vuetifyLocale = locale;
    locale.current.value = i18n.global.locale.value;
}

export function setUILanguage(configured?: string | null) {
    const language = resolveUILanguage(configured);
    i18n.global.locale.value = language;
    if (vuetifyLocale) vuetifyLocale.current.value = language;
    document.documentElement.lang = language;
}

/** Translation outside of components */
export const t = i18n.global.t;
