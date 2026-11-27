package general

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/gemini"
)

func NewAskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ask [prompt]",
		Aliases: []string{"query", "ask-api"},
		Short:   "Ask Gemini a question",
		Long: `Ask Gemini a question using the configured API interface.

CMDR uses the Gemini Interactions API by default. Use --model for a one-time
model override, "ask models" to discover available compatible models, or
"ask config --model MODEL" to change the saved default. API keys are never shown.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, _ := cmd.Flags().GetString("query")
			if strings.TrimSpace(query) == "" {
				query = strings.Join(args, " ")
			}
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("provide a query with --query/-q or as positional text")
			}

			cfg, warnings, err := gemini.LoadConfig()
			if err != nil {
				return err
			}
			for _, warning := range warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", warning)
			}

			modelOverride, _ := cmd.Flags().GetString("model")
			model := cfg.Model
			if strings.TrimSpace(modelOverride) != "" {
				model = strings.TrimSpace(modelOverride)
			}

			files, _ := cmd.Flags().GetStringArray("file")
			for _, file := range files {
				fileContent, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("read file %s: %w", file, err)
				}
				query += "\n" + string(fileContent)
			}

			text, err := gemini.NewClient().Generate(context.Background(), cfg, model, query)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), text)
			return nil
		},
	}

	cmd.Flags().StringP("model", "m", "", "Use a Gemini model for this request only")
	cmd.Flags().StringP("query", "q", "", "The query to send to Gemini")
	cmd.Flags().StringArrayP("file", "f", nil, "Add a file to the query")

	modelsCmd := &cobra.Command{
		Use:   "models",
		Short: "List Gemini models compatible with the configured API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			refresh, _ := cmd.Flags().GetBool("refresh")
			cfg, warnings, err := gemini.LoadConfig()
			if err != nil {
				return err
			}
			for _, warning := range warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", warning)
			}
			models, err := gemini.NewClient().ListModels(context.Background(), cfg, refresh)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-42s %-28s %-12s %s\n", "MODEL ID", "DISPLAY NAME", "TOKENS", "CAPABILITIES")
			for _, model := range models {
				name := model.DisplayName
				if strings.TrimSpace(name) == "" {
					name = "-"
				}
				tokens := "-"
				if model.InputTokenLimit > 0 || model.OutputTokenLimit > 0 {
					tokens = fmt.Sprintf("%d/%d", model.InputTokenLimit, model.OutputTokenLimit)
				}
				capabilities := strings.Join(model.SupportedGenerationMethods, ", ")
				if capabilities == "" {
					capabilities = "text generation"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-42s %-28s %-12s %s\n", model.ID(), name, tokens, capabilities)
			}
			return nil
		},
	}
	modelsCmd.Flags().Bool("refresh", false, "Refresh model data instead of using the cache")

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect or change Gemini configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, warnings, err := gemini.LoadConfig()
			if err != nil {
				return err
			}
			for _, warning := range warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", warning)
			}

			model, _ := cmd.Flags().GetString("model")
			if strings.TrimSpace(model) != "" {
				cfg.Model = strings.TrimPrefix(strings.TrimSpace(model), "models/")
				if err := gemini.SaveConfig(cfg); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Default Gemini model changed to: %s\n", cfg.Model)
				return nil
			}

			keyStatus := "not configured"
			if strings.TrimSpace(cfg.APIKey) != "" && cfg.APIKey != "your_gemini_api_key_here" {
				keyStatus = "configured"
			} else if strings.TrimSpace(os.Getenv("GEMINI_API_KEY")) != "" {
				keyStatus = "configured through GEMINI_API_KEY"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Provider: Gemini\nBase URL: %s\nAPI: %s\nModel: %s\nAPI key: %s\n", cfg.BaseURL, cfg.API, cfg.Model, keyStatus)
			return nil
		},
	}
	configCmd.Flags().StringP("model", "m", "", "Set the persistent default Gemini model")

	cmd.AddCommand(modelsCmd, configCmd)
	return cmd
}
