package util

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type ResJson map[string]interface{}
type HandlerFunc func(c context.Context, ctx app.RequestContext)
