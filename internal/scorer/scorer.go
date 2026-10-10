// Package scorer assigns a "naming crime" score to a freshly created file and
// maps that score onto one of the aggression severities used by the reminder
// engine. The scale intentionally runs backwards from a sane naming
// convention: structured names score low, improvisation scores high.
package scorer

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"name-it.vino/internal/session"
)

// Assessment is the verdict for a single file.
type Assessment struct {
	Path       string
	Score      int
	Confidence int
	Severity   string
	Reasons    []string
}

var (
	versionRE   = regexp.MustCompile(`(?i)(^|[_-])v\d{1,3}$`)
	numericRE   = regexp.MustCompile(`(?i)(^|[^0-9])([0-9]{1,3})$`)
	copyRE      = regexp.MustCompile(`(?i)(^|[_-])(copy|copia)(\d*)$`)
	dateRE      = regexp.MustCompile(`(?i)(^|[_-])\d{4}[-_]\d{2}[-_]\d{2}$`)
	generatedRE = regexp.MustCompile(`^(IMG|DSC|Screenshot)[ _-]?\d+`)
)

// suspicious collects tokens that betray a name nobody will be able to
// recognise next week. It mixes English and Italian because the author's
// working files tend to wander between the two.
var suspicious = map[string]bool{
	"final": true, "finale": true, "def": true, "definitivo": true, "definitive": true,
	"ultimo": true, "last": true, "new": true, "nuovo": true, "copy": true, "copia": true,
	"test": true, "prova": true, "temp": true, "temporary": true, "old": true, "vecchio": true,
	"latest": true, "vero": true, "true": true, "real": true, "actual": true, "boh": true,
	"aaaa": true, "xxx": true,
}

// structuredPrefixes mark filenames that come from a deliberate production
// scheme (scene03_take02) rather than careless naming. Numbers glued to these
// words are legitimate, so they are not treated as improvised suffixes.
var structuredPrefixes = []string{"scene", "shot", "take", "chapter", "char", "bg"}

// Assess scores a single file. Files that are machine-generated or already
// dated are excused outright.
func Assess(path string, recent []session.FileRecord) Assessment {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	low := strings.ToLower(name)

	if generatedRE.MatchString(name) || dateRE.MatchString(low) {
		return Assessment{Path: path, Score: 0, Confidence: 5, Severity: "ignore"}
	}

	score := 0
	confidence := 72
	reasons := []string{}
	tokens := splitTokens(low)

	suspiciousCount := 0
	for _, t := range tokens {
		if suspicious[t] {
			suspiciousCount++
		}
	}
	if suspiciousCount > 0 {
		score += 10 * suspiciousCount
		confidence += 8
		reasons = append(reasons, "suspicious naming token")
	}

	for i := 1; i < len(tokens); i++ {
		if tokens[i] == tokens[i-1] {
			score += 35
			confidence += 15
			reasons = append(reasons, "repeated token")
			break
		}
	}

	if copyRE.MatchString(low) {
		score += 20
		confidence += 8
		reasons = append(reasons, "copy pattern")
	}

	// A versioned name is a sign of someone who actually keeps their files in
	// order, so it earns a meaningful discount.
	if versionRE.MatchString(low) {
		score -= 35
		reasons = append(reasons, "structured version")
	}

	if numericRE.MatchString(low) && !versionRE.MatchString(low) && !hasStructuredToken(tokens) {
		score += 18
		confidence += 5
		reasons = append(reasons, "improvised numeric suffix")
	}

	if escalation(filepath.Base(path), recent) {
		score += 30
		confidence += 15
		reasons = append(reasons, "naming getting worse")
	}

	if strings.TrimSpace(name) == "" {
		score = 100
		confidence = 100
	}

	if confidence > 99 {
		confidence = 99
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return Assessment{
		Path:       path,
		Score:      score,
		Confidence: confidence,
		Severity:   severityFor(score),
		Reasons:    reasons,
	}
}

// severityFor maps a numeric score onto the aggression tier used to pick a
// message. Thresholds are deliberately low: even a mild crime should sting a
// little.
func severityFor(score int) string {
	switch {
	case score >= 85:
		return "maximum"
	case score >= 70:
		return "aggressive"
	case score >= 50:
		return "passive"
	case score >= 30:
		return "normal"
	case score >= 15:
		return "gentle"
	default:
		return "ignore"
	}
}

// splitTokens breaks a string on non-alphanumeric characters and on every
// letter/digit boundary, so that "finale2" becomes ["finale", "2"].
func splitTokens(s string) []string {
	var out []string

	var b strings.Builder
	prevDigit := false

	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}

		digit := unicode.IsDigit(r)
		if b.Len() > 0 && digit != prevDigit {
			flush()
		}

		b.WriteRune(r)
		prevDigit = digit
	}

	flush()

	return out
}

// hasStructuredToken reports whether any token belongs to a recognised
// production naming scheme, which suppresses the improvised-suffix penalty.
func hasStructuredToken(tokens []string) bool {
	for _, t := range tokens {
		for _, prefix := range structuredPrefixes {
			if strings.HasPrefix(t, prefix) {
				return true
			}
		}
	}

	return false
}

// stem reduces a filename to its comparable core by dropping a trailing number
// and an optional "vN" version marker. It is used to group files that are
// really the same document getting worse over time.
func stem(s string) string {
	tokens := splitTokens(strings.ToLower(s))
	if len(tokens) == 0 {
		return ""
	}

	if _, err := strconv.Atoi(tokens[len(tokens)-1]); err == nil {
		tokens = tokens[:len(tokens)-1]
	}

	if len(tokens) > 0 && strings.HasPrefix(tokens[len(tokens)-1], "v") && len(tokens[len(tokens)-1]) > 1 {
		if _, err := strconv.Atoi(tokens[len(tokens)-1][1:]); err == nil {
			tokens = tokens[:len(tokens)-1]
		}
	}

	return strings.Join(tokens, "_")
}

// escalation detects a document that is being re-saved under the same base name
// with only the trailing number changing. Two previous offenders are enough to
// make it official.
func escalation(name string, recent []session.FileRecord) bool {
	ext := filepath.Ext(name)

	base := stem(strings.TrimSuffix(name, ext))
	if base == "" {
		return false
	}

	count := 0
	last := 0

	for _, r := range recent {
		rExt := filepath.Ext(r.Name)
		if rExt != ext {
			continue
		}

		if stem(strings.TrimSuffix(r.Name, rExt)) == base {
			count++
			if n := trailingNumber(r.Name); n > last {
				last = n
			}
		}
	}

	return count >= 2 && last > 0
}

// trailingNumber returns the numeric suffix of a filename, or 0 when there is
// none.
func trailingNumber(name string) int {
	base := strings.TrimSuffix(name, filepath.Ext(name))

	m := numericRE.FindStringSubmatch(strings.ToLower(base))
	if len(m) == 0 {
		return 0
	}

	n, _ := strconv.Atoi(m[len(m)-1])

	return n
}
