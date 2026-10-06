// Types of the global $t/$te helpers that vue-i18n adds to every component template.
// vue-tsc 1.x resolves component instances through @vue/runtime-core, newer versions through vue.
import type {ComposerTranslation} from 'vue-i18n';

declare module '@vue/runtime-core' {
    interface ComponentCustomProperties {
        $t: ComposerTranslation;
        $te: (key: string) => boolean;
    }
}

declare module 'vue' {
    interface ComponentCustomProperties {
        $t: ComposerTranslation;
        $te: (key: string) => boolean;
    }
}

export {};
