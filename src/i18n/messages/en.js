export default {
    cli: {
        intro: ' ai-commits ',
        provider: 'Provider',
        notConfigured: 'not configured',
        stagingAll: 'Staging all changes',
        detectingFiles: 'Detecting staged files',
        noChanges: 'No staged changes found. Stage your changes manually, or use --all',
        filesDetected: {
            one: 'Detected {{count}} staged file:',
            other: 'Detected {{count}} staged files:'
        },
        analyzing: 'The AI is analyzing your changes',
        changesAnalyzed: 'Changes analyzed',
        useMessage: 'Use this commit message?',
        commitCancelled: 'Commit cancelled',
        commitSuccess: 'Successfully committed!',
        errorGenerating: 'Error: Could not generate commit message.',
        stagingFile: 'Staging file: {{file}}',
        noFileSpecified: 'Error: No file specified after --file flag. Usage: aicommits --file <file_path>',
        fileStageError: 'Error staging file {{file}}: {{error}}',
        version: 'Version: {{version}}'
    },
    config: {
        saved: '✔ {{key}}={{value}} saved successfully',
        error: 'Error saving config:',
        notFound: '✖ {{key}} not found in config. Use appropriate flag to set it.',
        invalidLocale: 'Invalid locale. Supported values are: {{values}}',
        invalidProvider: 'Invalid provider. Use: openai or gemini',
        errorReadFile: 'Error reading config file:'
    },
    dryRun: {
        title: '🔍 Dry Run Mode',
        divider: '==============',
        stagedFiles: '📋 Staged Files:',
        commitMessage: '💬 Commit Message:',
        diffPreview: '📊 Diff Preview:'
    },
    confirm: {
        yes: 'Yes',
        no: 'No',
    }
};