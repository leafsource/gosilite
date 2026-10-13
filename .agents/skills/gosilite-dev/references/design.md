# 架构与设计模式

以本仓库实际代码为准，不套用其他框架记忆。Handler/Service/Repository 细则见 `coding.md`。

## 1. 分层 + 单向依赖

```text
web (TDesign Vue3)
        │  HTTP JSON {code,msg,data}
        ▼
api/handler  ──►  application/service  ──►  application/dal/repository  ──►  application/infra
  Hertz           具体 Service 类型          GORM Query/SQL                PG / Redis / RustFS
        │               │
   api/model DTO    errorx + types/errno
        │
application.Module (Init + RegisterRoutes) ──► 挂载 api/router 生成物
```

- Handler：绑定校验、调用 Service、映射响应；**不写业务规则、不碰 GORM**。
- Service：唯一持有业务规则、权限判断、事务边界；构造时注入 Repository / Cache / Storage。
- Repository：接口在 service 可见，实现用 gorm gen query；只做持久化。
- Infra：连接、缓存、对象存储；业务代码不直接 `redis.New` / `gorm.Open`。
- 路由与中间件挂载在 `api/router/**`；权限用 `api/middleware/permission`。

## 2. IDL 契约边界

- 公开 HTTP 契约以 `idl/**/*.thrift` 为源；`api/model`、`api/router` 由 `hz`/`thriftgo` 生成。
- JSON 字段名 **camelCase**（`accessToken`、`expiresIn`），与 `web/src/api` 一致。
- 生成文件禁止手改；调整字段先改 IDL 再生成。
- 手写扩展放在 handler / service / repository，不塞进生成目录。

## 3. Module 插件

- 契约：`application.Module`（`Name` / `Init` / `RegisterRoutes`）。
- 装配：`Bootstrap` 初始化 infra 并 Init 全部模块；`MountRoutes` 按序挂路由。
- 模块 `InitService` 建具体 Service 并安装包级门面；Handler 不依赖 `I*Service`。
- 新域 = 新 Module + `builtinModules()` 一行；路由来自 IDL/hz 生成的 `Register`。
- 内置：`application/framework`（user/role/menu/system）。

## 4. 统一错误与响应

成功：

```json
{ "code": 200, "msg": "success", "data": { ... } }
```

业务失败（HTTP 仍为 200，除非鉴权/参数）：

```json
{ "code": 700000003, "msg": "invalid username or password..." }
```

- 业务错误只用 `errorx.New` / `errorx.WrapByCode`，错误码在 `types/errno` 注册。
- 词表按域划分（用户 700 000 000 段），消息模板支持 `{param}`。
- 401/400 仅用于未认证、非法参数；其余业务语义用 `code` 表达。
- Handler 只用 `pkg/utils/response`：`SuccessWithData` / `Error` / `Unauthorized` / `InvalidParamRequestResponse`。

## 5. 会话（OAuth2 风格）

- 登录/刷新返回 `SessionTokenResponse`：`accessToken`、`tokenType`、`expiresIn`、`refreshToken`。
- **禁止** `token` + `expire` 旧字段对。
- Access = Hertz JWT（`Authorization: Bearer`）；Refresh = Redis，rotation + 按用户 generation 吊销。
- 登出：吊销 refresh + access denylist（TTL ≤ 24h）。
- 改密成功后吊销该用户全部 refresh token。
- 前端 `useUserStore` 对并发 401 做单飞 refresh，避免轮换竞态。
- 刷新是独立业务路由 `/user/refresh`，不是 JWT 库默认 token 延长。

## 6. Repository + GORM Gen

- 模型与 query 在 `application/<domain>/dal/{model,query}` 由 gen 生成；repository 手写。
- Repository 接口由 service 包依赖，便于测试替身。
- `RowsAffected == 0` 或 `gorm.ErrRecordNotFound` 映射 `errorx` 业务错误。
- 列表强制 `page`/`pageSize`，禁止无边界全量查询。
- 字段结构化；禁止 JSON blob 顶替业务列。

## 7. 中间件链（入口固定）

顺序：`ContextCache → pprof(可选) → gzip(可选) → RequestID → AccessLog → Recovery → CORS → route`。

- Recovery：panic → `R.Error`，打 stack。
- RequestID：`x-request-id`。
- 权限在路由 `middleware.go` 按需挂：`permission.Default()` / `permission.New(条件...)`。
- 业务模块不得调用 `h.Spin()`；组装只在 `main.go`（本项目 `server.New` + 显式 Recovery）。

## 8. 构造器与生命周期

- `NewXxxRepository(db)`、`InitService(ctx, db, oss, cache)`。
- 全局仅保留框架级单例（`conf`、`permission.Security`、SVC 门面）；新代码显式注入。
- 入口负责配置、infra、优雅退出；库内不泄漏裸 goroutine。

## 9. 前端契约

- API：`web/src/utils/request` + `web/src/api/*.ts`。
- 类型与 IDL 字段一致；envelope 按 `{code,msg,data}` 解包。
- 登录态：Pinia `user` store；路由权限 `store/modules/permission`；业务页 `web/src/pages/<feature>/`。
