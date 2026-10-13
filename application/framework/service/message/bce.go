package message

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BCE Auth V1（百度智能云鉴权，https://cloud.baidu.com/doc/Reference/s/njwvz1yfu）
// 认证字符串：bce-auth-v1/{accessKeyId}/{timestamp}/{expirationPeriodInSeconds}/{signedHeaders}/{signature}
// SigningKey  = HMAC-SHA256-HEX(sk, authStringPrefix)
// signature   = HMAC-SHA256-HEX(SigningKey, CanonicalRequest)
// CanonicalRequest = Method \n CanonicalURI \n CanonicalQueryString \n CanonicalHeaders

const bceAuthVersion = "bce-auth-v1"

// uriEncode RFC 3986 percent-encoding（非保留字符原样，十六进制大写）
func uriEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '_', c == '-', c == '~', c == '.':
			b.WriteByte(c)
		case c == '/':
			if encodeSlash {
				b.WriteString("%2F")
			} else {
				b.WriteByte(c)
			}
		default:
			b.WriteString("%")
			b.WriteString(strings.ToUpper(hex.EncodeToString([]byte{c})))
		}
	}
	return b.String()
}

// uriEncodeExceptSlash 路径编码（斜杠不编码）
func uriEncodeExceptSlash(s string) string { return uriEncode(s, false) }

// bceCanonicalHeaders 规范头域：
// 名称小写、值去首尾空白，转为 UriEncode(name):UriEncode(value)；
// 输出行按整串字典序以 \n 连接；signedHeaders 按名称字典序以 ; 连接。
func bceCanonicalHeaders(headers map[string]string) (canonical string, signedHeaders string) {
	type hdr struct{ enc, name string }
	list := make([]hdr, 0, len(headers))
	names := make([]string, 0, len(headers))
	for name, value := range headers {
		n := strings.ToLower(strings.TrimSpace(name))
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		list = append(list, hdr{enc: uriEncode(n, true) + ":" + uriEncode(v, true), name: n})
		names = append(names, n)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].enc < list[j].enc })
	sort.Strings(names)
	lines := make([]string, 0, len(list))
	for _, h := range list {
		lines = append(lines, h.enc)
	}
	return strings.Join(lines, "\n"), strings.Join(names, ";")
}

// bceCanonicalRequest 规范请求
func bceCanonicalRequest(method, path, query string, headers map[string]string) (canonicalRequest, signedHeaders string) {
	ch, sh := bceCanonicalHeaders(headers)
	return method + "\n" + uriEncodeExceptSlash(path) + "\n" + query + "\n" + ch, sh
}

// bceSigningKey 派生签名密钥（hex 小写）
func bceSigningKey(sk, authStringPrefix string) string {
	return hmacHex(sk, authStringPrefix)
}

// bceSignature 签名摘要（hex 小写）
func bceSignature(signingKey, canonicalRequest string) string {
	return hmacHex(signingKey, canonicalRequest)
}

func hmacHex(key, message string) string {
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(message))
	return hex.EncodeToString(m.Sum(nil))
}

// bceAuthStringPrefix 前缀字符串 bce-auth-v1/{ak}/{UTC时间}/{有效期秒}
func bceAuthStringPrefix(accessKeyID string, ts time.Time, expirationSeconds int) string {
	return bceAuthVersion + "/" + accessKeyID + "/" + ts.UTC().Format("2006-01-02T15:04:05Z") + "/" + strconv.Itoa(expirationSeconds)
}

// bceAuthString 生成完整 Authorization 认证字符串
func bceAuthString(accessKeyID, secretAccessKey string, ts time.Time, expirationSeconds int, method, path, query string, headers map[string]string) string {
	prefix := bceAuthStringPrefix(accessKeyID, ts, expirationSeconds)
	canonical, signedHeaders := bceCanonicalRequest(method, path, query, headers)
	signature := bceSignature(bceSigningKey(secretAccessKey, prefix), canonical)
	return prefix + "/" + signedHeaders + "/" + signature
}
