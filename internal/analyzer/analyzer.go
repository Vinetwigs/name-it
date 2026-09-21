package analyzer

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type Input struct {
	Path        string
	RecentNames []string
	RecentStems []string
}

type Signals struct {
	SuspiciousWord    bool
	RepeatedToken     bool
	SloppyNumber      bool
	StructuredVersion bool
	CopyPattern       bool
	GeneratedFilename bool
	SessionEscalation bool
	Confidence        int
	Reasons           []string
}

var suspiciousWords = map[string]bool{
	"final": true, "finale": true, "def": true, "definitivo": true, "definitive": true,
	"ultimo": true, "last": true, "new": true, "nuovo": true, "copy": true, "copia": true,
	"test": true, "prova": true, "temp": true, "temporary": true, "old": true, "vecchio": true,
	"latest": true, "vero": true, "true": true, "real": true, "actual": true, "bello": true,
	"bellissimo": true, "ok": true, "buono": true,
}

var versionRe = regexp.MustCompile(`(?i)(?:^|[_-])v\d{1,3}$`)
var numberRe = regexp.MustCompile(`(?i)(?:^|[_-])\d{1,3}$`)
var trailingDigitsRe = regexp.MustCompile(`^(.*?)(\d{1,3})$`)
var copyRe = regexp.MustCompile(`(?i)(?:[_ -]?(?:copy|copia))(?:[_ -]?\d+)?$`)
var repeatedSepRe = regexp.MustCompile(`[_ -]+`)
var generatedRe = regexp.MustCompile(`(?i)^(?:IMG|DSC|PXL|Screenshot|Screen Shot|VID|MOV|DSC_|GOPR)[ _-]?\d{3,}$`)
var tempRe = regexp.MustCompile(`(?i)^(?:\.?~lock|~\$|\.tmp|\.temp|.*autosave.*)$`)

func Analyze(in Input) Signals {
	name := filepath.Base(in.Path)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	lower := strings.ToLower(stem)
	tokens := tokenize(lower)
	s := Signals{Confidence: 100}

	if generatedRe.MatchString(stem) || tempRe.MatchString(name) || strings.HasPrefix(name, ".") {
		s.GeneratedFilename = true
		s.Confidence = 8
		s.Reasons = append(s.Reasons, "generated or temporary filename")
		return s
	}

	if versionRe.MatchString(stem) {
		s.StructuredVersion = true
		s.Reasons = append(s.Reasons, "structured version")
	}
	if copyRe.MatchString(stem) || strings.Contains(lower, "_copy") || strings.Contains(lower, "_copia") {
		s.CopyPattern = true
		s.Reasons = append(s.Reasons, "copy pattern")
	}
	if hasSuspiciousWord(tokens) {
		s.SuspiciousWord = true
		s.Reasons = append(s.Reasons, "suspicious naming token")
	}
	if repeatedToken(tokens) {
		s.RepeatedToken = true
		s.Reasons = append(s.Reasons, "repeated token")
	}
	if numberRe.MatchString(stem) || sloppyTrailingNumber(stem) {
		if !s.StructuredVersion && !looksLikeStructuredToken(stem) {
			s.SloppyNumber = true
			s.Reasons = append(s.Reasons, "improvised numeric suffix")
		}
	}
	s.SessionEscalation = escalation(stem, in.RecentStems)
	if s.SessionEscalation {
		s.Reasons = append(s.Reasons, "naming getting worse")
	}

	if s.GeneratedFilename {
		s.Confidence = 5
	}
	if strings.TrimSpace(stem) == "" {
		s.Confidence = 0
	}
	return s
}

func tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' || r == '.' || unicode.IsPunct(r) })
}
func hasSuspiciousWord(tokens []string) bool {
	for _, t := range tokens {
		if suspiciousWords[t] {
			return true
		}
	}
	return false
}
func repeatedToken(tokens []string) bool {
	for i := 1; i < len(tokens); i++ {
		if tokens[i] != "" && tokens[i] == tokens[i-1] {
			return true
		}
	}
	return false
}
func sloppyTrailingNumber(stem string) bool {
	m := trailingDigitsRe.FindStringSubmatch(stem)
	return len(m) == 3 && m[1] != ""
}
func looksLikeStructuredToken(stem string) bool {
	ts := tokenize(strings.ToLower(stem))
	if len(ts) < 2 {
		return false
	}
	for _, t := range ts {
		if strings.HasPrefix(t, "scene") || strings.HasPrefix(t, "shot") || strings.HasPrefix(t, "take") || strings.HasPrefix(t, "chapter") || strings.HasPrefix(t, "char") || strings.HasPrefix(t, "bg") {
			return true
		}
	}
	return false
}
func escalation(stem string, recent []string) bool {
	if len(recent) < 2 {
		return false
	}
	lower := strings.ToLower(stem)
	base := stripChaos(lower)
	matches := 0
	for _, r := range recent {
		if stripChaos(strings.ToLower(r)) == base {
			matches++
		}
	}
	if matches < 2 {
		return false
	}
	return hasBadToken(lower) || sloppyTrailingNumber(lower)
}
func stripChaos(s string) string {
	s = versionRe.ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i)(?:_?finale?|_?final|_?definitivo|_?definitive|_?new|_?copy|_?copia|_?test|_?prova)$`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?:[_ -]?\d{1,3})$`).ReplaceAllString(s, "")
	return strings.Trim(s, "_ -")
}
func hasBadToken(s string) bool {
	return hasSuspiciousWord(tokenize(s)) || strings.Contains(s, "_") && repeatedToken(tokenize(s))
}
