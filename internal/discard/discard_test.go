// SPDX-FileCopyrightText: Copyright 2023 Prasad Tengse
// SPDX-License-Identifier: MIT

package discard_test

import (
	"context"
	"log/slog"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/tprasadtp/go-autotune/internal/discard"
)

func TestDiscardHandler(t *testing.T) {
	t.Run("always-disabled", func(t *testing.T) {
		handler := discard.NewHandler()
		tt := []slog.Level{math.MinInt, -255, -8, -4, 0, 4, 8, 90, 99, 255, math.MaxInt}
		for i := range tt {
			if handler.Enabled(context.Background(), tt[i]) {
				t.Errorf("discard handler Enabled returned true Level=%d", tt[i])
			}
		}
	})

	t.Run("nil-err-on-handle", func(t *testing.T) {
		handler := discard.NewHandler()
		err := handler.Handle(context.Background(),
			slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0))
		if err != nil {
			t.Errorf("Handle() must return nil")
		}
	})

	t.Run("returns-receiver-with-attr", func(t *testing.T) {
		handler := discard.NewHandler()
		rv := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})
		if !reflect.DeepEqual(rv, handler) {
			t.Error("discard handler WithAttrs did not return the receiver")
		}
	})

	t.Run("comparable", func(t *testing.T) {
		h1 := discard.NewHandler()
		h2 := discard.NewHandler()
		if h1 != h2 {
			t.Errorf("Returned values are not comparable")
		}
	})

	t.Run("returns-receiver-with-group", func(t *testing.T) {
		handler := discard.NewHandler()
		rv := handler.WithGroup("foo")
		if !reflect.DeepEqual(rv, handler) {
			t.Error("discard handler WithGroup did not return the receiver")
		}
	})
}
