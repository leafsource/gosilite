package message

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// ===== fake 仓储 =====

type fakeMessageRepo struct {
	configs   map[int64]*model.SystemMessageConfig
	keys      map[int64]*model.SystemMessageKey
	templates map[int64]*model.SystemMessageTemplate
	logs      []*model.SystemMessageLog
	nextID    int64

	configErr error
	keyExists map[string]int64
}

func newFakeRepo() *fakeMessageRepo {
	return &fakeMessageRepo{
		configs:   map[int64]*model.SystemMessageConfig{},
		keys:      map[int64]*model.SystemMessageKey{},
		templates: map[int64]*model.SystemMessageTemplate{},
		keyExists: map[string]int64{},
		nextID:    100,
	}
}

func (f *fakeMessageRepo) ConfigPage(context.Context, int, int, string, *string, *int16) ([]*model.SystemMessageConfig, int64, error) {
	return nil, 0, nil
}
func (f *fakeMessageRepo) ConfigGet(_ context.Context, id int64) (*model.SystemMessageConfig, error) {
	if f.configErr != nil {
		return nil, f.configErr
	}
	if m, ok := f.configs[id]; ok {
		return m, nil
	}
	return nil, errors.New("record not found")
}
func (f *fakeMessageRepo) ConfigCreate(_ context.Context, m *model.SystemMessageConfig) error {
	f.nextID++
	m.ID = f.nextID
	f.configs[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) ConfigSave(_ context.Context, m *model.SystemMessageConfig) error {
	f.configs[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) ConfigDelete(_ context.Context, id int64) (int64, error) {
	if _, ok := f.configs[id]; !ok {
		return 0, nil
	}
	delete(f.configs, id)
	return 1, nil
}

func (f *fakeMessageRepo) KeyPage(context.Context, int, int, string) ([]*model.SystemMessageKey, int64, error) {
	return nil, 0, nil
}
func (f *fakeMessageRepo) KeyGet(_ context.Context, id int64) (*model.SystemMessageKey, error) {
	if m, ok := f.keys[id]; ok {
		return m, nil
	}
	return nil, errors.New("record not found")
}
func (f *fakeMessageRepo) KeyCreate(_ context.Context, m *model.SystemMessageKey) error {
	f.nextID++
	m.ID = f.nextID
	f.keys[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) KeySave(_ context.Context, m *model.SystemMessageKey) error {
	f.keys[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) KeyDelete(_ context.Context, id int64) (int64, error) {
	if _, ok := f.keys[id]; !ok {
		return 0, nil
	}
	delete(f.keys, id)
	return 1, nil
}
func (f *fakeMessageRepo) KeyExists(_ context.Context, messageKey string, excludeID int64) (bool, error) {
	id, ok := f.keyExists[messageKey]
	if !ok || id == excludeID {
		return false, nil
	}
	return true, nil
}

func (f *fakeMessageRepo) TemplatePage(context.Context, int, int, string, *string, *int64, *int16) ([]*model.SystemMessageTemplate, int64, error) {
	return nil, 0, nil
}
func (f *fakeMessageRepo) TemplateGet(_ context.Context, id int64) (*model.SystemMessageTemplate, error) {
	if m, ok := f.templates[id]; ok {
		return m, nil
	}
	return nil, errors.New("record not found")
}
func (f *fakeMessageRepo) TemplateCreate(_ context.Context, m *model.SystemMessageTemplate) error {
	f.nextID++
	m.ID = f.nextID
	f.templates[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) TemplateSave(_ context.Context, m *model.SystemMessageTemplate) error {
	f.templates[m.ID] = m
	return nil
}
func (f *fakeMessageRepo) TemplateDelete(_ context.Context, id int64) (int64, error) {
	if _, ok := f.templates[id]; !ok {
		return 0, nil
	}
	delete(f.templates, id)
	return 1, nil
}
func (f *fakeMessageRepo) TemplateGetEnabled(_ context.Context, id int64) (*model.SystemMessageTemplate, error) {
	m, ok := f.templates[id]
	if !ok || m.Status != 1 {
		return nil, errors.New("record not found")
	}
	return m, nil
}
func (f *fakeMessageRepo) TemplateGetEnabledByKey(_ context.Context, messageKey, messageType string) (*model.SystemMessageTemplate, error) {
	var hit *model.SystemMessageTemplate
	for _, m := range f.templates {
		if m.MessageKey == messageKey && m.MessageType == messageType && m.Status == 1 {
			if hit == nil || m.ID > hit.ID {
				hit = m
			}
		}
	}
	if hit == nil {
		return nil, errors.New("record not found")
	}
	return hit, nil
}

func (f *fakeMessageRepo) LogPage(context.Context, int, int, string, *string, *string, *int16) ([]*model.SystemMessageLog, int64, error) {
	return nil, 0, nil
}
func (f *fakeMessageRepo) LogGet(_ context.Context, id int64) (*model.SystemMessageLog, error) {
	for _, m := range f.logs {
		if m.ID == id {
			return m, nil
		}
	}
	return nil, errors.New("record not found")
}
func (f *fakeMessageRepo) LogInsert(_ context.Context, m *model.SystemMessageLog) error {
	f.nextID++
	m.ID = f.nextID
	f.logs = append(f.logs, m)
	return nil
}
func (f *fakeMessageRepo) LogRemove(_ context.Context, id int64) (int64, error) {
	for i, m := range f.logs {
		if m.ID == id {
			f.logs = append(f.logs[:i], f.logs[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}
func (f *fakeMessageRepo) LogClear(context.Context) (int64, error) {
	n := int64(len(f.logs))
	f.logs = nil
	return n, nil
}

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

// ===== 渲染 =====

func TestRenderOutputVarsSubstitutesInput(t *testing.T) {
	tpl := &model.SystemMessageTemplate{
		Title:    "【测试】${company}验证码",
		Content:  "您的验证码是${code}，${minutes}分钟内有效。",
		VarsJson: `[{"key":"code","value":"${input_code}"},{"key":"minutes","value":"${input_minutes:5}"}]`,
	}
	p, err := render(tpl, map[string]string{"input_code": "123456", "company": "小米"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Content != "您的验证码是123456，5分钟内有效。" {
		t.Fatalf("content = %q", p.Content)
	}
	if p.Title != "【测试】小米验证码" {
		t.Fatalf("title = %q", p.Title)
	}
	if p.OutputVars["code"] != "123456" || p.OutputVars["minutes"] != "5" {
		t.Fatalf("outputVars = %v", p.OutputVars)
	}
}

func TestRenderKeepsUnresolvedPlaceholder(t *testing.T) {
	tpl := &model.SystemMessageTemplate{Content: "hi ${missing} and ${x:default}"}
	p, err := render(tpl, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Content != "hi ${missing} and default" {
		t.Fatalf("content = %q", p.Content)
	}
}

func TestRenderOutputVarChainsInputVars(t *testing.T) {
	tpl := &model.SystemMessageTemplate{
		Content:  "code=${code}",
		VarsJson: `[{"key":"code","value":"pre-${raw}"}]`,
	}
	p, err := render(tpl, map[string]string{"raw": "999"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Content != "code=pre-999" {
		t.Fatalf("content = %q", p.Content)
	}
}

// ===== 分发 =====

func TestHandlerDispatchByMessageType(t *testing.T) {
	h, err := defaultHandlerFor("MAIL")
	if err != nil {
		t.Fatalf("MAIL dispatch: %v", err)
	}
	if _, ok := h.(*MailHandler); !ok {
		t.Fatalf("want *MailHandler, got %T", h)
	}
	h, err = defaultHandlerFor("SMS")
	if err != nil {
		t.Fatalf("SMS dispatch: %v", err)
	}
	if _, ok := h.(*SmsHandler); !ok {
		t.Fatalf("want *SmsHandler, got %T", h)
	}
	_, err = defaultHandlerFor("PUSH")
	if errCode(err) != errno.ErrMessageSupplierNotSupported {
		t.Fatalf("want ErrMessageSupplierNotSupported, got %v", err)
	}
}

// ===== 发送：失败落日志 =====

func sendFixture(repo *fakeMessageRepo) {
	repo.configs[1] = &model.SystemMessageConfig{
		ID: 1, Title: "百度短信", MessageType: "SMS", SupplierType: "TENCENT",
		Status: 1, ConfigJson: `{}`,
	}
	repo.templates[1] = &model.SystemMessageTemplate{
		ID: 1, TemplateName: "验证码", MessageConfigID: 1,
		MessageKey: "OTP", MessageType: "SMS", TemplateMode: "TEMPLATE_ID",
		Content:  "您的验证码${code}",
		VarsJson: `[{"key":"code","value":"${v}"}]`,
		Status:   1,
	}
}

func TestSendUnknownSupplierWritesFailureLog(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	svc := NewService(repo)

	results, err := svc.Send(context.Background(), SendReq{
		TemplateID: 1,
		Accounts:   []string{"13800138000"},
		Vars:       map[string]string{"v": "123456"},
	})
	if errCode(err) != errno.ErrMessageSupplierNotSupported {
		t.Fatalf("want ErrMessageSupplierNotSupported, got %v", err)
	}
	if len(results) != 1 || results[0].LogID == 0 {
		t.Fatalf("want one result with log id, got %+v", results)
	}
	if len(repo.logs) != 1 {
		t.Fatalf("want 1 log row, got %d", len(repo.logs))
	}
	row := repo.logs[0]
	if row.IsSuccess == nil || *row.IsSuccess != 0 {
		t.Fatalf("want is_success=0, got %+v", row.IsSuccess)
	}
	if !strings.Contains(row.ResponseBody, "短信供应商未接入") {
		t.Fatalf("response body should carry clear error, got %q", row.ResponseBody)
	}
	if row.CostTime == nil || row.LogTime.IsZero() {
		t.Fatalf("cost_time/log_time must be set: %+v", row)
	}
	if row.Content != "您的验证码123456" {
		t.Fatalf("log content should be rendered, got %q", row.Content)
	}
}

type fakeSendHandler struct {
	success bool
	resp    string
	err     error
	account string
	payload *renderedPayload
}

func (f *fakeSendHandler) Send(_ context.Context, account string, p *renderedPayload, _ *model.SystemMessageTemplate, _ *model.SystemMessageConfig) (bool, string, error) {
	f.account = account
	f.payload = p
	return f.success, f.resp, f.err
}

func TestSendSuccessWritesLogWithResult(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	repo.configs[1].SupplierType = "MOCK"
	svc := NewService(repo)

	fake := &fakeSendHandler{success: true, resp: "OK-XYZ"}
	old := handlerFactory
	handlerFactory = func(string) (sendHandler, error) { return fake, nil }
	defer func() { handlerFactory = old }()

	results, err := svc.Send(context.Background(), SendReq{
		TemplateID: 1,
		Accounts:   []string{"a@b.com"},
		Vars:       map[string]string{"v": "999"},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(results) != 1 || !results[0].IsSuccess || results[0].ResponseBody != "OK-XYZ" {
		t.Fatalf("results = %+v", results)
	}
	if fake.account != "a@b.com" || fake.payload.Content != "您的验证码999" {
		t.Fatalf("handler got account=%q content=%q", fake.account, fake.payload.Content)
	}
	if len(repo.logs) != 1 || *repo.logs[0].IsSuccess != 1 || repo.logs[0].ResponseBody != "OK-XYZ" {
		t.Fatalf("log = %+v", repo.logs)
	}
	if results[0].LogID != repo.logs[0].ID {
		t.Fatalf("logId mismatch: %d vs %d", results[0].LogID, repo.logs[0].ID)
	}
}

func TestSendConfigMissingWritesLog(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	delete(repo.configs, 1)
	svc := NewService(repo)

	_, err := svc.Send(context.Background(), SendReq{TemplateID: 1, Accounts: []string{"x"}})
	if errCode(err) != errno.ErrMessageConfigNotFound {
		t.Fatalf("want ErrMessageConfigNotFound, got %v", err)
	}
	if len(repo.logs) != 1 || *repo.logs[0].IsSuccess != 0 {
		t.Fatalf("want failure log, got %+v", repo.logs)
	}
}

func TestSendDisabledTemplateNoLog(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	repo.templates[1].Status = 0
	svc := NewService(repo)

	_, err := svc.Send(context.Background(), SendReq{TemplateID: 1, Accounts: []string{"x"}})
	if errCode(err) != errno.ErrMessageTemplateNotFound {
		t.Fatalf("want ErrMessageTemplateNotFound, got %v", err)
	}
	if len(repo.logs) != 0 {
		t.Fatalf("lookup failure must not write log")
	}
}

func TestSendByKeyAndType(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	svc := NewService(repo)

	fake := &fakeSendHandler{success: true, resp: "OK"}
	old := handlerFactory
	handlerFactory = func(string) (sendHandler, error) { return fake, nil }
	defer func() { handlerFactory = old }()

	results, err := svc.Send(context.Background(), SendReq{
		MessageKey: "OTP", MessageType: "sms",
		Accounts: []string{"13800138000"},
		Vars:     map[string]string{"v": "1"},
	})
	if err != nil || len(results) != 1 {
		t.Fatalf("err=%v results=%+v", err, results)
	}
}

func TestSendRequiresAccount(t *testing.T) {
	repo := newFakeRepo()
	sendFixture(repo)
	svc := NewService(repo)
	_, err := svc.Send(context.Background(), SendReq{TemplateID: 1, Accounts: []string{"  "}})
	if errCode(err) != errno.ErrMessageInvalidParam {
		t.Fatalf("want ErrMessageInvalidParam, got %v", err)
	}
	if len(repo.logs) != 0 {
		t.Fatalf("no log expected")
	}
}

// ===== mail config 校验 =====

func TestMailConfigValidation(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"empty", ``, "邮件配置为空"},
		{"bad json", `{`, "邮件配置不是合法 JSON"},
		{"no host", `{"port":25,"from":"a@b.com"}`, "SMTP服务器域名不能为空"},
		{"bad port", `{"host":"smtp.qq.com","port":70000,"from":"a@b.com"}`, "SMTP端口不合法"},
		{"no from", `{"host":"smtp.qq.com","port":465}`, "发送方不能为空"},
		{"bad from", `{"host":"smtp.qq.com","port":465,"from":"not-an-email"}`, "发送方格式错误"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseMailConfig(tc.raw)
			if err == nil {
				t.Fatalf("want error")
			}
			if !strings.Contains(errorx.ErrorWithoutStack(err), tc.wantErr) {
				t.Fatalf("err = %q, want contain %q", errorx.ErrorWithoutStack(err), tc.wantErr)
			}
		})
	}
}

func TestMailConfigValidAndDefaults(t *testing.T) {
	c, err := parseMailConfig(`{"host":"smtp.qq.com","from":"张三 <a@b.com>"}`)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if c.Port != 465 {
		t.Fatalf("default port want 465, got %d", c.Port)
	}
	if !c.implicitTLS() {
		t.Fatalf("port 465 should use implicit TLS")
	}
	if !c.starttls() {
		t.Fatalf("starttls default true")
	}
	if c.needAuth() {
		t.Fatalf("auth default false when user empty")
	}
	if envelopeAddress(c.From) != "a@b.com" {
		t.Fatalf("envelope = %q", envelopeAddress(c.From))
	}

	c2, err := parseMailConfig(`{"host":"smtp.example.com","port":25,"from":"a@b.com","sslEnable":true,"user":"u","starttlsEnable":false}`)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !c2.implicitTLS() {
		t.Fatalf("sslEnable=true should use implicit TLS")
	}
	if c2.starttls() {
		t.Fatalf("starttlsEnable=false")
	}
	if !c2.needAuth() {
		t.Fatalf("user filled => auth")
	}
}

func TestMailHTMLDetection(t *testing.T) {
	if !isHTML("<p>hello</p>") {
		t.Fatalf("html should be detected")
	}
	if !isHTML("<br/>") {
		t.Fatalf("self-closing tag should be detected")
	}
	if isHTML("just plain text") {
		t.Fatalf("plain text should not be html")
	}
}

// ===== 供应商 / 字段元数据 =====

func TestSuppliersMeta(t *testing.T) {
	mail := Suppliers("MAIL")
	if len(mail) != 1 || mail[0].Code != "MAIL" || mail[0].Label != "邮件" {
		t.Fatalf("mail suppliers = %+v", mail)
	}
	sms := Suppliers("sms")
	if len(sms) != 1 || sms[0].Code != "BAIDU" || sms[0].Label != "百度云短信" {
		t.Fatalf("sms suppliers = %+v", sms)
	}
	if len(Suppliers("PUSH")) != 0 {
		t.Fatalf("unknown type should be empty")
	}
}

func TestFieldConfigsJSON(t *testing.T) {
	raw := FieldConfigsJSON()
	var data map[string]*MessageFieldConfig
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("field-configs json invalid: %v", err)
	}
	mail, ok := data["MAIL"]
	if !ok {
		t.Fatalf("missing MAIL: %s", raw)
	}
	host, ok := mail.SupplierConfig["host"]
	if !ok || !host.Required || host.Label != "SMTP服务器域名" {
		t.Fatalf("host field = %+v", host)
	}
	port := mail.SupplierConfig["port"]
	if port == nil || port.Component != "input-number" {
		t.Fatalf("port field = %+v", port)
	}
	if v, _ := port.Value.(float64); v != 465 {
		t.Fatalf("port default = %v", port.Value)
	}
	pass := mail.SupplierConfig["pass"]
	if pass == nil || pass.Visible != "auth" || pass.InputProps == nil || pass.InputProps.Type != "password" {
		t.Fatalf("pass field = %+v", pass)
	}
	if mail.TemplateMode.SupportTemplateID || !mail.TemplateMode.SupportTemplateContent {
		t.Fatalf("mail templateMode = %+v", mail.TemplateMode)
	}
	baidu, ok := data["BAIDU"]
	if !ok {
		t.Fatalf("missing BAIDU: %s", raw)
	}
	for _, key := range []string{"accessKeyId", "secretAccessKey", "templateId", "signName", "endpoint"} {
		f, ok := baidu.SupplierConfig[key]
		if !ok {
			t.Fatalf("missing baidu field %s", key)
		}
		if f.Label == "" || f.Component == "" {
			t.Fatalf("field %s missing label/component: %+v", key, f)
		}
	}
	if baidu.SupplierConfig["secretAccessKey"].InputProps.Type != "password" {
		t.Fatalf("secretAccessKey should be password input")
	}
}

// ===== 百度请求体构造 =====

func TestBuildBaiduSmsBodyV3(t *testing.T) {
	tpl := &model.SystemMessageTemplate{TemplateMode: "TEMPLATE_ID"}
	p := &renderedPayload{OutputVars: map[string]string{"code": "1234"}}
	body, err := buildBaiduSmsBody("https://sms.bj.baidubce.com/api/v3/sendSms", "13800138000", "sms-tmpl-1", "sms-sign-1", tpl, p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["mobile"] != "13800138000" || m["template"] != "sms-tmpl-1" || m["signatureId"] != "sms-sign-1" {
		t.Fatalf("v3 body = %s", body)
	}
	if _, ok := m["contentVar"].(map[string]any); !ok {
		t.Fatalf("v3 contentVar should be object: %s", body)
	}
}

func TestBuildBaiduSmsBodyLegacy(t *testing.T) {
	tpl := &model.SystemMessageTemplate{TemplateMode: "TEMPLATE_ID"}
	p := &renderedPayload{OutputVars: map[string]string{"code": "1234"}}
	body, err := buildBaiduSmsBody(baiduSmsDefaultEndpoint, "13800138000", "sms-tmpl-1", "sms-sign-1", tpl, p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["phoneNumber"] != "13800138000" || m["templateId"] != "sms-tmpl-1" || m["signature"] != "sms-sign-1" {
		t.Fatalf("legacy body = %s", body)
	}
	if s, ok := m["contentVar"].(string); !ok || !strings.Contains(s, "1234") {
		t.Fatalf("legacy contentVar should be json string: %s", body)
	}
}

func TestBuildBaiduSmsBodyRejectsContentMode(t *testing.T) {
	tpl := &model.SystemMessageTemplate{TemplateMode: "TEMPLATE_CONTENT"}
	p := &renderedPayload{OutputVars: map[string]string{}}
	_, err := buildBaiduSmsBody(baiduSmsDefaultEndpoint, "138", "t", "s", tpl, p)
	if errCode(err) != errno.ErrMessageSendFailed {
		t.Fatalf("want ErrMessageSendFailed, got %v", err)
	}
}

// ===== CRUD 校验 =====

func TestConfigCreateValidates(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	err := svc.ConfigCreate(context.Background(), &model.SystemMessageConfig{Title: " ", MessageType: "MAIL"})
	if errCode(err) != errno.ErrMessageInvalidParam {
		t.Fatalf("want ErrMessageInvalidParam, got %v", err)
	}
	err = svc.ConfigCreate(context.Background(), &model.SystemMessageConfig{
		Title: "t", MessageType: "PUSH", SupplierType: "MAIL", Status: 1,
	})
	if errCode(err) != errno.ErrMessageInvalidParam {
		t.Fatalf("want invalid message type, got %v", err)
	}
	err = svc.ConfigCreate(context.Background(), &model.SystemMessageConfig{
		Title: "t", MessageType: "MAIL", SupplierType: "MAIL", Status: 1, ConfigJson: `{bad`,
	})
	if errCode(err) != errno.ErrMessageInvalidParam {
		t.Fatalf("want invalid json error, got %v", err)
	}
	if err := svc.ConfigCreate(context.Background(), &model.SystemMessageConfig{
		Title: "t", MessageType: "MAIL", SupplierType: "MAIL", Status: 1, ConfigJson: `{"host":"h"}`,
	}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestKeyDuplicateRejected(t *testing.T) {
	repo := newFakeRepo()
	repo.keyExists["OTP"] = 7
	svc := NewService(repo)
	err := svc.KeyCreate(context.Background(), &model.SystemMessageKey{Name: "n", MessageKey: "OTP"})
	if errCode(err) != errno.ErrMessageKeyExists {
		t.Fatalf("want ErrMessageKeyExists, got %v", err)
	}
}

func TestTemplateGetMissing(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	if _, err := svc.TemplateGet(context.Background(), 9); errCode(err) != errno.ErrMessageTemplateNotFound {
		t.Fatalf("want ErrMessageTemplateNotFound, got %v", err)
	}
}

func TestTemplateCreateValidatesMode(t *testing.T) {
	repo := newFakeRepo()
	repo.keys[1] = &model.SystemMessageKey{ID: 1, Name: "n", MessageKey: "K"}
	svc := NewService(repo)
	err := svc.TemplateCreate(context.Background(), &model.SystemMessageTemplate{
		TemplateName: "t", MessageConfigID: 1, MessageKeyID: 1, MessageKey: "K",
		MessageType: "SMS", TemplateMode: "BAD", Status: 1,
	})
	if errCode(err) != errno.ErrMessageInvalidParam {
		t.Fatalf("want ErrMessageInvalidParam, got %v", err)
	}
}

func TestTemplateDerivesMessageKeyFromID(t *testing.T) {
	repo := newFakeRepo()
	repo.keys[7] = &model.SystemMessageKey{ID: 7, Name: "n", MessageKey: "OTP_CODE"}
	svc := NewService(repo)
	m := &model.SystemMessageTemplate{
		TemplateName: "t", MessageConfigID: 1, MessageKeyID: 7, MessageKey: "",
		MessageType: "MAIL", TemplateMode: "TEMPLATE_CONTENT", Status: 1,
	}
	if err := svc.TemplateCreate(context.Background(), m); err != nil {
		t.Fatalf("create: %v", err)
	}
	if m.MessageKey != "OTP_CODE" {
		t.Fatalf("messageKey not derived: %q", m.MessageKey)
	}
	// 常量不存在 → 明确错误
	err := svc.TemplateCreate(context.Background(), &model.SystemMessageTemplate{
		TemplateName: "t2", MessageConfigID: 1, MessageKeyID: 99, MessageKey: "",
		MessageType: "MAIL", TemplateMode: "TEMPLATE_CONTENT", Status: 1,
	})
	if errCode(err) != errno.ErrMessageKeyNotFound {
		t.Fatalf("want ErrMessageKeyNotFound, got %v", err)
	}
}

func TestLogClearDelegates(t *testing.T) {
	repo := newFakeRepo()
	repo.logs = append(repo.logs, &model.SystemMessageLog{ID: 1}, &model.SystemMessageLog{ID: 2})
	svc := NewService(repo)
	n, err := svc.LogClear(context.Background())
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestLogRemoveDelegates(t *testing.T) {
	repo := newFakeRepo()
	repo.logs = append(repo.logs, &model.SystemMessageLog{ID: 1}, &model.SystemMessageLog{ID: 2})
	svc := NewService(repo)
	n, err := svc.LogRemove(context.Background(), 1)
	if err != nil || n != 1 || len(repo.logs) != 1 {
		t.Fatalf("n=%d logs=%d err=%v", n, len(repo.logs), err)
	}
	if _, err := svc.LogRemove(context.Background(), 99); errCode(err) != errno.ErrMessageLogNotFound {
		t.Fatalf("want ErrMessageLogNotFound, got %v", err)
	}
}
