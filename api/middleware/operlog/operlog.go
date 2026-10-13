// Package operlog 是日志采集中间件：在 requestid 之后、recovery 之前注册，
// 请求结束后做一次轻量分类（操作 / 异常 / 登录），投递到有界异步队列。
// 热路径不做 JSON 反序列化、不阻塞、不分配大对象（报文 copy 仅写请求发生）。
package operlog

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gosilite/api/middleware/permission"
	"gosilite/application/base/ctxutil"
	"gosilite/application/base/logq"
	"gosilite/application/base/useragent"
	"gosilite/pkg/errorx"
	"gosilite/types/consts"

	"github.com/cloudwego/hertz/pkg/app"
)

// PanicKey 供 recovery 中间件回写 panic 信息（见 main.registerMiddleware）。
const PanicKey = "operlog_panic"

// PanicInfo recovery 捕获的 panic 上下文。
type PanicInfo struct {
	Err   any
	Stack []byte
}

const (
	maxBodyBytes   = 16 << 10 // 报文落库上限 16KB
	maxHeaderBytes = 8 << 10  // header 序列化上限
	pathLogin      = "/user/login"
	pathLogout     = "/user/logout"
)

// MW 采集请求日志并分类入队。
func MW() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		method := string(c.Request.Header.Method())
		// 分类/skip 用纯 path；落库 URI 为原始请求行（path + query，查询参数对排错有价值）
		reqPath := string(c.URI().Path())
		reqURI := string(c.URI().RequestURI())
		if skipPath(reqPath, method) {
			c.Next(ctx)
			return
		}

		start := time.Now()
		c.Next(ctx)
		cost := int32(time.Since(start).Milliseconds())

		s := logq.Current()
		if s == nil {
			return
		}
		now := time.Now()
		traceID := traceIDOf(c)
		userID, userName := int64(0), ""
		if u := ctxutil.GetLoginUser(ctx); u != nil {
			userID, userName = u.ID, u.Username
		}
		clientIP := c.ClientIP()
		httpStatus := c.Response.StatusCode()
		code, msg, hasEnvelope := peekEnvelope(c.Response.Body())

		// ===== 登录/登出专表：request body 永不落库 =====
		isLogout := reqPath == pathLogout && method == "POST"
		if (reqPath == pathLogin || isLogout) && method == "POST" {
			if userID == 0 {
				if v, ok := c.Get(permission.LoginUserIDKey); ok {
					if id, ok := v.(int64); ok {
						userID = id
					}
				}
			}
			// 登录：用户名从请求体提取（账号名）；登出：身份已由会话解析，
			// 同样取账号（Account）以与登录行口径一致
			if isLogout {
				if u := ctxutil.GetLoginUser(ctx); u != nil && u.Account != "" {
					userName = u.Account
				}
			}
			if userName == "" && !isLogout {
				userName = peekLoginUsername(c.Request.Body())
			}
			status := int8(0)
			fail := ""
			if hasEnvelope && code == httpStatusOK && httpStatus < 400 {
				if isLogout {
					status = 2 // 登出成功
				}
			} else {
				status = 1
				fail = msg
				if fail == "" {
					if isLogout {
						fail = "登出失败"
					} else {
						fail = "登录失败"
					}
				}
			}
			ua := string(c.GetHeader("User-Agent"))
			browser, osName := useragent.Parse(ua)
			if len(ua) > 512 {
				ua = ua[:512]
			}
			s.EnqueueLogin(&logq.LoginEntry{
				TraceID: traceID, UserID: userID, UserName: userName,
				Status: status, FailReason: fail,
				ClientIP: clientIP, UserAgent: ua, Browser: browser, OS: osName,
				CreatedAt: now,
			})
			return
		}

		// ===== panic（recovery 已写响应并回填 PanicKey） =====
		var panicInfo *PanicInfo
		if v, ok := c.Get(PanicKey); ok {
			panicInfo, _ = v.(*PanicInfo)
		}
		businessStack := ""
		if v, ok := c.Get(errorx.RequestStackKey); ok {
			businessStack, _ = v.(string)
		}

		headers := dumpHeaders(c)
		operStatus, toErrorLog := classify(httpStatus, code, hasEnvelope, panicInfo != nil)

		// 操作日志恒记（panic 也留操作痕迹）
		entry := &logq.OperEntry{
			TraceID: traceID, UserID: userID, UserName: userName,
			OperType:   operTypeOf(method, reqPath),
			RequestURI: reqURI, HTTPMethod: method, Module: moduleOf(reqPath),
			OperStatus:     operStatus,
			RequestHeaders: headers,
			Remark:         msg,
			CostMs:         cost, ClientIP: clientIP, CreatedAt: now,
		}
		if operStatus == 1 && msg != "" {
			entry.ErrorMsg = msg
		}
		// 报文全量落库（含查询类 GET；压缩响应除外）；登录/登出已走专表短路
		if bodyReadable(c) {
			entry.RequestBody = clipAndMask(c.Request.Body())
			entry.ResponseBody = clipAndMask(c.Response.Body())
		}
		s.EnqueueOper(entry)

		if toErrorLog {
			e := &logq.ErrorEntry{
				TraceID: traceID, UserID: userID, UserName: userName,
				ErrType: 1, ErrLevel: 1, ErrCode: code, ErrMsg: msg,
				RequestURI: reqURI, HTTPMethod: method, Module: moduleOf(reqPath),
				RequestHeaders: headers,
				CostMs:         cost, ClientIP: clientIP, CreatedAt: now,
			}
			if bodyReadable(c) {
				e.RequestBody = clipAndMask(c.Request.Body())
			}
			if panicInfo != nil {
				e.ErrType, e.ErrLevel = 2, 2
				e.ErrMsg = "panic: " + panicText(panicInfo.Err)
				e.Stack = clipStack(panicInfo.Stack)
				if e.ErrCode == 0 {
					e.ErrCode = 500
				}
			} else if businessStack != "" {
				e.Stack = clipStack([]byte(businessStack))
			} else if e.ErrMsg == "" {
				e.ErrMsg = "http status " + itoa(httpStatus)
			}
			s.EnqueueError(e)
		}
	}
}

const httpStatusOK = 200

// skipPath 过滤健康检查 / 文档 / 预检。
func skipPath(path, method string) bool {
	if method == "OPTIONS" {
		return true
	}
	if path == "/ping" || path == "/" {
		return true
	}
	return strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/metrics") || strings.HasPrefix(path, "/docs")
}

func traceIDOf(c *app.RequestContext) string {
	if v := c.Response.Header.Peek(consts.RequestHeaderIdKey); len(v) > 0 {
		return string(v)
	}
	return string(c.Request.Header.Peek(consts.RequestHeaderIdKey))
}

// classify 决定操作状态与是否写异常日志。
// 客户端拒绝（400/401/403/404、参数/鉴权类 code）只进操作日志失败态，不进异常表。
func classify(httpStatus int, code int32, hasEnvelope, panicked bool) (int8, bool) {
	success := hasEnvelope && code == httpStatusOK && httpStatus < 400
	operStatus := int8(0)
	if !success {
		operStatus = 1
	}
	if panicked {
		return operStatus, true
	}
	if success {
		return operStatus, false
	}
	switch httpStatus {
	case 400, 401, 403, 404:
		return operStatus, false
	}
	switch code {
	case 0, 400, 401, 403, 404, 700012006: // 400 参数 / 401 认证 / 403 禁止 / 404 / ErrUserAuthenticationFailed
		return operStatus, false
	}
	return operStatus, true
}

// peekEnvelope 从响应体字节级提取 {code,msg}，避免整包 unmarshal。
func peekEnvelope(body []byte) (int32, string, bool) {
	if len(body) == 0 || body[0] != '{' {
		return 0, "", false
	}
	var code int32
	found := false
	if i := bytes.Index(body, []byte(`"code":`)); i >= 0 {
		p := body[i+7:]
		j := 0
		for j < len(p) && p[j] >= '0' && p[j] <= '9' {
			j++
		}
		if j > 0 {
			v := int32(0)
			for k := 0; k < j; k++ {
				v = v*10 + int32(p[k]-'0')
			}
			code, found = v, true
		}
	}
	msg := ""
	if i := bytes.Index(body, []byte(`"msg":"`)); i >= 0 {
		p := body[i+7:]
		if end := bytes.IndexByte(p, '"'); end > 0 {
			if end > 512 {
				end = 512
			}
			msg = unescapeJSONString(p[:end])
		}
	}
	return code, msg, found
}

// bodyReadable 压缩响应不做报文落库（读到的是压缩流）。
func bodyReadable(c *app.RequestContext) bool {
	return len(c.Response.Header.Peek("Content-Encoding")) == 0
}

// dumpHeaders 序列化请求头，敏感头打码，超限截断。
func dumpHeaders(c *app.RequestContext) string {
	var b strings.Builder
	b.Grow(256)
	c.Request.Header.VisitAll(func(k, v []byte) {
		if b.Len() >= maxHeaderBytes {
			return
		}
		key := string(k)
		if isSensitiveHeader(key) {
			v = []byte("******")
		}
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(string(v))
		b.WriteByte('\n')
	})
	return b.String()
}

func isSensitiveHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "cookie", "set-cookie", "x-api-key":
		return true
	}
	return false
}

// clipAndMask 截断到 16KB 并对敏感字段字节级掩码。
func clipAndMask(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if len(b) > maxBodyBytes {
		b = b[:maxBodyBytes]
	}
	return string(maskSecrets(b))
}

var secretKeys = [][]byte{
	[]byte(`"password":`),
	[]byte(`"oldPassword":`),
	[]byte(`"newPassword":`),
	[]byte(`"confirmPassword":`),
	[]byte(`"accessToken":`),
	[]byte(`"refreshToken":`),
}

type maskRange struct{ start, end int }

// maskSecrets 将敏感 key 的值替换为 ******（字节级，不做 JSON 解析）。
func maskSecrets(src []byte) []byte {
	var ranges []maskRange
	for _, k := range secretKeys {
		from := 0
		for from < len(src) {
			i := bytes.Index(src[from:], k)
			if i < 0 {
				break
			}
			i += from
			vs := i + len(k)
			p := vs
			for p < len(src) && (src[p] == ' ' || src[p] == '\t') {
				p++
			}
			ve := -1
			if p < len(src) && src[p] == '"' {
				q := p + 1
				for q < len(src) {
					if src[q] == '\\' {
						q += 2
						continue
					}
					if src[q] == '"' {
						ve = q + 1
						break
					}
					q++
				}
			} else if p < len(src) {
				q := p
				for q < len(src) && src[q] != ',' && src[q] != '}' && src[q] != ']' {
					q++
				}
				ve = q
			}
			if ve < 0 {
				from = vs
				continue
			}
			ranges = append(ranges, maskRange{start: p, end: ve})
			from = ve
		}
	}
	if len(ranges) == 0 {
		return src
	}
	// 合并构建（区间通常 0~2 个）
	out := make([]byte, 0, len(src))
	last := 0
	for _, r := range ranges {
		if r.start < last {
			continue
		}
		out = append(out, src[last:r.start]...)
		out = append(out, []byte("******")...)
		last = r.end
	}
	out = append(out, src[last:]...)
	return out
}

// unescapeJSONString 解 JSON 字符串转义（主要处理 \uXXXX 中文消息）。
func unescapeJSONString(s []byte) string {
	if bytes.IndexByte(s, '\\') < 0 {
		return string(s)
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'u':
			if i+4 < len(s) {
				v := 0
				for k := 1; k <= 4; k++ {
					v = v*16 + hexVal(s[i+k])
				}
				b.WriteRune(rune(v))
				i += 4
			}
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

// peekLoginUsername 从登录请求体提取用户名（不落报文）。
func peekLoginUsername(body []byte) string {
	if i := bytes.Index(body, []byte(`"username":"`)); i >= 0 {
		p := body[i+12:]
		if j := bytes.IndexByte(p, '"'); j > 0 && j <= 64 {
			return string(p[:j])
		}
	}
	if i := bytes.Index(body, []byte("username=")); i >= 0 {
		p := body[i+9:]
		j := bytes.IndexAny(p, "&; ")
		if j < 0 {
			j = len(p)
		}
		if j > 0 && j <= 64 {
			return string(p[:j])
		}
	}
	return ""
}

func operTypeOf(method, path string) int8 {
	if strings.Contains(path, "/export") {
		return 6
	}
	if strings.Contains(path, "/import") {
		return 5
	}
	if strings.Contains(path, "/upload") {
		return 7
	}
	switch method {
	case "POST":
		return 1
	case "PUT", "PATCH":
		return 2
	case "DELETE":
		return 3
	case "GET":
		return 4
	}
	return 0
}

func moduleOf(path string) string {
	p := strings.TrimPrefix(path, "/")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return p
}

func clipStack(stack []byte) string {
	const maxStack = 8 << 10
	if len(stack) > maxStack {
		stack = stack[:maxStack]
	}
	return string(stack)
}

func panicText(v any) string {
	// 冷路径（异常分支）才做格式化
	s := fmt.Sprint(v)
	if len(s) > 512 {
		s = s[:512]
	}
	return s
}

func itoa(v int) string { return strconv.Itoa(v) }
