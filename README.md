# CLAI - Command Line AI Assistant

A CLI tool that converts natural language into shell commands using **fully offline** local LLM models via llamafile. No API keys, no internet required after initial setup, complete privacy.

## Features

- 🔒 **100% Offline** - All AI processing happens locally on your machine
- 🎯 **Natural Language to Shell Commands** - Describe what you want, get executable commands
- 🔄 **Multiple Model Options** - Choose between different models based on your needs:
  - **Gemma 3 1B** (1.32 GB) - Fast, low resource usage, good for simple commands
  - **Llama 3.2 3B** (2.62 GB) - Balanced performance and accuracy
  - **Gemma 3 4B** (3.50 GB) - Best accuracy, requires more resources
- 🚀 **Zero Configuration** - Downloads and configures everything automatically on first run
- 🔐 **Privacy First** - Your commands never leave your machine

## Installation

### Prerequisites

- Go 1.25+ installed
- Linux or macOS
- Sufficient disk space for the model you choose (1.3 - 3.5 GB)

### Building from Source

```bash
git clone https://github.com/samanar/clai.git
cd clai
go build -o clai
```

### Moving to PATH (optional)

```bash
sudo mv clai /usr/local/bin/
```

## Usage

### First Run

On first run, CLAI will automatically:

1. Download the llamafile runtime (~293 MB)
2. Download your selected model (default: Gemma 3 1B)
3. Create a config file at `~/.local/share/clai/config/config.yml` (Linux) or `~/Library/Application Support/Clai/config/config.yml` (macOS)

```bash
clai "list all files in current directory"
```

### Switching Models

To change which model you're using:

```bash
clai config
```

This will show an interactive menu to select from available models. The choice is saved and persists across sessions.

### Examples

```bash
# File operations
clai "find all python files modified in the last 7 days"
clai "compress the logs folder into a tar.gz archive"
clai "show the size of each subdirectory"

# System information
clai "show me disk usage"
clai "list all running docker containers"
clai "find processes using port 8080"

# Git operations
clai "show uncommitted changes"
clai "create a new branch called feature-x"
```

## How It Works

1. **Input**: You provide a natural language description of what you want to do
2. **Processing**: CLAI sends your request to the local LLM model running via llamafile
3. **Output**: The model generates shell command(s) with explanations
4. **Offline**: Everything happens on your machine - no data is sent to external servers

## File Locations

### Linux

- **Binary**: `~/.local/share/clai/bin/llamafile`
- **Models**: `~/.local/share/clai/models/`
- **Config**: `~/.local/share/clai/config/config.yml`

### macOS

- **Binary**: `~/Library/Application Support/Clai/bin/llamafile`
- **Models**: `~/Library/Application Support/Clai/models/`
- **Config**: `~/Library/Application Support/Clai/config/config.yml`

## Available Models

| Model | Size | Resource Usage | Accuracy | Best For |
|-------|------|---------------|----------|----------|
| Gemma 3 1B | 1.32 GB | Low | Good | Quick tasks, limited hardware |
| Llama 3.2 3B | 2.62 GB | Moderate | Better | General purpose usage |
| Gemma 3 4B | 3.50 GB | High | Best | Complex commands, ample resources |

## Configuration

The config file (`config.yml`) is automatically created on first run:

```yaml
model: "gemma-3-1b-it-q6.llamafile"
```

You can manually edit this file or use `clai config` to change models interactively.

## Privacy & Security

- **No telemetry** - CLAI doesn't collect or send any usage data
- **No internet required** - After downloading models, works completely offline
- **Local processing** - All AI inference happens on your machine
- **Open source** - Inspect the code, build it yourself

## Technology Stack

- **Go** - Core application
- **Cobra** - CLI framework
- **Bubble Tea** - Terminal UI components
- **llamafile** - Local LLM runtime by Mozilla
- **Models** - Gemma 3 (Google) and Llama 3.2 (Meta)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- [llamafile](https://github.com/Mozilla-Ocho/llamafile) by Mozilla for making local LLM execution simple
- Google's Gemma models and Meta's Llama models for powerful open-source LLMs
