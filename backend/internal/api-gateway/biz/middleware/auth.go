package middleware

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthMiddleWare() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		token := ctx.Request.Header.Get("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": 401,
				"msg":  "未授权，请先登录",
			})
			ctx.Abort() // 终止后续 Handler 执行
			return
		}
		
		// [模拟鉴权逻辑]：在真实的业务中，这里应该解析 JWT token 拿到 UserID
		// 为了使通用接口可以跑通并打通后台组件，我们这里设定固定的测试账号 UserID: 10001
		ctx.Set("user_id", int64(10001))

		// 验证通过，继续执行下一个 Handler (如 CreateOrder)
		ctx.Next(c)
	}
}
