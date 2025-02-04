import {createAIProvider} from './providers/index.js';
import {getConfig} from './config.js';

export async function makeAPIRequest(diff, options = {}) {
  try {
    const providerType = getConfig('AI_PROVIDER', 'gemini');
    const apiKey = getConfig(`${providerType.toUpperCase()}_KEY`);

    const aiProvider = createAIProvider(providerType, {
      apiKey
    });

    return await aiProvider.generateCommitMessage(diff, options);
  } catch (error) {
    console.error('API Request Error:', error);
    throw error;
  }
}