export const COMMIT_TYPES = {
    docs: 'Documentation only changes',
    style: 'Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)',
    refactor: 'A code change that neither fixes a bug nor adds a feature',
    perf: 'A code change that improves performance',
    test: 'Adding missing tests or correcting existing tests',
    build: 'Changes that affect the build system or external dependencies',
    ci: 'Changes to our CI configuration files and scripts',
    chore: 'Other changes that don\'t modify src or test files',
    feat: 'A new feature',
    fix: 'A bug fix'
};

export class PromptBuilder {
    constructor(options = {}) {
        this.options = {
            maxLength: 100,
            language: 'english',
            context: {},
            style: 'concise',
            ...options
        };
    }

    generateCommitMessagePrompt(diff, customOptions = {}) {
        const mergedOptions = { ...this.options, ...customOptions };

        const typeDescriptions = Object.entries(COMMIT_TYPES)
            .map(([type, description]) => `- ${type}: '${description}'`)
            .join('\n');

        return `Generate a git commit message with the following specifications:
Language: ${mergedOptions.language}
Maximum Length: ${mergedOptions.maxLength} characters
Style: ${mergedOptions.style}

Commit Type Guidelines:
${typeDescriptions}

Additional Context:
${JSON.stringify(mergedOptions.context, null, 2)}

Rules:
- Write in present tense
- Be descriptive yet concise
- Exclude unnecessary details
- Directly representable as a git commit message

Analyze the following code changes and generate an appropriate commit message:
${diff}`;
    }

    generateGenericPrompt(template, data) {
        if (typeof template !== 'string' || typeof data !== 'object') {
            throw new Error('Invalid template or data');
        }

        const interpolatedTemplate = template.replace(/\{\{(\w+)\}\}/g, (match, key) =>
            data[key] !== undefined ? data[key] : match
        );

        return interpolatedTemplate;
    }

    validateMessageLength(message, maxLength) {
        const length = maxLength || this.options.maxLength;

        if (message.length > length) {
            const truncated = message.substring(0, length).trim() + '…';
            return truncated;
        }
        return message;
    }
}

export function createCommitPrompt(diff, options = {}) {
    const promptBuilder = new PromptBuilder(options);
    return promptBuilder.generateCommitMessagePrompt(diff);
}