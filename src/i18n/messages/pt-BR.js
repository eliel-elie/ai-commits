export default {
    cli: {
        intro: ' ai-commits ',
        provider: 'Provedor',
        notConfigured: 'não configurado',
        stagingAll: 'Preparando todas as alterações',
        detectingFiles: 'Detectando arquivos',
        noChanges: 'Nenhuma alteração preparada encontrada. Prepare suas alterações manualmente ou use --all',
        filesDetected: {
            one: 'Detectado {{count}} arquivo preparado:',
            other: 'Detectados {{count}} arquivos preparados:'
        },
        analyzing: 'A IA está analisando suas alterações',
        changesAnalyzed: 'Alterações analisadas',
        useMessage: 'Usar esta mensagem de commit?',
        commitCancelled: 'Commit cancelado',
        commitSuccess: '✔ Commit realizado com sucesso!',
        errorGenerating: 'Erro: Não foi possível gerar a mensagem de commit.'
    },
    config: {
        saved: '✔ {{key}}={{value}} salvo com sucesso',
        error: 'Erro ao salvar configuração:',
        notFound: '✖ {{key}} não encontrado na configuração. Use a flag apropriada para configurá-lo.',
        invalidLocale: 'Idioma inválido. Valores suportados: {{values}}',
        invalidProvider: 'Provider inválido. Use: openai ou gemini'
    },
    dryRun: {
        title: '🔍 Modo Simulação',
        divider: '==============',
        stagedFiles: '📋 Arquivos Preparados:',
        commitMessage: '💬 Mensagem de Commit:',
        diffPreview: '📊 Preview das Alterações:'
    }
};