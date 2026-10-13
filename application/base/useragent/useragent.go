// Package useragent 从 User-Agent 轻量解析浏览器与操作系统。
// 供 operlog 日志采集与在线用户会话登记共用（application 内共享，
// api 层可引用；反向禁止）。
package useragent

import "strings"

// Parse 返回浏览器与操作系统名（包含匹配，未识别返回空串）。
func Parse(ua string) (browser, osName string) {
	switch {
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	case strings.Contains(ua, "curl/"):
		browser = "curl"
	case strings.Contains(ua, "Postman"):
		browser = "Postman"
	}
	switch {
	case strings.Contains(ua, "Windows"):
		osName = "Windows"
	case strings.Contains(ua, "Android"):
		osName = "Android"
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"):
		osName = "iOS"
	case strings.Contains(ua, "Mac OS X"), strings.Contains(ua, "Macintosh"):
		osName = "macOS"
	case strings.Contains(ua, "Linux"):
		osName = "Linux"
	}
	return browser, osName
}
