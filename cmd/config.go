package cmd

import (
	"fmt"

	"github.com/samanar/clai/model"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage clai configuration",
	Long: `Manage clai configuration settings including model selection and preferences.

Use subcommands to view or modify specific configuration options.`,
}

// setModelCmd represents the set-model command
var setModelCmd = &cobra.Command{
	Use:   "set-model",
	Short: "Change the LLM model used by clai",
	Long: `Interactively select which LLM model to use for generating commands.

This will display a menu with available models showing their size and accuracy trade-offs.
Your selection will be saved and used for all future clai invocations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := model.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Show current model
		fmt.Printf("Current model: %s\n\n", cfg.Model)
		fmt.Println("Select a new model:")

		// Use the interactive prompt
		if err := cfg.UpdatePrompt(); err != nil {
			return fmt.Errorf("failed to update model: %w", err)
		}

		// Reload to show the updated model
		cfg, err = model.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}

		fmt.Printf("\n✓ Model successfully changed to: %s\n", cfg.Model)
		fmt.Println("\nNote: The new model will be downloaded automatically on next use if not already present.")

		return nil
	},
}

// showConfigCmd represents the show command
var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display the current clai configuration including the selected model.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := model.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		configPath, err := cfg.FullPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		fmt.Println("Current Configuration:")
		fmt.Println("─────────────────────")
		fmt.Printf("Model: %s\n", cfg.Model)
		fmt.Printf("Config file: %s\n", configPath)

		// Find and display model details
		for _, m := range model.AllModels {
			if m.Filename == cfg.Model.String() {
				fmt.Printf("Description: %s\n", m.Description)
				fmt.Printf("Download size: %s\n", m.DownloadSize)
				break
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setModelCmd)
	configCmd.AddCommand(showConfigCmd)
}
