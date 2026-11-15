package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eliel-elie/ai-commits/internal/config"
	"github.com/eliel-elie/ai-commits/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global ai-commits configuration (~/.ai-commits)",
	}

	cmd.AddCommand(
		newConfigSetCmd(),
		newConfigGetCmd(),
		newConfigListCmd(),
		newConfigPathCmd(),
	)

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value (provider, model, locale, max-length, max-tokens, base-url, gemini-key, openai-key, deepseek-key, groq-key, openrouter-key)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			canonical, err := config.SetGlobal(args[0], args[1])
			if err != nil {
				return err
			}
			displayVal := args[1]
			if strings.HasSuffix(canonical, "_KEY") {
				displayVal = config.MaskSecret(displayVal)
			}
			ui.PrintSuccess(fmt.Sprintf("Saved %s = %s", canonical, displayVal))
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value from ~/.ai-commits",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := config.GetGlobalRaw()
			if err != nil {
				return err
			}
			val, ok := raw[args[0]]
			if !ok {
				return fmt.Errorf("key %q not set", args[0])
			}
			if strings.HasSuffix(strings.ToUpper(args[0]), "_KEY") {
				if s, ok := val.(string); ok {
					val = config.MaskSecret(s)
				}
			}
			fmt.Println(val)
			return nil
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all values in ~/.ai-commits (with API keys masked)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := config.GetGlobalRaw()
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				ui.PrintInfo("No global configuration found.")
				return nil
			}
			keys := make([]string, 0, len(raw))
			for k := range raw {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				v := raw[k]
				if strings.HasSuffix(strings.ToUpper(k), "_KEY") {
					if s, ok := v.(string); ok {
						v = config.MaskSecret(s)
					}
				}
				fmt.Printf("%s = %v\n", k, v)
			}
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the global configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.GlobalConfigPath()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}
}
