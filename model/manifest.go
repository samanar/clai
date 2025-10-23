package model

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/samanar/clai/components"
)

type Asset struct {
	URL        string
	Filename   string
	Executable bool
	BaseFolder string
}

type Manifest struct {
	Llama Asset
	Model Asset
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

func NewManifest() Manifest {
	var llama Asset
	llama = Asset{
		URL:        "https://github.com/Mozilla-Ocho/llamafile/releases/download/0.9.3/llamafile-0.9.3",
		Filename:   "llamafile",
		Executable: true,
		BaseFolder: "bin",
	}

	model := Asset{
		URL:        "https://huggingface.co/Mozilla/gemma-3-1b-it-llamafile/resolve/main/google_gemma-3-1b-it-Q6_K.llamafile?download=true",
		Filename:   "gemma-3-1b-it-q6.llamafile",
		Executable: false,
		BaseFolder: "models",
	}

	return Manifest{Llama: llama, Model: model}
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
