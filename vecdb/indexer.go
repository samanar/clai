package vecdb

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samanar/clai/components"
	"github.com/samanar/clai/man"
)

// ManPage represents a parsed man page
type ManPage struct {
	Name        string
	Section     int
	Description string
	Content     string
	FilePath    string
}

// IndexManPages indexes all man pages from the Man model
func (v *VecDB) IndexManPages(m *man.Man) error {
	// Check if already indexed
	indexed, err := v.IsIndexed()
	if err != nil {
		return err
	}
	if indexed {
		return nil // Already indexed
	}

	if len(m.ManFiles) == 0 {
		return fmt.Errorf("no man files found")
	}

	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO man_pages (name, section, description, content, file_path)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	doneChan := make(chan error, 1)
	cancelChan := make(chan struct{})

	// Start spinner in goroutine
	go func() {
		components.ShowSpinner("Indexing man pages...", doneChan, cancelChan)
	}()

	// Monitor for cancellation
	go func() {
		<-cancelChan
		os.Exit(1)
	}()

	// Run the command

	indexed_count := 0
	for _, filePath := range m.ManFiles {
		manPage, err := parseManPage(filePath)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		_, err = stmt.Exec(
			manPage.Name,
			manPage.Section,
			manPage.Description,
			manPage.Content,
			manPage.FilePath,
		)
		if err == nil {
			indexed_count++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Signal spinner to stop
	doneChan <- err
	close(doneChan)

	fmt.Printf("Indexed %d man pages\n", indexed_count)
	return nil
}

// parseManPage extracts information from a man page file
// Handles both .gz and plain files using the man.ReadManFileInChunks
func parseManPage(filePath string) (*ManPage, error) {
	// Read entire file content using man package
	var contentBuf bytes.Buffer

	for chunk := range man.ReadManFileInChunks(filePath, 4096) {
		if chunk.Err != nil {
			return nil, fmt.Errorf("error reading %s: %w", filePath, chunk.Err)
		}
		contentBuf.Write(chunk.Data)
	}

	rawContent := contentBuf.Bytes()

	// Convert groff/troff to plain text using col -b
	content, err := convertToPlainText(rawContent)
	if err != nil {
		// If conversion fails, use raw content
		content = string(rawContent)
	}

	name := extractManName(filePath)
	section := extractSectionFromPath(filePath)
	description := extractDescription(content)

	return &ManPage{
		Name:        name,
		Section:     section,
		Description: description,
		Content:     content,
		FilePath:    filePath,
	}, nil
}

// convertToPlainText converts groff/troff format to plain text
func convertToPlainText(content []byte) (string, error) {
	cmd := exec.Command("col", "-b")
	cmd.Stdin = bytes.NewReader(content)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// extractManName extracts the command name from file path
func extractManName(filePath string) string {
	base := filepath.Base(filePath)
	// Remove .gz extension if present
	base = strings.TrimSuffix(base, ".gz")
	// Remove section number (e.g., .1, .8, .3p)
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}
	return base
}

// extractSectionFromPath extracts section number from file path
func extractSectionFromPath(filePath string) int {
	// Use the man package's extractSection function
	sectionStr := extractSectionString(filePath)
	if sectionStr == "" {
		return 0
	}

	// Parse just the numeric part (e.g., "3p" -> 3)
	if len(sectionStr) > 0 && sectionStr[0] >= '0' && sectionStr[0] <= '9' {
		return int(sectionStr[0] - '0')
	}
	return 0
}

// extractSectionString extracts section string from file path
func extractSectionString(path string) string {
	base := filepath.Base(path)
	// Try .N.gz first
	if strings.HasSuffix(base, ".gz") {
		base = strings.TrimSuffix(base, ".gz")
	}
	// Extract .N pattern
	if idx := strings.LastIndex(base, "."); idx > 0 {
		return base[idx+1:]
	}
	return ""
}

// extractDescription extracts the NAME section description
func extractDescription(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inNameSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "NAME") {
			inNameSection = true
			continue
		}

		if inNameSection {
			// Stop at next section
			if strings.HasPrefix(line, "SYNOPSIS") ||
				strings.HasPrefix(line, "DESCRIPTION") {
				break
			}

			// Extract description (usually format: "command - description")
			if strings.Contains(line, "-") {
				parts := strings.SplitN(line, "-", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}

			// Return any non-empty line in NAME section
			if line != "" {
				return line
			}
		}
	}

	return ""
}
