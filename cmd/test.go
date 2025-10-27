package cmd

import (
	"github.com/samanar/clai/man"
	"github.com/samanar/clai/vecdb"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// Create Man instance (discovers all man files)
		m, err := man.NewMan()
		if err != nil {
			panic(err)
		}

		// Create/open vector database
		db, err := vecdb.NewVecDB()
		if err != nil {
			panic(err)
		}
		defer db.Close()

		// Index all man pages (only runs once)
		err = db.IndexManPages(&m)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
