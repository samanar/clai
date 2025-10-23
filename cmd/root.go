/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path"

	"github.com/samanar/clai/components"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clai",
	Short: "A brief description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		// manifest := bootstrap.DefaultManifest()
		components.Download("https://github.com/samanar/gitkit/releases/download/v1.1.0/gitkit-linux-amd64", path.Join(".", "gitkit"), true)
		// components.Download(manifest.Model.URL, path.Join(".", manifest.Model.Filename))
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
