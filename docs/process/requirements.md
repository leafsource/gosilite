# 需求开发流程（Requirements Loop）

与 **Compose Next** 对齐：一个 feature 一份文档，需求 → 设计 → 编码 → 验收 → 归档可被 Agent 完整驱动。

## 闭环总览

```text
需求澄清 (Grill / Question)
    ↓
docs/compose/spec/<feature>.md   ← Problem / Design / Out of Scope / Tasks
    ↓
设计契约：IDL DTO · 错误码 · 路由 · 表 · 权限 · 前端类型
    ↓
实现顺序：idl → hz generate → repo → service → handler/router → web api
    ↓
测试：service → http → 会话/权限 → 前端 build
    ↓
验收：go test/vet/build && cd web && npm run build
    ↓
评审 → spec status: delivered
```

## 1. 需求澄清

一次只收敛一个决策轴；可枚举选项用选择题。必须确认：

- 业务对象与读写用例
- 对外 HTTP 路径与 JSON 契约（camelCase）
- 是否改会话/权限（若是，必须兼容 `SessionTokenResponse`）
- 错误码用例（invalid / not_found / session / permission…）
- 超出范围项（写进 Out of Scope）

## 2. Spec 文档

路径：`docs/compose/spec/<feature>.md`（不要加日期到文件名）。

```markdown
---
feature: <feature-name>
status: designed | in-progress | delivered
updated: YYYY-MM-DD
branch: <branch-name>
commits: <short-base-sha>..<short-head-sha>
---

# <Feature Name>

## Report

## [S1] Problem
## [S2] Design
## [S3] Out of Scope

## Tasks
- [ ] T1: <工作> — acceptance: <可观察结果> (covers: S2)
```

- `[Sn]` 锚点稳定，不重编号。
- 任务是最小可验证工作项；`covers:` 必填；依赖用 `depends:` 且无环。
- 设计段只写已敲定契约，不写探索过程。

## 3. 契约先行

写代码前定下：

1. Domain 字段与状态机
2. IDL struct / API 路由（`api.get` / `api.post`）
3. JSON 字段名（camelCase）与错误码
4. 表结构 / gorm 模型字段
5. 前端 `web/src/api` 类型与 store 影响面

## 4. 实现顺序

1. IDL + `hz update`（生成 model/router）
2. `types/errno` 错误码
3. dal model/query + repository + 测试
4. service + 测试（登录/会话类必测轮换）
5. handler + router middleware 权限
6. `web/src/api` + store/页面
7. 冒烟：`curl` 登录/业务接口

每完成一层就跑该层测试，不攒到最后。

## 5. 验收与评审

- 验收四连必须全绿（见 `docs/patterns/testing.md`）。
- 评审关注：契约是否兑现、分层是否被破坏、会话字段是否 OAuth2 风格、测试是否测行为。
- 通过后把 spec 标记 `delivered`，勾选 Tasks，填写 Report。

## 6. 变更

已交付 feature 扩展时：

- 在原 spec 追加 Amendment 或新开 feature 文档。
- 不重编号已有 `[Sn]` 锚点。
- 先补失败测试再改实现。
- 若破坏 JSON 契约，必须同时改 IDL、前端类型、本技能契约表。

## 7. Compose Next 操作映射

| Compose 阶段 | 本项目落点 |
|--------------|------------|
| Grill | Question 收敛契约/范围 |
| Workspace | 功能分支或 `.worktrees/<slug>`（默认不要在 main 直接改） |
| Spec | `docs/compose/spec/<feature>.md` |
| Implement | 上文「实现顺序」 |
| Verify | 验收四连 + 登录冒烟 |
| Review | 独立 subagent 按 Tasks 验收 |
| Finalize | spec `delivered` + Report |

## 8. Agent 执行注意

- 编码前读 `AGENTS.md`、`.agents/skills/gosilite-dev/SKILL.md`、相关 IDL/Service。
- 不虚构 Hertz/生成器 API；不确定就查 `docs/hertz/api-index.md` 或官方文档。
- 不要恢复 Gin、`cmd/gosilite`、`token`/`expire` 旧会话字段、MinIO。
- 工作区已有大量重建未提交改动时，禁止 `git reset/clean/checkout` 破坏现场。
