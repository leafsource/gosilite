# 需求闭环（Requirements Loop）

与 Compose Next 对齐：一个 feature 一份文档，需求 → 设计 → 编码 → 验收 → 归档可被 Agent 完整驱动。

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
验收：./script/check.sh （WEB=1 可带前端）
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

### 1.1 核心原则

1. **先推断，后询问**：能从上下文/代码库/惯例推出的，绝不问用户。
2. **能假设就假设，但必须标注**：有合理默认值的先假设，明确告诉用户「我按 X 处理」。
3. **只问卡住的**：只有「无法推断 + 影响执行」的才问，且带选项。
4. **给出方案而非只给结论**：必要时给 A/B/C 选项 + 代价。
5. **主动指出问题**：需求不合理、有风险、有更简单方案，要主动说。

### 1.2 需求解析（内部完成，不输出）

对每个评估维度做三分类：

- 【可推断】能从上下文推出 → 自己补全
- 【可默认】有合理默认值 → 假设 + 标注
- 【必须问】无法推断 + 影响执行 → 加入询问列表

评估维度：

1. 功能目标、输入输出、业务规则、验收标准、优先级
2. 技术选型、实现难度、依赖、兼容性
3. 代码规范、测试、性能、安全、错误处理、可维护性
4. 改动范围、影响面、交付形式、验证方式、回滚方案

### 1.3 信息推断（项目落点）

主动读取：项目文件结构、现有代码与配置、`go.mod` / `web/package.json`、现有模块命名与模式、`AGENTS.md` 与 README、需求措辞和上下文。

推断优先级：**项目现有惯例 > 行业最佳实践 > 通用默认值**。本项目惯例即：Hertz 分层、`{code,msg,data}` 信封、camelCase、`SessionTokenResponse`、`types/errno`（见 SKILL.md「项目事实」表）。

### 1.4 假设标注

对「可默认」的项，输出时明确标注：

- 格式：「假设：……（如有偏差请纠正）」
- 例：「假设：按项目惯例用 IDL 先行 + 错误码注册到 `types/errno`，返回 `{code,msg,data}`」

### 1.5 关键提问

只问同时满足以下条件的：

- 无法推断
- 不同选择导致完全不同的实现
- 不问就会做错

提问要求：

- 数量 ≤ 3，按影响大小排序
- 带选项（A/B/C 或 是/否）
- 说明为什么问（影响什么）

### 1.6 澄清输出模板

```text
【我的理解】
（1–2 句复述需求）

【我的假设】（如有偏差请纠正）
- 技术栈 / 契约：……
- 输入输出：……
- 其他：……

【需你确认】（关键问题，带选项，≤3 个）
1. ……？A. …… B. ……
2. ……？是 / 否

【方案】
- 推荐：……（理由）
- 备选：……（代价：……）
- 风险 / 影响面 / 前置条件：……
```

示例——用户说「帮我写个用户登录接口」：

> 【我的理解】在现有项目新增用户登录接口。
> 【我的假设】读了项目，复用现有 User 模型与 JWT 机制；输入 username + password；返回 `SessionTokenResponse`（`accessToken`/`tokenType`/`expiresIn`/`refreshToken`）；密码用 `pkg/bcrypt` 校验；自动含输入校验、错误码注册、单测。
> 【需你确认】1. 登录失败需要限流吗？（影响是否用 Redis 计数，A. 先不做 B. 做）2. 需要「记住我」延长 refresh 吗？（影响 expiresIn 策略，A. 否 B. 是）
> 【方案】推荐按现有会话契约走 MVP，复用现有机制；风险：限流若后补需改 service 分支。

### 1.7 行为准则

应该做：先读代码/上下文再下判断；能推断的不问；假设要标注；问题要少、要关键、带选项；主动指出不合理处并给替代方案；说清风险、代价、不确定性。

禁止做：把需求确认清单逐条甩给用户；能推断还去问；开放式提问（「你想要什么？」）；不标注假设就默认执行；闷头写代码不做需求评估；对不合理需求直接拒绝而不给替代方案；过度设计（需求只要 A，别做 A+B+C）。

### 1.8 特殊场景

- **需求太模糊**：不逐条问。先按最可能的理解给方案 + 标注假设，再问 1–2 个最关键分叉点。
- **需求不合理**：先说明为什么不合理（技术/逻辑/资源），再给替代方案让用户选。
- **需求有冲突**：指出冲突点，给出取舍建议，请用户拍板。
- **信息不足但可假设**：按最合理假设给完整方案，标注「如果 X 不是这样，方案会变成 Y」，让用户低成本纠偏。

## 2. Spec 文档

路径：`docs/compose/spec/<feature>.md`（文件名不加日期）。

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

- 验收必须全绿（见 `references/testing.md`）。
- 评审关注：契约是否兑现、分层是否被破坏、会话字段是否 OAuth2 风格、测试是否测行为。
- 通过后把 spec 标记 `delivered`，勾选 Tasks，填写 Report。

## 6. 变更

已交付 feature 扩展时：

- 在原 spec 追加 Amendment 或新开 feature 文档。
- 不重编号已有 `[Sn]` 锚点。
- 先补失败测试再改实现。
- 若破坏 JSON 契约，必须同时改 IDL、前端类型、skill 契约表。

## 7. Compose Next 操作映射

| Compose 阶段 | 本项目落点 |
|--------------|------------|
| Grill | Question 收敛契约/范围 |
| Workspace | 功能分支或 `.worktrees/<slug>`（默认不要在 main 直接改） |
| Spec | `docs/compose/spec/<feature>.md` |
| Implement | 上文「实现顺序」 |
| Verify | 验收命令 + 登录冒烟 |
| Review | 独立 subagent 按 Tasks 验收 |
| Finalize | spec `delivered` + Report |

## 8. Agent 执行注意

- 编码前读 `AGENTS.md`、本 skill、相关 IDL/Service。
- 不虚构 Hertz/生成器 API；不确定就查 `references/hertz-api.md` 或官方文档。
- 不要恢复 Gin、`cmd/gosilite`、`token`/`expire` 旧会话字段、MinIO。
- 工作区已有大量重建未提交改动时，禁止 `git reset/clean/checkout` 破坏现场。
