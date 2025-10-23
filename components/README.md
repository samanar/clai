# Components

This folder contains reusable UI components for the CLI application.

## Download Progress Component

A beautiful animated TUI download progress component built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

### Features

- ✨ Animated progress bar with gradient colors
- 📊 Real-time download progress updates
- 🎨 Styled with lipgloss for a polished look
- ⚡ Lightweight and efficient
- 🛡️ Error handling and directory creation
- ⌨️ Keyboard controls (q or Ctrl+C to cancel)

### Usage

#### Method 1: Using the convenience function

```go
package main

import (
    "fmt"
    "github.com/samanar/clai/components"
)

func main() {
    err := components.Download(
        "https://example.com/file.zip",
        "/path/to/save/file.zip",
    )
    if err != nil {
        fmt.Printf("Download failed: %v\n", err)
    }
}
```

#### Method 2: Using the model directly

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/samanar/clai/components"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    model := components.NewDownloadModel(
        "https://example.com/file.zip",
        "/path/to/save/file.zip",
    )
    
    p := tea.NewProgram(model)
    finalModel, err := p.Run()
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
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
```

### Example

Try it out with the included example:

```bash
go run examples/download_example.go https://github.com/charmbracelet/bubbletea/archive/refs/tags/v1.3.10.tar.gz /tmp/bubbletea.tar.gz
```

### API

#### `NewDownloadModel(url, destPath string) DownloadModel`

Creates a new download model.

**Parameters:**
- `url`: The URL to download from
- `destPath`: The local path where the file should be saved

**Returns:** A `DownloadModel` ready to be used with Bubble Tea

#### `Download(url, destPath string) error`

Convenience function that creates a model, runs the program, and returns any error.

**Parameters:**
- `url`: The URL to download from
- `destPath`: The local path where the file should be saved

**Returns:** `error` if the download fails, `nil` on success

#### `(m DownloadModel) Error() error`

Returns the error if the download failed.

### Dependencies

- [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions

### Screenshot

The component displays:
- File name being downloaded
- Animated progress bar with percentage
- Success/error messages
- Keyboard shortcut hints
