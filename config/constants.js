import path from 'path';
import os from 'os';

export const CONFIG_FILE = path.join(os.homedir(), '.ai-commits');
export const DEFAULT_LOCALE = 'en';

export const SUPPORTED_LOCALES = {
    'en': 'english',
    'pt-BR': 'portuguese'
};
