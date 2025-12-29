import https from 'https';
import { BaseAIProvider } from './baseProvider.js';

export class OpenAIProvider extends BaseAIProvider {
    get defaultModel() {
        return 'gpt-5-mini';
    }

    async _generateMessage(prompt, options) {
        const model = this.config.model.trim().toLowerCase();
        const isRestricted = model.startsWith('o1') || model.startsWith('gpt-5');
        
        const systemContent = prompt.split("Analyze the")[0];
        
        let messages;
        if (isRestricted) {
             messages = [
                {
                    role: 'user',
                    content: systemContent + "\n\nAnalyze the" + prompt.split("Analyze the")[1]
                }
            ];
        } else {
             messages = [
                {
                    role: 'system',
                    content: systemContent
                },
                {
                    role: 'user',
                    content: prompt
                }
            ];
        }

        const payload = {
            model: this.config.model,
            messages: messages,
            max_completion_tokens: 5000
        };

        if (!isRestricted) {
            payload.temperature = 0.7;
        }

        const data = JSON.stringify(payload);

        return new Promise((resolve, reject) => {
            const requestOptions = {
                hostname: 'api.openai.com',
                path: '/v1/chat/completions',
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.config.apiKey}`,
                    'Content-Type': 'application/json'
                },
                rejectUnauthorized: false
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
                        error.message = `[Model: ${this.config.model}] ${error.message}`;
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