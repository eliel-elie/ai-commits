import https from 'https';
import { BaseAIProvider } from './baseProvider.js';

export class GeminiProvider extends BaseAIProvider {
    get defaultModel() {
        return 'gemini-1.5-flash';
    }

    async _generateMessage(prompt, options) {
        const data = JSON.stringify({
            contents: [
                {
                    parts: [{ text: prompt }]
                }
            ],
            generationConfig: {
                temperature: 0.7,
                maxOutputTokens: 200
            }
        });

        return new Promise((resolve, reject) => {
            const requestOptions = {
                hostname: 'generativelanguage.googleapis.com',
                path: `/v1beta/models/${this.config.model}:generateContent?key=${this.config.apiKey}`,
                method: 'POST',
                headers: {
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
                            (parsedResponse) => parsedResponse.candidates[0]?.content?.parts[0]?.text
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