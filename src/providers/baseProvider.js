import { createCommitPrompt } from '../utils/promptGenerator.js';

export class BaseAIProvider {
    constructor(config = {}) {
        this.config = {
            apiKey: config.apiKey,
            model: config.model ? config.model : this.defaultModel,
            ...config
        };
    }

    get defaultModel() {
        throw new Error('Subclasses must implement defaultModel');
    }

    async generateCommitMessage(diff, options = {}) {
        const prompt = createCommitPrompt(diff, options);
        return this._generateMessage(prompt, options);
    }

    async _generateMessage(prompt, options) {
        throw new Error('Subclasses must implement _generateMessage method');
    }

    _handleAPIResponse(responseData, extractMessageFn) {
        try {
            const parsedResponse = JSON.parse(responseData);

            if (parsedResponse.error) {
                throw new Error(`API Error: ${parsedResponse.error.message}`);
            }

            const message = extractMessageFn(parsedResponse);
            return message ? message.trim() : null;
        } catch (error) {
            console.error('Raw response:', responseData);
            throw new Error(`Failed to parse API response: ${error.message}`);
        }
    }
}