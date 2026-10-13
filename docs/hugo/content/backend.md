---
title: "后端"
weight: 3
---

gosilite 后端是单一 CloudWeGo Hertz 服务（`GO_ENV=test go run .`，端口 8888）：按 Handler → Service → Repository → infra 分层，业务以 Module 插件挂载，HTTP 契约全部由 IDL 经 `hz` 生成，响应统一 `{code,msg,data}` 信封。本文先讲全项目通用约定，再逐模块给出「一句话职责 + 关键表/路由/权限码/入口文件」，最后是生成边界与迁移。全部结论来自仓库代码与 SQL，可与[概览](/docs/overview/)、[快速开始](/docs/quick-start/)互相印证。

## 先行约定

### 分层职责

| 层 | 位置 | 职责 |
|---|---|---|
| Handler | `api/handler/` | 绑定/校验参数、调 Service、经 `pkg/utils/response` 回包、写 Swagger 注解 |
| 中间件 | `api/middleware/` | 登录与权限（`permission`）、操作日志（`operlog`）、请求级缓存（`ctxcache`） |
| Service | `application/framework/service/` | 业务规则；**具体类型** `Service` + `NewService`/`InitService`，禁止 `I*Service` 接口 |
| Repository | `application/framework/dal/repository/` | GORM 仓储，Handler 不得直接碰库 |
| 模型/查询 | `application/framework/dal/model`、`dal/query` | gorm gen 生成物 |
| infra | `application/infra/` | PG / Redis / RustFS / 图片 / GORM 钩子 |
| 装配 | `application/module.go`、`application/application.go` | Module 契约与 Bootstrap |

### Module 插件机制

- `Module` 接口 = `Name` + `Init(ctx, deps)` + `RegisterRoutes(r)`（`application/module.go:12`）。
- `builtinModules()` 是唯一注册点，后端业务全部收口在一个 `framework.NewModule()`（`application/application.go:21-23`）；`Bootstrap` 先 `infra.Init` 再逐模块 `Init`，`MountRoutes` 统一挂路由。
- hz 生成的 `api/router/register.go:GeneratedRegister` 只供 hz 追加代码，**运行时不调用**；路由由 Module 的 `RegisterRoutes` 挂载。

### 响应信封与错误码

- 统一 `{"code": int32, "msg": string, "data": …}`；成功 `code=200, msg="success"`。业务错误经 `errorx.StatusError` 由 `response.Error` 输出：**HTTP 状态仍为 200，错误语义由 code 表达**。
- HTTP 的 i64 标识字段在 IDL 源头使用 `json:"field,string"`；前端将其作为字符串传递，避免 JavaScript `Number` 精度丢失。
- 错误码集中在 `types/errno/*.go`，按域分段并在 `init()` 注册文案：Passport `700xxxxxx`（如 `700012006` 认证失败，**前端依赖、勿改**，`user.go:7`）、系统 `701000001+`（`system.go:7-18`）、对象存储 `701000034+`（`oss_config.go:6-11`）等。

### Swagger 注解（强制）

- 每个对外 Handler 必须写 `@Summary` `@Tags` `@Produce` `@Success` `@Router`，有参数再写 `@Param`（GET 用 query、其余用 body，类型写 `api/model` 包名）。
- 改注解后执行 `swag init -g main.go -o ./docs/swagger --parseDependency --parseInternal`；`main.go` 空导入 `_ "gosilite/docs/swagger"`，UI 地址 `GET /swagger/index.html`（`main.go:116` 注册 `/swagger/*any`，免登录）。

## 认证与会话（service/user）

一句话：签发 JWT access token + Redis refresh token，支持轮换、吊销与在线会话管理。

| 项 | 事实 |
|---|---|
| 路由 | `POST /user/login`、`POST /user/refresh` **免登录**（`api/router/framework/user/middleware.go:26,58` 返回 nil）；`GET /user/info`、`GET /user/profile`、`POST /user/logout`、`POST /user/password`、`PUT /user/profile/avatar` 挂 `permission.Default()`（仅校验登录） |
| 会话字段 | `accessToken` / `tokenType` / `expiresIn` / `refreshToken`，camelCase（`idl/framework/user.thrift:53-57`） |
| 登录防护 | YAML `login_security` 提供默认值；系统参数 `sys.login.max_failures`、`sys.login.failure_window_minutes`、`sys.login.lock_minutes` 可即时覆盖。15 分钟内默认 5 次错误会临时锁定 30 分钟；累计 10 次错误永久锁定，管理员重置密码后解除。所有失败统一返回凭据错误，不暴露锁定状态。 |
| refresh 存储 | Redis，前缀 `gosilite:auth:refresh:`（`types/consts/consts.go:13`）；`SessionStore` 签发/校验/吊销（`service/user/session.go:39,55,77`） |
| 轮换与吊销 | 刷新即轮换并 `MoveOnline` 迁移在线会话；`RevokeUserRefreshTokens` 按用户代际整批吊销（改密时调用，`session.go:84`）；access 可进黑名单 `DenyAccessToken`（`session.go:89`，前缀 `gosilite:auth:access_deny:`） |
| 在线会话 | Redis `gosilite:auth:online_sessions`（无数据表，见 `0015` 迁移注释），供在线用户页与强退使用 |
| 身份注入 | 登录校验通过后 `ctxutil.SetLoginUser` 写入 context（`application/base/ctxutil/api_auth.go:9`）；启动时 `permission.Register(conf.JWT)` 装载 JWT 配置（`main.go`） |

入口文件：`application/framework/service/user/session.go`、`service_impl.go`；Handler `api/handler/framework/user/user_controller.go`。

## 用户与 RBAC（service/user、role、menu）

一句话：用户、角色、菜单的 CRUD 与授权；菜单树是前端侧栏和可访问路由的唯一来源。

| 项 | 事实 |
|---|---|
| 路由 | `/system/user/*`（list/create/update/remove/status/roles）、`/system/role/*`（含 `GET/PUT /system/role/authorize`）、`/system/menu/*`（tree/list/CRUD）、`GET /menu/routes`（挂 `permission.Default()`，登录即可，见 `menu/middleware.go:11-13`） |
| 权限码 | `system:user:list/add/edit/remove`、`system:role:list/add/edit/remove/assign`、`system:menu:list/add/edit/remove`（出处 `api/router/framework/system/middleware.go`） |
| 表 | `system_user`、`system_user_role`、`system_role`、`system_role_permission`、`system_menu`、`system_role_menu`、`system_user_post`、`system_role_dept`（均位于 v0.1.0 最终 schema） |
| 数据范围 | `system_role.data_scope`：1 全部 / 2 本部门 / 3 本部门及以下 / 4 仅本人 / 5 自定义（`0002_system_rbac.sql:52`） |
| 鉴权实现 | `permission.Default()` = 仅登录；`permission.HPermissions(...)` = 权限码校验，支持 `*` 通配（`api/middleware/permission/permission.go:196,204,232`）；码模式统一为 `system:<域>:<动作>` |

## 组织：部门与岗位（service/dept、post）

一句话：树形部门与岗位维护，并作为角色数据范围与用户归属的载体。

| 项 | 事实 |
|---|---|
| 路由 | `/system/dept/tree|list` 与增删改、`PUT /system/dept/leader`；`/system/post/list` 与增删改（`idl/framework/system.thrift:267-278`） |
| 权限码 | `system:dept:list/add/edit/remove`、`system:post:list/add/edit/remove` |
| 表 | `system_dept`、`system_post`（`0002_system_rbac.sql`），关联 `system_user_post`、`system_role_dept`；部门/岗位种子数据同在 0002 |
| 组织架构页 | 菜单 44「组织架构」，权限码 `system:org:list`（`0007_org_contacts.sql:18`） |

## 字典（service/dict）

一句话：维护下拉框/枚举类的基础数据，供前后端统一取值。

| 项 | 事实 |
|---|---|
| 路由 | `/system/dict/type/*`、`/system/dict/group/*`、`/system/dict/data/*`（含 `GET /system/dict/data/options` 取选项，`idl/framework/system_dict.thrift:148-162`） |
| 权限码 | `system:dict:list/add/edit/remove` |
| 表 | `system_dict_type`、`system_dict_data`（`0003_system_dict_param.sql`）、`system_dict_group`（`0008_dict_group.sql`） |

## 系统参数（service/config）

一句话：键值对运行时参数，含一份匿名可读的全局参数子集。

| 项 | 事实 |
|---|---|
| 路由 | `/system/config/list|key` 与增删改（权限码 `system:config:list/add/edit/remove`，`syscfg/middleware.go`）；`GET /system/config/global` **免登录**（`syscfgglobal/middleware.go` 全部返回 nil） |
| 表 | `system_config`（`0003_system_dict_param.sql`，消费方式见 `0006_config_consumption.sql`） |
| 匿名口径 | `config_type=global` 且启用才从 global 接口出（`service/config/service.go:77-78`） |

## 通知公告（service/notice）

一句话：管理员发布公告，登录用户在通知中心收件箱阅读。

| 项 | 事实 |
|---|---|
| 管理路由 | `/system/notice/list|id` 与增删改，权限码 `system:notice:list/add/edit/remove`（`notice/middleware.go`） |
| 消费路由 | `/notice/inbox`、`/notice/detail`、`/notice/unread-count`、`/notice/read`——仅登录、不占管理权限码（`notice/middleware.go:46` 注释） |
| 表 | `system_notice`（`0011_notice.sql`）、`system_notice_read`（`0012_notice_read.sql`） |

## 日志（service/syslog）

一句话：操作日志、异常日志、登录日志三类落库，全部由 `operlog` 中间件异步写入。

| 项 | 事实 |
|---|---|
| 采集 | `api/middleware/operlog/operlog.go:40` 的 `MW()` 挂在全局（`main.go:205`）：请求结束后分类（操作/异常/登录），投递有界异步队列；登录/登出走专表且 **request body 永不落库**（`operlog.go:69`），敏感报文脱敏 |
| 路由 | `/system/operlog/list|id|clean`、`/system/errorlog/list|id|handle|clean`、`/system/loginlog/list`（`idl/framework/system_log.thrift:150-157`） |
| 权限码 | 操作日志 `system:operlog:list/handle/clean`；异常日志**复用** `system:operlog:*`（`log/middleware.go:26-36`）；登录日志 `system:loginlog:list` |
| 表 | `system_oper_log`、`system_error_log`、`system_login_log`（均建于 `0013_log.sql`） |
| 关停排空 | `main.go` 注册 `OnShutdown` 刷队列（配合 `application/base/logq` 的 `Flush`） |

## 监控（service/monitor）

一句话：在线用户、数据监控、缓存监控三个管理页 + 指标出口 `/metrics`，接口文档独立成页。

| 项 | 事实 |
|---|---|
| 在线用户 | `GET /system/online/list`、`POST /system/online/force`，权限码 `system:online:list/force`；数据来自 Redis 在线会话，菜单 4701（`0015`） |
| 数据监控 | `GET /system/monitor/datasource|server|sql|firewall|uri|http|storage`，全部 `system:monitor:list`（`monitor/middleware.go:40-99`）；菜单 4702（`0016`） |
| 缓存监控 | `GET /system/cache/prefix/list`、`/system/cache/key/list`、`/system/cache/value`，删除/清空 `POST /system/cache/key/delete`、`/system/cache/prefix/clear`；`system:cachemonitor:list/delete`；菜单 4703（`0017`） |
| 系统接口页 | 菜单 4704，权限码 `system:apimonitor:list`，页面内嵌 Swagger UI，位于系统工具（component `tool/api/index`） |
| `/metrics` | Prometheus 指标走**独立端口**：`monitor.enable_metrics`、`metrics_addr: ":9091"`、`metrics_path: "/metrics"`（`conf/conf.yaml:42-44`），与业务 8888 隔离；底座在 `application/base/metrics` |
| 入口文件 | `application/framework/service/monitor/service.go`（页面数据）+ `cache.go`（缓存探针）；Handler `api/handler/framework/monitor/` |

## 存储（service/ossconfig、ossfile、ossrule + common.file）

一句话：存储配置、文件台账、访问规则三件套，外加全站统一的上传/下载口 `/common/file/*`。

| 组件 | 关键事实 |
|---|---|
| 存储配置 ossconfig | 路由 `/system/oss/config/*`（含 `status` 启停、`test` 连通性测试）；权限码 `system:ossConfig:list/add/edit/remove/test`（驼峰，`ossconfig/middleware.go`）；表 `system_oss_config`（`0021_oss.sql`）；默认配置不可删（`errno/oss_config.go:8`） |
| 文件台账 ossfile | 路由 `/system/oss/list|id` 与 `PUT/DELETE`；权限码 `system:oss:list/edit/remove`（另有 `system:oss:lock` 按钮码，`0021_oss.sql:168`）；表 `system_oss`；上传上限 20MB（`service.go:24`）；含旧存储管理吸收来的桶视图 |
| 访问规则 ossrule | 路由 `/system/oss/rule/*`（含 `PUT .../overwrite` 覆盖开关、`DELETE .../cache` 刷新缓存）；权限码 `system:ossRule:list/add/edit/remove/refresh`；表 `system_oss_rule` |
| 统一文件 common.file | `POST /common/file/upload`（登录即可，multipart 字段 `file`，≤20MB）、`GET /common/file/download`（按台账 id 回流附件，**非 JSON 信封**）——`common_file_controller.go:16-71`；台账由 `ossfile.Service.Upload` 自动记录 |
| 动态切换 | `infra/storage/dynamic.Controller` 按当前启用的配置动态解析客户端，热切换时 `Invalidate()`（`dynamic/dynamic.go:75-127`） |

## 定时调度（service/scheduler）

一段已编译的 Job 可由管理员配置 cron、超时、重试和全局并发实例上限；每次运行保存不可变快照、持久化日志，并通过 SSE 回放和实时展示。

| 项 | 事实 |
|---|---|
| 路由 | `/system/scheduler/job/*`、`/system/scheduler/run/*` 与日志 SSE；权限码 `system:scheduler:*` |
| 表 | `system_job`、`system_job_run`、`system_job_run_log` |
| 集群安全 | Redis 执行租约 + 数据库容量槽；停止为协作式取消，不强杀 goroutine |
| Job 组织 | 稳定 key 在 `types/jobkeys/`；框架 Job 在 `application/framework/jobs/`；领域 Job 在 `application/<domain>/jobs/` |
| 日志保留 | `scheduler.log_retention_days`，默认 30 天；仅清理执行日志，不删除运行摘要 |

## 基础设施 infra

一句话：`application/infra` 提供连接与资源底座，`application/base` 提供请求上下文与可观测底座。

| 包 | 职责与出处 |
|---|---|
| `infra/orm` | GORM + PG 连接；池参数 `max_idle_conns`/`max_open_conns`（`conf/conf.go` Database 段） |
| `infra/cache` | Redis 客户端接口 `CacheClient`（`cache/cache.go:19`），字符串/哈希/列表/管道等分接口，实现在 `impl/` |
| `infra/storage` | 对象存储抽象 `storage.go` + S3 实现（RustFS 兼容），`dynamic/` 动态路由，`internal/` 内部工具 |
| `infra/imagex` | 图片上传鉴权与取资源选项（`upload_auth_opt.go`、`get_resource_opt.go`） |
| `infra/hook` | GORM 插件 `BaseModelPlugin`：`beforeCreate/Update/Delete` 统一填充模型基字段（`grom_hook.go:18-52`） |
| `base/ctxutil` | `LoginUser` 的 context 存取与超管判定（`api_auth.go`） |
| `base/logq` | 日志有界异步队列：`Set` 安装 Sink、`Flush` 关停排空（`logq.go:84-96`） |
| `base/metrics` | 进程内 Prometheus 指标底座 + `/metrics` 文本出口（`metrics.go:52`）与 GORM SQL 统计 |
| `base/useragent` | UA 解析为浏览器/系统（`useragent.go:9`），供日志记录 |

## 生成边界

| 要改什么 | 动哪里 | 怎么再生 |
|---|---|---|
| HTTP 契约/路由 | `idl/framework/*.thrift` | `hz update -I idl -idl <file> -customize_package idl/package.yaml` |
| 表结构 | `script/schema/v0.1.0.sql` | 在空库初始化后重跑 gorm gen / `script/codegen` |
| 表模型与查询 | 生成源（表 + `script/codegen/codegen.go`） | gorm gen → `dal/model/*.gen.go`、`dal/query/*.gen.go` |
| CRUD/Service 初稿 | `script/codegen/templates/*.tmpl` | `script/codegen`（**只生成不存在的文件**，之后归手写） |
| 业务规则/装配/前端 | service、module、`web/src` | 不生成，直接改 |

规则（AGENTS.md 二）：

- 标 `DO NOT EDIT` / `*.gen.go` 的文件（`dal/query/*.gen.go`、`dal/model/*.gen.go`、`api/model/**`）**禁止手改**；只改生成源后重新生成。
- `api/router/**` 的 `Register` 必须由 hz 从 IDL 生成，禁止手写；生成后的 `middleware.go` 允许补权限挂载。
- 顺序永远是：**先 IDL 后生成，先表后 gen，先 Repository 后 Service 后 Handler。**

## v0.1.0 数据库基线

- 最终基线是 `script/schema/v0.1.0.sql`，包含当前全部 DDL、索引、约束及必要系统种子。
- 只允许在空 PostgreSQL 16 数据库中运行 `./script/init-schema.sh`；脚本发现核心表即拒绝执行，不会删除或覆盖数据。
- v0.1.0 删除历史增量迁移链与旧 `postgres.sql`，不支持既有开发库就地升级。旧库需先备份，再创建空库初始化。
- 新增模块涉及表结构时，更新当前 schema，在干净库演练后再执行 gorm/codegen。发布与回滚边界见[发布 v0.1.0](/docs/release/)。

命令速查、配置表与权限码模式汇总见[参考](/docs/reference/)；前端对接方式见[前端](/docs/frontend/)。
