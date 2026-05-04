package middleware

import (
	"context"
	"strings"

	"full_backend_practice/pkg/response"
	token "full_backend_practice/pkg/token"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthMiddleWare() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(ctx, 401, "未授权，请先登录")
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(ctx, 401, "Token 格式错误")
			ctx.Abort()
			return
		}
		claims, err := token.ParseToken(parts[1], token.AccessSecret)
		if err != nil {
			response.Error(ctx, 401, "Token 无效")
			ctx.Abort()
			return
		}
		ctx.Set("user_id", claims.UserID)
		ctx.Next(c)
	}
}
