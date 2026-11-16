package ui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliel-elie/ai-commits/internal/i18n"
)

// Action represents the user's choice after reviewing a generated commit message.
type Action string

const (
	ActionConfirm    Action = "confirm"
	ActionEdit       Action = "edit"
	ActionRegenerate Action = "regenerate"
	ActionCancel     Action = "cancel"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("36")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("36")).
			Padding(0, 1).
			Margin(1, 0)
)

// PrintHeader renders the CLI banner and active provider/directory info.
func PrintHeader(appName, providerLabel, providerName, modelName, workDir string) {
	fmt.Println(titleStyle.Render(appName))
	meta := fmt.Sprintf("%s: %s (%s)", providerLabel, providerName, modelName)
	if workDir != "" {
		meta += fmt.Sprintf("  |  dir: %s", workDir)
	}
	fmt.Println(infoStyle.Render(meta))
}

// PrintFiles displays the list of staged files.
func PrintFiles(header string, files []string) {
	fmt.Println(successStyle.Render("✔ " + header))
	for _, f := range files {
		fmt.Println("  " + fileStyle.Render(f))
	}
}

// PrintUnstagedFiles displays the list of unstaged files with a warning header.
func PrintUnstagedFiles(header string, files []string) {
	fmt.Println(warnStyle.Render("! " + header))
	for _, f := range files {
		fmt.Println("  " + fileStyle.Render(f))
	}
}

// PrintMessageBox renders the generated commit message inside a bordered box.
func PrintMessageBox(message string) {
	fmt.Println(boxStyle.Render(message))
}

// PrintSuccess prints a green success line.
func PrintSuccess(text string) {
	fmt.Println(successStyle.Render("✔ " + text))
}

// PrintWarn prints a yellow warning line.
func PrintWarn(text string) {
	fmt.Println(warnStyle.Render("! " + text))
}

// PrintInfo prints a subtle info line.
func PrintInfo(text string) {
	fmt.Println(infoStyle.Render(text))
}

// PrintError prints a red error line.
func PrintError(text string) {
	fmt.Println(errorStyle.Render("✖ " + text))
}

// RunWithSpinner executes fn while showing an animated terminal spinner.
func RunWithSpinner(title string, fn func() error) error {
	var runErr error
	err := spinner.New().
		Title(title).
		Action(func() {
			runErr = fn()
		}).
		Run()
	if err != nil {
		return err
	}
	return runErr
}

// ChooseAction prompts the user to confirm, edit, regenerate, or cancel.
func ChooseAction(m i18n.Messages) (Action, error) {
	var choice Action
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Action]().
				Title(m.ActionPrompt).
				Options(
					huh.NewOption(m.ActionConfirm, ActionConfirm),
					huh.NewOption(m.ActionEdit, ActionEdit),
					huh.NewOption(m.ActionRegenerate, ActionRegenerate),
					huh.NewOption(m.ActionCancel, ActionCancel),
				).
				Value(&choice),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ActionCancel, nil
		}
		return ActionCancel, err
	}
	return choice, nil
}

// EditMessage opens a multi-line text editor in the terminal prefilled with initial.
func EditMessage(m i18n.Messages, initial string) (string, bool, error) {
	val := initial
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title(m.EditPrompt).
				Value(&val),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", false, nil
		}
		return "", false, err
	}
	return val, true, nil
}

// AskHint prompts for an optional hint before regenerating.
func AskHint(m i18n.Messages) (string, bool, error) {
	var hint string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(m.HintPrompt).
				Value(&hint),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", false, nil
		}
		return "", false, err
	}
	return hint, true, nil
}

// ConfirmStage asks the user whether to stage detected unstaged changes and continue.
func ConfirmStage(title string) (bool, error) {
	confirm := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Value(&confirm),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, err
	}
	return confirm, nil
}
