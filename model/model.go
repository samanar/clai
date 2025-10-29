package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/samanar/clai/components"
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

func (m *Model) GetEmbeddingAsset() Asset {
	return m.manifest.Embedding
}

func (m *Model) EnsureAssets() error {
	if err := m.GetLlamaAsset().Ensure(); err != nil {
		return err
	}
	if err := m.GetModelAsset().Ensure(); err != nil {
		return err
	}

	if err := m.GetEmbeddingAsset().Ensure(); err != nil {
		return err
	}
	return nil
}

// GenerateEmbedding generates an embedding vector for the given text using the embedding model
func (m *Model) GenerateEmbedding(text string) ([]float32, error) {
	// Truncate text if too long
	if len(text) > 8000 {
		text = text[:8000]
	}

	// Clean text
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	llamaPath, err := m.GetLlamaAsset().FullPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get llamafile path: %w", err)
	}

	embeddingPath, err := m.GetEmbeddingAsset().FullPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding model path: %w", err)
	}

	args := []string{
		llamaPath,
		"-m", embeddingPath,
		"--embedding",
		"--log-disable",
		"-p", text,
		"--n-predict", "0",
		"--threads", "4",
		"-ngl", "999",
		"--ctx-size", "8192",
		"--batch-size", "8192",
	}

	cmd := exec.Command("/bin/bash", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("embedding generation failed: %v\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Parse space-separated float values
	floatStrings := strings.Fields(output)
	if len(floatStrings) == 0 {
		return nil, fmt.Errorf("no embedding values found in output")
	}

	var embedding []float32
	for _, floatStr := range floatStrings {
		val, err := strconv.ParseFloat(floatStr, 32)
		if err != nil {
			continue // Skip invalid values
		}
		embedding = append(embedding, float32(val))
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("no valid embedding values parsed")
	}

	return embedding, nil
}

// GenerateQueryEmbedding generates an embedding optimized for user queries
func (m *Model) GenerateQueryEmbedding(query string) ([]float32, error) {
	return m.GenerateEmbedding(query)
}

func (m *Model) Ask(userInput string) ([]Result, error) {
	// Get relevant man pages using vector search
	manReference, err := m.buildManReferenceWithVectorSearch(userInput)
	fmt.Println(manReference)
	if err != nil {
		// Fall back to keyword-based search if vector search fails
		// manReference = buildManReference(userInput)
	}

	prompt := buildPrompt(userInput, manReference)
	fmt.Println("--------------------")
	fmt.Println(prompt)
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
		"--threads", "4", // Limit CPU threads
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "/bin/bash", llamaArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create spinner with channel
	doneChan := make(chan error, 1)
	cancelChan := make(chan struct{})

	// Start spinner in goroutine
	go func() {
		components.ShowSpinner("Generating commands...", doneChan, cancelChan)
	}()

	// Monitor for cancellation
	go func() {
		<-cancelChan
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cancel() // Cancel the context
	}()

	// Run the command
	err = cmd.Run()

	// Signal spinner to stop
	doneChan <- err
	close(doneChan)

	if err != nil {
		return nil, fmt.Errorf("llamafile failed: %v\nstderr: %s", err, stderr.String())
	}

	raw := strings.TrimSpace(stdout.String())

	// Parse the JSON response into an array of Result objects
	var results []Result
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v\nraw: %s", err, raw)
	}

	return results, nil
}

func buildPrompt(userInput, manReference string) string {
	var builder strings.Builder
	builder.WriteString("Generate shell commands as JSON array.\n\n")
	builder.WriteString(fmt.Sprintf("Task: %s\n\n", userInput))

	if manReference != "" {
		builder.WriteString("Reference material from relevant man pages:\n")
		builder.WriteString("[MANPAGE EXCERPT]\n")
		builder.WriteString(manReference)
		builder.WriteString("\n[/MANPAGE EXCERPT]\n\n")
	}

	builder.WriteString("Rules:\n")
	builder.WriteString("- Return 1-4 real Linux commands only\n")
	builder.WriteString("- Use actual commands\n")
	builder.WriteString("- Most common solution first\n")
	builder.WriteString("- Args as separate array elements\n\n")
	builder.WriteString("JSON format:\n")
	builder.WriteString(`[{"cmd":"command","args":["arg1","arg2"],"explain":"description"}]`)
	builder.WriteString("\n\nOutput:")

	return builder.String()
}
