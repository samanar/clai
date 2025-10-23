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
}

func NewModel() Model {
	return Model{
		manifest: NewManifest(),
	}
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
- Return 1-3 real Linux commands only
- Use actual commands: ls, find, tar, zip, sha256sum, docker, etc.
- Most common solution first
- Args as separate array elements

JSON format:
[{"cmd":"command","args":["arg1","arg2"],"explain":"description"}]

Output:`, userInput)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	llamaFilePath, err := m.GetLlamaAsset().FullPath()
	if err != nil {
		panic(err)
	}
	modelPath, err := m.GetModelAsset().FullPath()
	if err != nil {
		panic(err)
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
		"--grammar-file", tmp.Name(),
		"-p", prompt,
		"--temp", "0.3", // Lower temperature for more consistent results
		"--n-predict", "800", // Enough for multiple commands
		"--ctx-size", "4096",
		"--repeat-penalty", "1.1",
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
