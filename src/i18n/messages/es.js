export default {
    cli: {
        intro: ' ai-commits ',
        provider: 'Proveedor',
        notConfigured: 'no configurado',
        stagingAll: 'Preparando todos los cambios',
        detectingFiles: 'Detectando archivos',
        noChanges: 'No se encontraron cambios preparados. Prepare sus cambios manualmente o use --all',
        filesDetected: {
            one: 'Detectado {{count}} archivo preparado:',
            other: 'Detectados {{count}} archivos preparados:'
        },
        analyzing: 'La IA está analizando sus cambios',
        changesAnalyzed: 'Cambios analizados',
        useMessage: '¿Usar este mensaje de commit?',
        commitCancelled: 'Commit cancelado',
        commitSuccess: '✔ ¡Commit realizado con éxito!',
        errorGenerating: 'Error: No se pudo generar el mensaje de commit.',
        stagingFile: 'Preparando archivo: {{file}}',
        noFileSpecified: 'Error: No se especificó ningún archivo después de la bandera --file. Uso: aicommits --file <ruta_del_archivo>',
        fileStageError: 'Error al preparar el archivo {{file}}: {{error}}',
        version: 'Versión: {{version}}'
    },
    config: {
        saved: '✔ {{key}}={{value}} guardado con éxito',
        error: 'Error al guardar la configuración:',
        notFound: '✖ {{key}} no encontrado en la configuración. Use la bandera apropiada para configurarlo.',
        invalidLocale: 'Locale inválido. Valores soportados: {{values}}',
        invalidProvider: 'Proveedor inválido. Use: openai o gemini',
        errorReadFile: 'Error al leer el archivo de configuración:'
    },
    dryRun: {
        title: '🔍 Modo Simulación',
        divider: '==============',
        stagedFiles: '📋 Archivos Preparados:',
        commitMessage: '💬 Mensaje de Commit:',
        diffPreview: '📊 Vista Previa de los Cambios:'
    },
    confirm: {
        yes: 'Sí',
        no: 'No',
    },
    git: {
        reStaging: 'El gancho de pre-confirmación modificó algunos archivos. Cambios en la puesta en escena...',
        failedPreCommit: 'El gancho de pre-confirmación falló. Corrija los problemas antes de continuar.',
        error: 'Error al ejecutar el gancho de pre-confirmación:',
        errorStagedFiles: 'Error al recuperar archivos almacenados:',
        errorCommit: 'Error al confirmar los cambios:'
    }
};