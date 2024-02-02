// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package shared

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*DiscardHandler)(nil)

// Event is an alias for [log/slog.Record].
type Event = slog.Record

// DiscardHandler is a [log/slog.DiscardHandler] which discards all events,
// attributes and groups written to it and is always disabled.
type DiscardHandler struct{}

// NewDiscardHandler returns a new [Handler].
func NewDiscardHandler() DiscardHandler {
	return DiscardHandler{}
}

// Enabled always returns false.
func (d DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

// Handle should never be called, as [DiscardHandler.Enabled] always returns false.
func (d DiscardHandler) Handle(_ context.Context, _ Event) error {
	return nil
}

// WithAttrs always discards all attrs provided.
func (d DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return d
}

// WithAttrs always discards the group provided.
func (d DiscardHandler) WithGroup(_ string) slog.Handler {
	return d
}
