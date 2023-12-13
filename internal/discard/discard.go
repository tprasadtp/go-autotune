// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package discard

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*Handler)(nil)

// Event is an alias for [log/slog.Record].
type Event = slog.Record

// Handler is a [log/slog.Handler] which discards all events,
// attributes and groups written to it and is always disabled.
type Handler struct{}

// NewHandler returns a new [Handler].
func NewHandler() Handler {
	return Handler{}
}

// Enabled always returns false.
func (d Handler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

// Handle should never be called, as [DiscardHandler.Enabled] always returns false.
func (d Handler) Handle(_ context.Context, _ Event) error {
	return nil
}

// WithAttrs always discards all attrs provided.
func (d Handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return d
}

// WithAttrs always discards the group provided.
func (d Handler) WithGroup(_ string) slog.Handler {
	return d
}
