# 编码模式（Coding Patterns）

供人和 Agent 共同遵守；审查与 skill 检查清单与此对齐。范围是当前 Hertz + GORM Gen + Vue3/TDesign 项目，**不是**旧版 stdlib/httpx/dbx 样板。

## 包与命名

### 前端文件命名（强制）

1. **单文件组件 / 类**：`PascalCase.vue`（如 `STable.vue`、`TDesignSelect.vue`）
2. **目录下主文件**：`index.ts` / `index.vue`（按目录导入）
3. **接口定义文件**：`camelCase`（如 `list.ts`、`dictModel.ts`、`configModel.ts`）
4. **资源 / 样式文件**：`kebab-case`（如 `s-table.css`、`login-banner.png`）

### 通用组件 s-*（强制）

- 同一 UI **第 3 处**复用 → 抽到 `web/src/components/s-<name>/`，组件名 `S<Name>`（范例：`s-table` / `STable`）。
- 仅模块内使用 → 放该模块 `components/`，不进全局。
- `t-*` 组件 API 经 TDesign MCP `get-component-docs` 核对；细则见 skill `frontend-dev`。

示例：

```text
web/src/components/s-table/index.vue   # 组件主文件
web/src/components/s-table/sTable.ts   # 接口/类型（camelCase）
web/src/components/s-table/s-table.css
web/src/api/system/dict.ts               # API
web/src/api/system/dictModel.ts          # 类型
```

### Go 侧

- 框架包短且表义：`conf` / `pkg/errorx` / `pkg/utils/response` / `types/errno` / `types/consts`。
- 业务按域：`application/framework/...`、`application/agent/...`；生成模型包名来自 IDL `namespace`。
- 构造函数 `NewXxx`；初始化门面 `InitService`；可选配置 `WithXxx`。
- 只导出跨包需要的标识符。接口名可用 `IXxx`（与现有 repository 一致）或 `XxxRepository`，同一模块内统一。

## 函数与参数

- `context.Context` 永远是第一个参数。
- Hertz Handler 签名固定：`func(ctx context.Context, c *app.RequestContext)`。
- 返回 `(T, error)`；错误不吞、不裸 panic（编程错误除外）。
- 业务错误：`errorx.New(errno.XxxCode, errorx.KV("msg", ...))`；包装底层错误用 `errorx.WrapByCode`。
- 需要稳定给前端的错误码写进 `types/errno` 并 `code.Register`。

## 校验与绑定

- HTTP 入参用 `c.BindAndValidate(&req)`；失败 `R.InvalidParamRequestResponse`。
- Thrift/hz 生成的 tag（`json:"field,required"`）由生成器维护；业务默认值放 Service。
- Service 入口再做业务校验（长度、状态机、权限），错误码 `errno.ErrUserInvalidParamCode` 等。
- 字符串业务字段按需 `TrimSpace`；枚举显式解析。

## SQL 与持久化

- Repository 使用 gorm gen query API；复杂查询可写在 repository 方法内，但 SQL 不出现在 Handler/Service。
- 不拼接用户输入进 SQL；排序/筛选字段白名单。
- 更新/删除用主键；影响行数为 0 时返回 not found 类错误。
- 分页：`page` 从 1 起，`pageSize` 上限在 Service 约束。

## 响应与 JSON

- 成功：`R.SuccessWithData(ctx, c, data)`。
- 业务失败：`R.Error(ctx, c, err)`。
- 鉴权失败：`R.Unauthorized(c, msg)`。
- JSON 字段 camelCase；与 `web/src/api/*.ts` 接口逐字段对齐。
- 会话字段使用 `accessToken` / `tokenType` / `expiresIn` / `refreshToken`。

## 前端（web/）

### TDesign MCP（强制）

写/改任何 `t-*` 组件前，必须用 TDesign MCP（配置：仓库根 `.mcp.json`，`tdesign-mcp-server`）：

- `get-component-docs`：查 props / events / slots —— **禁止凭记忆写组件 API**
- `get-component-list`：选型
- `get-component-changelog`：升级组件库前
- `get-component-dom`：调 DOM/样式时

### 约定

- 请求封装改 `web/src/utils/request`，业务只写 `web/src/api/<feature>.ts`。
- 新类型放 `api` 旁或 `web/src/types`；不要在组件里内联重复定义 DTO。
- 登录/刷新相关逻辑集中在 `store/modules/user.ts`；页面不直接碰 token 存储。
- 文案走 `web/src/locales`；避免硬编码多处中文。
- 提交前：`npm run build`（含 `vue-tsc`）。

## 并发与生命周期

- 不在业务包起裸 goroutine 泄漏；后台任务用可控生命周期或 `taskgroup`。
- 连接池参数在 `conf` + `application/infra` 设置。
- 测试与进程都要能释放资源。

## 注释

- 默认不写注释；只解释非显然约束（生成边界、会话轮换、错误码不可变更）。
- 不写「本函数用于…」式叙事注释。
- 禁止保留 `// todo edit custom code` 等生成噪音作为长期状态；要么实现要么删掉。

## 生成代码纪律

- `api/model/**`、部分 `api/router/**` 头注释为 `DO NOT EDIT`：只改 IDL 后重新生成。
- `api/router/**/middleware.go` 允许改挂载条件，但保持生成骨架。
- 提交前确认生成物与 IDL 同步，避免手改被下次 `hz update` 冲掉。
