
import { OpenAIProvider } from './openai.js';
import { GeminiProvider } from './gemini.js';

const PROVIDERS = {
    openai: OpenAIProvider,
    gemini: GeminiProvider
};

export function createAIProvider(type, config) {
    const ProviderClass = PROVIDERS[type.toLowerCase()];

    if (!ProviderClass) {
        throw new Error(`Unsupported AI provider: ${type}`);
    }

    return new ProviderClass(config);
}

export {
    OpenAIProvider,
    GeminiProvider
};