---
name: gosilite-dev
description: Use when developing, reviewing, or extending web API features with the gosilite Go + Vue3 framework — Hertz Handler/Service/Repository layering, IDL + hz codegen, OAuth2-style session tokens, unified {code,msg,data} envelope, PostgreSQL/Redis/RustFS infra, TDesign Vue3 frontend, and the Compose-aligned requirements loop. Trigger on gosilite, Hertz API, user module, "用 gosilite 开发接口", or adding a business module.
---

# gosilite 开发技能

把 Web 功能需求变成可验收的 Go + Vue3 代码：**IDL 契约 + 分层实现 + 统一响应 + 前后端联调**。

## 何时使用

- 在本仓库新增/修改 HTTP API 业务功能
- 调整用户会话、权限、登录/刷新/登出契约
- 修 bug、补测试、补 schema
- 评审 gosilite 风格的代码
- 为 Agent 驱动开发扩展脚手架、文档或 skill

## 项目事实（禁止凭记忆写旧框架）

| 项 | 当前实际 |
|----|----------|
| 模块路径 | `gosilite` |
| HTTP 框架 | CloudWeGo Hertz（禁用 Gin / 旧标准库业务 Handler） |
| 入口 | 仓库根目录 `main.go`，`go run .` |
| 配置 | `conf/<env>/conf.yaml`，`GO_ENV=test\|dev\|prod` |
| 响应信封 | `{"code":<int32>,"msg":string,"data"?:any}` |
| 错误 | `pkg/errorx` + `types/errno` 数字码，经 `R.Error` 输出 |
| 会话契约 | `SessionTokenResponse`：`accessToken` / `tokenType` / `expiresIn` / `refreshToken`（camelCase） |
| 数据库 | PostgreSQL + GORM + gen（`application/*/dal`） |
| 缓存 | Redis（refresh token / 吊销） |
| 对象存储 | RustFS（S3 兼容，默认 host `19000`） |
| 前端 | `web/` = TDesign Vue3 Starter（Pinia + Axios + vue-router） |
| 生成代码 | `hz` / thriftgo，落在 `api/model`、`api/router` |

## 硬性约定

1. **分层单向**：`api/handler → application/service → application/dal/repository → application/infra`，禁止跨层；跨域装配只经 `application.Module`。
2. **Service 无接口**：`type Service struct` + `NewService`；`InitService` 只安装包级门面；Repository 保留接口。
2. **错误只用 `errorx`**：`errorx.New(errno.Xxx, errorx.KV("k","v"))` / `errorx.WrapByCode`；Handler 不吞错，统一 `R.Error`。
3. **响应只走 `pkg/utils/response`**：`SuccessWithData` / `Error` / `Unauthorized` / `InvalidParamRequestResponse`；禁止 Handler 手写 JSON。
4. **契约源是 IDL**：`idl/**/*.thrift` 先改，再 `hz update`；不手改 `api/model` 生成文件。
5. **Hertz API 以官方文档为准**：`https://cloudwego.cn/zh/docs/hertz/`；绑定用 `BindAndValidate`，不发明 Gin 风格 API。
6. **会话用 OAuth2 风格字段**：禁用 `token`/`expire` 混用；JSON 字段 camelCase 与前端 `web/src/api/user.ts` 对齐。
7. **验收**：
   ```bash
   ./script/check.sh          # go test + vet + build
   WEB=1 ./script/check.sh    # 需要前端时
   ```

## 插件化 Module

- 契约：`application.Module` — `Name` / `Init(ctx, deps)` / `RegisterRoutes(h)`。
- 注册：`application.builtinModules()`；`main` → `Bootstrap` + `MountRoutes`。
- 内置：`application/framework`（user/role/menu 路由）、`application/agent`（draft 路由）。
- 新业务域：写 `application/<域>/module.go` 实现 Module，挂进注册表；Service 用**具体类型**，不要 `I*Service`。

## 标准开发流程（与 Compose Next 对齐）

1. **定契约**：路由、DTO 字段（camelCase）、错误码、表、权限。
2. **写 spec**（功能值得成文时）：`docs/compose/spec/<feature>.md`，含 `[S1] Problem` / `[S2] Design` / `[S3] Out of Scope` / `Tasks`。
3. **改 IDL** → `hz update -I idl -idl idl/<domain>/<file>.thrift -customize_package idl/package.yaml`。
4. **Repo → Service → Handler → Router**，每层可测。
5. **前端接入**：`web/src/api/*.ts` 类型与 envelope、`store` / 登录态、必要页面。
6. **测试 + 验收四连**，修到全绿。
7. **评审**（需要时）→ spec `status: delivered`。

## 新增业务模块骨架

以 `framework/user` 为参考实现，新模块目录：

```text
idl/<domain>/<feature>.thrift
api/model/<domain>/<feature>.go          # hz/thriftgo 生成，勿手改
api/handler/<domain>/<feature>_controller.go
api/router/<domain>/<feature>.go         # 生成
api/router/<domain>/middleware.go        # 生成后可改权限挂载
application/<domain>/service/<feature>/  # service.go + service_impl.go
application/<domain>/dal/model/          # gorm gen 模型
application/<domain>/dal/query/          # gorm gen query
application/<domain>/dal/repository/     # 接口 + 实现
types/errno/<feature>.go                 # 错误码注册
web/src/api/<feature>.ts                 # 前端契约
```

组装：Service 在 `application/.../service.go` 的 `InitService` 注入 `db` / cache / storage；Handler 只调 `XxxSVC`。

## 会话与认证（当前实现）

登录成功返回：

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "accessToken": "…",
    "tokenType": "Bearer",
    "expiresIn": 3600,
    "refreshToken": "…"
  }
}
```

- 路由：`POST /user/login`、`POST /user/refresh`、`POST /user/logout`、`POST /user/password`、`GET /user/info`、`GET /user/profile`
- Access token：Hertz JWT，`Authorization: Bearer <accessToken>`
- Refresh token：Redis 存储，支持 rotation / revoke；改密后吊销该用户全部 refresh
- 登出：可选 `refreshToken` + 吊销当前 access token（denylist）
- 前端：`web/src/api/user.ts` + `web/src/store/modules/user.ts`（单飞 refresh）

新增认证相关接口必须同时改：IDL、Handler、`web/src/api`、本技能中的契约表。

## 错误模型

- 业务错误码在 `types/errno` 用 `code.Register` 注册模板消息。
- `errorx.StatusError` → HTTP 200 + `{"code":<业务码>,"msg":...}`（非 2xx 仅用于鉴权失败 401 / 参数 400）。
- 稳定性敏感错误 `code.WithAffectStability(false)`。
- 前端依赖的错误码（如 `ErrUserAuthenticationFailed = 700012006`）**禁止改号**。

## 配置与基础设施

- 配置结构见 `conf/conf.go`：`server` / `jwt` / `database` / `redis` / `storage`。
- 本地依赖：`docker compose up -d` → PostgreSQL 5432、Redis 6379、RustFS 19000/19001。
- JWT secret 必须 ≥32 字符且显式配置。
- 密码只存 `pkg/bcrypt` 哈希。
- 存储兼容 RustFS：HeadBucket 失败允许 CreateBucket，错误码兼容 `NotFound`/`NoSuchBucket`/`404`。

## 测试要求

- Service：可注入仓库/缓存；密码、会话、权限分支要有用例。
- Handler/HTTP：断言 status + `code/msg/data` envelope，不只测 mock 调用。
- 修 bug 先写失败测试再改。
- 详见 `docs/patterns/testing.md`。

## 必读文档

- `AGENTS.md` — 目录边界与硬规则
- `README.md` — 启动与命令
- `docs/patterns/design.md` — 架构与设计模式
- `docs/patterns/coding.md` — 编码规范
- `docs/patterns/testing.md` — 测试规范
- `docs/process/requirements.md` — Compose 需求闭环
- `docs/hertz/api-index.md` — Hertz 速查
- `application/framework/` + `api/handler/framework/user/` — 端到端参考

## 检查清单

- [ ] 无 Handler 业务逻辑 / 无 Repository 外泄漏 SQL
- [ ] JSON 契约 camelCase，与 `web/src/api` 一致；无 `token`/`expire` 旧字段混用
- [ ] 响应只经 `R.SuccessWithData` / `R.Error`，信封 `{code,msg,data}`
- [ ] 错误码在 `types/errno` 注册，禁止魔法数字散落
- [ ] IDL 变更已重新生成，生成文件无手改
- [ ] 列表接口分页；管理接口挂 `permission` 中间件
- [ ] `go test ./...`、`go vet ./...`、`go build ./...`、`cd web && npm run build` 全绿