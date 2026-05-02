package processor

import (
	"strings"
	"testing"
)

func TestParse_ValidJSON(t *testing.T) {
	data := []byte(`{"slug":"test","name":"Test","tasks":[]}`)
	pb, err := newTestParser().Parse(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pb.Slug != "test" {
		t.Errorf("expected slug=test, got %s", pb.Slug)
	}
	if pb.Name != "Test" {
		t.Errorf("expected name=Test, got %s", pb.Name)
	}
}

func TestParse_EmptyData(t *testing.T) {
	_, err := newTestParser().Parse([]byte{}, "test.json")
	if err == nil {
		t.Error("expected error for empty data")
	}
	if !strings.Contains(err.Error(), "empty data") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := newTestParser().Parse([]byte(`{"slug": "bad"`), "test.json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "JSON syntax error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_UnsupportedFormat(t *testing.T) {
	_, err := newTestParser().Parse([]byte(`slug: test`), "test.yaml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParse_UppercaseExtension(t *testing.T) {
	data := []byte(`{"slug":"test","name":"Test","tasks":[]}`)
	_, err := newTestParser().Parse(data, "test.JSON")
	if err != nil {
		t.Errorf("uppercase .JSON extension should parse as json, got: %v", err)
	}
}
