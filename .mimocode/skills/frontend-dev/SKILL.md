---
name: frontend-dev
description: Use when developing Vue3/TDesign mid-backend frontend — shared s-* components, file naming, TDesign MCP docs, and high-frequency task UX (filter query, batch ops, form dialog vs drawer, preview, onboarding, import, status flow, task guidance). Triggers on writing/editing web/ .vue, t-* widgets, 通用组件, s-组件, 列表筛选/查询, 批量操作, 表单弹窗/抽屉选择, 新手引导, 数据导入, 状态流转, pro-table/s-table, TDesign MCP.
---

# 前端开发（frontend-dev）

中后台前端：组件分层（s-* / 业务）、命名、TDesign MCP、高频任务交互设计。

## Important

1. **第 3 处重复 UI → 抽 `s-*`**，禁止继续复制。
2. **`t-*` API 必须先 TDesign MCP `get-component-docs`**，禁止凭记忆。
3. 交付前 `WEB=1 ./script/check.sh`。
4. **弹出表单容器由字段数决定**：新增/编辑表单弹出时按 `t-form-item` 计数（不含底部按钮行），**字段 >5 → 用 `t-drawer`（抽屉）**，字段 ≤5 → 用 `t-dialog`（对话框）；禁止把多字段表单塞进对话框拉宽拉高。

## 何时自动加载

满足任一即用（与 AGENTS「Skills 自动使用」一致）：

- 新增/修改 `web/**/*.vue`、`components/**`
- 使用或排查 `t-*`
- 同一 UI 结构将出现第 3 次拷贝
- 设计列表筛选、批量、导入、引导、状态流转等高频任务
- 组件命名/目录归属不确定

## Step 1 组件分层

| 类型 | 位置 | 命名 | 条件 |
|------|------|------|------|
| 通用 | `web/src/components/s-<name>/` | 目录 `s-<name>`，组件 `S<Name>`（如 `STable` / `s-table`） | ≥3 处业务使用 |
| 业务 | `web/src/pages/<module>/components/` | 模块内 PascalCase 或 `index.vue` | 仅 1–2 业务依赖 |
| 原生 | 直接 `t-*` | 不包一层 | 先 MCP 查 docs |

现成通用组件：`s-table`（查询表格：搜索 / 工具栏 / 列显隐 / 分页 / 刷新 / 选择框 / `selection-change` / `#batch` 批量操作插槽）。

## Step 2 文件命名（强制）

1. 单文件组件/类：`PascalCase.vue`（`STable.vue`）
2. 目录主文件：`index.ts` / `index.vue`
3. 接口定义：`camelCase`（`sTable.ts`、`dictModel.ts`）
4. 资源/样式：`kebab-case`（`s-table.css`）

示例：

```text
web/src/components/s-table/index.vue
web/src/components/s-table/sTable.ts
web/src/components/s-table/s-table.css
web/src/api/system/dict.ts
web/src/api/system/dictModel.ts
```

## Step 3 TDesign MCP（强制）

配置：仓库根 `.mcp.json`（`npx -y tdesign-mcp-server@latest`）。

```json
{
  "mcpServers": {
    "tdesign-mcp-server": {
      "command": "npx",
      "args": ["-y", "tdesign-mcp-server@latest"]
    }
  }
}
```

| Tool | 何时用 |
|------|--------|
| `get-component-docs` | 写/改任何 `t-*` **之前** |
| `get-component-list` | 选型不确定 |
| `get-component-dom` | 样式/DOM/选择器问题 |
| `get-component-changelog` | 升级 `tdesign-vue-next` 前 |

MCP 不可用时：显式标注「未查 MCP」，优先读 `node_modules/tdesign-vue-next` 类型定义，禁止凭记忆写 API。

## Step 4 高频任务设计

先定整体框架与页面布局，再按业务选任务模式。完整定义、流程与组件清单见 **`references/high-frequency-tasks.md`**。

| 任务 | 模式选择要点 |
|------|----------------|
| **数据筛选与查询** | 范围明确、少改 → **A 查询后生效**（条件→查询→结果→清空）；反复改 → **B 立即生效**；项多 → **折叠条件**；个性化 → **条件模版** |
| **表格批量操作** | 批量为主/高频 → **A 所见即所得**（按钮常显，未选禁用）；次级 → **B 选择后触发** |
| **效果预览** | 配置多/有延迟 → **A 异步触发**（表单→预览按钮）；局部实时 → **B 同步预览** |
| **新手指引** | 必读聚焦 → **A 阻断式**（对话框）；辅助 → **B 非阻断式**（气泡/弹层）；反复看 → **主动触发入口** |
| **数据导入** | 轻量单文件 → **A 逐个导入**；多文件复杂 → **B 批量上传后导入** |
| **状态流转** | 信息复杂需核验 → **A 详情页流转**；摘要即可决策 → **B 列表内流转** |
| **任务引导** | 开始前准备 → **A 开始前**（步骤条/提示）；结果接续操作 → **B 结束后**（结果页引导） |

实现时：

- 列表筛选优先落在 **`s-table` 搜索区**（keyword + `searchFields`，支持 `input` / `select` / `tree-select`；不要再在页面上自建筛选条）。
- 查询后生效 = 点「查询」才 `emit('search')`；立即生效 = 控件 `@change` 即触发。
- 批量：`s-table` 内置选择框（`:selectable="false"` 可关）；选中后工具栏行最左侧出现「已选 N 项」批量区，批量按钮放 `#batch` 插槽（作用域：`selected` / `selected-keys` / `clear`），选择变化事件 `selection-change`；次级批量可用底部操作条。
- 状态标签/步骤用 TDesign `t-tag` / `t-steps`，样式先查 MCP docs。

## 检查清单

- [ ] 第 3 处重复已抽 `s-*`（或 spec 写明暂不抽）
- [ ] 通用在 `components/s-*`，业务在模块 `components/`
- [ ] `t-*` 已经 MCP `get-component-docs`
- [ ] 高频任务模式与 `references/high-frequency-tasks.md` 一致
- [ ] 弹出表单：`t-form-item` >5 用 `t-drawer`，≤5 用 `t-dialog`
- [ ] 命名四条；`WEB=1 ./script/check.sh` 绿
