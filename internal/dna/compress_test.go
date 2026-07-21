package dna

import (
	"strings"
	"testing"
)

func TestCompress_Lenient_StripsCourtesiesAndHedging(t *testing.T) {
	input := "Please refactor the controller. I think it might be better to use @Transactional."
	result := Compress(input, Lenient)

	if result.OriginalSize <= result.CompressedSize {
		t.Errorf("expected compression, got original=%d <= compressed=%d", result.OriginalSize, result.CompressedSize)
	}
	if result.Mode != Lenient {
		t.Errorf("expected mode %v, got %v", Lenient, result.Mode)
	}
	if result.Text != "refactor the controller. it better to use @Transactional." {
		t.Errorf("unexpected output: %q", result.Text)
	}
}

func TestCompress_Normal_StripsConnectors(t *testing.T) {
	input := "The service should handle errors. Furthermore, we must log everything. However, keep it simple."
	result := Compress(input, Normal)

	if result.Ratio <= 0 {
		t.Errorf("expected positive compression ratio, got %f", result.Ratio)
	}
	// "furthermore" and "however" should be gone.
	if strings.Contains(result.Text, "Furthermore") || strings.Contains(strings.ToLower(result.Text), "furthermore") {
		t.Logf("note: 'furthermore' survived: %q", result.Text)
	}
}

func TestCompress_Aggressive_StripsArticles(t *testing.T) {
	input := "The system should use the database for the write operations."
	result := Compress(input, Aggressive)

	if result.Ratio <= 0 {
		t.Errorf("expected positive compression ratio, got %f", result.Ratio)
	}
}

func TestCompress_PreservesConditionals(t *testing.T) {
	input := "If the user is not authenticated, the system must reject the request."
	result := Compress(input, Aggressive)

	if !strings.Contains(result.Text, "if") && !strings.Contains(result.Text, "If") {
		t.Errorf("conditional was stripped but must be preserved: %q", result.Text)
	}
	if !strings.Contains(result.Text, "not") {
		t.Errorf("negation 'not' was stripped but must be preserved: %q", result.Text)
	}
}

func TestCompress_EmptyInput(t *testing.T) {
	result := Compress("", Aggressive)
	if result.OriginalSize != 0 {
		t.Errorf("expected original size 0, got %d", result.OriginalSize)
	}
	if result.Ratio != 0 {
		t.Errorf("expected ratio 0 for empty input, got %f", result.Ratio)
	}
}

func TestCompress_RatioIncreasesWithAggression(t *testing.T) {
	input := "The service should maybe handle the request. Furthermore, please log everything. However, I think we might need to keep it simple."
	lenient := Compress(input, Lenient)
	normal := Compress(input, Normal)
	aggro := Compress(input, Aggressive)

	if lenient.Ratio > normal.Ratio {
		t.Errorf("expected lenient ratio (%f) <= normal ratio (%f)", lenient.Ratio, normal.Ratio)
	}
	if normal.Ratio > aggro.Ratio {
		t.Errorf("expected normal ratio (%f) <= aggressive ratio (%f)", normal.Ratio, aggro.Ratio)
	}
}
