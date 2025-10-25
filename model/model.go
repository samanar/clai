package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args"`
	Explain string   `json:"explain"`
}

type Model struct {
	manifest Manifest
	Config   Config
}

func NewModel() (Model, error) {
	config, err := NewConfig()
	if err != nil {
		return Model{}, err
	}
	manifest, err := NewManifest()
	if err != nil {
		return Model{}, err
	}
	return Model{
		manifest: manifest,
		Config:   config,
	}, nil
}

func (m *Model) GetModelAsset() Asset {
	return m.manifest.Model
}

func (m *Model) GetLlamaAsset() Asset {
	return m.manifest.Llama
}

func (m *Model) EnsureAssets() error {
	if err := m.GetLlamaAsset().Ensure(); err != nil {
		return err
	}
	if err := m.GetModelAsset().Ensure(); err != nil {
		return err
	}
	return nil
}

func (m *Model) Ask(userInput string) ([]Result, error) {
	// Get current working directory for context

	prompt := fmt.Sprintf(`Generate shell commands as JSON array.

Task: %s

Rules:
- Return 1-4 real Linux commands only
- Use actual commands
- Most common solution first
- Args as separate array elements

JSON format:
[{"cmd":"command","args":["arg1","arg2"],"explain":"description"}]

Output:`, userInput)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	llamaFilePath, err := m.GetLlamaAsset().FullPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get llamafile path: %v", err)
	}
	modelPath, err := m.GetModelAsset().FullPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get model path: %v", err)
	}

	// GBNF grammar for JSON array of command objects
	// More flexible grammar that allows for proper JSON structure
	gbnf := `root ::= ws "[" ws (object (ws "," ws object)*)? ws "]" ws
object ::= "{" ws "\"cmd\"" ws ":" ws string ws "," ws "\"args\"" ws ":" ws array ws "," ws "\"explain\"" ws ":" ws string ws "}"
array ::= "[" ws (string (ws "," ws string)*)? ws "]"
string ::= "\"" char* "\""
char ::= [^"\\] | "\\" (["\\/bfnrt] | "u" [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F] [0-9a-fA-F])
ws ::= [ \t\n\r]*`

	tmp, err := os.CreateTemp("", "command_*.gbnf")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(gbnf); err != nil {
		panic(err)
	}
	tmp.Close()

	llamaArgs := []string{
		llamaFilePath,
		"-m", modelPath,
		"--no-display-prompt",
		"--fast",
		"-ngl", "32", // Enable GPU layers if available
		"--mlock", // Lock model in memory
		"--grammar-file", tmp.Name(),
		"-p", prompt,
		"--temp", "0.3", // Lower temperature for more consistent results
		"--n-predict", "400", // Reduced for faster processing
		"--ctx-size", "2048", // Reduced context size
		// "--repeat-penalty", "1.1",
		"--threads", "4", // Limit CPU threads
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "/bin/bash", llamaArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("llamafile failed: %v\nstderr: %s", err, stderr.String())
	}

	raw := strings.TrimSpace(stdout.String())
	fmt.Println("user iput:\n", userInput)
	fmt.Println("LLM Response:\n", raw)

	// Parse the JSON response into an array of Result objects
	var results []Result
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v\nraw: %s", err, raw)
	}

	return results, nil
}
