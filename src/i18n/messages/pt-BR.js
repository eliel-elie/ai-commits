import {getStagedFiles} from "../../git.js";

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
        errorGenerating: 'Erro: Não foi possível gerar a mensagem de commit.',
        noFileSpecified: 'Erro: Nenhum arquivo especificado após a flag --file. Use: aicommits --file <caminho_do_arquivo>',
        stagingFile: 'Preparando arquivo: {{file}}',
        fileStageError: 'Erro ao preparar o arquivo {{file}}: {{error}}',
        version: 'Versão: {{version}}'
    },
    config: {
        saved: '✔ {{key}}={{value}} salvo com sucesso',
        error: 'Erro ao salvar configuração:',
        notFound: '✖ {{key}} não encontrado na configuração. Use a flag apropriada para configurá-lo.',
        invalidLocale: 'Idioma inválido. Valores suportados: {{values}}',
        invalidProvider: 'Provider inválido. Use: openai ou gemini',
        errorReadFile: 'Erro ao ler o arquivo de configuração:'
    },
    dryRun: {
        title: '🔍 Modo Simulação',
        divider: '==============',
        stagedFiles: '📋 Arquivos Preparados:',
        commitMessage: '💬 Mensagem de Commit:',
        diffPreview: '📊 Preview das Alterações:'
    },
    confirm: {
        yes: 'Sim',
        no: 'Não',
    },
    git: {
        reStaging: 'Pre-commit hook modificou alguns arquivos. Re-staging alterações...',
        failedPreCommit: 'Pre-commit hook falhou. Corriga os problemas antes de continuar.',
        error: 'Erro ao executar pre-commit hook:',
        errorStagedFiles: 'Erro ao recuperar arquivos mofificados:',
        errorCommit: 'Erro ao confirmar alterações:'
    }
};