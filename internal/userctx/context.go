package userctx

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey contextKey = "userID"

var ErrUserIDMissing = errors.New("user id not found in context")

// WithUserID returns a new context containing the provided userID.
func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID extracts the userID from context. Returns ErrUserIDMissing if value not present.
func GetUserID(ctx context.Context) (uint, error) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, ErrUserIDMissing
	}

	userID, ok := val.(uint)
	if !ok {
		return 0, ErrUserIDMissing
	}

	return userID, nil
}
