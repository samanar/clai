# CLAI - AI Coding Assistant Instructions

## Project Overview

CLAI is a CLI tool that downloads and runs local LLM models (via llamafile) to interpret natural language commands and generate shell commands. Built with Go, Cobra CLI framework, and Bubble Tea TUI components.

## Architecture

### Core Components

- **`cmd/root.go`**: Cobra CLI entry point - currently hardcoded to run a single example query
- **`model/`**: LLM integration layer
  - `model.go`: Executes llamafile with prompts, expects JSON array responses with cmd/args/explain
  - `asset.go`: Asset management - downloads llamafile binary and configurable LLM models to platform-specific app data dirs
  - `config.go`: Configuration management - YAML-based config file for model selection
- **`components/download.go`**: Bubble Tea TUI for download progress with real-time updates via channels

### Data Flow

1. User runs `clai` → `cmd/root.go` executes
2. `Model.EnsureAssets()` downloads missing files (llamafile + selected model) to `~/.local/share/clai/` (Linux) or `~/Library/Application Support/Clai/` (macOS)
3. `Model.Ask()` spawns llamafile subprocess with structured prompt + GBNF grammar, parses JSON array response
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

- **Linux**: `~/.local/share/clai/{bin,models,config}/`
- **macOS**: `~/Library/Application Support/Clai/{bin,models,config}/`
- **Config**: `config.yml` with selected model
- **Models**: Multiple options (Gemma 3 1B/4B, Llama 3.2 3B)
- **Downloads**: llamafile-0.9.3 binary + selected model file

## Project-Specific Patterns

### Asset Management Pattern

Assets use a struct-based approach in `model/asset.go`:

```go
Asset{URL, Filename, Description, DownloadSize, Executable, BaseFolder}
```

The `Ensure()` method checks existence, downloads if missing, and sets executable permissions. Platform-specific paths via `AppDataDir()`.

### Configuration Pattern

YAML-based config in `model/config.go`:

```yaml
model: "gemma-3-1b-it-q6.llamafile"
```

- Auto-creates on first run
- Supports model switching via `UpdatePrompt()`
- Stored in platform-specific config directory

### LLM Invocation

- **Prompt format**: JSON array schema instruction + task → expects `[{"cmd":string,"args":[string],"explain":string}]`
- **GBNF Grammar**: Custom grammar file enforces JSON array structure for reliable parsing
- **Timeout**: 30 seconds via `context.WithTimeout`
- **Temperature**: 0.3 for consistent responses
- **Max tokens**: 800 (`--n-predict`)
- **Context size**: 4096 tokens
- **Repeat penalty**: 1.1

### Bubble Tea TUI Pattern

`components/download.go` demonstrates channel-based progress updates:

- Background goroutine writes to `progressChan`
- `waitForProgress()` command pulls from channel
- `Update()` processes messages and re-queues `waitForProgress` until complete

## Dependencies

- **Cobra**: CLI framework (no subcommands implemented yet, only root command)
- **Bubble Tea**: TUI framework for interactive progress displays
- **Lipgloss**: Styling for TUI components
- **YAML v3**: Configuration file parsing

## Known Issues & TODOs

- Hardcoded llamafile path override removed but may need verification
- No actual CLI flags implemented (toggle flag defined but unused)
- Copyright placeholders not filled in
- No tests present
- Config creation uses hardcoded default model

## Adding New Features

- **New commands**: Add to `cmd/` as Cobra subcommands, call `Execute()` in root
- **New models**: Add to `AllModels` slice in `asset.go`, update `ModelType` constants
- **Custom prompts**: Modify template in `Model.Ask()` - ensure GBNF grammar matches JSON schema
- **Config options**: Extend `Config` struct in `config.go`, update YAML marshaling
