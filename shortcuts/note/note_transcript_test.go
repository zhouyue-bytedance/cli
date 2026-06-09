// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package note

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

func TestNoteTranscriptRequiresUnifiedNote(t *testing.T) {
	factory, stdout, _, reg := noteShortcutTestFactory(t)
	reg.Register(noteDetailStub("note_normal", displayTypeNormal))

	err := runNoteShortcut(t, NoteTranscript, []string{"+transcript", "--note-id", "note_normal", "--output", "out.md", "--as", "user"}, factory, stdout)
	if err == nil {
		t.Fatal("expected non-unified note to fail")
	}
	if got := err.Error(); !strings.Contains(got, "not a unified note") || !strings.Contains(got, "note_display_type=normal") {
		t.Fatalf("err = %q, want non-unified guidance", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestNoteTranscriptFetchesUnifiedNote(t *testing.T) {
	factory, stdout, _, reg := noteShortcutTestFactory(t)
	dir := t.TempDir()
	cmdutil.TestChdir(t, dir)

	reg.Register(noteDetailStub("note_unified", displayTypeUnified))
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/vc/v1/notes/note_unified/unified_note_transcript?format=markdown&locale=zh_cn&page_size=200",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"has_more": false,
				"transcript": map[string]interface{}{
					"markdown": "# transcript\n",
				},
			},
		},
	})

	err := runNoteShortcut(t, NoteTranscript, []string{"+transcript", "--note-id", "note_unified", "--as", "user"}, factory, stdout)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "notes", "note_unified", "unified_transcript.md"))
	if err != nil {
		t.Fatalf("ReadFile transcript err=%v", err)
	}
	if string(content) != "# transcript\n" {
		t.Fatalf("transcript = %q, want %q", string(content), "# transcript\n")
	}
	data := decodeNoteEnvelope(t, stdout)
	if data["note_id"] != "note_unified" || data["size_bytes"] != float64(len(content)) {
		t.Fatalf("unexpected output: %#v", data)
	}
}

func noteShortcutTestFactory(t *testing.T) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer, *httpmock.Registry) {
	t.Helper()
	config := &core.CliConfig{
		AppID:      "test-app-" + strings.ReplaceAll(strings.ToLower(t.Name()), "/", "-"),
		AppSecret:  "test-secret",
		Brand:      core.BrandFeishu,
		UserOpenId: "ou_testuser",
	}
	return cmdutil.TestFactory(t, config)
}

func runNoteShortcut(t *testing.T, shortcut common.Shortcut, args []string, factory *cmdutil.Factory, stdout *bytes.Buffer) error {
	t.Helper()
	parent := &cobra.Command{Use: "note"}
	shortcut.Mount(parent, factory)
	parent.SetArgs(args)
	parent.SilenceErrors = true
	parent.SilenceUsage = true
	stdout.Reset()
	if stderr, ok := factory.IOStreams.ErrOut.(*bytes.Buffer); ok {
		stderr.Reset()
	}
	return parent.ExecuteContext(context.Background())
}

func noteDetailStub(noteID string, displayType int) *httpmock.Stub {
	return &httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/vc/v1/notes/" + noteID,
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"note": map[string]interface{}{
					"note_display_type": displayType,
					"artifacts": []interface{}{
						map[string]interface{}{"artifact_type": artifactTypeVerbatim, "doc_token": "doc_verbatim"},
					},
				},
			},
		},
	}
}

func decodeNoteEnvelope(t *testing.T, stdout *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var envelope map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode stdout: %v\nstdout=%s", err, stdout.String())
	}
	if data, _ := envelope["data"].(map[string]interface{}); data != nil {
		return data
	}
	return envelope
}
