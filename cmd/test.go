package cmd

import (
	"fmt"

	"github.com/samanar/clai/man"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		man, err := man.NewMan()
		if err != nil {
			panic(err)
		}
		fmt.Println(man.RootPaths)
		fmt.Println("-----------------------------------")
		fmt.Println(len(man.ManFiles))
		fmt.Println(man.ManFiles[0])
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
