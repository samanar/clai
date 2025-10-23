package components

// Example usage of the download component:
//
// To use this component in your application:
//
// import (
//     "github.com/samanar/clai/components"
//     tea "github.com/charmbracelet/bubbletea"
// )
//
// func main() {
//     url := "https://example.com/file.zip"
//     destPath := "/path/to/save/file.zip"
//
//     model := components.NewDownloadModel(url, destPath)
//     p := tea.NewProgram(model)
//
//     if _, err := p.Run(); err != nil {
//         fmt.Printf("Error: %v\n", err)
//         os.Exit(1)
//     }
// }
//
// Or use the convenience function:
//
// func main() {
//     err := components.Download("https://example.com/file.zip", "/path/to/save/file.zip")
//     if err != nil {
//         fmt.Printf("Download failed: %v\n", err)
//         os.Exit(1)
//     }
// }
