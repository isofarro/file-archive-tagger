package fileops

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Regex for finding camelCase words
	camelCase = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

// NormalizeFilename normalizes a filename according to the rules
func NormalizeFilename(filename string) string {
	// Get the extension
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace '&' with 'and'
	name = strings.ReplaceAll(name, "&", "and")

	// Remove apostrophes
	name = strings.ReplaceAll(name, "'", "")

	// Split camelCase into words
	name = camelCase.ReplaceAllString(name, "${1}-${2}")
	name = strings.ToLower(name)

	// Replace special characters with '-'
	name = replaceSpecialChars(name)

	// Reduce multiple '-' to single '-'
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")

	// Trim leading/trailing '-'
	name = strings.Trim(name, "-")

	return name + ext
}

// replaceSpecialChars replaces special characters with '-'
func replaceSpecialChars(s string) string {
	var result strings.Builder
	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsNumber(ch) || ch == '-' {
			result.WriteRune(ch)
		} else {
			result.WriteRune('-')
		}
	}
	return result.String()
}