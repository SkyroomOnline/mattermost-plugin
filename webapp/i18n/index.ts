import * as fa from './fa.json';
import * as en from './en.json';

export function getTranslations(locale: string): {[key: string]: string} {
    switch (locale) {
    case 'fa':
        return fa;
    case 'en':
        return en;
    }
    return {};
}

