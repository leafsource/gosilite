package response

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gosilite/pkg/errorx"
	"gosilite/pkg/jsonid"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Result struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// writeJSON 手动序列化并经 jsonid 转换（int64 id → 字符串），再写回响应
func writeJSON(c *app.RequestContext, status int, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		hlog.Errorf("[Response] marshal failed: %v", err)
		c.SetStatusCode(http.StatusInternalServerError)
		c.Response.Header.SetContentType("application/json; charset=utf-8")
		c.SetBodyString(`{"code":500,"msg":"internal server error"}`)
		return
	}
	c.SetStatusCode(status)
	c.Response.Header.SetContentType("application/json; charset=utf-8")
	c.SetBodyString(string(jsonid.TransformOut(b)))
}

func InvalidParamRequestResponse(c *app.RequestContext, errMsg string) {
	c.Abort()
	writeJSON(c, http.StatusBadRequest, Result{Code: http.StatusBadRequest, Msg: errMsg})
}

func Unauthorized(c *app.RequestContext, errMsg string) {
	c.Abort()
	writeJSON(c, http.StatusUnauthorized, Result{Code: http.StatusUnauthorized, Msg: errMsg})
}

func Error(ctx context.Context, c *app.RequestContext, err error) {
	var customErr errorx.StatusError

	if errors.As(err, &customErr) && customErr.Code() != 0 {
		hlog.CtxWarnf(ctx, "[ErrorX] error:  %v %v \n", customErr.Code(), err)
		c.Abort()
		writeJSON(c, http.StatusOK, Result{Code: customErr.Code(), Msg: customErr.Msg()})
		return
	}

	hlog.CtxErrorf(ctx, "[InternalError]  error: %v \n", err)
	c.Abort()
	writeJSON(c, http.StatusInternalServerError, Result{Code: 500, Msg: "internal server error"})
}

// SendErrResponse  pack error response
func Failed(ctx context.Context, c *app.RequestContext, code int32, err error) {
	writeJSON(c, consts.StatusOK, &Result{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	})
}

// SendSuccessResponse  pack success response
func SuccessWithData(ctx context.Context, c *app.RequestContext, data interface{}) {
	writeJSON(c, consts.StatusOK, &Result{
		Code: consts.StatusOK,
		Msg:  "success",
		Data: data,
	})
}
