// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package markdown

import (
	"github.com/larksuite/cli/errs"
)

func markdownValidationError(format string, args ...any) *errs.ValidationError {
	return errs.NewValidationError(errs.SubtypeInvalidArgument, format, args...)
}

func markdownValidationParamError(param, format string, args ...any) *errs.ValidationError {
	return markdownValidationError(format, args...).WithParam(param)
}

func markdownInvalidParam(name, reason string) errs.InvalidParam {
	return errs.InvalidParam{Name: name, Reason: reason}
}

// wrapMarkdownDownloadError classifies a download failure. An already-typed
// error keeps its carrier — type, subtype, code and extensions — so callers see
// the upstream classification: a validation problem passes through verbatim,
// any other problem gains a "download failed" prefix for operation context.
// An untyped error becomes a network transport error carrying the original as
// its cause.
func wrapMarkdownDownloadError(err error) error {
	if p, ok := errs.ProblemOf(err); ok {
		if p.Category != errs.CategoryValidation {
			p.Message = "download failed: " + p.Message
		}
		return err
	}
	return errs.NewNetworkError(errs.SubtypeNetworkTransport, "download failed: %s", err).WithCause(err)
}
