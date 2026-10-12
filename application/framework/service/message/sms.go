package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SmsHandler 短信通道抽象：按 supplier 分发；未接入的 supplier 返回明确错误
type SmsHandler struct{}

// baiduSmsConfig config_json（百度云短信凭据）
type baiduSmsConfig struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	TemplateId      string `json:"templateId"`
	SignName        string `json:"signName"`
	Endpoint        string `json:"endpoint"`
}

const (
	// baiduSmsDefaultEndpoint spec 默认值（legacy 路径）
	baiduSmsDefaultEndpoint = "https://sms.bj.baidubce.com/v1/sms"
	// baiduSmsV3Path 官方现行发送接口（cloud.baidu.com/doc/SMS/s/lkijy5wvf）
	baiduSmsV3Path = "/api/v3/sendSms"
	bceAuthExpiry  = 1800
)

// Send 单账号发送
func (h *SmsHandler) Send(ctx context.Context, account string, p *renderedPayload, tpl *model.SystemMessageTemplate, cfg *model.SystemMessageConfig) (bool, string, error) {
	switch strings.ToUpper(strings.TrimSpace(cfg.SupplierType)) {
	case "BAIDU":
		return sendBaiduSms(ctx, account, p, tpl, cfg)
	default:
		return false, "", errorx.New(errno.ErrMessageSupplierNotSupported, errorx.KV("supplier", cfg.SupplierType))
	}
}

// sendBaiduSms 百度云短信：BCE Auth V1 签名 + HTTPS JSON
// 核对自官方文档：
//   - 发送接口 POST /api/v3/sendSms，body {mobile, template, signatureId, contentVar(JsonObject)}
//     响应 {requestId, code:"1000", message, data:[{code,message,mobile,messageId}]}
//   - 服务域名 sms.bj.baidubce.com / sms.su.baidubce.com
//   - 认证 bce-auth-v1/...，签名仅用 host;x-bce-date 两个头域
func sendBaiduSms(ctx context.Context, account string, p *renderedPayload, tpl *model.SystemMessageTemplate, cfg *model.SystemMessageConfig) (bool, string, error) {
	bc, err := parseBaiduSmsConfig(cfg.ConfigJson)
	if err != nil {
		return false, "", err
	}
	templateID := strings.TrimSpace(tpl.TemplateID)
	if templateID == "" {
		templateID = bc.TemplateId
	}
	if templateID == "" {
		return false, "", errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "百度云短信缺少模板ID（模板或配置均未填写）"))
	}
	signName := strings.TrimSpace(tpl.Signature)
	if signName == "" {
		signName = bc.SignName
	}
	if signName == "" {
		return false, "", errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "百度云短信缺少签名ID（模板或配置均未填写）"))
	}

	endpoint := strings.TrimSpace(bc.Endpoint)
	if endpoint == "" {
		endpoint = baiduSmsDefaultEndpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return false, "", errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "endpoint 不是合法 URL"))
	}

	body, err := buildBaiduSmsBody(endpoint, account, templateID, signName, tpl, p)
	if err != nil {
		return false, "", err
	}

	respBody, err := doBaiduSmsRequest(ctx, bc, u, body)
	if err != nil {
		return false, respBody, err
	}
	return true, respBody, nil
}

func parseBaiduSmsConfig(raw string) (*baiduSmsConfig, error) {
	var cfg baiduSmsConfig
	if strings.TrimSpace(raw) == "" {
		return nil, errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "短信配置为空"))
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "短信配置不是合法 JSON"))
	}
	if cfg.AccessKeyId == "" || cfg.SecretAccessKey == "" {
		return nil, errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "短信配置缺少 accessKeyId/secretAccessKey"))
	}
	return &cfg, nil
}

// buildBaiduSmsBody 按 endpoint 路径选择请求体：
//   - /api/v3/sendSms（官方现行）：{mobile, template, signatureId, contentVar(object)}
//   - 其他（spec 默认 legacy /v1/sms）：{phoneNumber, templateId, signature, contentVar(JSON字符串)}
func buildBaiduSmsBody(endpoint, account, templateID, signName string, tpl *model.SystemMessageTemplate, p *renderedPayload) ([]byte, error) {
	if tpl.TemplateMode == "TEMPLATE_CONTENT" {
		return nil, errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "百度云短信仅支持模板ID模式（TEMPLATE_ID）"))
	}
	u, _ := url.Parse(endpoint)
	if strings.Contains(u.Path, baiduSmsV3Path) {
		return json.Marshal(map[string]any{
			"mobile":      account,
			"template":    templateID,
			"signatureId": signName,
			"contentVar":  p.OutputVars,
		})
	}
	contentVar, err := json.Marshal(p.OutputVars)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"phoneNumber": account,
		"templateId":  templateID,
		"signature":   signName,
		"contentVar":  string(contentVar),
	})
}

// doBaiduSmsRequest 发送请求并校验响应
func doBaiduSmsRequest(ctx context.Context, bc *baiduSmsConfig, u *url.URL, body []byte) (string, error) {
	now := time.Now().UTC()
	date := now.Format("2006-01-02T15:04:05Z")
	headers := map[string]string{
		"host":       u.Host,
		"x-bce-date": date,
	}
	auth := bceAuthString(bc.AccessKeyId, bc.SecretAccessKey, now, bceAuthExpiry, http.MethodPost, u.Path, u.RawQuery, headers)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "构造请求失败: "+err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-bce-date", date)
	req.Header.Set("Authorization", auth)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "请求百度云短信失败: "+err.Error()))
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	respBody := string(raw)

	var parsed struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return truncate(respBody, 1000), errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", fmt.Sprintf("响应解析失败(HTTP %d)", resp.StatusCode)))
	}
	if parsed.Code != "1000" {
		return truncate(respBody, 1000), errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "百度云短信返回 "+parsed.Code+": "+parsed.Message))
	}
	return truncate(respBody, 1000), nil
}
