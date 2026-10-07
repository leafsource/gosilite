# 编码规范

审查与检查清单与此对齐。范围是当前 Hertz + GORM Gen + Vue3/TDesign，不是旧 stdlib/httpx/dbx 样板。

## 包与命名（Go）

- 框架包短且表义：`conf` / `pkg/errorx` / `pkg/utils/response` / `types/errno` / `types/consts`。
- 业务按域：`application/framework/...`、`application/agent/...`；生成模型包名来自 IDL `namespace`。
- 构造函数 `NewXxx`；初始化门面 `InitService`；可选配置 `WithXxx`。
- 只导出跨包需要的标识符。Repository 接口名 `IXxx` 或 `XxxRepository`，同一模块内统一。
- **Service 是具体类型**（`type Service struct` + `NewService`），禁止 `I*Service`。

## 函数与参数

- `context.Context` 永远是第一个参数。
- Hertz Handler：`func(ctx context.Context, c *app.RequestContext)`。
- 返回 `(T, error)`；错误不吞、不裸 panic（编程错误除外）。
- 业务错误：`errorx.New(errno.XxxCode, errorx.KV("msg", ...))`；包装底层用 `errorx.WrapByCode`。
- 需要稳定给前端的错误码写进 `types/errno` 并 `code.Register`。

## 校验与绑定

- HTTP 入参 `c.BindAndValidate(&req)`；失败 `R.InvalidParamRequestResponse`。
- 生成 tag（`json:"field,required"`）由生成器维护；业务默认值放 Service。
- Service 入口再做业务校验（长度、状态机、权限）。
- 字符串业务字段按需 `TrimSpace`；枚举显式解析。

## SQL 与持久化

- Repository 用 gorm gen query；复杂查询写在 repository 方法内；SQL 不出现在 Handler/Service。
- 不拼接用户输入进 SQL；排序/筛选字段白名单。
- 更新/删除用主键；影响行数 0 返回 not found 类错误。
- 分页：`page` 从 1 起，`pageSize` 上限在 Service 约束。

## 响应与 JSON

- 成功：`R.SuccessWithData(ctx, c, data)`。
- 业务失败：`R.Error(ctx, c, err)`。
- 鉴权失败：`R.Unauthorized(c, msg)`。
- JSON 字段 camelCase，与 `web/src/api/*.ts` 逐字段对齐。
- 会话字段：`accessToken` / `tokenType` / `expiresIn` / `refreshToken`。

## 模型转换 convert.go

所有 **dal model ↔ DTO/VO** 字段映射集中在对应区域包的 `convert.go`，禁止内联散落在 controller/service 正文：

- Handler 包：`api/handler/<domain>/convert.go` — model → 响应 DTO、请求 DTO → model。
- Service 包（确有转换时）：`application/<domain>/service/<feature>/convert.go`。
- 粒度按**包**一个文件；包内 feature 多时可 `convert_<feature>.go`，仍归该包统一管理。
- 命名：`toXxxDTO` / `fromXxxRequest` / `toXxxDTOs`（列表映射，避免 Handler 手写循环 append）。
- 转换函数只做字段赋值：不含业务逻辑、不查库、不返回 error；派生计算放 Service。

## 前端（web/）编码约定

命名（强制）：

1. 单文件组件/类：`PascalCase.vue`（`STable.vue`）
2. 目录主文件：`index.ts` / `index.vue`
3. 接口定义：`camelCase`（`sTable.ts`、`dictModel.ts`）
4. 资源/样式：`kebab-case`（`s-table.css`）

组件与请求：

- 通用组件 ≥3 处 → `web/src/components/s-<name>/`（`S<Name>`）；仅模块内 → 该模块 `components/`。
- 请求封装改 `web/src/utils/request`，业务只写 `web/src/api/<feature>.ts`。
- 新类型放 `api` 旁或 `web/src/types`；不要在组件内联重复 DTO。
- 登录/刷新集中在 `store/modules/user.ts`；页面不直接碰 token 存储。
- 文案走 `web/src/locales`；避免硬编码多处中文。
- 提交前 `npm run build`（含 `vue-tsc`）。
- `t-*` API 必须经 TDesign MCP `get-component-docs` 核对（见 frontend-dev skill）。

## 并发与生命周期

- 不在业务包起裸 goroutine 泄漏；后台任务可控生命周期或 `taskgroup`。
- 连接池参数在 `conf` + `application/infra`。
- 测试与进程都要能释放资源。

## 注释

- 默认不写注释；只解释非显然约束（生成边界、会话轮换、错误码不可变更）。
- 不写「本函数用于…」式叙事注释。
- 禁止长期保留 `// todo edit custom code` 等生成噪音。

## 生成代码纪律

- `api/model/**`、部分 `api/router/**` 头注释 `DO NOT EDIT`：只改 IDL 后重新生成。
- `api/router/**/middleware.go` 允许改权限挂载条件，保持生成骨架。
- 提交前确认生成物与 IDL 同步，避免手改被下次 `hz update` 冲掉。
