---
feature: system-message
status: delivered
updated: 2026-10-12
branch: develop
commits:
---

# 消息管理（配置/常量/模板/发送记录 · 复刻 ruoyi-tdesign）

## Report

### 验收记录（2026-10-12，收尾验收）

**S2.6.1 迁移与菜单**：`0023_message.sql` 于 dev 库连跑两遍幂等（第二次全部 `INSERT 0 0`）；菜单树 19 行就位（M 50 + C 5001-5004 + F 15 枚含新增 `500403 system:messageLog:remove`），`system` 角色绑定 19 条；四个 component 均指向真实存在的 `web/src/pages/system/message/*/index.vue`，`defineOptions` 名与 `genRouteName` 一致。

**S2.6.1/2 API 冒烟（33 项全过）**：四资源 list/get/create/update/delete 走通统一信封；`GET /system/message/suppliers?messageType=MAIL|SMS` 返回 邮件/百度云短信；`field-configs` 下发 ruoyi 形态 JSON。send-test：
- MAIL（无效 SMTP `127.0.0.1:25999`）→ `701000066 连接SMTP服务器失败: connection refused`，`system_message_log` 新增 `is_success=0` 记录，`response_body` 无栈污染；
- SMS（未接入 supplier ALIYUN）→ `701000065 短信供应商未接入: ALIYUN`，同样落记录；
- 渲染语义核对：title 用输入变量（`测试标题-张三`）、content 用 varsJson 输出变量，与 ruoyi 一致；
- `DELETE /system/message/log` 行删除 → `data:1`，重复删 → `701000064 message log not found`；`DELETE .../clear` → 清零。

**S2.6.3 质量门**：`WEB=1 ./script/check.sh` → `check OK`（go test/vet/build + web build + hugo docs）；消息包单测含 `TestLogRemoveDelegates` 等新增用例全绿；log 页 eslint `--max-warnings 0` 与 `vue-tsc --noEmit` exit 0。

**S2.6.1 浏览器冒烟（IAB 实测）**：四页均可打开（config/key/template/log，表格+查询区+新增按钮齐全）；侧栏 消息管理→四子页在位，**发送记录图标实渲染 `t-icon-history`**（修复前 `file-history` 非 TDesign 图标，svg 空引用）；常量页 UI CRUD 闭环：新增 `UI冒烟常量` → 行出现 → 行删除确认 → 行消失，无前端报错；log 页行删除按钮与失败记录行可见。

**验收期发现并修复的缺口**：
1. 发送记录菜单图标 `file-history` 不在 TDesign 图标集 → 改 `history`（迁移 + dev 库同步 UPDATE）。
2. 前端 log 页行删除/批量删除调用 `DELETE /system/message/log`，后端无此路由（裸 404）→ 按用户拍板补齐：IDL `LogRemove` → hz update → handler/service/repository → middleware 挂 `system:messageLog:remove` → 迁移补 F 500403 + 角色绑定 → 前端行/批量删除按钮改挂 `remove` 码 → swag 重新生成。

**API 冒烟脚本口径备忘**（复用时注意）：消息模块 `status` 语义 **1=正常 0=停用**（与本仓 `sys_status` 相反）；标量 i64 标识字段（`id/messageConfigId/messageKeyId/messageTemplateId/logId`）由 IDL `go.tag` 生成 `json:",string"`，body **必须传 JSON string**，裸数字会 400；`suppliers` 查询参数是 `messageType` 不是 `type`；log 详情前端走 query `?id=`。数组型 i64 id 仍由 jsonid 逐项转换。

### 独立评审

独立 reviewer 只读核查（对照 ruoyi 原版 + 本仓惯例），结论 **需修复** → 已全部处置：

| # | 级别 | 问题 | 处置 |
|---|------|------|------|
| 1 | blocker | UI 新建模板必失败：前端只提交 `messageKeyId`、`messageKey` 恒空，后端未复刻 ruoyi 的派生逻辑 | **已修**：`deriveTemplateKey` 在 create/update 前按 `messageKeyId` 查常量回填 `messageKey`（不存在 → `701000062`）；单测 `TestTemplateDerivesMessageKeyFromID`；curl 按 UI payload 复验通过 |
| 2 | minor | `messageTemplateId/messageKeyId/messageConfigId` 的前端 `Number()` 转换会超 2^53 丢精度 | **已修并升级为全仓规则**：IDL `go.tag` 生成 `json:",string"`；前端全程 string；jsonid 仅保留数组 id |
| 3 | minor | 查询区较 ruoyi 缺 templateMode/supplierType/时间范围等筛选 | **遗留**：keyword 可部分替代，见「后续项」 |
| 4 | minor | suppliers/field-configs 挂 `messageConfig:list`，log/template 页调用会被 403 静默吞掉 | **已修**：middleware 改 OR 多码（suppliers: config/template/log list；field-configs: config/template list） |
| 5 | minor | `@Tags "消息管理"` 不在 AGENTS 约定枚举 | **已修**：AGENTS.md Tags 枚举补「消息管理」 |

评审同时确认无问题：BCE 官方向量单测在位、渲染语义与 ruoyi 一致、失败落日志/删除路径实测通、F 码 ↔ middleware ↔ v-permission 三方一致、defineOptions 四名与 genRouteName 一致、四页结构与 ruoyi 对齐。

### 后续项（不在本次交付范围）

1. **标量 ID 精度迁移已交付**：2026-10-12 UI 实际复现 `messageKeyId …353920` 被 `Number()` 舍入为 `…354000` → message key not found；现已将全仓 12 份 IDL 的 88 个标量 `id/*Id` 用 `go.tag` 生成 `json:",string"`，前端全程 string，Swagger 同步为 string。`pkg/jsonid` 仅保留 `list<i64>` 数组 id（Go 的 `,string` 不支持 slice）；AGENTS 已固化新增字段规则。
2. **查询筛选补齐**：log 页 templateMode/supplierType/时间范围、config 页平台筛选、template 页模板类型筛选（需同步 IDL ListRequest）。

## [S1] Problem

系统管理下缺少消息通道能力：无法配置短信/邮箱发送源、无法以「常量 key + 模板」的方式管理业务消息、也没有发送记录留痕。需求：照 `/Users/leafsource/tmp-project/ruoyi-tdesign` 的 ruoyi-message 模块与 4 个前端页面复刻到 gosilite（Hertz + Vue3/TDesign）。

## [S2] Design

### S2.1 范围与发送能力（用户拍板）

- **全量复刻 4 个功能页**：消息配置、消息常量、消息模板、消息发送记录（页面结构/交互对齐 ruoyi 版，技术栈适配本仓）。
- **通道抽象层** `MessageSendHandler`（按 message_type 分发）：
  - **MAIL：真实发送**，Go 标准库 `net/smtp`（支持 STARTTLS 与 465 implicit TLS，零新依赖）；
  - **SMS：抽象层就位，优先实现百度云短信**（BCE Auth V1 签名 + HTTPS JSON API，标准库实现，不引 SDK）；未接入的 supplier 调用返回明确错误「短信供应商未接入」。
- 每次发送（含测试发送）**无论成败都写发送记录**（is_success / response_body / cost_time）。
- 不复刻 ruoyi 的 `refreshCache`（本仓无该缓存体系）、不复刻导出按钮（本仓无导出惯例）。

### S2.2 数据（迁移 `script/migrations/0023_message.sql`，对齐 0011 风格）

四表（BIGSERIAL 主键 + created_at/updated_at；去掉 ruoyi 的 create_dept/create_by/update_by）：

| 表 | 关键字段（业务列照 ruoyi，见参考 SQL sys_message_*） |
|----|----|
| `system_message_config` | title, message_type(SMS/MAIL), supplier_type, config_json, status, remark；索引 (message_type,status) |
| `system_message_key` | name, message_key, remark |
| `system_message_template` | template_name, message_config_id, message_key_id, message_key, message_type, template_mode, title, signature, template_id, content, vars_json, status, remark；索引 (message_key_id)、(message_key,message_type,status) |
| `system_message_log` | message_template_id, message_key, message_template_name, message_type, template_mode, account, title, template_id, content, message_config_title, supplier_type, is_success, response_body, log_time, cost_time；索引 (message_template_id) |

字典（挂「系统」分组，仿 0011）：`sys_message_type`（MAIL 邮件 / SMS 短信）、`sys_message_template_mode`（TEMPLATE_ID 模板ID / TEMPLATE_CONTENT 模板内容，对应 ruoyi 同名字典）。

菜单（系统管理 id=3 下，id 取空闲段 **50**，实现时以 `SELECT` 确认无占用）：

| id | 类型 | 内容 |
|----|------|------|
| 50 | M | 消息管理，path `message`，icon `chat`，component `layout`，permission_code `page:root:message`，perms `''`，sort = 现有子菜单 max+1 |
| 5001 | C | 消息配置 → `system/message/config/index`，perms `system:messageConfig:list`（驼峰沿 ruoyi，本仓已有 ossConfig 先例） |
| 5002 | C | 消息常量 → `system/message/key/index`，perms `system:messageKey:list` |
| 5003 | C | 消息模板 → `system/message/template/index`，perms `system:messageTemplate:list` |
| 5004 | C | 发送记录 → `system/message/log/index`，perms `system:messageLog:list` |
| F | F | 每 C 下 query/add/edit/remove（记录页为 query/remove/clear，id 形如 `500101`；记录删除 `500403` = `system:messageLog:remove`）；权限码 `system:messageXxx:add` 等 |

`system/admin` 角色绑定全部新菜单 id（仿 0015/0022）；幂等 `WHERE NOT EXISTS`。

### S2.3 契约（新 IDL `idl/framework/system_message.thrift` → hz update）

- 四资源分页 CRUD：`GET /system/message/{config|key|template|log}/list|get`、`POST/PUT/DELETE` 同名路径（路径风格与粒度对齐本仓 `system_notice.thrift`，字段名沿 ruoyi：messageConfigId 等 camelCase）。
- `GET /system/message/suppliers?type=MAIL|SMS` → 该类型可用 supplier 列表（贴 ruoyi `getSupplierType`）。
- `GET /system/message/field-configs` → supplier → 动态表单字段定义（移植 ruoyi `MessageFieldConfig`/`MessageData` + `getMessageFieldConfigs`，config 页动态渲染依赖它）。
- `POST /system/message/template/send-test`：body `{messageTemplateId, account, vars}` → 走发送服务（对应 ruoyi `sendMessageTest`）。
- `DELETE /system/message/log`：删除单条发送记录（返回删除条数，权限 `system:messageLog:remove`）；`DELETE /system/message/log/clear`（记录页清空，需 `system:messageLog:clear`）。
- 统一信封 `{code,msg,data}`；分页同本仓；每个 Handler 补全 swag 注解。

### S2.4 发送服务（`application/framework/service/message/`）

- `Service.Send(ctx, SendReq{TemplateID 或 Key+Type, Account[], Vars map})`：查模板（status=启用）→ 找 config（启用）→ **渲染**（占位符格式与 ruoyi `BaseMessageSendHandler` 一致，实现前读参考代码确认，预期 `${var}`）→ 按 type 选 handler → 计时 → 写 `system_message_log`（成功/失败都写）。
- `MailHandler`：`config_json` 解析为 `{host,port,user,password,from,ssl}`；465 用 `tls.Dial` + `smtp.NewClient`，其余 STARTTLS/明文按 ruoyi MailAccount 语义；正文 HTML 正则判定（对齐 ruoyi）。
- `BaiduSmsHandler`：supplier `BAIDU`；实现 BCE Auth V1 签名（HMAC-SHA256，官方文档为契约，实现前 fetch 核对），POST `config_json.endpoint`（默认 `https://sms.bj.baidubce.com/v1/sms`）；凭据 `{accessKeyId,secretAccessKey,templateId,signName}` 存 config_json。**无真实凭据时的验收以签名单测 + 请求构造单测为准**。
- 其余 SMS supplier → 返回 `errorx`「短信供应商未接入」，log 记 is_success=0。
- `service_test.go`：渲染、类型分发、失败落日志、mail 地址/端口校验、百度签名向量。

### S2.5 前端（移植 4 页）

- `web/src/api/system/message{Config,Key,Template,Log}.ts` + models：对齐本仓 request/信封/分页返回（非 ruoyi 的 R/TableDataInfo 形状，字段名不变）。
- 页面 `web/src/pages/system/message/{config,key,template,log}/index.vue`：结构、列、弹窗表单、查询区对齐 ruoyi 原页（参考文件路径见 S2.1）；**config 页动态字段表单**移植 `evalFieldConfigs` 机制；template 页保留「测试发送」弹窗（account + vars 键值编辑）。
- 技术适配：`t-*` API 先过 TDesign MCP；列表用本仓 `s-table` 惯例（若与移植冲突以还原 ruoyi 行为优先，记录取舍）；页面 `defineOptions({name})` = `genRouteName('/system/message/xxx')`（SystemMessageConfig 等）；按钮权限 `v-permission`。
- 菜单 component 与文件一一对应，上线前核对。

### S2.6 验收

1. [x] 迁移 dev 库两遍幂等；菜单树：系统管理 → 消息管理 → 四页 + F 按钮；四页浏览器可开、CRUD 走通（配置/常量/模板）。
2. [x] `send-test`：MAIL 用本地可验证配置（或无效 SMTP）调用 → 返回明确结果且**发送记录新增一条**（成败均可，以记录存在为准）；SMS 未接入 supplier → 错误信息明确且同样落记录。
3. [x] 单测：渲染/签名/分发绿；`WEB=1 ./script/check.sh` 全绿。
4. [x] 独立 reviewer 按「复刻对齐 + 本仓惯例 + 无臆测（百度 API 以官方文档为据）」核查。

## [S3] Out of Scope

- 除百度云外的其他短信供应商实现（仅留抽象层与错误分支）。
- 业务侧按 key 自动触发发送的调用点（本次只交付 Send 服务方法 + 测试发送 API，业务接入后续按需）。
- ruoyi 的 refreshCache、导出、i18n 文案体系；敏感词表。
- 真实短信/邮件凭据的端到端投递验收（无供应商账号，以单测与日志落库代替）。

## Tasks

- [x] T1: 迁移 0023_message.sql（4 表 + 2 字典 + 菜单 50 系 + F + 角色绑定）— acceptance: dev 库连跑两遍幂等；菜单树四 C + F 就位且 component 均指向将创建的文件 (covers: S2.2)
- [x] T2: IDL system_message.thrift + `hz update` + swagger 注解 — acceptance: `api/model`/`api/router` 生成无手改；`/swagger/index.html` 可见全部新路由与注解 (covers: S2.3)
- [x] T3: 发送服务与通道 handler（渲染 + MAIL net/smtp + 百度 BCE 签名 + 失败落日志）+ 单测 — acceptance: `go test` 新增用例全绿；send-test 对未接入 supplier 返回明确错误且 log 表有记录 (covers: S2.4; depends: T2)
- [x] T4: 四资源 CRUD service/repository/handler — acceptance: curl 四组 list/get/add/edit/del 走通信封；权限码与迁移一致 (covers: S2.3; depends: T2)
- [x] T5: 前端 api/model + 4 页移植（含动态字段表单与测试发送弹窗）— acceptance: eslint/vue-tsc 绿；浏览器四页 CRUD 冒烟、defineOptions 名与路由一致 (covers: S2.5; depends: T1, T4)
- [x] T6: 端到端验收——菜单/四页冒烟、send-test 落记录、`WEB=1 ./script/check.sh`、独立 reviewer — acceptance: S2.6 四条全过 (covers: S2.6; depends: T3, T5)
