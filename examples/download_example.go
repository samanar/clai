package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samanar/clai/components"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run examples/download_example.go <url> <destination-path>")
		fmt.Println("\nExample:")
		fmt.Println("  go run examples/download_example.go https://github.com/charmbracelet/bubbletea/archive/refs/tags/v1.3.10.tar.gz /tmp/bubbletea.tar.gz")
		os.Exit(1)
	}

	url := os.Args[1]
	destPath := os.Args[2]

	model := components.NewDownloadModel(url, destPath)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Check if download was successful
	if m, ok := finalModel.(components.DownloadModel); ok {
		if m.Error() != nil {
			fmt.Printf("Download failed: %v\n", m.Error())
			os.Exit(1)
		}
	}
}
