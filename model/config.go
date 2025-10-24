package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samanar/clai/components"
	"gopkg.in/yaml.v3"
)

const CONFIG_FILE_NAME = "config.yml"
const CONFIG_FILE_BASE_FOLDER = "config"

type Config struct {
	Model ModelType `yaml:"model"`
}

func NewConfig() (Config, error) {
	cfg := Config{}
	err := cfg.Load()
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (cfg *Config) BasePath() (string, error) {
	appDataDir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, CONFIG_FILE_BASE_FOLDER), nil
}

func (cfg *Config) FullPath() (string, error) {
	base, err := cfg.BasePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, CONFIG_FILE_NAME), nil
}

func (cfg *Config) Ensure() error {
	configPath, err := cfg.FullPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), os.ModePerm); err != nil {
			return err
		}
		cfg.Create()
	}
	return nil
}

func (cfg *Config) Create() error {
	cfgPath, err := cfg.FullPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return nil // Config already exists
	}

	claiConfig := Config{}

	// Ask user inputs with defaults
	claiConfig.Model = ModelGemma3_1B

	// Save YAML file
	data, err := yaml.Marshal(&claiConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (cfg *Config) Load() error {
	err := cfg.Ensure()
	if err != nil {
		return err
	}

	configPath, err := cfg.FullPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var claiConfig Config
	if err := yaml.Unmarshal(data, &claiConfig); err != nil {
		return err
	}
	*cfg = claiConfig
	return nil
}

func (cfg *Config) Save() error {
	configPath, err := cfg.FullPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) UpdatePrompt() error {
	options := []components.SelectOption{}
	for _, model := range AllModels {
		options = append(options, components.SelectOption{
			Title:       model.Filename,
			Description: fmt.Sprintf("%s (Size: %s)", model.Description, model.DownloadSize),
			Value:       model.Filename,
		})
	}
	selected, err := components.Select(options)
	if err != nil {
		return err
	}
	cfg.Model = ToModelType(selected)
	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}
