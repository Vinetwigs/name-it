// Package pattern turns a human-friendly naming template into a regular
// expression and checks filenames against it.
//
// The template language is deliberately small:
//
//	filename   any non-empty filename stem
//	extension  the extension without the dot
//	X          one or more digits
//	YYYY       four-digit year
//	MM         two-digit month
//	DD         two-digit day
//
// Anything else is treated as literal text. So, for example,
//
//	filename_vX.extension        ->  report_v1.txt, invoice_v42.pdf
//	YYYY-MM-DD_filename_vX.extension ->  2026-09-21_report_v3.txt
package pattern

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Matches reports whether filename satisfies the given pattern template.
func Matches(filename string, pattern string) bool {
	filename = filepath.Base(filename)
	pattern = strings.TrimSpace(pattern)

	if filename == "" || pattern == "" {
		return false
	}

	expression := patternToRegex(pattern)

	matched, err := regexp.MatchString(
		"^"+expression+"$",
		filename,
	)

	if err != nil {
		return false
	}

	return matched
}

func patternToRegex(pattern string) string {
	var result strings.Builder

	for i := 0; i < len(pattern); {
		switch {

		case strings.HasPrefix(pattern[i:], "YYYY"):
			result.WriteString(`\d{4}`)
			i += 4

		case strings.HasPrefix(pattern[i:], "extension"):
			result.WriteString(`[A-Za-z0-9]+`)
			i += len("extension")

		case strings.HasPrefix(pattern[i:], "filename"):
			result.WriteString(`[^.]+`)
			i += len("filename")

		case strings.HasPrefix(pattern[i:], "MM"):
			result.WriteString(`(?:0[1-9]|1[0-2])`)
			i += 2

		case strings.HasPrefix(pattern[i:], "DD"):
			result.WriteString(`(?:0[1-9]|[12][0-9]|3[01])`)
			i += 2

		case pattern[i] == 'X':
			result.WriteString(`\d+`)
			i++

		default:
			result.WriteString(regexp.QuoteMeta(string(pattern[i])))
			i++
		}
	}

	return result.String()
}

func DefaultPattern() string {
	return "filename_vX.extension"
}

func DatePattern() string {
	return "YYYY-MM-DD_filename_vX.extension"
}

func ExampleFilename(pattern string) string {
	now := time.Now()

	example := pattern

	example = strings.ReplaceAll(
		example,
		"YYYY",
		now.Format("2006"),
	)

	example = strings.ReplaceAll(
		example,
		"MM",
		now.Format("01"),
	)

	example = strings.ReplaceAll(
		example,
		"DD",
		now.Format("02"),
	)

	example = strings.ReplaceAll(
		example,
		"filename",
		"report",
	)

	example = strings.ReplaceAll(
		example,
		"extension",
		"txt",
	)

	example = strings.ReplaceAll(
		example,
		"X",
		"1",
	)

	return example
}
