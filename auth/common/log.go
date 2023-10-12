package common

import (
	"context"

	"golang.org/x/exp/slog"
)

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	requestID := ctx.Value(ContextKeyRequestID)
	args = append(args, slog.Any("request_id", requestID))
	slog.Log(ctx, level, msg, args)
}
