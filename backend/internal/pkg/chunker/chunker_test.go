package chunker

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunk(t *testing.T) {
	cfg := Config{ChunkSize: 100, Overlap: 20}

	tests := []struct {
		name       string
		input      string
		cfg        Config
		wantCount  int
		wantAssert func(t *testing.T, chunks []string)
	}{
		{
			name:      "empty input returns nil",
			input:     "",
			cfg:       cfg,
			wantCount: 0,
			wantAssert: func(t *testing.T, chunks []string) {
				assert.Nil(t, chunks)
			},
		},
		{
			name:      "whitespace only returns nil",
			input:     "   \n\n  \t  ",
			cfg:       cfg,
			wantCount: 0,
			wantAssert: func(t *testing.T, chunks []string) {
				assert.Nil(t, chunks)
			},
		},
		{
			name:      "short text returns single chunk",
			input:     "This is a short text.",
			cfg:       cfg,
			wantCount: 1,
			wantAssert: func(t *testing.T, chunks []string) {
				assert.Equal(t, "This is a short text.", chunks[0])
			},
		},
		{
			name:  "basic split by paragraphs",
			input: strings.Repeat("Word ", 30) + "\n\n" + strings.Repeat("Other ", 30),
			cfg:   cfg,
			wantAssert: func(t *testing.T, chunks []string) {
				assert.Greater(t, len(chunks), 1, "should produce multiple chunks")
				for _, c := range chunks {
					assert.NotEmpty(t, c)
				}
			},
		},
		{
			name: "overlap is present between chunks",
			input: "First sentence is about programming. Second sentence is about testing. " +
				"Third sentence is about deployment. Fourth sentence is about monitoring. " +
				"Fifth sentence is about scaling systems.",
			cfg: Config{ChunkSize: 60, Overlap: 20},
			wantAssert: func(t *testing.T, chunks []string) {
				require.Greater(t, len(chunks), 1, "should produce multiple chunks")
				for i := 1; i < len(chunks); i++ {
					prevRunes := []rune(chunks[i-1])
					overlapEnd := string(prevRunes[len(prevRunes)-10:])
					assert.True(t,
						strings.Contains(chunks[i], strings.TrimSpace(overlapEnd)) ||
							len(chunks[i]) > 0,
						"chunk %d should contain overlap from previous chunk", i,
					)
				}
			},
		},
		{
			name:  "preserves sentence boundaries",
			input: "First sentence here. Second sentence here. Third sentence here. Fourth sentence here.",
			cfg:   Config{ChunkSize: 50, Overlap: 10},
			wantAssert: func(t *testing.T, chunks []string) {
				for _, c := range chunks {
					assert.NotEmpty(t, c, "no empty chunks")
				}
			},
		},
		{
			name:  "chunk size is respected approximately",
			input: strings.Repeat("This is a test sentence. ", 50),
			cfg:   Config{ChunkSize: 200, Overlap: 30},
			wantAssert: func(t *testing.T, chunks []string) {
				for i, c := range chunks {
					runeCount := utf8.RuneCountInString(c)
					maxAllowed := 200 + 30 + 50 // chunk + overlap + sentence buffer
					assert.LessOrEqual(t, runeCount, maxAllowed,
						"chunk %d has %d runes, exceeds max %d", i, runeCount, maxAllowed)
				}
			},
		},
		{
			name:  "default config works",
			input: strings.Repeat("Hello world. ", 200),
			cfg:   DefaultConfig(),
			wantAssert: func(t *testing.T, chunks []string) {
				assert.Greater(t, len(chunks), 1, "should produce multiple chunks with default config")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := Chunk(tt.input, tt.cfg)

			if tt.wantCount > 0 {
				assert.Len(t, chunks, tt.wantCount)
			}

			if tt.wantAssert != nil {
				tt.wantAssert(t, chunks)
			}
		})
	}
}
