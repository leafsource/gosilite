package message

import (
	"encoding/json"
	"strings"
)

// ========== 元数据：suppliers / field-configs（移植 ruoyi MessageData + MailFieldConfigs） ==========

// SupplierOption 供应商选项
type SupplierOption struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// Suppliers 按消息类型返回可用供应商（spec S2.3 getSupplierType）
func Suppliers(messageType string) []SupplierOption {
	switch strings.ToUpper(strings.TrimSpace(messageType)) {
	case "MAIL":
		return []SupplierOption{{Code: "MAIL", Label: "邮件"}}
	case "SMS":
		return []SupplierOption{{Code: "BAIDU", Label: "百度云短信"}}
	default:
		return []SupplierOption{}
	}
}

// ========== FieldConfig JSON 结构（字段命名对齐 ruoyi org.dromara.common.core.ui.FieldConfig） ==========

// FormRule 校验规则（对齐 ruoyi FormRule，null 字段省略）
type FormRule struct {
	Bool     *bool    `json:"boolean,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Message  string   `json:"message,omitempty"`
	Trigger  string   `json:"trigger,omitempty"`
	Required *bool    `json:"required,omitempty"`
	Enums    []string `json:"enum,omitempty"`
}

// FieldInputProps 输入框组件属性（对齐 ruoyi FieldInputProps 默认值）
type FieldInputProps struct {
	Autofocus            bool   `json:"autofocus"`
	Borderless           bool   `json:"borderless"`
	Clearable            bool   `json:"clearable"`
	ShowClearIconOnEmpty bool   `json:"showClearIconOnEmpty"`
	ShowLimitNumber      bool   `json:"showLimitNumber"`
	Type                 string `json:"type,omitempty"`
}

// FieldInputNumberProps 数字输入框组件属性
type FieldInputNumberProps struct {
	AllowInputOverLimit bool   `json:"allowInputOverLimit"`
	AutoWidth           bool   `json:"autoWidth"`
	DecimalPlaces       int    `json:"decimalPlaces"`
	Min                 *int64 `json:"min,omitempty"`
	Max                 *int64 `json:"max,omitempty"`
}

// FieldSwitchProps 开关组件属性
type FieldSwitchProps struct{}

// FieldConfig 动态表单字段定义
type FieldConfig struct {
	Value            any                    `json:"value,omitempty"`
	Label            string                 `json:"label"`
	Component        string                 `json:"component"`
	Help             string                 `json:"help,omitempty"`
	Required         bool                   `json:"required"`
	Visible          string                 `json:"visible,omitempty"`
	Rules            []FormRule             `json:"rules,omitempty"`
	InputProps       *FieldInputProps       `json:"inputProps,omitempty"`
	InputNumberProps *FieldInputNumberProps `json:"inputNumberProps,omitempty"`
	SwitchProps      *FieldSwitchProps      `json:"switchProps,omitempty"`
}

// TemplateModeMeta 模板模式支持情况（对齐 ruoyi TemplateMode）
type TemplateModeMeta struct {
	SupportTemplateID      bool `json:"supportTemplateId"`
	SupportTemplateContent bool `json:"supportTemplateContent"`
}

// MessageFieldConfig supplier → 字段定义（对齐 ruoyi MessageFieldConfig）
type MessageFieldConfig struct {
	SupplierConfig map[string]*FieldConfig `json:"supplierConfig"`
	MessageType    string                  `json:"messageType"`
	TemplateMode   TemplateModeMeta        `json:"templateMode"`
}

func inputField(label string, required bool) *FieldConfig {
	return &FieldConfig{
		Label:     label,
		Component: "input",
		Required:  required,
		InputProps: &FieldInputProps{
			Autofocus:            false,
			Borderless:           false,
			Clearable:            true,
			ShowClearIconOnEmpty: false,
			ShowLimitNumber:      false,
		},
	}
}

func passwordField(label, help string, required bool) *FieldConfig {
	f := inputField(label, required)
	f.Help = help
	f.InputProps.Type = "password"
	return f
}

func numberField(label string, required bool, value int64, min, max *int64) *FieldConfig {
	return &FieldConfig{
		Value:     value,
		Label:     label,
		Component: "input-number",
		Required:  required,
		InputNumberProps: &FieldInputNumberProps{
			AllowInputOverLimit: false,
			AutoWidth:           true,
			DecimalPlaces:       0,
			Min:                 min,
			Max:                 max,
		},
	}
}

func switchField(label string, required bool, value bool) *FieldConfig {
	return &FieldConfig{
		Value:       value,
		Label:       label,
		Component:   "switch",
		Required:    required,
		SwitchProps: &FieldSwitchProps{},
	}
}

// mailFieldConfigs 移植 ruoyi MailFieldConfigs（字段/label/help/required/visible/默认值完整语义）
func mailFieldConfigs() *MessageFieldConfig {
	min0, max65535, minNeg := int64(0), int64(65535), int64(-1)
	fields := map[string]*FieldConfig{
		"host": {
			Label:      "SMTP服务器域名",
			Component:  "input",
			Required:   true,
			InputProps: &FieldInputProps{Clearable: true},
		},
		"port": numberField("SMTP服务端口", true, 465, &min0, &max65535),
		"auth": switchField("用户名密码验证", true, false),
		"user": func() *FieldConfig {
			f := inputField("用户名", true)
			f.Help = "发件人邮箱地址（如果使用foxmail邮箱，此处为qq号）"
			f.Visible = "auth"
			return f
		}(),
		"pass": func() *FieldConfig {
			f := passwordField("密码", "注意，某些邮箱需要为SMTP服务单独设置密码，详情查看相关帮助", true)
			f.Visible = "auth"
			return f
		}(),
		"from": func() *FieldConfig {
			f := inputField("发送方", true)
			f.Help = "邮件中显示的发件人姓名，遵循RFC-822标准。支持以下形式：1.user@xxx.xx 2.name&nbsp;&lt;user@xxx.xx&gt;"
			f.Rules = []FormRule{{
				Pattern: `^([^ \f\n\r\t\\v<]+@[^ \f\n\r\t\\v<>]+\.[^ \f\n\r\t\\v>]+$)|(\S+ <[^ \f\n\r\t\\v<]+@[^ \f\n\r\t\\v<>]+\.[^ \f\n\r\t\\v>]+>$)`,
				Message: "发送方格式错误",
				Trigger: "change",
			}}
			return f
		}(),
		"debug":               switchField("调试模式", true, false),
		"splitlongparameters": switchField("超长参数是否切分", true, false),
		"encodefilename":      switchField("使用charset编码", true, true),
		"starttlsEnable": func() *FieldConfig {
			f := switchField("STARTTLS安全连接", true, true)
			f.Help = "使用 STARTTLS安全连接，STARTTLS是对纯文本通信协议的扩展。它将纯文本连接升级为加密连接（TLS或SSL）， 而不是使用一个单独的加密通信端口。"
			return f
		}(),
		"sslEnable": switchField("SSL安全连接", true, true),
		"sslProtocols": func() *FieldConfig {
			f := inputField("SSL协议", false)
			f.Help = "SSL协议，多个协议用空格分隔"
			return f
		}(),
		"socketFactoryClass": func() *FieldConfig {
			f := inputField("socketFactoryClass", true)
			f.Value = "javax.net.ssl.SSLSocketFactory"
			f.Help = "指定实现javax.net.SocketFactory接口的类的名称,这个类将被用于创建SMTP的套接字"
			return f
		}(),
		"socketFactoryFallback": func() *FieldConfig {
			f := switchField("创建套接字工厂类", true, true)
			f.Help = "如果设置为true,未能创建一个套接字使用指定的套接字工厂类将导致使用java.net.Socket创建的套接字类, 默认值为true"
			return f
		}(),
		"socketFactoryPort": func() *FieldConfig {
			f := numberField("Socket端口", true, 465, &min0, &max65535)
			f.Help = "指定的端口连接到在使用指定的套接字工厂。如果没有设置,将使用默认端口"
			return f
		}(),
		"timeout":           numberFieldWithHelp("SMTP超时时长", "SMTP超时时长，单位毫秒，缺省值不超时", false, &minNeg, nil),
		"connectionTimeout": numberFieldWithHelp("Socket连接超时值", "Socket连接超时值，单位毫秒，缺省值不超时", false, &minNeg, nil),
		"writeTimeout":      numberFieldWithHelp("Socket写出超时值", "Socket写出超时值，单位毫秒，缺省值不超时", false, &minNeg, nil),
	}
	return &MessageFieldConfig{
		SupplierConfig: fields,
		MessageType:    "MAIL",
		TemplateMode:   TemplateModeMeta{SupportTemplateID: false, SupportTemplateContent: true},
	}
}

func numberFieldWithHelp(label, help string, required bool, min, max *int64) *FieldConfig {
	f := numberField(label, required, 0, min, max)
	f.Value = nil // ruoyi 未设置默认值（min=-1 的输入框为空）
	f.Help = help
	return f
}

// baiduSmsFieldConfigs 百度云短信字段（自设计：input/password）
func baiduSmsFieldConfigs() *MessageFieldConfig {
	fields := map[string]*FieldConfig{
		"accessKeyId":     inputField("Access Key ID", true),
		"secretAccessKey": passwordField("Secret AccessKey", "百度智能云 AK/SK 中的 Secret AccessKey", true),
		"templateId": func() *FieldConfig {
			f := inputField("默认模板ID", true)
			f.Help = "短信模板ID（sms-tmpl-xxx），模板未填模板ID时使用"
			return f
		}(),
		"signName": func() *FieldConfig {
			f := inputField("默认签名ID", true)
			f.Help = "短信签名ID（sms-sign-xxx），模板未填签名时使用"
			return f
		}(),
		"endpoint": func() *FieldConfig {
			f := inputField("接口地址", false)
			f.Value = baiduSmsDefaultEndpoint
			f.Help = "默认 https://sms.bj.baidubce.com/v1/sms；官方现行发送接口为 https://sms.bj.baidubce.com/api/v3/sendSms"
			return f
		}(),
	}
	return &MessageFieldConfig{
		SupplierConfig: fields,
		MessageType:    "SMS",
		TemplateMode:   TemplateModeMeta{SupportTemplateID: true, SupportTemplateContent: false},
	}
}

// FieldConfigsJSON supplier → 字段定义 map（前端 evalFieldConfigs 消费）
func FieldConfigsJSON() string {
	data := map[string]*MessageFieldConfig{
		"MAIL":  mailFieldConfigs(),
		"BAIDU": baiduSmsFieldConfigs(),
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}
