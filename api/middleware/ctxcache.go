package middleware

import (
	"context"
	"gosilite/pkg/ctxcache"

	"github.com/cloudwego/hertz/pkg/app"
)

func ContextCacheMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		c = ctxcache.Init(c)
		ctx.Next(c)
	}
}
