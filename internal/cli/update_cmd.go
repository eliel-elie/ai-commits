package cli

import (
	"fmt"

	"github.com/eliel-elie/ai-commits/internal/ui"
	"github.com/eliel-elie/ai-commits/internal/updater"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for updates and upgrade aic to the latest GitHub release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if checkOnly {
				var latest string
				var hasUpdate bool
				err := ui.RunWithSpinner("Checking for updates...", func() error {
					var e error
					latest, hasUpdate, e = updater.CheckOnly(ctx, Version)
					return e
				})
				if err != nil {
					return err
				}
				if hasUpdate {
					ui.PrintWarn(fmt.Sprintf("Update available: v%s -> %s (run \"aic update\")", Version, latest))
				} else {
					ui.PrintSuccess(fmt.Sprintf("You are already on the latest version (v%s)", Version))
				}
				return nil
			}

			var latest string
			var updated bool
			err := ui.RunWithSpinner("Checking and downloading update...", func() error {
				var e error
				latest, updated, e = updater.SelfUpdate(ctx, Version)
				return e
			})
			if err != nil {
				return err
			}

			if !updated {
				ui.PrintSuccess(fmt.Sprintf("You are already on the latest version (v%s)", Version))
				return nil
			}

			ui.PrintSuccess(fmt.Sprintf("Updated aic from v%s to %s!", Version, latest))
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check if a new version is available without installing")
	return cmd
}
