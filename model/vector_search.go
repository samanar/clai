package model

import (
	"fmt"
	"strings"

	"github.com/samanar/clai/vecdb"
)

// buildManReferenceWithVectorSearch uses vector embeddings to find relevant man pages
func (m *Model) buildManReferenceWithVectorSearch(userInput string) (string, error) {
	// Open vector database
	db, err := vecdb.NewVecDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Check if database is indexed
	count, err := db.GetManPageCount()
	if err != nil || count == 0 {
		return "", fmt.Errorf("database not indexed")
	}

	// Generate embedding for user query
	queryEmbedding, err := m.GenerateQueryEmbedding(userInput)
	if err != nil {
		// Fall back to keyword search
		return "", fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for similar man pages
	results, err := db.SearchByEmbedding(queryEmbedding, 3)
	if err != nil || len(results) == 0 {
		return "", fmt.Errorf("vector search returned no results: %w", err)
	}

	// Build reference from top results
	var builder strings.Builder
	for i, result := range results {
		if i > 0 {
			builder.WriteString("\n---\n")
		}
		builder.WriteString(fmt.Sprintf("COMMAND: %s(%d)\n", result.Name, result.Section))
		builder.WriteString(fmt.Sprintf("Description: %s\n\n", result.Description))

		// Extract relevant sections (limit to 1500 chars per page)
		relevant := extractRelevantSections(result.Content, 1500)
		builder.WriteString(relevant)
	}

	return builder.String(), nil
}

// extractRelevantSections extracts key sections from man page content
func extractRelevantSections(content string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var extracted []string
	inRelevantSection := false
	currentLength := 0

	sections := []string{"SYNOPSIS", "DESCRIPTION", "OPTIONS"}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we hit a relevant section
		for _, sec := range sections {
			if strings.HasPrefix(trimmed, sec) {
				inRelevantSection = true
				extracted = append(extracted, line)
				currentLength += len(line) + 1
				break
			}
		}

		// Stop at less relevant sections
		if strings.HasPrefix(trimmed, "SEE ALSO") ||
			strings.HasPrefix(trimmed, "AUTHOR") {
			break
		}

		// Add content from relevant sections
		if inRelevantSection {
			if currentLength+len(line) > maxLength {
				extracted = append(extracted, "[truncated...]")
				break
			}
			extracted = append(extracted, line)
			currentLength += len(line) + 1
		}
	}

	return strings.Join(extracted, "\n")
}
