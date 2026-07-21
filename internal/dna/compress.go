package dna

import (
	"regexp"
	"strings"
)

// CompressionMode controls how aggressively Caveman filters.
type CompressionMode int

const (
	// Lenient applies stages 1+2 only (~40% reduction).
	Lenient CompressionMode = iota
	// Normal applies stages 1-3 (~55% reduction).
	Normal
	// Aggressive applies all 4 stages (~65% reduction).
	Aggressive
)

func (m CompressionMode) String() string {
	switch m {
	case Lenient:
		return "lenient"
	case Normal:
		return "normal"
	case Aggressive:
		return "aggressive"
	default:
		return "unknown"
	}
}

// CompressResult holds the outcome of a Caveman compression pass.
type CompressResult struct {
	OriginalSize   int
	CompressedSize int
	Ratio          float64
	Mode           CompressionMode
	Text           string
}

var (
	// Stage 1: strip courtesies — case-insensitive.
	courtesiesRe = regexp.MustCompile(`(?i)\b(?:please|kindly|thank\s*you|i['’]?d\s+like\s+to)\b\s*`)

	// Stage 2: strip hedging — case-insensitive.
	hedgingRe = regexp.MustCompile(`(?i)\b(?:maybe|perhaps|i\s+think|might\s+be|possibly|presumably|arguably|typically|supposedly)\b\s*`)

	// Stage 3: strip connectors — case-insensitive. Preserves and/or/but.
	connectorsRe = regexp.MustCompile(`(?i)\b(?:however|furthermore|moreover|additionally|nevertheless|nonetheless|consequently|thereby|henceforth)\b\s*`)

	// Stage 4: strip articles where they don't change meaning.
	articlesRe = regexp.MustCompile(`(?i)\b(?:a|an|the)\b\s*`)
)

// Compress applies the Caveman token filter pipeline.
func Compress(text string, mode CompressionMode) CompressResult {
	original := []byte(text)
	origSize := len(original)

	out := text

	// Stages always applied.
	out = courtesiesRe.ReplaceAllString(out, "")
	out = hedgingRe.ReplaceAllString(out, "")

	if mode >= Normal {
		out = connectorsRe.ReplaceAllString(out, "")
	}
	if mode >= Aggressive {
		out = articlesRe.ReplaceAllString(out, "")
	}

	// Normalise whitespace.
	out = whitespaceRe.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)

	compSize := len(out)
	ratio := 1.0 - float64(compSize)/float64(origSize)
	if origSize == 0 {
		ratio = 0
	}

	return CompressResult{
		OriginalSize:   origSize,
		CompressedSize: compSize,
		Ratio:          ratio,
		Mode:           mode,
		Text:           out,
	}
}

var whitespaceRe = regexp.MustCompile(`\s+`)
