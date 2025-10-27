package cmd

import (
	"fmt"
	"os"

	"github.com/samanar/clai/man"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := man.NewMan()
		if err != nil {
			panic(err)
		}
		fmt.Println(m.RootPaths)
		fmt.Println("-----------------------------------")
		fmt.Println(len(m.ManFiles))
		fmt.Println(m.ManFiles[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing man files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d man files across %d paths\n", len(m.ManFiles), len(m.RootPaths))

		const chunkSize = 64 * 1024
		for _, f := range m.ManFiles {
			for ch := range man.ReadManFileInChunks(f, chunkSize) {
				if ch.Err != nil {
					fmt.Fprintf(os.Stderr, "read error on %s: %v\n", ch.Path, ch.Err)
					break
				}
				preview := ch.Data
				if len(preview) > 80 {
					preview = preview[:80]
				}
				fmt.Printf("[%s (sec %s)] chunk %d: %q\n", ch.Path, ch.Section, ch.Index, string(preview))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
