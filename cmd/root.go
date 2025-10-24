/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/samanar/clai/model"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clai",
	Short: "A brief description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		model, err := model.NewModel()
		if err != nil {
			panic(err)
		}
		if err := model.EnsureAssets(); err != nil {
			panic(err)
		}
		userInput := "compress folder to zip very fast "
		model.Ask(userInput)

		userInput = "get sha256 check sum of a file"
		model.Ask(userInput)

		userInput = "get sha256 check sum of a file"
		model.Ask(userInput)

		userInput = "get 5 largest files in a directory"
		model.Ask(userInput)

		userInput = "delete all unused images in docker "
		model.Ask(userInput)
		// for i, result := range results {
		// 	cmd.Printf("Command %d:\n", i+1)
		// 	cmd.Printf("  Command: %s\n", result.Cmd)
		// 	cmd.Printf("  Args: %v\n", result.Args)
		// 	cmd.Printf("  Explanation: %s\n", result.Explain)
		// }
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
