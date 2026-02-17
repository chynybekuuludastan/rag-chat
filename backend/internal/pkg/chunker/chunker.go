package chunker

import (
	"strings"
	"unicode/utf8"
)

type Config struct {
	ChunkSize int
	Overlap   int
}

func DefaultConfig() Config {
	return Config{
		ChunkSize: 800,
		Overlap:   100,
	}
}

func Chunk(text string, cfg Config) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if utf8.RuneCountInString(text) <= cfg.ChunkSize {
		return []string{text}
	}

	paragraphs := splitParagraphs(text)
	segments := splitLargeParagraphs(paragraphs, cfg.ChunkSize)
	chunks := mergeIntoChunks(segments, cfg.ChunkSize)
	chunks = addOverlap(chunks, cfg.Overlap)

	return chunks
}

func splitParagraphs(text string) []string {
	raw := strings.Split(text, "\n\n")
	var result []string
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitLargeParagraphs(paragraphs []string, maxSize int) []string {
	var result []string
	for _, p := range paragraphs {
		if utf8.RuneCountInString(p) <= maxSize {
			result = append(result, p)
			continue
		}
		sentences := splitSentences(p)
		result = append(result, sentences...)
	}
	return result
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])

		if isSentenceEnd(runes, i) {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		sentences = append(sentences, s)
	}

	return sentences
}

func isSentenceEnd(runes []rune, i int) bool {
	r := runes[i]
	if r != '.' && r != '!' && r != '?' {
		return false
	}
	if i+1 < len(runes) && runes[i+1] == ' ' {
		return true
	}
	if i+1 == len(runes) {
		return true
	}
	return false
}

func mergeIntoChunks(segments []string, maxSize int) []string {
	var chunks []string
	var current strings.Builder

	for _, seg := range segments {
		segLen := utf8.RuneCountInString(seg)
		currentLen := utf8.RuneCountInString(current.String())

		if currentLen > 0 && currentLen+1+segLen > maxSize {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}

		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(seg)
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		chunks = append(chunks, s)
	}

	return chunks
}

func addOverlap(chunks []string, overlap int) []string {
	if len(chunks) <= 1 || overlap <= 0 {
		return chunks
	}

	result := make([]string, len(chunks))
	result[0] = chunks[0]

	for i := 1; i < len(chunks); i++ {
		prevRunes := []rune(chunks[i-1])
		overlapSize := overlap
		if overlapSize > len(prevRunes) {
			overlapSize = len(prevRunes)
		}

		overlapText := string(prevRunes[len(prevRunes)-overlapSize:])
		result[i] = strings.TrimSpace(overlapText + " " + chunks[i])
	}

	return result
}
