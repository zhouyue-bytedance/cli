// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package note

import (
	"encoding/json"
	"testing"
)

// These tests were relocated from shortcuts/vc/vc_notes_test.go together with
// the note-detail parsing helpers they cover.

func TestParseLooseInt(t *testing.T) {
	tests := []struct {
		input any
		want  int
	}{
		{float64(1), 1},
		{float64(2), 2},
		{json.Number("3"), 3},
		{"unknown", 0},
		{nil, 0},
	}
	for _, tt := range tests {
		got := parseLooseInt(tt.input)
		if got != tt.want {
			t.Errorf("parseLooseInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestExtractArtifactTokens(t *testing.T) {
	artifacts := []any{
		map[string]any{"doc_token": "main_doc", "artifact_type": float64(1)},
		map[string]any{"doc_token": "verbatim_doc", "artifact_type": float64(2)},
		map[string]any{"doc_token": "unknown_doc", "artifact_type": float64(99)},
		nil,
	}
	noteDoc, verbatimDoc := extractArtifactTokens(artifacts)
	if noteDoc != "main_doc" {
		t.Errorf("noteDoc = %q, want %q", noteDoc, "main_doc")
	}
	if verbatimDoc != "verbatim_doc" {
		t.Errorf("verbatimDoc = %q, want %q", verbatimDoc, "verbatim_doc")
	}
}

func TestExtractArtifactTokens_Empty(t *testing.T) {
	noteDoc, verbatimDoc := extractArtifactTokens(nil)
	if noteDoc != "" || verbatimDoc != "" {
		t.Errorf("expected empty tokens for nil input, got %q, %q", noteDoc, verbatimDoc)
	}
}

func TestExtractDocTokens(t *testing.T) {
	refs := []any{
		map[string]any{"doc_token": "shared1"},
		map[string]any{"doc_token": "shared2"},
		map[string]any{"doc_token": ""},
		map[string]any{},
		nil,
	}
	tokens := extractDocTokens(refs)
	if len(tokens) != 2 || tokens[0] != "shared1" || tokens[1] != "shared2" {
		t.Errorf("extractDocTokens = %v, want [shared1 shared2]", tokens)
	}
}

func TestExtractDocTokens_Empty(t *testing.T) {
	tokens := extractDocTokens(nil)
	if tokens != nil {
		t.Errorf("expected nil for nil input, got %v", tokens)
	}
}
