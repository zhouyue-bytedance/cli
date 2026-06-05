// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

// Package note owns the Note domain: querying note detail and the unified
// (three-in-one) transcript by a known note_id. The vc domain locates a
// note_id from meeting context and delegates note-detail parsing here, so the
// parsing logic lives in exactly one place.
package note

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

// NoNoteReadPermissionCode is returned when the caller lacks read permission
// for the requested note.
const NoNoteReadPermissionCode = 121005

// artifact_type enum from the note detail API.
const (
	artifactTypeMainDoc  = 1 // main note document
	artifactTypeVerbatim = 2 // verbatim transcript
)

// note_display_type enum (i32) from the note detail API. Surfaced to callers as
// a stable string so Agents route on a name, not a magic number.
const (
	displayTypeNormal  = 1
	displayTypeUnified = 2
)

// Detail is the parsed note detail shared by `note +detail` and `vc +notes`.
type Detail struct {
	NoteID           string
	CreatorID        string
	CreateTime       string
	DisplayType      string // unknown | normal | unified
	NoteDocToken     string
	VerbatimDocToken string
	SharedDocTokens  []string
}

// FetchDetail queries GET /open-apis/vc/v1/notes/{note_id} and parses the note
// object. The raw API error is returned untouched so callers map the
// no-permission case to their own user-facing message.
func FetchDetail(_ context.Context, runtime *common.RuntimeContext, noteID string) (*Detail, error) {
	data, err := runtime.DoAPIJSON(http.MethodGet, fmt.Sprintf("/open-apis/vc/v1/notes/%s", validate.EncodePathSegment(noteID)), nil, nil)
	if err != nil {
		return nil, err
	}
	noteObj, _ := data["note"].(map[string]any)
	if noteObj == nil {
		return nil, output.ErrAPI(0, "note detail is empty", nil)
	}
	noteDoc, verbatimDoc := extractArtifactTokens(common.GetSlice(noteObj, "artifacts"))
	return &Detail{
		NoteID:           noteID,
		CreatorID:        common.GetString(noteObj, "creator_id"),
		CreateTime:       common.FormatTime(noteObj["create_time"]),
		DisplayType:      displayTypeString(displayTypeValue(noteObj)),
		NoteDocToken:     noteDoc,
		VerbatimDocToken: verbatimDoc,
		SharedDocTokens:  extractDocTokens(common.GetSlice(noteObj, "references")),
	}, nil
}

// ToMap renders the detail as the field map consumed by `vc +notes`, keeping
// the historical key set (shared_doc_tokens omitted when empty) and adding the
// note_id / note_display_type fields.
func (d *Detail) ToMap() map[string]any {
	m := map[string]any{
		"note_id":            d.NoteID,
		"note_display_type":  d.DisplayType,
		"creator_id":         d.CreatorID,
		"create_time":        d.CreateTime,
		"note_doc_token":     d.NoteDocToken,
		"verbatim_doc_token": d.VerbatimDocToken,
	}
	if len(d.SharedDocTokens) > 0 {
		m["shared_doc_tokens"] = d.SharedDocTokens
	}
	return m
}

// displayTypeValue reads the display-type field, tolerating either the
// documented note_display_type key or a bare display_type fallback.
func displayTypeValue(note map[string]any) any {
	if v, ok := note["note_display_type"]; ok {
		return v
	}
	return note["display_type"]
}

func displayTypeString(v any) string {
	switch parseLooseInt(v) {
	case displayTypeNormal:
		return "normal"
	case displayTypeUnified:
		return "unified"
	default:
		return "unknown"
	}
}

// extractArtifactTokens picks main-doc and verbatim-doc tokens from artifacts.
func extractArtifactTokens(artifacts []any) (noteDoc, verbatimDoc string) {
	for _, a := range artifacts {
		artifact, _ := a.(map[string]any)
		if artifact == nil {
			continue
		}
		docToken, _ := artifact["doc_token"].(string)
		switch parseLooseInt(artifact["artifact_type"]) {
		case artifactTypeMainDoc:
			noteDoc = docToken
		case artifactTypeVerbatim:
			verbatimDoc = docToken
		}
	}
	return
}

// extractDocTokens collects doc_token values from a list of reference objects.
func extractDocTokens(refs []any) []string {
	var tokens []string
	for _, s := range refs {
		source, _ := s.(map[string]any)
		if source == nil {
			continue
		}
		if docToken, _ := source["doc_token"].(string); docToken != "" {
			tokens = append(tokens, docToken)
		}
	}
	return tokens
}

// parseLooseInt extracts an int from the varying JSON number representations
// DoAPIJSON may yield (json.Number, float64, or int).
func parseLooseInt(v any) int {
	switch n := v.(type) {
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// parseLooseInt64 is the int64 variant used for pagination cursors.
func parseLooseInt64(v any) int64 {
	switch n := v.(type) {
	case json.Number:
		i, _ := n.Int64()
		return i
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}
