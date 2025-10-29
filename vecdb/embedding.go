package vecdb

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
)

// ExtractKeyContent extracts relevant parts of man page for embedding
func ExtractKeyContent(content string, maxChars int) string {
	lines := strings.Split(content, "\n")
	var keyLines []string
	inRelevantSection := false
	currentChars := 0

	sections := []string{"NAME", "SYNOPSIS", "DESCRIPTION", "OPTIONS"}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		isSection := false
		for _, sec := range sections {
			if strings.HasPrefix(trimmed, sec) {
				inRelevantSection = true
				isSection = true
				keyLines = append(keyLines, line)
				currentChars += len(line)
				break
			}
		}

		if isSection {
			continue
		}

		if strings.HasPrefix(trimmed, "SEE ALSO") ||
			strings.HasPrefix(trimmed, "AUTHOR") {
			break
		}

		if inRelevantSection && trimmed != "" {
			if currentChars+len(line) > maxChars {
				break
			}
			keyLines = append(keyLines, line)
			currentChars += len(line)
		}
	}

	return strings.Join(keyLines, "\n")
}

// Float32SliceToBytes converts float32 slice to bytes for storage
func Float32SliceToBytes(floats []float32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, floats)
	return buf.Bytes()
}

// BytesToFloat32Slice converts bytes to float32 slice
func BytesToFloat32Slice(data []byte) []float32 {
	floatCount := len(data) / 4
	if floatCount == 0 {
		return nil
	}

	floats := make([]float32, floatCount)
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &floats)
	return floats
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// NormalizeVector normalizes vector to unit length
func NormalizeVector(v []float32) []float32 {
	var sum float64
	for _, val := range v {
		sum += float64(val) * float64(val)
	}

	if sum == 0 {
		return v
	}

	norm := math.Sqrt(sum)
	normalized := make([]float32, len(v))
	for i, val := range v {
		normalized[i] = float32(float64(val) / norm)
	}
	return normalized
}
