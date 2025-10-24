/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/samanar/clai/model"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clai [query]",
	Short: "Convert natural language to shell commands using local AI",
	Long: `CLAI is a CLI tool that converts natural language into shell commands 
using fully offline local LLM models. No API keys, no internet required.

Examples:
  clai list all python files
  clai compress the logs folder
  clai show disk usage

Special commands:
  clai config show         - Show current configuration
  clai config set-model    - Change the LLM model`,
	Args:                       cobra.ArbitraryArgs,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{UnknownFlags: true},
	DisableFlagParsing:         false,
	SuggestionsMinimumDistance: 2,
	DisableSuggestions:         true,
	SilenceErrors:              true,
	Run: func(cmd *cobra.Command, args []string) {
		// Join all arguments into a single string as user input
		if len(args) == 0 {
			cmd.Println("Error: Please provide a query")
			cmd.Println("Usage: clai \"your query here\"")
			os.Exit(1)
		}
		fmt.Println(args)

		userInput := strings.Join(args, " ")

		m, err := model.NewModel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing model: %v\n", err)
			os.Exit(1)
		}

		if err := m.EnsureAssets(); err != nil {
			fmt.Fprintf(os.Stderr, "Error ensuring assets: %v\n", err)
			os.Exit(1)
		}

		results, err := m.Ask(userInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing query: %v\n", err)
			os.Exit(1)
		}

		// Display results
		if len(results) == 0 {
			cmd.Println("No commands generated")
			return
		}

		cmd.Println("\nGenerated Commands:")
		cmd.Println(strings.Repeat("─", 60))

		for i, result := range results {
			cmd.Printf("\n%d. %s\n", i+1, result.Explain)

			// Build full command string
			fullCmd := result.Cmd
			if len(result.Args) > 0 {
				fullCmd += " " + strings.Join(result.Args, " ")
			}

			cmd.Printf("   $ %s\n", fullCmd)
		}

		cmd.Println()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.clai.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
