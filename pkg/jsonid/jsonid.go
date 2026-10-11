// Package jsonid 在 JSON 编解码边界转换 int64 id：
// 响应侧把 id 类 key 的数字字面量序列化为字符串；请求侧把 id 类 key 的
// 纯数字字符串还原为数字字面量。
//
// 背景：GORM 钩子为新建行生成雪花 int64（bwmarrin/snowflake，低位不对齐），
// 值超过 2^53 后 JS JSON.parse 必然丢精度，导致前端 id 与库 id 不一致。
// 按 key 转换（而非按大小）保证同一字段类型恒定：id 类字段出站一律 string、
// 入站一律交给 Go int64 精确解析（encoding/json 对 int64 目标用原始数字解析，无精度损失）。
package jsonid

import (
	"bytes"
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// idKeys 出入站统一处理的键；数组型键的元素逐个处理
var idKeys = map[string]struct{}{
	"id": {}, "parentId": {}, "userId": {}, "deptId": {},
	"roleId": {}, "menuId": {}, "postId": {}, "leaderId": {},
	"roleIds": {}, "postIds": {}, "menuIds": {}, "deptIds": {}, "ids": {},
	"fileId": {},
}

func isIDKey(k string) bool {
	_, ok := idKeys[k]
	return ok
}

func isWS(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// isInteger 字面量是否为整数（id 只会是整数；小数不动）
func isInteger(num []byte) bool {
	i := 0
	if i < len(num) && num[i] == '-' {
		i++
	}
	if i >= len(num) {
		return false
	}
	for ; i < len(num); i++ {
		if num[i] < '0' || num[i] > '9' {
			return false
		}
	}
	return true
}

func isDigits(s []byte) bool {
	i := 0
	if i < len(s) && s[i] == '-' {
		i++
	}
	if i >= len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// scanString 返回从 data[i]（必须是 '"'）开始的字符串结束位置（闭合引号下标）与是否合法
func scanString(data []byte, i int) int {
	for j := i + 1; j < len(data); j++ {
		if data[j] == '\\' {
			j++
			continue
		}
		if data[j] == '"' {
			return j
		}
	}
	return -1
}

// TransformOut 响应侧：id 类 key 的整数 → 字符串
func TransformOut(data []byte) []byte {
	out := make([]byte, 0, len(data)+len(data)/8)
	var lastKey string
	inStr := false
	for i := 0; i < len(data); {
		c := data[i]
		if inStr {
			if c == '\\' && i+1 < len(data) {
				out = append(out, data[i], data[i+1])
				i += 2
				continue
			}
			if c == '"' {
				inStr = false
			}
			out = append(out, c)
			i++
			continue
		}
		if c == '"' {
			end := scanString(data, i)
			if end < 0 {
				return append(out, data[i:]...)
			}
			k := end + 1
			for k < len(data) && isWS(data[k]) {
				k++
			}
			if k < len(data) && data[k] == ':' {
				lastKey = string(data[i+1 : end])
			}
			out = append(out, data[i:end+1]...)
			i = end + 1
			continue
		}
		if c == '-' || (c >= '0' && c <= '9') {
			j := i
			for j < len(data) && (data[j] == '-' || data[j] == '+' || data[j] == '.' || data[j] == 'e' ||
				data[j] == 'E' || (data[j] >= '0' && data[j] <= '9')) {
				j++
			}
			num := data[i:j]
			if isIDKey(lastKey) && isInteger(num) {
				out = append(out, '"')
				out = append(out, num...)
				out = append(out, '"')
			} else {
				out = append(out, num...)
			}
			// 不重置 lastKey：数组型 id 键（roleIds 等）的后续元素仍归属同一 key
			i = j
			continue
		}
		out = append(out, c)
		i++
	}
	return out
}

// TransformIn 请求侧：id 类 key 的纯数字字符串 → 数字
func TransformIn(data []byte) []byte {
	out := make([]byte, 0, len(data))
	var lastKey string
	var strStart = -1 // 字符串内容起始（开引号之后）；-1 = 不在字符串内
	for i := 0; i < len(data); {
		c := data[i]
		if strStart >= 0 {
			if c == '\\' && i+1 < len(data) {
				i += 2
				continue
			}
			if c == '"' {
				raw := data[strStart:i]
				end := i
				k := end + 1
				for k < len(data) && isWS(data[k]) {
					k++
				}
				isKey := k < len(data) && data[k] == ':'
				switch {
				case isKey:
					lastKey = string(raw)
					out = append(out, '"')
					out = append(out, raw...)
					out = append(out, '"')
				case isIDKey(lastKey) && isDigits(raw):
					// 数字字符串 → 数字（原始数字字符，无精度损失）
					out = append(out, raw...)
				default:
					out = append(out, '"')
					out = append(out, raw...)
					out = append(out, '"')
				}
				strStart = -1
				i = end + 1
				continue
			}
			i++
			continue
		}
		if c == '"' {
			strStart = i + 1
			i++
			continue
		}
		out = append(out, c)
		i++
	}
	return out
}

// Middleware 请求体 JSON 预处理：id 类 key 的字符串数字还原为数字，
// 使前端统一以字符串持有雪花 id 也能通过 int64 绑定。
func Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ct := c.Request.Header.ContentType()
		if bytes.Contains(ct, []byte("application/json")) {
			switch strings.ToUpper(string(c.Request.Method())) {
			case "POST", "PUT", "PATCH", "DELETE":
				if body, err := c.Body(); err == nil && len(body) > 0 {
					if nb := TransformIn(body); len(nb) != len(body) {
						c.Request.SetBody(nb)
					}
				}
			}
		}
		c.Next(ctx)
	}
}
