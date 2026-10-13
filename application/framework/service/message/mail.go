package message

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// MailHandler 邮件通道（Go 标准库 net/smtp 真实发送）
type MailHandler struct{}

// mailConfig config_json（字段语义对齐 ruoyi MailFieldConfigs / hutool MailAccount）
type mailConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Auth           *bool  `json:"auth"`
	User           string `json:"user"`
	Pass           string `json:"pass"`
	From           string `json:"from"`
	StarttlsEnable *bool  `json:"starttlsEnable"`
	SslEnable      *bool  `json:"sslEnable"`
}

// fromPattern 对齐 ruoyi MailFieldConfigs.from 的 FormRule.pattern
var fromPattern = regexp.MustCompile(`^([^ \f\n\r\t\\v<]+@[^ \f\n\r\t\\v<>]+\.[^ \f\n\r\t\\v>]+$)|(\S+ <[^ \f\n\r\t\\v<]+@[^ \f\n\r\t\\v<>]+\.[^ \f\n\r\t\\v>]+>$)`)

// htmlPattern 对齐 ruoyi MailMessageSendHandler.HTML 正则
var htmlPattern = regexp.MustCompile(`.*((<[a-zA-Z]+[^>]*>[^<>]*</[a-zA-Z]+>)+|(<[a-zA-Z]+[^>]*/>)+).*`)

func parseMailConfig(raw string) (*mailConfig, error) {
	var cfg mailConfig
	if strings.TrimSpace(raw) == "" {
		return nil, errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "邮件配置为空"))
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "邮件配置不是合法 JSON"))
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *mailConfig) validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "SMTP服务器域名不能为空"))
	}
	if c.Port == 0 {
		c.Port = 465
	}
	if c.Port < 1 || c.Port > 65535 {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", fmt.Sprintf("SMTP端口不合法: %d", c.Port)))
	}
	if strings.TrimSpace(c.From) == "" {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "发送方不能为空"))
	}
	if !fromPattern.MatchString(c.From) {
		return errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "发送方格式错误"))
	}
	return nil
}

// implicitTLS 465 端口或 sslEnable=true 走隐式 TLS（对齐 spec S2.4 / hutool sslEnable 语义）
func (c *mailConfig) implicitTLS() bool {
	if c.Port == 465 {
		return true
	}
	return c.SslEnable != nil && *c.SslEnable
}

// starttls 是否启用 STARTTLS（缺省 true，对齐 ruoyi starttlsEnable 默认值）
func (c *mailConfig) starttls() bool {
	if c.StarttlsEnable != nil {
		return *c.StarttlsEnable
	}
	return true
}

// needAuth 是否需要用户名密码（显式 auth 优先，缺省按是否填写 user 判断）
func (c *mailConfig) needAuth() bool {
	if c.Auth != nil {
		return *c.Auth
	}
	return c.User != ""
}

// isHTML 对齐 ruoyi MailMessageSendHandler.isHtml（整串匹配）
func isHTML(content string) bool {
	return htmlPattern.MatchString(content)
}

// Send 单账号发送
func (h *MailHandler) Send(ctx context.Context, account string, p *renderedPayload, _ *model.SystemMessageTemplate, cfg *model.SystemMessageConfig) (bool, string, error) {
	mc, err := parseMailConfig(cfg.ConfigJson)
	if err != nil {
		return false, "", err
	}
	addr := fmt.Sprintf("%s:%d", mc.Host, mc.Port)

	c, err := dialSMTP(ctx, mc, addr)
	if err != nil {
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "连接SMTP服务器失败: "+err.Error()))
	}
	defer c.Close()

	if mc.starttls() && !mc.implicitTLS() {
		if err := c.StartTLS(&tls.Config{ServerName: mc.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "STARTTLS失败: "+err.Error()))
		}
	}
	if mc.needAuth() {
		if mc.User == "" {
			return false, "", errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "已启用验证但用户名为空"))
		}
		auth := smtp.PlainAuth("", mc.User, mc.Pass, mc.Host)
		if err := c.Auth(auth); err != nil {
			return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "SMTP认证失败: "+err.Error()))
		}
	}
	if err := c.Mail(envelopeAddress(mc.From)); err != nil {
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "MAIL FROM失败: "+err.Error()))
	}
	if err := c.Rcpt(account); err != nil {
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "RCPT TO失败: "+err.Error()))
	}
	w, err := c.Data()
	if err != nil {
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "DATA失败: "+err.Error()))
	}
	msg := buildMIME(mc.From, account, p)
	if _, err := io.WriteString(w, msg); err != nil {
		_ = w.Close()
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "写入邮件内容失败: "+err.Error()))
	}
	if err := w.Close(); err != nil {
		return false, "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "提交邮件失败: "+err.Error()))
	}
	_ = c.Quit()
	return true, "OK", nil
}

func dialSMTP(ctx context.Context, mc *mailConfig, addr string) (*smtp.Client, error) {
	dialer := net.Dialer{Timeout: 15 * time.Second}
	if mc.implicitTLS() {
		tlsCfg := &tls.Config{ServerName: mc.Host, MinVersion: tls.VersionTLS12}
		conn, err := tls.DialWithDialer(&dialer, "tcp", addr, tlsCfg)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, mc.Host)
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, mc.Host)
}

// envelopeAddress 提取 `Name <user@dom>` 中的邮箱地址作为 SMTP envelope
func envelopeAddress(from string) string {
	if i := strings.LastIndex(from, "<"); i >= 0 && strings.HasSuffix(from, ">") {
		return from[i+1 : len(from)-1]
	}
	return from
}

func buildMIME(from, to string, p *renderedPayload) string {
	contentType := "text/plain; charset=UTF-8"
	if isHTML(p.Content) {
		contentType = "text/html; charset=UTF-8"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", p.Title))
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: %s\r\n\r\n", contentType)
	b.WriteString(p.Content)
	return b.String()
}
