# AI Commits 🤖📝

Intelligent commit message generator using different AI providers


## Key Features
- ✅ Supports multiple AI providers (Gemini, OpenAI, and more)
- 🔧 Simple CLI configuration
- 🌍 Multi-language support (English, Portuguese, and others)
- 💡 Context-aware message generation
- 📦 Modular and extensible architecture

## Installation

```bash
# Global install
npm install -g ai-commits

# via npx
npx ai-commits
```

## Configuration

```bash
# Set AI provider
aicommits config set AI_PROVIDER=gemini

# Set API keys
aicommits config set GEMINI_KEY=your_key_here

# Other configurations
aicommits config set locale=pt

```

## Supported Providers

|Provider	| Default Model	       |Environment Variable|
|----------|----------------------|--------------------|
|Gemini	| gemini-1.5-flash	 |GEMINI_KEY|
|OpenAI	| gpt-3.5-turbo	       |OPENAI_KEY|

## Usage

```bash
# Commit with staged files
aicommits
# or 
aic # alias

# Commit including all modified files
aicommits --all

# Specify provider
aicommits config set AI_PROVIDER=gemini

```

## Provider Configuration

### Google Gemini

1. Obtain an API Key from [Google AI Studio](https://aistudio.google.com/).

2. Configure it:

```bash
aicommits config set AI_PROVIDER=gemini
aicommits config set GEMINI_KEY=your_key
```

### OpenAI

1. Obtain an API Key from [OpenAI Dashboard](https://platform.openai.com/).

2. Configure it:

```bash
aicommits config set AI_PROVIDER=openai
aicommits config set OPENAI_KEY=your_key
```

## FAQ

```bash
aicommits config set AI_PROVIDER=novo_provedor
# Ex: aicommits config set AI_PROVIDER=openai
```

### Where is my API key stored?

Credentials are stored locally at ```~/.ai-commits```


** The inspiration for this project was based on [Nutlope/aicommits](https://github.com/Nutlope/aicommits).