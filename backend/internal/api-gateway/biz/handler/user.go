


type UserRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

func Register(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(consts.StatusBadRequset, utils.H{
			"msg": err.Error(),
		})
		return
	}
}