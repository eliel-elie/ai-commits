import https from 'https';
import { BaseAIProvider } from './baseProvider.js';

export class OpenAIProvider extends BaseAIProvider {
    get defaultModel() {
        return 'gpt-3.5-turbo';
    }

    async _generateMessage(prompt, options) {
        const data = JSON.stringify({
            model: this.config.model,
            messages: [
                {
                    role: 'system',
                    content: prompt.split("Analyze the")[0]
                },
                {
                    role: 'user',
                    content: prompt
                }
            ],
            temperature: 0.7,
            max_tokens: 200
        });

        return new Promise((resolve, reject) => {
            const requestOptions = {
                hostname: 'api.openai.com',
                path: '/v1/chat/completions',
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.config.apiKey}`,
                    'Content-Type': 'application/json'
                }
            };

            const req = https.request(requestOptions, (res) => {
                let responseData = '';

                res.on('data', (chunk) => responseData += chunk);

                res.on('end', () => {
                    try {
                        const message = this._handleAPIResponse(
                            responseData,
                            (parsedResponse) => parsedResponse.choices[0]?.message?.content
                        );
                        resolve(message);
                    } catch (error) {
                        reject(error);
                    }
                });
            });

            req.on('error', (error) => {
                console.error('Request error:', error);
                reject(error);
            });

            req.write(data);
            req.end();
        });
    }
}