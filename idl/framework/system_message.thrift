namespace go framework.message

// ===== 消息配置 =====

struct MessageConfigListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
    // 缺省=全部；MAIL / SMS
    4: optional string messageType
    // 缺省=全部；1 正常 0 停用
    5: optional i32 status
}

struct MessageConfigDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string title
    // MAIL / SMS
    3: string messageType
    // 供应商标识：MAIL / BAIDU
    4: string supplierType
    // 通道凭据与参数 JSON
    5: string configJson
    // 1 正常 0 停用
    6: i32 status
    7: string remark
    8: string createdAt
    9: string updatedAt
}

struct MessageConfigSaveRequest {
    1: optional i64 id (go.tag = "json:\"id,string\"")
    2: string title
    3: string messageType
    4: string supplierType
    5: string configJson
    6: i32 status
    7: string remark
}

struct MessageConfigListResponse {
    1: list<MessageConfigDTO> list
    2: i64 total
}

struct MessageIdRequest {
    1: i64 id (go.tag = "json:\"id,string\"")
}

// ===== 消息常量 =====

struct MessageKeyListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
}

struct MessageKeyDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string name
    3: string messageKey
    4: string remark
    5: string createdAt
    6: string updatedAt
}

struct MessageKeySaveRequest {
    1: optional i64 id (go.tag = "json:\"id,string\"")
    2: string name
    3: string messageKey
    4: string remark
}

struct MessageKeyListResponse {
    1: list<MessageKeyDTO> list
    2: i64 total
}

// ===== 消息模板 =====

struct MessageTemplateListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
    4: optional string messageType
    5: optional i64 messageKeyId (go.tag = "json:\"messageKeyId,string\"")
    // 缺省=全部；1 正常 0 停用
    6: optional i32 status
}

struct MessageTemplateDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string templateName
    3: i64 messageConfigId (go.tag = "json:\"messageConfigId,string\"")
    4: i64 messageKeyId (go.tag = "json:\"messageKeyId,string\"")
    5: string messageKey
    6: string messageType
    // TEMPLATE_ID / TEMPLATE_CONTENT
    7: string templateMode
    8: string title
    9: string signature
    10: string templateId
    11: string content
    // 输入变量定义 JSON
    12: string varsJson
    13: i32 status
    14: string remark
    15: string createdAt
    16: string updatedAt
}

struct MessageTemplateSaveRequest {
    1: optional i64 id (go.tag = "json:\"id,string\"")
    2: string templateName
    3: i64 messageConfigId (go.tag = "json:\"messageConfigId,string\"")
    4: i64 messageKeyId (go.tag = "json:\"messageKeyId,string\"")
    5: string messageKey
    6: string messageType
    7: string templateMode
    8: string title
    9: string signature
    10: string templateId
    11: string content
    12: string varsJson
    13: i32 status
    14: string remark
}

struct MessageTemplateListResponse {
    1: list<MessageTemplateDTO> list
    2: i64 total
}

// ===== 发送记录 =====

struct MessageLogListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
    4: optional string messageType
    5: optional string messageKey
    // 缺省=全部；1 成功 0 失败
    6: optional i32 isSuccess
}

struct MessageLogDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: i64 messageTemplateId (go.tag = "json:\"messageTemplateId,string\"")
    3: string messageKey
    4: string messageTemplateName
    5: string messageType
    6: string templateMode
    7: string account
    8: string title
    9: string templateId
    10: string content
    11: string messageConfigTitle
    12: string supplierType
    13: i32 isSuccess
    14: string responseBody
    15: string logTime
    16: i64 costTime
}

struct MessageLogListResponse {
    1: list<MessageLogDTO> list
    2: i64 total
}

// ===== 测试发送 / 元数据 =====

struct MessageSendTestRequest {
    1: i64 messageTemplateId (go.tag = "json:\"messageTemplateId,string\"")
    // 收件账号（邮箱地址或手机号）
    2: string account
    3: map<string, string> vars
}

struct MessageSendTestResponse {
    1: bool isSuccess
    2: string responseBody
    // 毫秒
    3: i64 costTime
    4: i64 logId (go.tag = "json:\"logId,string\"")
}

struct MessageSuppliersRequest {
    // MAIL / SMS
    1: string messageType
}

struct MessageSupplierOption {
    1: string code
    2: string label
}

struct MessageSuppliersResponse {
    1: list<MessageSupplierOption> list
}

// supplier → 动态表单字段定义（FieldConfig JSON，前端 evalFieldConfigs 消费）
struct MessageFieldConfigsResponse {
    1: string json
}

service SystemMessageController {
    // ---- 消息配置 ----
    MessageConfigListResponse ConfigList(1: MessageConfigListRequest req)(api.get = "/system/message/config/list")
    MessageConfigDTO ConfigGet(1: MessageIdRequest req)(api.get = "/system/message/config")
    MessageConfigDTO ConfigCreate(1: MessageConfigSaveRequest req)(api.post = "/system/message/config")
    MessageConfigDTO ConfigUpdate(1: MessageConfigSaveRequest req)(api.put = "/system/message/config")
    i32 ConfigRemove(1: MessageIdRequest req)(api.delete = "/system/message/config")

    // ---- 消息常量 ----
    MessageKeyListResponse KeyList(1: MessageKeyListRequest req)(api.get = "/system/message/key/list")
    MessageKeyDTO KeyGet(1: MessageIdRequest req)(api.get = "/system/message/key")
    MessageKeyDTO KeyCreate(1: MessageKeySaveRequest req)(api.post = "/system/message/key")
    MessageKeyDTO KeyUpdate(1: MessageKeySaveRequest req)(api.put = "/system/message/key")
    i32 KeyRemove(1: MessageIdRequest req)(api.delete = "/system/message/key")

    // ---- 消息模板 ----
    MessageTemplateListResponse TemplateList(1: MessageTemplateListRequest req)(api.get = "/system/message/template/list")
    MessageTemplateDTO TemplateGet(1: MessageIdRequest req)(api.get = "/system/message/template")
    MessageTemplateDTO TemplateCreate(1: MessageTemplateSaveRequest req)(api.post = "/system/message/template")
    MessageTemplateDTO TemplateUpdate(1: MessageTemplateSaveRequest req)(api.put = "/system/message/template")
    i32 TemplateRemove(1: MessageIdRequest req)(api.delete = "/system/message/template")
    // 测试发送（走发送服务，无论成败都落发送记录）
    MessageSendTestResponse TemplateSendTest(1: MessageSendTestRequest req)(api.post = "/system/message/template/send-test")

    // ---- 发送记录 ----
    MessageLogListResponse LogList(1: MessageLogListRequest req)(api.get = "/system/message/log/list")
    MessageLogDTO LogGet(1: MessageIdRequest req)(api.get = "/system/message/log")
    // 删除单条发送记录，返回删除条数
    i64 LogRemove(1: MessageIdRequest req)(api.delete = "/system/message/log")
    // 清空发送记录，返回删除条数
    i64 LogClear()(api.delete = "/system/message/log/clear")

    // ---- 元数据 ----
    MessageSuppliersResponse Suppliers(1: MessageSuppliersRequest req)(api.get = "/system/message/suppliers")
    MessageFieldConfigsResponse FieldConfigs()(api.get = "/system/message/field-configs")
}
