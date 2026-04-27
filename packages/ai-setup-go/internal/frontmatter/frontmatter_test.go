package frontmatter

import (
	"testing"
)

func TestExtractFrontmatter_ParsesValidFrontmatter(t *testing.T) {
	t.Parallel()

	content := []byte(`---
title: Test Page
tags:
  - go
  - test
---
This is the body content.
`)
	fm, body, err := ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("ExtractFrontmatter: %v", err)
	}

	if len(fm) == 0 {
		t.Fatal("frontmatter map is empty")
	}
	if fm["title"] != "Test Page" {
		t.Errorf("title = %v, want Test Page", fm["title"])
	}

	bodyStr := string(body)
	if bodyStr == "" {
		t.Error("body is empty")
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	t.Parallel()

	content := []byte("Just regular content without frontmatter.\n")
	fm, body, err := ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("ExtractFrontmatter: %v", err)
	}

	if len(fm) != 0 {
		t.Errorf("expected empty map, got %v", fm)
	}
	if string(body) != string(content) {
		t.Error("body should equal original content when no frontmatter")
	}
}

func TestHasFrontmatter(t *testing.T) {
	t.Parallel()

	withFM := []byte("---\ntitle: Test\n---\nBody here\n")
	withoutFM := []byte("No frontmatter here\n")

	if !HasFrontmatter(withFM) {
		t.Error("HasFrontmatter = false for content with frontmatter")
	}
	if HasFrontmatter(withoutFM) {
		t.Error("HasFrontmatter = true for content without frontmatter")
	}
}

func TestStripFrontmatter(t *testing.T) {
	t.Parallel()

	content := []byte("---\ntitle: Test\n---\nBody content\n")
	result := StripFrontmatter(content)

	if HasFrontmatter(result) {
		t.Error("StripFrontmatter did not remove frontmatter")
	}
}

func TestExtractField(t *testing.T) {
	t.Parallel()

	fm := map[string]any{
		"title": "My Title",
		"count": 42,
	}

	if got := ExtractField(fm, "title"); got != "My Title" {
		t.Errorf("ExtractField(title) = %q, want My Title", got)
	}
	if got := ExtractField(fm, "count"); got != "" {
		t.Errorf("ExtractField(count) = %q, want empty (not a string)", got)
	}
	if got := ExtractField(fm, "missing"); got != "" {
		t.Errorf("ExtractField(missing) = %q, want empty", got)
	}
}

func TestSplitYamlFrontmatter(t *testing.T) {
	t.Parallel()

	content := "---\ntitle: Hello\n---\nBody text"
	fmBody, body := SplitYamlFrontmatter(content)

	if fmBody != "title: Hello" {
		t.Errorf("frontmatter body = %q, want %q", fmBody, "title: Hello")
	}
	if body != "Body text" {
		t.Errorf("body = %q, want Body text", body)
	}
}

func TestSplitYamlFrontmatter_NoFrontmatter(t *testing.T) {
	t.Parallel()

	content := "Just text, no frontmatter"
	fmBody, body := SplitYamlFrontmatter(content)

	if fmBody != "" {
		t.Errorf("frontmatter body = %q, want empty", fmBody)
	}
	if body != content {
		t.Errorf("body = %q, want %q", body, content)
	}
}

func TestExtractSchemaVersion(t *testing.T) {
	t.Parallel()

	version, err := ExtractSchemaVersion([]byte("---\nschema_version: 1\nartifact_type: spec_plan\n---\nbody\n"))
	if err != nil {
		t.Fatalf("ExtractSchemaVersion: %v", err)
	}
	if version != 1 {
		t.Fatalf("schema version = %d, want 1", version)
	}
}

func TestHasSpec006Metadata(t *testing.T) {
	t.Parallel()

	withMetadata := []byte("---\nschema_version: 1\nartifact_type: spec_plan\n---\nbody\n")
	withoutMetadata := []byte("---\ntitle: Test\n---\nbody\n")

	if !HasSpec006Metadata(withMetadata) {
		t.Fatal("expected HasSpec006Metadata to return true")
	}
	if HasSpec006Metadata(withoutMetadata) {
		t.Fatal("expected HasSpec006Metadata to return false")
	}
}
