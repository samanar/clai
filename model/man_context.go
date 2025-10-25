package model

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

const (
	maxReferenceCommands   = 2
	maxReferenceCharacters = 2000
	manCommandTimeout      = 5 * time.Second
)

func buildManReference(userInput string) string {
	ctx, cancel := context.WithTimeout(context.Background(), manCommandTimeout)
	defer cancel()

	keywords := extractKeywords(userInput)
	if len(keywords) == 0 {
		return ""
	}

	commandCandidates := selectCommandCandidates(ctx, keywords)
	if len(commandCandidates) == 0 {
		return ""
	}

	var snippets []string
	totalChars := 0
	for _, cmdName := range commandCandidates {
		excerpt, err := fetchManExcerpt(ctx, cmdName)
		if err != nil || excerpt == "" {
			continue
		}
		section := fmt.Sprintf("COMMAND: %s\n%s", cmdName, excerpt)
		snippets = append(snippets, section)
		totalChars += len(section)
		if totalChars >= maxReferenceCharacters {
			break
		}
	}

	return strings.TrimSpace(strings.Join(snippets, "\n\n"))
}

func extractKeywords(input string) []string {
	fields := strings.FieldsFunc(strings.ToLower(input), func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		switch r {
		case '-', '_':
			return false
		default:
			return true
		}
	})
	seen := make(map[string]struct{})
	var keywords []string
	for _, field := range fields {
		if len(field) == 0 {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		keywords = append(keywords, field)
	}
	return keywords
}

func selectCommandCandidates(ctx context.Context, keywords []string) []string {
	var candidates []string
	seen := make(map[string]struct{})

	for _, token := range keywords {
		if len(candidates) >= maxReferenceCommands {
			break
		}
		if !isLikelyCommand(token) {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		if hasManPage(ctx, token) {
			candidates = append(candidates, token)
			seen[token] = struct{}{}
		}
	}

	if len(candidates) >= maxReferenceCommands {
		return candidates
	}

	for _, token := range keywords {
		if len(candidates) >= maxReferenceCommands {
			break
		}
		for _, match := range searchApropos(ctx, token, maxReferenceCommands-len(candidates)) {
			if _, ok := seen[match]; ok {
				continue
			}
			candidates = append(candidates, match)
			seen[match] = struct{}{}
			if len(candidates) >= maxReferenceCommands {
				break
			}
		}
	}

	return candidates
}

func isLikelyCommand(token string) bool {
	if len(token) == 0 || len(token) > 32 {
		return false
	}
	for _, r := range token {
		if unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func hasManPage(ctx context.Context, topic string) bool {
	cmd := exec.CommandContext(ctx, "man", "-w", topic)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func fetchManExcerpt(ctx context.Context, command string) (string, error) {
	shellCmd := fmt.Sprintf("LANG=C man %s | col -bx", shellEscape(command))
	cmd := exec.CommandContext(ctx, "/bin/bash", "-lc", shellCmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return sliceManSections(stdout.String()), nil
}

func searchApropos(ctx context.Context, keyword string, limit int) []string {
	if limit <= 0 || len(keyword) == 0 {
		return nil
	}
	shellCmd := fmt.Sprintf("LANG=C man -k %s", shellEscape(keyword))
	cmd := exec.CommandContext(ctx, "/bin/bash", "-lc", shellCmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil
	}

	lines := strings.Split(stdout.String(), "\n")
	var matches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.Contains(line, "nothing appropriate") {
			continue
		}
		name := parseAproposLine(line)
		if name == "" {
			continue
		}
		matches = append(matches, name)
		if len(matches) >= limit {
			break
		}
	}
	return matches
}

func parseAproposLine(line string) string {
	idx := strings.Index(line, " (")
	if idx <= 0 {
		return ""
	}
	return strings.TrimSpace(line[:idx])
}

func sliceManSections(manText string) string {
	lines := strings.Split(manText, "\n")
	if len(lines) == 0 {
		return ""
	}

	allowedSections := map[string]struct{}{
		"NAME":        {},
		"SYNOPSIS":    {},
		"DESCRIPTION": {},
		"OVERVIEW":    {},
		"OPTIONS":     {},
		"EXAMPLES":    {},
	}

	var builder strings.Builder
	sectionsCaptured := 0
	currentSection := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if currentSection != "" {
				builder.WriteString("\n")
			}
			continue
		}

		if isSectionHeader(trimmed) {
			if _, ok := allowedSections[trimmed]; ok {
				if builder.Len() > 0 {
					builder.WriteString("\n")
				}
				builder.WriteString(trimmed)
				builder.WriteString("\n")
				currentSection = trimmed
				sectionsCaptured++
				continue
			}
			if sectionsCaptured > 0 {
				break
			}
			continue
		}

		if currentSection != "" {
			builder.WriteString(line)
			builder.WriteString("\n")
			if builder.Len() >= maxReferenceCharacters {
				break
			}
		}
	}

	if builder.Len() == 0 {
		fallback := fallbackExcerpt(lines)
		builder.WriteString(fallback)
	}

	return strings.TrimSpace(builder.String())
}

func fallbackExcerpt(lines []string) string {
	var builder strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		builder.WriteString(line)
		builder.WriteString("\n")
		if builder.Len() >= maxReferenceCharacters {
			break
		}
	}
	return strings.TrimSpace(builder.String())
}

func isSectionHeader(line string) bool {
	if len(line) == 0 || len(line) > 40 {
		return false
	}
	for _, r := range line {
		switch {
		case r == ' ':
			continue
		case r == '-':
			continue
		case unicode.IsDigit(r):
			continue
		case unicode.IsLetter(r):
			if !unicode.IsUpper(r) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func shellEscape(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
