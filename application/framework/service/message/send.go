package message

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SendReq 发送请求：TemplateID 或 (MessageKey + MessageType) 二选一
type SendReq struct {
	TemplateID  int64
	MessageKey  string
	MessageType string
	Accounts    []string
	Vars        map[string]string
}

// SendResult 单账号发送结果（每次发送都会落一条 system_message_log）
type SendResult struct {
	Account      string
	IsSuccess    bool
	ResponseBody string
	// CostTime 毫秒
	CostTime int64
	LogID    int64
}

// sendHandler 通道发送处理（按 config.message_type 分发）
type sendHandler interface {
	Send(ctx context.Context, account string, p *renderedPayload, tpl *model.SystemMessageTemplate, cfg *model.SystemMessageConfig) (success bool, responseBody string, err error)
}

// renderedPayload 渲染后的消息体
type renderedPayload struct {
	// Title 已按输入变量渲染的标题
	Title string
	// Content 已按模板变量渲染的正文
	Content string
	// OutputVars 模板变量渲染结果（短信模板ID模式使用）
	OutputVars map[string]string
}

// Send 按模板发送；无论成败每个账号都写发送记录。
// 模板/配置查找阶段失败返回 error 不落记录（无模板信息可记）；
// 渲染或通道发送失败先落 is_success=0 再返回 error。
func (s *Service) Send(ctx context.Context, req SendReq) ([]SendResult, error) {
	accounts := make([]string, 0, len(req.Accounts))
	for _, a := range req.Accounts {
		if a = strings.TrimSpace(a); a != "" {
			accounts = append(accounts, a)
		}
	}
	if len(accounts) == 0 {
		return nil, errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "发送账号不能为空"))
	}

	tpl, err := s.resolveTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	// 查配置：失败也要落记录（模板信息已具备）
	cfg, cfgErr := s.repo.ConfigGet(ctx, tpl.MessageConfigID)
	if cfgErr != nil || cfg == nil {
		failure := errorx.New(errno.ErrMessageConfigNotFound, errorx.KV("id", formatID(tpl.MessageConfigID)))
		s.writeLogs(ctx, tpl, nil, accounts, "", false, errorx.ErrorWithoutStack(failure), 0)
		return nil, failure
	}
	if cfg.Status != 1 {
		failure := errorx.New(errno.ErrMessageSendFailed, errorx.KV("reason", "消息配置已停用"))
		s.writeLogs(ctx, tpl, cfg, accounts, "", false, errorx.ErrorWithoutStack(failure), 0)
		return nil, failure
	}

	payload, err := render(tpl, req.Vars)
	if err != nil {
		s.writeLogs(ctx, tpl, cfg, accounts, "", false, errorx.ErrorWithoutStack(err), 0)
		return nil, err
	}

	handler, err := handlerFactory(cfg.MessageType)
	if err != nil {
		s.writeLogs(ctx, tpl, cfg, accounts, payload, false, errorx.ErrorWithoutStack(err), 0)
		return nil, err
	}

	results := make([]SendResult, 0, len(accounts))
	var firstErr error
	for _, account := range accounts {
		start := time.Now()
		success, respBody, sendErr := handler.Send(ctx, account, payload, tpl, cfg)
		cost := time.Since(start).Milliseconds()
		logID := s.writeLogs(ctx, tpl, cfg, []string{account}, payload, success, firstNonEmpty(sendErr, respBody, "OK"), cost)
		res := SendResult{
			Account:      account,
			IsSuccess:    success,
			ResponseBody: respBody,
			CostTime:     cost,
			LogID:        logID,
		}
		if sendErr != nil && !success {
			res.ResponseBody = errorx.ErrorWithoutStack(sendErr)
			if firstErr == nil {
				firstErr = sendErr
			}
		}
		results = append(results, res)
	}
	return results, firstErr
}

func (s *Service) resolveTemplate(ctx context.Context, req SendReq) (*model.SystemMessageTemplate, error) {
	if req.TemplateID > 0 {
		tpl, err := s.repo.TemplateGetEnabled(ctx, req.TemplateID)
		if err != nil || tpl == nil {
			return nil, errorx.New(errno.ErrMessageTemplateNotFound, errorx.KV("id", formatID(req.TemplateID)))
		}
		return tpl, nil
	}
	key := strings.TrimSpace(req.MessageKey)
	msgType := strings.ToUpper(strings.TrimSpace(req.MessageType))
	if key == "" {
		return nil, errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "TemplateID 或 Key+Type 必须指定其一"))
	}
	if msgType != "MAIL" && msgType != "SMS" {
		return nil, errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "按 Key 发送必须指定 messageType（MAIL/SMS）"))
	}
	tpl, err := s.repo.TemplateGetEnabledByKey(ctx, key, msgType)
	if err != nil || tpl == nil {
		return nil, errorx.New(errno.ErrMessageTemplateNotFound, errorx.KV("id", key))
	}
	return tpl, nil
}

// handlerFactory 通道分发入口（单测可替换）
var handlerFactory = defaultHandlerFor

func defaultHandlerFor(messageType string) (sendHandler, error) {
	switch strings.ToUpper(strings.TrimSpace(messageType)) {
	case "MAIL":
		return &MailHandler{}, nil
	case "SMS":
		return &SmsHandler{}, nil
	default:
		return nil, errorx.New(errno.ErrMessageSupplierNotSupported, errorx.KV("supplier", messageType))
	}
}

// writeLogs 逐账号写发送记录，返回最后写入的 log id（失败返回 0）
func (s *Service) writeLogs(ctx context.Context, tpl *model.SystemMessageTemplate, cfg *model.SystemMessageConfig, accounts []string, payload any, success bool, respBody string, cost int64) int64 {
	var (
		title, content string
		outVars        map[string]string
	)
	switch v := payload.(type) {
	case *renderedPayload:
		if v != nil {
			title, content, outVars = v.Title, v.Content, v.OutputVars
		}
	case string:
		content = v
	}
	_ = outVars
	var lastID int64
	isSuccess := int16(0)
	if success {
		isSuccess = 1
	}
	for _, account := range accounts {
		row := &model.SystemMessageLog{
			MessageTemplateID:   tpl.ID,
			MessageKey:          tpl.MessageKey,
			MessageTemplateName: tpl.TemplateName,
			MessageType:         tpl.MessageType,
			TemplateMode:        tpl.TemplateMode,
			Account:             account,
			Title:               title,
			TemplateID:          tpl.TemplateID,
			Content:             content,
			SupplierType:        tplSuppliedBy(cfg),
			IsSuccess:           &isSuccess,
			ResponseBody:        truncate(respBody, 1000),
			LogTime:             time.Now(),
		}
		if cfg != nil {
			row.MessageConfigTitle = cfg.Title
		}
		c := cost
		row.CostTime = &c
		if err := s.repo.LogInsert(ctx, row); err == nil {
			lastID = row.ID
		}
	}
	return lastID
}

func tplSuppliedBy(cfg *model.SystemMessageConfig) string {
	if cfg == nil {
		return ""
	}
	return cfg.SupplierType
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func firstNonEmpty(err error, ss ...string) string {
	if err != nil {
		return errorx.ErrorWithoutStack(err)
	}
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// ========== 渲染（占位符 ${var} / ${var:default}，对齐 ruoyi BaseMessageSendHandler） ==========

var placeholderRe = regexp.MustCompile(`\$\{([^{}]+)\}`)

type templateVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// replacePlaceholders 替换 ${key} 与 ${key:default}；不可解析的占位符原样保留
func replacePlaceholders(s string, lookup func(key string) (string, bool)) string {
	return placeholderRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[2 : len(match)-1]
		key, def := inner, ""
		hasDef := false
		if i := strings.Index(inner, ":"); i >= 0 {
			key, def, hasDef = inner[:i], inner[i+1:], true
		}
		if v, ok := lookup(key); ok {
			return v
		}
		if hasDef {
			return def
		}
		return match
	})
}

func inputLookup(vars map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := vars[key]
		return v, ok
	}
}

// renderOutputVars 解析 vars_json 并把变量值中的占位符替换为输入变量（保持声明顺序）
func renderOutputVars(varsJSON string, input map[string]string) (map[string]string, error) {
	out := make(map[string]string)
	raw := strings.TrimSpace(varsJSON)
	if raw == "" || raw == "null" {
		return out, nil
	}
	var vars []templateVar
	if err := json.Unmarshal([]byte(raw), &vars); err != nil {
		// 兼容对象形状 {"k":"v"}（宽松解析）
		var obj map[string]string
		if err2 := json.Unmarshal([]byte(raw), &obj); err2 != nil {
			return nil, errorx.New(errno.ErrMessageInvalidParam, errorx.KV("msg", "varsJson 不是合法的变量数组"))
		}
		for k, v := range obj {
			out[k] = replacePlaceholders(v, inputLookup(input))
		}
		return out, nil
	}
	for _, v := range vars {
		if v.Key == "" {
			continue
		}
		out[v.Key] = replacePlaceholders(v.Value, inputLookup(input))
	}
	return out, nil
}

// render 渲染标题/正文/模板变量（title 用输入变量，content 用输出变量，语义对齐 ruoyi）
func render(tpl *model.SystemMessageTemplate, input map[string]string) (*renderedPayload, error) {
	if input == nil {
		input = map[string]string{}
	}
	outVars, err := renderOutputVars(tpl.VarsJson, input)
	if err != nil {
		return nil, err
	}
	outLookup := func(key string) (string, bool) {
		v, ok := outVars[key]
		return v, ok
	}
	return &renderedPayload{
		Title:      replacePlaceholders(tpl.Title, inputLookup(input)),
		Content:    replacePlaceholders(tpl.Content, outLookup),
		OutputVars: outVars,
	}, nil
}
