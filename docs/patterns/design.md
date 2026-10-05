# 设计模式（Design Patterns）

gosilite 面向 Agent 高频生成与修改 Web 功能，模式刻意收敛到可重复、可校验的一组。**以本仓库实际代码为准**，不要套用其他框架的记忆。

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
- JSON 字段名使用 **camelCase**（`accessToken`、`expiresIn`），与前端 `web/src/api` 一致。
- 生成文件禁止手改；调整字段先改 IDL 再生成。
- 手写扩展放在 handler / service / repository，不塞进生成目录。

## 1.1 Module 插件

- 契约：`application.Module`（`Name` / `Init` / `RegisterRoutes`）。
- 装配：`Bootstrap` 初始化 infra 并 Init 全部模块；`MountRoutes` 按序挂路由。
- 模块自己 `InitService` 建具体 Service 并安装包级门面；Handler 不依赖 `I*Service`。
- 新域 = 新 Module + 注册表一行；路由仍来自 IDL/hz 生成的 `Register`。

## 3. 统一错误与响应模型

成功：

```json
{ "code": 200, "msg": "success", "data": { ... } }
```

业务失败（HTTP 仍为 200，除非鉴权/参数）：

```json
{ "code": 700000003, "msg": "invalid username or password..." }
```

- 全系统业务错误只用 `errorx.New` / `errorx.WrapByCode`，错误码在 `types/errno` 注册。
- 词表按域划分（用户 700 000 000 段），消息模板支持 `{param}`。
- 401/400 仅用于：未认证、非法参数；其余业务语义用 `code` 表达，便于前端统一处理。
- Handler 只使用 `pkg/utils/response` 的 `SuccessWithData` / `Error` / `Unauthorized` / `InvalidParamRequestResponse`。

## 4. 会话（OAuth2 风格）

- 登录 / 刷新返回 `SessionTokenResponse`：`accessToken`、`tokenType`、`expiresIn`、`refreshToken`。
- **禁止**再使用 `token` + `expire` 旧字段对。
- Access token = Hertz JWT（`Authorization: Bearer`）；Refresh token = Redis，支持 rotation 与按用户 generation 吊销。
- 登出：吊销 refresh + access denylist（TTL ≤ 24h）。
- 改密成功后必须吊销该用户全部 refresh token。
- 前端 `useUserStore` 对并发 401 做单飞 refresh，避免轮换竞态。

## 5. Repository + GORM Gen

- 模型与 query 在 `application/<domain>/dal/{model,query}` 由 gen 生成；repository 手写。
- Repository 接口由 service 包依赖，便于测试替身。
- `RowsAffected == 0` 或 `gorm.ErrRecordNotFound` 映射为 `errorx` 业务错误（如 `ErrUserResourceNotFound`）。
- 列表强制 `page`/`pageSize`，禁止无边界全量查询。
- DDL/模型字段结构化；禁止用 JSON blob 顶替业务列。

## 6. 中间件链（入口固定）

顺序：`ContextCache → pprof(可选) → gzip(可选) → RequestID → AccessLog → Recovery → CORS → route`。

- Recovery：panic → `R.Error`，打 stack。
- RequestID：`x-request-id`。
- 权限中间件在路由 `middleware.go` 按需挂载：`permission.Default()` / `permission.New(条件...)`。
- 业务模块不得调用 `h.Spin()`；组装只在 `main.go`。

## 7. 构造器与生命周期

- `NewXxxRepository(db)`、`InitService(ctx, db, oss, cache)`。
- 全局仅保留框架级单例（`conf`、`permission.Security`、SVC 门面）；新代码优先显式注入。
- 进程入口负责配置加载、infra 初始化、优雅退出；库内不泄漏裸 goroutine。

## 8. 前端契约

- API 封装：`web/src/utils/request`（Axios）+ `web/src/api/*.ts`。
- 类型与后端 IDL 字段一致；envelope 解包按 `{code,msg,data}`。
- 登录态：Pinia `user` store，persist `token`/`refreshToken`。
- 页面路由权限由 `store/modules/permission` 控制；业务页放 `web/src/pages/<feature>/`。
