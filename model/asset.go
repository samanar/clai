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
	Llama     Asset
	Model     Asset
	Embedding Asset
}

type ModelType string
type EmbeddingType string

const (
	ModelGemma3_1B ModelType = "gemma-3-1b-it-q6.llamafile"
	ModelLlama3_2B ModelType = "llama-3.2-3b-it-q6.llamafile"
	ModelGemma3_4B ModelType = "gemma-3-4b-it-q6.llamafile"
)

const (
	EmbeddingQwen3  EmbeddingType = "qwen3-embedding-0.6B-q8.gguf"
	EmbeddingGemma3 EmbeddingType = "gemma3-embedding-0.3B-BF16.gguf"
)

func (mt ModelType) String() string {
	return string(mt)
}

func (em EmbeddingType) String() string {
	return string(em)
}

func ToModelType(stringVal string) ModelType {
	switch stringVal {
	case ModelGemma3_1B.String():
		return ModelGemma3_1B
	case ModelLlama3_2B.String():
		return ModelLlama3_2B
	case ModelGemma3_4B.String():
		return ModelGemma3_4B
	default:
		return ModelGemma3_1B
	}
}

func ToEmbeddingType(stringVal string) EmbeddingType {
	switch stringVal {
	case EmbeddingQwen3.String():
		return EmbeddingQwen3
	case EmbeddingGemma3.String():
		return EmbeddingGemma3
	default:
		return EmbeddingQwen3
	}
}

var AllModels = []Asset{
	{
		URL:          "https://huggingface.co/Mozilla/gemma-3-1b-it-llamafile/resolve/main/google_gemma-3-1b-it-Q6_K.llamafile?download=true",
		Filename:     ModelGemma3_1B.String(),
		DownloadSize: "1.32 GB",
		Description:  "Gemma3 1B. low resource usage. low accuracy.",
	},
	{
		URL:          "https://huggingface.co/Mozilla/Llama-3.2-3B-Instruct-llamafile/resolve/main/Llama-3.2-3B-Instruct.Q6_K.llamafile?download=true",
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

var AllEmbeddingModels = []Asset{
	{
		URL:          "https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF/resolve/main/Qwen3-Embedding-0.6B-Q8_0.gguf",
		Filename:     EmbeddingQwen3.String(),
		DownloadSize: "639 MB",
		Description:  "Qwen3 Embedding 0.6B. high quality embeddings for various tasks.",
	},
	{
		URL:          "https://huggingface.co/unsloth/embeddinggemma-300m-GGUF/resolve/main/embeddinggemma-300M-BF16.gguf",
		Filename:     EmbeddingGemma3.String(),
		DownloadSize: "612 MB",
		Description:  "Gemma3 Embedding 0.3B. high quality embeddings for various tasks.",
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

func GetEmbeddingModel(embeddingType EmbeddingType) Asset {
	for _, model := range AllEmbeddingModels {
		if model.Filename == embeddingType.String() {
			return Asset{
				URL:          model.URL,
				Filename:     model.Filename,
				Description:  model.Description,
				DownloadSize: model.DownloadSize,
				Executable:   false,
				BaseFolder:   "embeddings",
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
	embedding := GetEmbeddingModel(config.Embedding)

	return Manifest{Llama: llama, Model: model, Embedding: embedding}, nil
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
