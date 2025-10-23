# CLAI - AI Coding Assistant Instructions

## Project Overview

CLAI is a CLI tool that downloads and runs local LLM models (via llamafile) to interpret natural language commands and generate shell commands. Built with Go, Cobra CLI framework, and Bubble Tea TUI components.

## Architecture

### Core Components

- **`cmd/root.go`**: Cobra CLI entry point - currently hardcoded to run a single example query
- **`model/`**: LLM integration layer
  - `model.go`: Executes llamafile with prompts, expects JSON responses with cmd/args/description/risk
  - `manifest.go`: Asset management - downloads llamafile binary and Gemma 3 1B model to platform-specific app data dirs
- **`components/download.go`**: Bubble Tea TUI for download progress with real-time updates via channels

### Data Flow

1. User runs `clai` → `cmd/root.go` executes
2. `Model.EnsureAssets()` downloads missing files (llamafile + model) to `~/.local/share/clai/` (Linux) or `~/Library/Application Support/Clai/` (macOS)
3. `Model.Ask()` spawns llamafile subprocess with structured prompt, parses JSON response
4. Downloads use `components.Download()` with Bubble Tea progress bars

## Development Workflow

### Running the Application

```bash
go run main.go  # Runs with hardcoded example in cmd/root.go
./clai          # After building
```

### Building

```bash
go build -o clai  # Creates binary in project root
```

### Asset Storage Locations

- **Linux**: `~/.local/share/clai/{bin,models}/`
- **macOS**: `~/Library/Application Support/Clai/{bin,models}/`
- **Downloads**: llamafile-0.9.3 binary + gemma-3-1b-it-q6.llamafile model

## Project-Specific Patterns

### Asset Management Pattern

Assets use a struct-based approach in `model/manifest.go`:

```go
Asset{URL, Filename, Executable, BaseFolder}
```

The `Ensure()` method checks existence, downloads if missing, and sets executable permissions. Platform-specific paths via `AppDataDir()`.

### LLM Invocation

- **Prompt format**: JSON schema instruction + task → expects `{"cmd":string,"args":[string],"explain":string,"risk":"low"|"med"|"high"}`
- **Timeout**: 30 seconds via `context.WithTimeout`
- **Temperature**: 0.2 for consistent responses
- **Max tokens**: 128 (`--n-predict`)

### Bubble Tea TUI Pattern

`components/download.go` demonstrates channel-based progress updates:

- Background goroutine writes to `progressChan`
- `waitForProgress()` command pulls from channel
- `Update()` processes messages and re-queues `waitForProgress` until complete

## Dependencies

- **Cobra**: CLI framework (no subcommands implemented yet, only root command)
- **Bubble Tea**: TUI framework for interactive progress displays
- **Lipgloss**: Styling for TUI components

## Known Issues & TODOs

- Hardcoded llamafile path override in `model.go:50`: `/home/samanar/.local/share/clai/bin/llamafile`
- No actual CLI flags implemented (toggle flag defined but unused)
- Copyright placeholders not filled in
- No error handling for failed JSON parsing from LLM
- No tests present

## Adding New Features

- **New commands**: Add to `cmd/` as Cobra subcommands, call `Execute()` in root
- **New asset types**: Extend `Manifest` struct in `model/manifest.go`
- **Custom prompts**: Modify template in `Model.Ask()` - ensure JSON output schema is preserved
