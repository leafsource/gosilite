package model

import "time"

// 消息管理四表手写模型（script/migrations/0023_message.sql，非 gen）

// SystemMessageConfig 消息配置表
type SystemMessageConfig struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	Title        string    `gorm:"column:title" json:"title"`
	MessageType  string    `gorm:"column:message_type" json:"messageType"`
	SupplierType string    `gorm:"column:supplier_type" json:"supplierType"`
	ConfigJson   string    `gorm:"column:config_json" json:"configJson"`
	Status       int16     `gorm:"column:status" json:"status"`
	Remark       string    `gorm:"column:remark" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 表名
func (SystemMessageConfig) TableName() string { return "system_message_config" }

// SystemMessageKey 消息常量表
type SystemMessageKey struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	MessageKey string    `gorm:"column:message_key" json:"messageKey"`
	Remark     string    `gorm:"column:remark" json:"remark"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 表名
func (SystemMessageKey) TableName() string { return "system_message_key" }

// SystemMessageTemplate 消息模板表
type SystemMessageTemplate struct {
	ID              int64     `gorm:"column:id;primaryKey" json:"id"`
	TemplateName    string    `gorm:"column:template_name" json:"templateName"`
	MessageConfigID int64     `gorm:"column:message_config_id" json:"messageConfigId"`
	MessageKeyID    int64     `gorm:"column:message_key_id" json:"messageKeyId"`
	MessageKey      string    `gorm:"column:message_key" json:"messageKey"`
	MessageType     string    `gorm:"column:message_type" json:"messageType"`
	TemplateMode    string    `gorm:"column:template_mode" json:"templateMode"`
	Title           string    `gorm:"column:title" json:"title"`
	Signature       string    `gorm:"column:signature" json:"signature"`
	TemplateID      string    `gorm:"column:template_id" json:"templateId"`
	Content         string    `gorm:"column:content" json:"content"`
	VarsJson        string    `gorm:"column:vars_json" json:"varsJson"`
	Status          int16     `gorm:"column:status" json:"status"`
	Remark          string    `gorm:"column:remark" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 表名
func (SystemMessageTemplate) TableName() string { return "system_message_template" }

// SystemMessageLog 消息发送记录表
type SystemMessageLog struct {
	ID                  int64     `gorm:"column:id;primaryKey" json:"id"`
	MessageTemplateID   int64     `gorm:"column:message_template_id" json:"messageTemplateId"`
	MessageKey          string    `gorm:"column:message_key" json:"messageKey"`
	MessageTemplateName string    `gorm:"column:message_template_name" json:"messageTemplateName"`
	MessageType         string    `gorm:"column:message_type" json:"messageType"`
	TemplateMode        string    `gorm:"column:template_mode" json:"templateMode"`
	Account             string    `gorm:"column:account" json:"account"`
	Title               string    `gorm:"column:title" json:"title"`
	TemplateID          string    `gorm:"column:template_id" json:"templateId"`
	Content             string    `gorm:"column:content" json:"content"`
	MessageConfigTitle  string    `gorm:"column:message_config_title" json:"messageConfigTitle"`
	SupplierType        string    `gorm:"column:supplier_type" json:"supplierType"`
	IsSuccess           *int16    `gorm:"column:is_success" json:"isSuccess"`
	ResponseBody        string    `gorm:"column:response_body" json:"responseBody"`
	LogTime             time.Time `gorm:"column:log_time" json:"logTime"`
	CostTime            *int64    `gorm:"column:cost_time" json:"costTime"`
}

// TableName 表名
func (SystemMessageLog) TableName() string { return "system_message_log" }
