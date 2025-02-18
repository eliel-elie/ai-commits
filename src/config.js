import fs from 'fs';
import { red, green } from 'kolorist';

import { 
  CONFIG_FILE, 
  SUPPORTED_LOCALES
} from '../config/constants.js';
import {t} from "./i18n/index.js";
import {log} from "@clack/prompts";

export function loadConfig() {
  try {
    if (fs.existsSync(CONFIG_FILE)) {
      return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8'));
    }
  } catch (error) {
    log.error(red(t('config.errorReadFile'), error.message));
  }
  return {};
}

export function updateConfig(key, value = null) {
  try {
    const config = loadConfig();

    if (value === null) {
      return config[key];
    }

    switch (key) {
      case 'locale':
        if (!SUPPORTED_LOCALES[value]) {
          log.error(red(t('config.invalidLocale', { values: Object.keys(SUPPORTED_LOCALES).join(', ') })));
          process.exit(1);
        }
        break;

      case 'provider':
        if (!['openai', 'gemini', 'deepseek'].includes(value)) {
          log.error(red(t('config.invalidProvider')));
          process.exit(1);
        }
        key = 'AI_PROVIDER';
        break;
    }

    if (key === 'AI_PROVIDER') {
      if (config.hasOwnProperty('model')) {
        delete config['model'];
      }
    }

    config[key] = value;
    fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
    log.message(green(t('config.saved',{ key: key, value: value})));

    return value;
  } catch (error) {
    log.error(red(t('config.error'), error.message));
    process.exit(1);
  }
}

export function getConfig(key, defaultValue = null) {
  const config = loadConfig();
  const value = config[key];
  if (!value && defaultValue === null) {
    log.error(red(t('config.notFound', { key: key })));
    process.exit(1);
  }
  return value || defaultValue;
}
