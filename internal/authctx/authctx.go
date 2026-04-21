package authctx

import "context"

type ctxKey struct{}

var userIDKey ctxKey

// WithUserID 将当前请求关联的用户 ID 写入 context（供 MCP / HTTP 下游共用）。
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserID 从 context 读取用户 ID；未设置或类型不符时 ok 为 false。
func UserID(ctx context.Context) (userID int64, ok bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}
