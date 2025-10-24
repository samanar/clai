package model

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/samanar/clai/components"
)

type Asset struct {
	URL          string
	Filename     string
	Description  string
	DownloadSize string
	Executable   bool
	BaseFolder   string
}

type Manifest struct {
	Llama Asset
	Model Asset
}

type ModelType string

const (
	ModelGemma3_1B ModelType = "gemma-3-1b-it-q6.llamafile"
	ModelLlama3_2B ModelType = "llama-3.2-3b-it-q6.llamafile"
	ModelGemma3_4B ModelType = "gemma-3-4b-it-q6.llamafile"
)

func (mt ModelType) String() string {
	return string(mt)
}

var AllModels = []Asset{
	{
		URL:          "https://huggingface.co/Mozilla/gemma-3-1b-it-llamafile/resolve/main/google_gemma-3-1b-it-Q6_K.llamafile?download=true",
		Filename:     ModelGemma3_1B.String(),
		DownloadSize: "1.32 GB",
		Description:  "Gemma3 1B. low resource usage. low accuracy.",
	},
	{
		URL:          "https://huggingface.co/Mozilla/Llama-3.2-3B-Instruct-llamafile/resolve/main/Llama-3.2-3B-Instruct.Q6_K.llamafile",
		Filename:     ModelLlama3_2B.String(),
		DownloadSize: "2.62 GB",
		Description:  "Llama 3.2 3B. moderate resource usage. better accuracy.",
	},
	{
		URL:          "https://huggingface.co/Mozilla/gemma-3-4b-it-llamafile/resolve/main/google_gemma-3-4b-it-Q6_K.llamafile?download=true",
		Filename:     ModelGemma3_4B.String(),
		DownloadSize: "3.50 GB",
		Description:  "Gemma3 4B. high resource usage. best accuracy.",
	},
}

func GetModel(modelType ModelType) Asset {
	for _, model := range AllModels {
		if model.Filename == modelType.String() {
			return Asset{
				URL:          model.URL,
				Filename:     model.Filename,
				Description:  model.Description,
				DownloadSize: model.DownloadSize,
				Executable:   false,
				BaseFolder:   "models",
			}
		}
	}
	return Asset{}
}

func (a Asset) BasePath() (string, error) {
	appDataDir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, a.BaseFolder), nil
}

func (a Asset) FullPath() (string, error) {
	base, err := a.BasePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, a.Filename), nil
}

func (a Asset) Ensure() error {
	fullPath, err := a.FullPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			return err
		}
		// Download the file
		if err := components.Download(a.URL, fullPath); err != nil {
			return err
		}
		if a.Executable {
			if err := os.Chmod(fullPath, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewManifest() (Manifest, error) {
	config, err := NewConfig()
	if err != nil {
		return Manifest{}, err
	}
	var llama Asset
	llama = Asset{
		URL:          "https://github.com/Mozilla-Ocho/llamafile/releases/download/0.9.3/llamafile-0.9.3",
		Filename:     "llamafile",
		DownloadSize: "293 MB",
		Executable:   true,
		BaseFolder:   "bin",
	}

	model := GetModel(config.Model)

	return Manifest{Llama: llama, Model: model}, nil
}

func AppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, "Library", "Application Support", "Clai"), nil
	}
	// default to Linux behaviour
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "clai"), nil
	}
	return filepath.Join(homeDir, ".local", "share", "clai"), nil
}
