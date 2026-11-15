package i18n

import (
	"fmt"
	"strings"
)

// Messages holds localized UI strings for a single locale.
type Messages struct {
	Intro            string
	ProviderLabel    string
	StagingAll       string
	StagingFiles     string
	DetectingFiles   string
	NoChanges        string
	FilesDetectedOne string
	FilesDetectedN   string
	Analyzing        string
	ChangesAnalyzed  string
	DryRunHeader     string
	ActionPrompt     string
	ActionConfirm    string
	ActionEdit       string
	ActionRegenerate string
	ActionCancel     string
	EditPrompt       string
	HintPrompt       string
	CommitCancelled  string
	CommitSuccess    string
	ConfigSaved      string
	NotConfigured    string
	MissingAPIKey    string
	TruncatedDiff    string
}

var catalog = map[string]Messages{
	"en": {
		Intro:            "ai-commits",
		ProviderLabel:    "Provider",
		StagingAll:       "Staging all changes...",
		StagingFiles:     "Staging %d specific file(s)...",
		DetectingFiles:   "Detecting staged files...",
		NoChanges:        "No staged changes detected. Stage your changes or use --all / --file.",
		FilesDetectedOne: "1 staged file detected:",
		FilesDetectedN:   "%d staged files detected:",
		Analyzing:        "Analyzing changes with AI...",
		ChangesAnalyzed:  "Changes analyzed",
		DryRunHeader:     "Generated commit message (dry run):",
		ActionPrompt:     "What would you like to do with this commit message?",
		ActionConfirm:    "Commit with this message",
		ActionEdit:       "Edit message",
		ActionRegenerate: "Regenerate message",
		ActionCancel:     "Cancel",
		EditPrompt:       "Edit the commit message:",
		HintPrompt:       "Optional hint for regeneration (leave empty to retry as-is):",
		CommitCancelled:  "Commit cancelled.",
		CommitSuccess:    "Successfully committed!",
		ConfigSaved:      "Saved %s = %s",
		NotConfigured:    "not configured",
		MissingAPIKey:    "API key not configured for provider %q. Set it via `aic config set %s-key <KEY>` or environment variable.",
		TruncatedDiff:    "Note: diff exceeded %d bytes and was truncated to fit context limits.",
	},
	"pt-BR": {
		Intro:            "ai-commits",
		ProviderLabel:    "Provedor",
		StagingAll:       "Adicionando todas as alterações ao stage...",
		StagingFiles:     "Adicionando %d arquivo(s) específico(s) ao stage...",
		DetectingFiles:   "Detectando arquivos no stage...",
		NoChanges:        "Nenhuma alteração no stage encontrada. Use `git add`, --all ou --file.",
		FilesDetectedOne: "1 arquivo no stage detectado:",
		FilesDetectedN:   "%d arquivos no stage detectados:",
		Analyzing:        "Analisando alterações com IA...",
		ChangesAnalyzed:  "Alterações analisadas",
		DryRunHeader:     "Mensagem de commit gerada (dry run):",
		ActionPrompt:     "O que você deseja fazer com esta mensagem de commit?",
		ActionConfirm:    "Commitar com esta mensagem",
		ActionEdit:       "Editar mensagem",
		ActionRegenerate: "Gerar novamente",
		ActionCancel:     "Cancelar",
		EditPrompt:       "Edite a mensagem de commit:",
		HintPrompt:       "Dica opcional para gerar novamente (vazio para tentar igual):",
		CommitCancelled:  "Commit cancelado.",
		CommitSuccess:    "Commit realizado com sucesso!",
		ConfigSaved:      "Configuração salva: %s = %s",
		NotConfigured:    "não configurado",
		MissingAPIKey:    "Chave de API não configurada para o provedor %q. Configure com `aic config set %s-key <CHAVE>` ou variável de ambiente.",
		TruncatedDiff:    "Aviso: o diff excedeu %d bytes e foi truncado para caber no limite de contexto.",
	},
	"es": {
		Intro:            "ai-commits",
		ProviderLabel:    "Proveedor",
		StagingAll:       "Agregando todos los cambios al stage...",
		StagingFiles:     "Agregando %d archivo(s) específico(s) al stage...",
		DetectingFiles:   "Detectando archivos en stage...",
		NoChanges:        "No se detectaron cambios en stage. Usa `git add`, --all o --file.",
		FilesDetectedOne: "1 archivo en stage detectado:",
		FilesDetectedN:   "%d archivos en stage detectados:",
		Analyzing:        "Analizando cambios con IA...",
		ChangesAnalyzed:  "Cambios analizados",
		DryRunHeader:     "Mensaje de commit generado (dry run):",
		ActionPrompt:     "¿Qué deseas hacer con este mensaje de commit?",
		ActionConfirm:    "Hacer commit con este mensaje",
		ActionEdit:       "Editar mensaje",
		ActionRegenerate: "Regenerar mensaje",
		ActionCancel:     "Cancelar",
		EditPrompt:       "Edita el mensaje de commit:",
		HintPrompt:       "Indicación opcional para regenerar (vacío para reintentar):",
		CommitCancelled:  "Commit cancelado.",
		CommitSuccess:    "¡Commit realizado con éxito!",
		ConfigSaved:      "Guardado %s = %s",
		NotConfigured:    "no configurado",
		MissingAPIKey:    "Falta la clave API para el proveedor %q. Configúrala con `aic config set %s-key <CLAVE>` o variable de entorno.",
		TruncatedDiff:    "Aviso: el diff superó %d bytes y fue truncado para ajustarse al límite de contexto.",
	},
}

// Get returns the Messages bundle for the given locale, defaulting to "en".
func Get(locale string) Messages {
	if m, ok := catalog[locale]; ok {
		return m
	}
	for k, m := range catalog {
		if strings.EqualFold(k, locale) {
			return m
		}
	}
	return catalog["en"]
}

// FilesDetected returns the singular or plural count message.
func (m Messages) FilesDetected(count int) string {
	if count == 1 {
		return m.FilesDetectedOne
	}
	return fmt.Sprintf(m.FilesDetectedN, count)
}
