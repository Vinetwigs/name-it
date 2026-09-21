package scorer

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"name-it.vino/internal/session"
)

type Assessment struct {
	Path       string
	Score      int
	Confidence int
	Severity   string
	Reasons    []string
}

var versionRE = regexp.MustCompile(`(?i)(^|[_-])v\d{1,3}$`)
var numericRE = regexp.MustCompile(`(?i)(^|[_-])\d{1,3}$`)
var copyRE = regexp.MustCompile(`(?i)(^|[_-])(copy|copia)(\d*)$`)
var dateRE = regexp.MustCompile(`(?i)(^|[_-])\d{4}[-_]\d{2}[-_]\d{2}$`)
var generatedRE = regexp.MustCompile(`^(IMG|DSC|Screenshot)[ _-]?\d+`)

var suspicious = map[string]bool{"final": true, "finale": true, "def": true, "definitivo": true, "definitive": true, "ultimo": true, "last": true, "new": true, "nuovo": true, "copy": true, "copia": true, "test": true, "prova": true, "temp": true, "temporary": true, "old": true, "vecchio": true, "latest": true, "vero": true, "true": true, "real": true, "actual": true, "boh": true, "aaaa": true, "xxx": true}

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
	if versionRE.MatchString(low) {
		score -= 35
		reasons = append(reasons, "structured version")
	}
	if numericRE.MatchString(low) && !versionRE.MatchString(low) {
		num := numericRE.FindStringSubmatch(low)
		if len(num) > 0 {
			_ = num
			score += 18
			confidence += 5
			reasons = append(reasons, "improvised numeric suffix")
		}
	}
	if escalation(name, recent) {
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
	sev := "ignore"
	switch {
	case score >= 85:
		sev = "maximum"
	case score >= 70:
		sev = "aggressive"
	case score >= 50:
		sev = "passive"
	case score >= 30:
		sev = "normal"
	case score >= 15:
		sev = "gentle"
	}
	return Assessment{Path: path, Score: score, Confidence: confidence, Severity: sev, Reasons: reasons}
}
func splitTokens(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
func stem(s string) string {
	t := splitTokens(strings.ToLower(s))
	if len(t) == 0 {
		return ""
	}
	last := t[len(t)-1]
	if _, e := strconv.Atoi(last); e == nil {
		t = t[:len(t)-1]
	}
	if len(t) > 0 && strings.HasPrefix(t[len(t)-1], "v") && len(t[len(t)-1]) > 1 {
		if _, e := strconv.Atoi(t[len(t)-1][1:]); e == nil {
			t = t[:len(t)-1]
		}
	}
	return strings.Join(t, "_")
}
func escalation(name string, recent []session.FileRecord) bool {
	base := stem(name)
	if base == "" {
		return false
	}
	count := 0
	last := 0
	for _, r := range recent {
		if filepath.Ext(r.Name) != filepath.Ext(name) {
			continue
		}
		if stem(r.Name) == base {
			count++
			if n := trailingNumber(r.Name); n > last {
				last = n
			}
		}
	}
	return count >= 2 && last > 0
}
func trailingNumber(name string) int {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	m := numericRE.FindStringSubmatch(strings.ToLower(base))
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(m[len(m)-1])
	return n
}
