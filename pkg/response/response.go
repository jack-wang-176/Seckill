package response

import (
	"github.com/cloudwego/hertz/pkg/app"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(ctx *app.RequestContext, data interface{}) {
	ctx.JSON(200, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}
func Error(ctx *app.RequestContext, code int, msg string) {
	ctx.JSON(200, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
func ServerError(ctx *app.RequestContext, err error) {
	ctx.JSON(500, Response{
		Code: 500,
		Msg:  "internal server error",
	})
}
