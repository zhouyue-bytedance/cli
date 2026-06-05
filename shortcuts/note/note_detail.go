// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT
//
// note +detail — get note metadata and document tokens by a known note_id.

package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

// NoteDetail queries note metadata, display type and document tokens by note_id.
var NoteDetail = common.Shortcut{
	Service:     "note",
	Command:     "+detail",
	Description: "Get note detail (display type, document tokens) by note_id",
	Risk:        "read",
	Scopes:      []string{"vc:note:read"},
	AuthTypes:   []string{"user"},
	Flags: []common.Flag{
		{Name: "note-id", Desc: "note ID", Required: true},
	},
	Validate: func(_ context.Context, runtime *common.RuntimeContext) error {
		noteID := strings.TrimSpace(runtime.Str("note-id"))
		if noteID == "" {
			return output.ErrValidation("--note-id is required")
		}
		if err := validate.ResourceName(noteID, "--note-id"); err != nil {
			return output.ErrValidation("%s", err)
		}
		return nil
	},
	DryRun: func(_ context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		noteID := strings.TrimSpace(runtime.Str("note-id"))
		return common.NewDryRunAPI().
			GET(fmt.Sprintf("/open-apis/vc/v1/notes/%s", validate.EncodePathSegment(noteID)))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		noteID := strings.TrimSpace(runtime.Str("note-id"))
		detail, err := FetchDetail(ctx, runtime, noteID)
		if err != nil {
			return mapNoteError(err)
		}
		runtime.OutFormat(map[string]any{"note": detail.ToMap()}, nil, nil)
		return nil
	},
}

// mapNoteError surfaces the no-permission case explicitly and passes through
// any other API error unchanged.
func mapNoteError(err error) error {
	var exitErr *output.ExitError
	if errors.As(err, &exitErr) && exitErr.Detail != nil && exitErr.Detail.Code == NoNoteReadPermissionCode {
		return output.ErrAPI(NoNoteReadPermissionCode, "no read permission for this note", nil)
	}
	return err
}
