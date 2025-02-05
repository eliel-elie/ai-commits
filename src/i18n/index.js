import { loadConfig } from '../config.js';
import en from './messages/en.js';
import ptBR from './messages/pt-BR.js';
import es from './messages/es.js';

const messages = {
    'en': en,
    'pt-BR': ptBR,
    'es': es
};

export function t(key, params = {}) {
    const config = loadConfig();
    const locale = config.locale || 'en';
    const translations = messages[locale] || messages['en'];

    let message = key.split('.').reduce((obj, k) => obj?.[k], translations);

    if (!message) {
        console.warn(`Translation missing: ${key}`);
        return key;
    }

    if (typeof message === 'object' && params.count !== undefined) {
        message = message[params.count === 1 ? 'one' : 'other'];
    }

    return message.replace(/\{\{(\w+)\}\}/g, (_, k) => params[k] ?? `{{${k}}}`);
}