# 编码模式（Coding Patterns）

供人和 agent 共同遵守；审查与 skill 检查清单与此对齐。

## 包与命名

- 框架包名短且表义：`httpx` / `dbx` / `mcpx` / `app`。
- 业务示例按域组织：`examples/tasks`，导出类型不带包名前缀（`tasks.Task` 而非 `tasks.TaskEntity`）。
- 构造函数统一 `NewXxx`；可选链式配置用 `WithXxx`，只在测试/少量场景使用。
- 只导出需要被跨包使用的标识符。

## 函数与参数

- `context.Context` 永远是第一个参数。
- 入参结构体优先于过长位置参数（≥3 个同层参数时）。
- 返回 `(T, error)`；错误不吞，不 panic（除编程错误）。
- 错误包装：`httpx.Internal("insert task", err)` / `httpx.Wrap(err, code, msg)`，保证 `errors.Is/As` 可回溯。

## 校验

- 校验放 Service 入口，错误码用 `httpx.Invalid`。
- 字符串字段 `TrimSpace` 后再判空；枚举用 `ParseXxx` 显式解析。
- HTTP 层用 `DecodeJSON`（`DisallowUnknownFields`）挡未知字段，但不写业务规则。

## SQL

- Repository 内 SQL 用常量字符串 + `bindPlaceholders` 生成方言占位符。
- 不拼接用户输入进 SQL；标识符引用只用于固定列名场景。
- `RowsAffected == 0` 映射为 `httpx.NotFound`（更新/删除）。

## 并发与生命周期

- 不在库内起裸 goroutine 泄漏；进程入口负责信号与 shutdown。
- `db.SetMaxOpenConns` 等在 `dbx.Open` 设保守默认值。
- 测试与进程都要 `Close` 数据库。

## 注释

- 默认不写注释；只解释非显然约束（多库方言、协议边界、幂等迁移）。
- 不写叙事型注释、不写「本函数用于…」。
