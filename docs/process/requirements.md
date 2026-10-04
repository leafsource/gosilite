# 需求开发流程（Requirements Loop）

与 Compose Next 对齐：一个 feature 一份文档，需求 → 设计 → 编码 → 验收 → 归档可被 agent 完整驱动。

## 闭环总览

```text
需求澄清 (Grill)
    ↓
docs/compose/spec/<feature>.md   ← Problem / Design / Tasks
    ↓
设计契约：DTO · 错误码 · 路由 · 表 · 页面字段 · 测试边界
    ↓
实现顺序：schema 版本 → repo → service → http / mcp → web
    ↓
测试：service → repo → http → mcp → 双通道 → web build
    ↓
验收：go test/vet/build + cd web && npm run build
    ↓
评审 → spec status: delivered
```

## 1. 需求澄清

一次只收敛一个决策轴；可枚举选项用选择题，避免空泛讨论。必须确认：

- 业务对象与读写用例
- 对外 HTTP 路径与 JSON 契约
- 是否需要 MCP 业务工具（默认：需要，与 HTTP 同一 Service）
- 是否需要 `web/` 页面（Vue3 + TDesign）及关键交互
- 错误码用例（invalid / not_found / conflict…）
- 超出范围项（写进 Out of Scope）

## 2. Spec 文档

路径：`docs/compose/spec/<feature>.md`

必备章节：`[S1] Problem`、`[S2] Design`、`[S3] Out of Scope`、`Tasks`。

任务写成最小可验证工作项，格式：

```markdown
- [ ] T1: <工作> — acceptance: <可观察结果> (covers: S2.x)
```

设计段落只记录已敲定契约，不写探索过程。

## 3. 契约先行

写代码前定下：

1. Domain 字段与状态机
2. DTO JSON 字段
3. 错误码与 HTTP 映射
4. 表结构与 schema 版本
5. MCP tool 名称与 inputSchema
6. 前端页面字段与 `web/src/types` 类型

## 4. 实现顺序

1. schema 版本文件（建表/索引）
2. Repository + 单测
3. Service + 单测（固定时钟）
4. HTTP handler + httptest
5. MCP tools + 报文测试
6. 组装（`examples/<feature>/app.go`）与入口
7. 前端 `web/`：types → api → TDesign 页面

每完成一层就跑该层测试，不攒到最后。

## 5. 验收与评审

- 后端三连与 `web` 的 `npm run build` 必须全绿。
- 评审关注：契约是否兑现、分层是否被破坏、测试是否测行为、前端 DTO 是否对齐。
- 通过后把 spec 标记 `delivered`，勾选 Tasks，填写 Report。

## 6. 变更

已交付 feature 需要扩展时：

- 在原 spec 追加 Amendment 或新开 feature 文档。
- 不重编号已有 `[Sn]` 锚点。
- 先补失败测试再改实现。
