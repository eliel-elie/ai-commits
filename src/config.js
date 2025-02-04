import fs from 'fs';
import { red, green } from 'kolorist';

import { 
  CONFIG_FILE, 
  SUPPORTED_LOCALES
} from '../config/constants.js';

export function loadConfig() {
  try {
    if (fs.existsSync(CONFIG_FILE)) {
      return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf8'));
    }
  } catch (error) {
    console.error(red('Error reading config file:', error.message));
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
          console.error(red(`Locale not supported. Supported values: ${Object.keys(SUPPORTED_LOCALES).join(', ')}`));
          process.exit(1);
        }
        break;

      case 'provider':
        if (!['openai', 'gemini'].includes(value)) {
          console.error(red('Provider not supported. Use: openai or gemini'));
          process.exit(1);
        }
        key = 'AI_PROVIDER';
        break;
    }

    config[key] = value;
    fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
    console.log(green(`✔ ${key}=${value} saved successfully`));

    return value;
  } catch (error) {
    console.error(red('Error saving configuration:', error.message));
    process.exit(1);
  }
}

export function saveConfig(key, value) {
  try {
    const config = loadConfig();

    if (key === 'locale' && !SUPPORTED_LOCALES[value]) {
      console.error(red(`Invalid locale. Supported values are: ${Object.keys(SUPPORTED_LOCALES).join(', ')}`));
      process.exit(1);
    }

    config[key] = value;

    if (key === 'AI_PROVIDER') {
      if (config.hasOwnProperty('model')) {
        delete config['model'];
      }
    }

    fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
    console.log(green(`✔ Successfully saved ${key} to config`));
  } catch (error) {
    console.error(red('Error saving config:', error.message));
    process.exit(1);
  }
}

export function getConfig(key, defaultValue = null) {
  const config = loadConfig();
  const value = config[key];
  if (!value && defaultValue === null) {
    console.error(red(`✖ ${key} not found in config. Use 'config set ${key}=<value>' to set it.`));
    process.exit(1);
  }
  return value || defaultValue;
}

export function handleConfig(args) {
  const command = args[3];
  const keyValue = args[4];

  if (!command || !keyValue) {
    console.error(red('Usage: config <set|get> <KEY=value|KEY>'));
    process.exit(1);
  }

  switch (command.toLowerCase()) {
    case 'set':
      const [key, value] = keyValue.split('=');
      if (!key || !value) {
        console.error(red('Usage: config set KEY=value'));
        process.exit(1);
      }
      saveConfig(key, value);
      break;
    case 'get':
      const configValue = getConfig(keyValue);
      console.log(`${keyValue}=${configValue}`);
      break;
    default:
      console.error(red('Invalid command. Use set or get.'));
      process.exit(1);
  }
  process.exit(0);
}
