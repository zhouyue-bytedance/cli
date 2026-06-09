// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/shortcuts/common"
)

// testDocxToken is a bare docx token that parseDocumentRef accepts, letting the
// validation tests reach the flag checks that run after --doc is resolved.
const testDocxToken = "doxcnDocErrorsTestToken"

// docValidateRuntime builds a RuntimeContext carrying only the flags a Doc
// Validate function reads. String values are applied (and marked Changed) only
// when non-empty; int values are always applied so Changed() reports true,
// mirroring how cobra records an explicitly supplied numeric flag.
func docValidateRuntime(t *testing.T, str map[string]string, bools map[string]bool, ints map[string]int) *common.RuntimeContext {
	t.Helper()
	cmd := &cobra.Command{Use: "docs"}
	fs := cmd.Flags()
	for name, val := range str {
		fs.String(name, "", "")
		if val != "" {
			if err := fs.Set(name, val); err != nil {
				t.Fatalf("set --%s=%q: %v", name, val, err)
			}
		}
	}
	for name, val := range bools {
		fs.Bool(name, false, "")
		if val {
			if err := fs.Set(name, "true"); err != nil {
				t.Fatalf("set --%s: %v", name, err)
			}
		}
	}
	for name, val := range ints {
		fs.Int(name, 0, "")
		if err := fs.Set(name, strconv.Itoa(val)); err != nil {
			t.Fatalf("set --%s=%d: %v", name, val, err)
		}
	}
	return common.TestNewRuntimeContext(cmd, nil)
}

// assertValidationContract pins the typed envelope every migrated Doc
// validation fault must emit: a *errs.ValidationError in CategoryValidation
// with the expected Subtype, and the offending flag in Param. wantParam == ""
// asserts the deliberate convention that multi-flag faults (mutual exclusion,
// "one of A or B") carry no single Param.
func assertValidationContract(t *testing.T, err error, wantSubtype errs.Subtype, wantParam string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var ve *errs.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("error type = %T, want *errs.ValidationError (%v)", err, err)
	}
	if ve.Category != errs.CategoryValidation {
		t.Errorf("category = %q, want %q", ve.Category, errs.CategoryValidation)
	}
	if ve.Subtype != wantSubtype {
		t.Errorf("subtype = %q, want %q", ve.Subtype, wantSubtype)
	}
	if ve.Param != wantParam {
		t.Errorf("param = %q, want %q", ve.Param, wantParam)
	}
}

func TestDocMediaInsertValidateContract(t *testing.T) {
	cases := []struct {
		name      string
		str       map[string]string
		bools     map[string]bool
		ints      map[string]int
		wantParam string
	}{
		{
			name:      "neither file nor clipboard",
			str:       map[string]string{"doc": testDocxToken},
			wantParam: "", // one-of-two flags: no single param
		},
		{
			name:      "file and clipboard together",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png"},
			bools:     map[string]bool{"from-clipboard": true},
			wantParam: "", // mutual exclusion: no single param
		},
		{
			name:      "non-docx document",
			str:       map[string]string{"doc": "https://example.larksuite.com/doc/xxxxxx", "file": "dummy.png"},
			wantParam: "--doc",
		},
		{
			name:      "blank selection",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "selection-with-ellipsis": "   "},
			wantParam: "--selection-with-ellipsis",
		},
		{
			name:      "before without selection",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png"},
			bools:     map[string]bool{"before": true},
			wantParam: "--before",
		},
		{
			name:      "invalid file-view",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "file-view": "bogus"},
			wantParam: "--file-view",
		},
		{
			name:      "file-view without type file",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "file-view": "card", "type": "image"},
			wantParam: "--file-view",
		},
		{
			name:      "dimensions with non-image type",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "type": "file"},
			ints:      map[string]int{"width": 100},
			wantParam: "", // spans --width/--height and --type: no single param
		},
		{
			name:      "non-positive width",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "type": "image"},
			ints:      map[string]int{"width": 0},
			wantParam: "--width",
		},
		{
			name:      "non-positive height",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "type": "image"},
			ints:      map[string]int{"height": 0},
			wantParam: "--height",
		},
		{
			name:      "width over maximum",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "type": "image"},
			ints:      map[string]int{"width": 10001},
			wantParam: "--width",
		},
		{
			name:      "height over maximum",
			str:       map[string]string{"doc": testDocxToken, "file": "dummy.png", "type": "image"},
			ints:      map[string]int{"height": 10001},
			wantParam: "--height",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := docValidateRuntime(t, tc.str, tc.bools, tc.ints)
			err := DocMediaInsert.Validate(context.Background(), rt)
			assertValidationContract(t, err, errs.SubtypeInvalidArgument, tc.wantParam)
		})
	}
}

func TestValidateCreateV2Contract(t *testing.T) {
	cases := []struct {
		name      string
		str       map[string]string
		wantParam string
	}{
		{
			name:      "content required",
			str:       map[string]string{},
			wantParam: "--content",
		},
		{
			name:      "parent token and position mutually exclusive",
			str:       map[string]string{"content": "<doc/>", "parent-token": "fldcnX", "parent-position": "my_library"},
			wantParam: "", // mutual exclusion: no single param
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := docValidateRuntime(t, tc.str, nil, nil)
			err := validateCreateV2(context.Background(), rt)
			assertValidationContract(t, err, errs.SubtypeInvalidArgument, tc.wantParam)
		})
	}
}

func TestValidateFetchV2Contract(t *testing.T) {
	cases := []struct {
		name      string
		str       map[string]string
		ints      map[string]int
		wantParam string
	}{
		{
			name:      "range mode without block ids",
			str:       map[string]string{"doc": testDocxToken, "detail": "simple", "scope": "range"},
			wantParam: "", // either --start-block-id or --end-block-id: no single param
		},
		{
			name:      "keyword mode without keyword",
			str:       map[string]string{"doc": testDocxToken, "detail": "simple", "scope": "keyword"},
			wantParam: "--keyword",
		},
		{
			name:      "section mode without start block id",
			str:       map[string]string{"doc": testDocxToken, "detail": "simple", "scope": "section"},
			wantParam: "--start-block-id",
		},
		{
			name:      "negative context-before",
			str:       map[string]string{"doc": testDocxToken, "detail": "simple", "scope": "outline"},
			ints:      map[string]int{"context-before": -1},
			wantParam: "--context-before",
		},
		{
			name:      "unknown scope",
			str:       map[string]string{"doc": testDocxToken, "detail": "simple", "scope": "bogus"},
			wantParam: "--scope",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := docValidateRuntime(t, tc.str, nil, tc.ints)
			err := validateFetchV2(context.Background(), rt)
			assertValidationContract(t, err, errs.SubtypeInvalidArgument, tc.wantParam)
		})
	}
}

// TestBuildDocsSearchRequestPreservesParseCause pins the --filter parse faults:
// the typed envelope carries Param --filter and chains the original parse error
// so errors.Is/Unwrap traversal keeps the underlying JSON/time-parse detail.
func TestBuildDocsSearchRequestPreservesParseCause(t *testing.T) {
	cases := []struct {
		name   string
		filter string
	}{
		{"invalid filter json", "{not json"},
		{"invalid open_time start", `{"open_time":{"start":"not-a-time"}}`},
		{"invalid open_time end", `{"open_time":{"end":"not-a-time"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildDocsSearchRequest("q", tc.filter, "", "15")
			var ve *errs.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("error type = %T, want *errs.ValidationError (%v)", err, err)
			}
			if ve.Subtype != errs.SubtypeInvalidArgument {
				t.Errorf("subtype = %q, want %q", ve.Subtype, errs.SubtypeInvalidArgument)
			}
			if ve.Param != "--filter" {
				t.Errorf("param = %q, want %q", ve.Param, "--filter")
			}
			if errors.Unwrap(ve) == nil {
				t.Error("parse error not chained: errors.Unwrap == nil")
			}
		})
	}
}

func TestValidateUpdateV2Contract(t *testing.T) {
	cases := []struct {
		name      string
		str       map[string]string
		wantParam string
	}{
		{
			name:      "command required",
			str:       map[string]string{"doc": testDocxToken},
			wantParam: "--command",
		},
		{
			name:      "invalid command",
			str:       map[string]string{"doc": testDocxToken, "command": "bogus"},
			wantParam: "--command",
		},
		{
			name:      "str_replace without pattern",
			str:       map[string]string{"doc": testDocxToken, "command": "str_replace"},
			wantParam: "--pattern",
		},
		{
			name:      "block_delete without block id",
			str:       map[string]string{"doc": testDocxToken, "command": "block_delete"},
			wantParam: "--block-id",
		},
		{
			name:      "block_insert_after without block id",
			str:       map[string]string{"doc": testDocxToken, "command": "block_insert_after"},
			wantParam: "--block-id",
		},
		{
			name:      "block_insert_after without content",
			str:       map[string]string{"doc": testDocxToken, "command": "block_insert_after", "block-id": "blkX"},
			wantParam: "--content",
		},
		{
			name:      "block_copy_insert_after without block id",
			str:       map[string]string{"doc": testDocxToken, "command": "block_copy_insert_after"},
			wantParam: "--block-id",
		},
		{
			name:      "block_copy_insert_after without src block ids",
			str:       map[string]string{"doc": testDocxToken, "command": "block_copy_insert_after", "block-id": "blkX"},
			wantParam: "--src-block-ids",
		},
		{
			name:      "block_move_after without block id",
			str:       map[string]string{"doc": testDocxToken, "command": "block_move_after"},
			wantParam: "--block-id",
		},
		{
			name:      "block_move_after without src block ids",
			str:       map[string]string{"doc": testDocxToken, "command": "block_move_after", "block-id": "blkX"},
			wantParam: "--src-block-ids",
		},
		{
			name:      "block_move_after rejects content",
			str:       map[string]string{"doc": testDocxToken, "command": "block_move_after", "block-id": "blkX", "src-block-ids": "blkY", "content": "x"},
			wantParam: "--content",
		},
		{
			name:      "block_replace without block id",
			str:       map[string]string{"doc": testDocxToken, "command": "block_replace"},
			wantParam: "--block-id",
		},
		{
			name:      "block_replace without content",
			str:       map[string]string{"doc": testDocxToken, "command": "block_replace", "block-id": "blkX"},
			wantParam: "--content",
		},
		{
			name:      "overwrite without content",
			str:       map[string]string{"doc": testDocxToken, "command": "overwrite"},
			wantParam: "--content",
		},
		{
			name:      "append without content",
			str:       map[string]string{"doc": testDocxToken, "command": "append"},
			wantParam: "--content",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := docValidateRuntime(t, tc.str, nil, nil)
			err := validateUpdateV2(context.Background(), rt)
			assertValidationContract(t, err, errs.SubtypeInvalidArgument, tc.wantParam)
		})
	}
}
