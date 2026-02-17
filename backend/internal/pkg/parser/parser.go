package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

var supportedExtensions = map[string]bool{
	".txt": true,
	".md":  true,
	".pdf": true,
}

func IsSupportedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedExtensions[ext]
}

func Parse(file io.Reader, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".txt":
		return parseTxt(file)
	case ".md":
		return parseMd(file)
	case ".pdf":
		return parsePdf(file)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}
