---
name: gosilite-dev
description: Use when developing, reviewing, or delivering web API features with the gosilite Go + Vue3 framework — Hertz Handler/Service/Repository layering, IDL + hz codegen, OAuth2-style session tokens, unified code/msg/data envelope, PostgreSQL/Redis/RustFS, requirements loop, and acceptance checks. Trigger on gosilite, Hertz API, user module, "用 gosilite 开发接口", 新增业务模块, 修 bug, 补测试, 改 IDL/错误码/会话, or reviewing this codebase. Also use for closed-loop feature delivery from requirement to check.sh green.
---

# gosilite 开发技能

把 Web 功能需求变成可验收的 Go + Vue3 代码：**IDL 契约 + 分层实现 + 统一响应 + 前后端联调 + 验收闭环**。

## Important（硬性约定）

1. **分层单向**：`api/handler → application/service → application/dal/repository → application/infra`，禁止跨层；跨域装配只经 `application.Module`。
2. **Service 无接口**：`type Service struct` + `NewService`；`InitService` 只安装包级门面；Repository 保留接口。
3. **错误只用 `errorx`**：`errorx.New(errno.Xxx, errorx.KV("k","v"))` / `errorx.WrapByCode`；Handler 不吞错，统一 `R.Error`。
4. **响应只走 `pkg/utils/response`**：`SuccessWithData` / `Error` / `Unauthorized` / `InvalidParamRequestResponse`；禁止 Handler 手写 JSON。
5. **契约源是 IDL**：`idl/**/*.thrift` 先改，再 `hz update`；不手改 `api/model` / `*.gen.go` 生成文件。
6. **Hertz API 以官方文档为准**：`https://cloudwego.cn/zh/docs/hertz/`；绑定用 `BindAndValidate`，不发明 Gin 风格 API。速查见 `references/hertz-api.md`。
7. **会话用 OAuth2 风格字段**：禁用 `token`/`expire`；JSON 字段 camelCase 与 `web/src/api` 对齐。
8. **模型转换集中 `convert.go`**：dal model ↔ DTO/VO 映射写在对应包 `convert.go`，禁止内联散落（细则 `references/coding.md`）。
9. **验收**：
   ```bash
   ./script/check.sh          # go test + vet + build
   WEB=1 ./script/check.sh    # 需要前端时
   ```
10. **需求澄清是强制门禁（Gate）**：任何功能变更在**编码前**必须按 `references/requirements.md` §1.6 输出【我的理解】【我的假设】【需你确认】并**等待用户确认**，未确认不得进入契约与实现（仅当用户显式说「跳过澄清直接做」，或已在对话中给出完整需求与验收标准时可豁免）。闭环：澄清 → 契约 → 实现 → 测试 → 验收；细则 `references/requirements.md` §1.0。

## 何时使用

- 新增/修改 HTTP API、业务模块、迁移、表结构
- 调整用户会话、权限、登录/刷新/登出契约
- 修 bug、补测试、改 IDL/错误码
- 评审 gosilite 风格代码
- 从需求到可交付的完整功能闭环

写 `web/` 时同时加载 `.agents/skills/frontend-dev/SKILL.md`。

## 项目事实（禁止凭记忆写旧框架）

| 项 | 当前实际 |
|----|----------|
| 模块路径 | `gosilite` |
| HTTP 框架 | CloudWeGo Hertz（禁用 Gin / 旧标准库业务 Handler） |
| 入口 | 仓库根 `main.go`，`go run .`（`server.New` + 显式 Recovery） |
| 配置 | `conf/<env>/conf.yaml`，`GO_ENV=test\|dev\|prod` |
| 响应信封 | `{"code":<int32>,"msg":string,"data"?:any}` |
| 错误 | `pkg/errorx` + `types/errno` 数字码，经 `R.Error` 输出 |
| 会话契约 | `SessionTokenResponse`：`accessToken` / `tokenType` / `expiresIn` / `refreshToken` |
| 数据库 | PostgreSQL + GORM + gen（`application/*/dal`） |
| 缓存 | Redis（refresh token / 吊销） |
| 对象存储 | RustFS（S3 兼容，默认 host `19000`） |
| 前端 | `web/` = TDesign Vue3 Starter（Pinia + Axios + vue-router） |
| 生成代码 | `hz` / thriftgo → `api/model`、`api/router`；gorm gen → `dal/model|query` |

## Instructions：标准开发流程

1. **澄清需求（强制门禁）**：按 `references/requirements.md` §1.6 输出理解/假设/关键提问（≤3）并**等待用户确认**，确认后才进第 2 步；能推断的不问，但假设必须列出。
2. **定契约**：路由、DTO 字段（camelCase）、错误码、表、权限码（`system:*:xxx`）。
3. **写 spec**（功能值得成文时）：`docs/compose/spec/<feature>.md`（Problem / Design / Out of Scope / Tasks）。
4. **迁移** `script/migrations/NNNN_*.sql`（幂等 DDL + 种子）。
5. **生成**：先改 IDL → `hz update -I idl -idl <file> -customize_package idl/package.yaml`；表模型 `go run ./script/codegen`。
6. **Repo → Service → Handler → Router**，每层可测；model↔DTO/VO 进 `convert.go`。**路由必须 IDL → hz 生成，禁止手写 `Register`**；免登录等特殊分组用独立 IDL 文件生成独立 `Register`/`middleware.go`，生成后仅允许在 `middleware.go` 挂权限。
7. **前端接入**：`web/src/api/*.ts`、store/页面；`t-*` 先 TDesign MCP `get-component-docs`。
8. **测试 + `./script/check.sh`**（前端 `WEB=1`），修到全绿。
9. **评审**（需要时）→ spec `status: delivered` + Report。

顺序约束：**先 IDL 后生成；先生成源后手改生成物（禁止直接改生成物）；先表后 gen；先 Repository 后 Service 后 Handler。** 改已交付契约必须同步：IDL、错误码、前端类型、本 skill 契约表。

详细设计模式见 `references/design.md`；编码细则 `references/coding.md`；测试 `references/testing.md`。

## 参考速查

### 插件化 Module

- 契约：`application.Module` — `Name` / `Init(ctx, deps)` / `RegisterRoutes(h)`。
- 注册：`application.builtinModules()`；`main` → `Bootstrap` + `MountRoutes`。
- 新业务域：`application/<域>/module.go` 实现 Module 挂进注册表；Service 用**具体类型**。

### 新增业务模块骨架

以 `framework/user` 为参考：

```text
idl/<domain>/<feature>.thrift
api/model/<domain>/<feature>.go          # 生成，勿手改
api/handler/<domain>/<feature>_controller.go
api/handler/<domain>/convert.go
api/router/<domain>/<feature>.go         # 生成
api/router/<domain>/middleware.go        # 生成后可改权限挂载
application/<domain>/service/<feature>/  # 具体 Service
application/<domain>/dal/model|query|repository/
types/errno/<feature>.go
web/src/api/<feature>.ts
```

组装：Service 在 `InitService` 注入 `db` / cache / storage；Handler 只调 `XxxSVC`。

### 会话与认证

登录成功 `data`：`accessToken` / `tokenType` / `expiresIn` / `refreshToken`。

- 路由：`POST /user/login|refresh|logout|password`，`GET /user/info|profile`（logout 挂 `permission.Default()`：登出为已认证动作，同时供登录日志记录登出用户）
- Access：Hertz JWT `Authorization: Bearer`；Refresh：Redis rotation/revoke；改密吊销全部 refresh
- 登出：吊销 refresh + access denylist
- 前端：`web/src/api/user.ts` + `store/modules/user.ts`（单飞 refresh）

新增认证接口必须同改：IDL、Handler、`web/src/api`、本 skill 契约表。

### 错误模型

- 错误码在 `types/errno` `code.Register`；业务失败 HTTP 200 + 业务 `code`。
- 401/400 仅鉴权/参数；前端依赖的错误码（如 `ErrUserAuthenticationFailed = 700012006`）**禁止改号**。
- 稳定性敏感错误 `code.WithAffectStability(false)`。

### 配置与基础设施

- 配置：`conf/conf.go` — `server` / `jwt` / `database` / `redis` / `storage`。
- 本地：`docker compose up -d` → PG 5432、Redis 6379、RustFS 19000/19001。
- JWT secret ≥32 字符；密码只存 `pkg/bcrypt`。
- 存储兼容 RustFS：HeadBucket 失败允许 CreateBucket；错误码兼容 `NotFound`/`NoSuchBucket`/`404`。

### 系统管理（system-rbac）

- 五域：部门树 / 岗位 / 用户 / 菜单 / 角色（data_scope + role_menu）。
- IDL：`idl/framework/system.thrift` → hz；Handler `api/handler/framework/system`。
- 迁移：`script/migrations/0002_system_rbac.sql`；路由 framework Module `RegisterRoutes`；`_systemMw` 挂 `permission.Default()`。
- 前端：`web/src/pages/system/*` + `web/src/api/system/*` + `router/modules/system.ts`。
- 字典分组（`0008_dict_group.sql`）：`system_dict_group` + `group_id`（0=未分组）；分组 CRUD `/system/dict/group/*`，复用 `system:dict:*`；字典页左分组 + 右类型表。

## Examples

**「帮我写个用户登录接口」** → 读 README/AGENTS/user service → 按模板输出理解/假设/≤3 个关键问 → 复用现有会话契约 → IDL/hz（若需）→ service 测试（含轮换）→ handler → `./script/check.sh` 绿。

**「列表接口加筛选」** → 查现有 handler/service/repository → 确认分页与字段白名单 → 改 Service 校验 + Repository 查询 → 补测试 → 前端 `web/src/api` 同步字段 → 验收。

## Troubleshooting

- **hz update 冲掉手改**：生成物勿手改；权限挂载只改 `middleware.go` 骨架内；改契约先改 IDL。
- **会话字段不一致**：必须 `accessToken`/`tokenType`/`expiresIn`/`refreshToken`；同步 IDL + 前端类型。
- **不确定 Hertz API**：查 `references/hertz-api.md` 或官方文档，禁止 Gin 记忆。
- **分层被破坏**：SQL 只在 repository；业务只在 service；Handler 只绑定/调用/回包。

## 检查清单

- [ ] **编码前已完成需求澄清并获用户确认**（或属 §1.0 豁免，已复述理解）
- [ ] 无 Handler 业务逻辑 / 无 Repository 外泄漏 SQL
- [ ] model↔DTO/VO 转换集中在 `convert.go`
- [ ] JSON camelCase，无 `token`/`expire` 旧字段
- [ ] 响应只经 `R.*`，信封 `{code,msg,data}`
- [ ] 错误码在 `types/errno` 注册，前端依赖码未改号
- [ ] IDL 变更已重新生成，生成文件无手改
- [ ] 列表分页；管理接口挂 `permission`
- [ ] 会话改动已测轮换/吊销/改密
- [ ] 前端 `t-*` 已 TDesign MCP 核对
- [ ] `./script/check.sh` 全绿（前端 `WEB=1`）
- [ ] spec（如有）status/tasks/Report 已更新
