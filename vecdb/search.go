package vecdb

import (
	"fmt"
	"strings"
)

// SearchResult represents a search result from the database
type SearchResult struct {
	Name        string
	Section     int
	Description string
	Content     string
	Relevance   float64
}

// SearchManPages performs a full-text search across man pages
func (v *VecDB) SearchManPages(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use FTS5 for full-text search with ranking
	sql := `
		SELECT 
			m.name,
			m.section,
			m.description,
			m.content,
			fts.rank
		FROM man_pages_fts fts
		JOIN man_pages m ON m.id = fts.rowid
		WHERE man_pages_fts MATCH ?
		ORDER BY fts.rank
		LIMIT ?
	`

	rows, err := v.db.Query(sql, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Name, &r.Section, &r.Description, &r.Content, &r.Relevance); err != nil {
			continue
		}
		results = append(results, r)
	}

	return results, nil
}

// FindManPage finds a specific man page by name and optional section
func (v *VecDB) FindManPage(name string, section int) (*SearchResult, error) {
	var sql string
	var args []interface{}

	if section > 0 {
		sql = `
			SELECT name, section, description, content, 0.0 as rank
			FROM man_pages
			WHERE name = ? AND section = ?
			LIMIT 1
		`
		args = []interface{}{name, section}
	} else {
		sql = `
			SELECT name, section, description, content, 0.0 as rank
			FROM man_pages
			WHERE name = ?
			ORDER BY section
			LIMIT 1
		`
		args = []interface{}{name}
	}

	var result SearchResult
	err := v.db.QueryRow(sql, args...).Scan(
		&result.Name,
		&result.Section,
		&result.Description,
		&result.Content,
		&result.Relevance,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ExtractRelevantSections extracts the most relevant sections from man page content
func ExtractRelevantSections(content string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 2000
	}

	sections := []string{"DESCRIPTION", "SYNOPSIS", "OPTIONS", "EXAMPLES"}
	var extracted strings.Builder

	lines := strings.Split(content, "\n")
	inRelevantSection := false
	currentLength := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we hit a relevant section
		isRelevantSection := false
		for _, sec := range sections {
			if strings.HasPrefix(trimmed, sec) {
				isRelevantSection = true
				break
			}
		}

		if isRelevantSection {
			inRelevantSection = true
			extracted.WriteString("\n")
			extracted.WriteString(line)
			extracted.WriteString("\n")
			currentLength += len(line) + 2
			continue
		}

		// Stop at other major sections
		if strings.HasPrefix(trimmed, "SEE ALSO") ||
			strings.HasPrefix(trimmed, "AUTHOR") ||
			strings.HasPrefix(trimmed, "COPYRIGHT") {
			inRelevantSection = false
			continue
		}

		// Add content from relevant sections
		if inRelevantSection {
			if currentLength+len(line) > maxLength {
				extracted.WriteString("\n[truncated...]")
				break
			}
			extracted.WriteString(line)
			extracted.WriteString("\n")
			currentLength += len(line) + 1
		}
	}

	return extracted.String()
}

// SearchRelevantManPages finds man pages relevant to a user query
func (v *VecDB) SearchRelevantManPages(userQuery string, maxResults int) ([]SearchResult, error) {
	// Extract potential command names from the query
	keywords := extractKeywords(userQuery)

	if len(keywords) == 0 {
		return nil, nil
	}

	// Build FTS5 query
	ftsQuery := strings.Join(keywords, " OR ")

	return v.SearchManPages(ftsQuery, maxResults)
}

// extractKeywords extracts potential command names from user query
func extractKeywords(query string) []string {
	commonCommands := []string{
		"ls", "cd", "grep", "find", "tar", "zip", "gzip", "docker", "git",
		"cp", "mv", "rm", "cat", "less", "ssh", "scp", "rsync", "curl", "wget",
		"sed", "awk", "sort", "uniq", "head", "tail", "chmod", "chown",
	}

	words := strings.Fields(strings.ToLower(query))
	var keywords []string

	for _, word := range words {
		// Check if word is a common command
		for _, cmd := range commonCommands {
			if strings.Contains(word, cmd) {
				keywords = append(keywords, cmd)
			}
		}
	}

	return keywords
}
