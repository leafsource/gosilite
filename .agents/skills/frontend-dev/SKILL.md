---
name: frontend-dev
description: Use for ANY frontend development in this repo (web/ Vue3 + TypeScript + TDesign) — new pages, features, components, bug fixes, styles, or UI changes. Loads on any mention of 前端开发, web 页面, 新增/修改 .vue, Vue3 组件, 页面/组件开发, 样式调整. Covers s-* shared components (survey web/src/components/s-* before building; extract on 3rd UI repeat and refactor all existing copies), file naming, TDesign MCP (get-component-docs before any t-*), and high-frequency task UX (列表筛选/查询, 批量操作, 表单弹窗/抽屉选择, 效果预览, 新手引导, 数据导入, 状态流转, 任务引导). Also triggers on t-* widgets, 通用组件/s-组件/s-table/pro-table, TDesign MCP 开启/连接, tdesign-mcp-server.
---

# 前端开发（frontend-dev）

中后台前端（Vue3 + TDesign）：组件分层（s-* / 业务）、文件命名、TDesign MCP 使用、高频任务交互设计。

## Important

1. **第 3 处重复 UI → 抽 `s-*`，禁止继续复制**：同一 UI 结构第 3 次出现时，必须先抽出 `s-*` 通用组件，**并把已存在的 3 处全部改为引用该 `s-*`**，之后一律复用；不得只抽新的、留下旧拷贝。
2. **开发新功能前先盘点通用组件**：先浏览 `web/src/components/s-*`（现成 `s-table` 等），有匹配组件就优先使用，禁止自建同类 UI，保持全站一致。
3. **前端开发必须开启 TDesign MCP 并优先使用**（Step 1）：写任何 `t-*` 之前先 `get-component-docs` 核对 API，禁止凭记忆。
4. 交付前 `WEB=1 ./script/check.sh`。
5. **弹出表单容器由字段数决定**：按 `t-form-item` 计数（不含底部按钮行），**字段 >5 → `t-drawer`（抽屉）**，**字段 ≤5 → `t-dialog`（对话框）**；禁止把多字段表单塞进对话框拉宽拉高。

## 何时自动加载

满足任一即用（与 AGENTS「Skills 自动使用」一致）：

- 新增/修改 `web/**/*.vue`、`components/**`
- 使用或排查 `t-*`
- 同一 UI 结构将出现第 3 次拷贝
- 设计列表筛选、批量、导入、引导、状态流转等高频任务
- 组件命名/目录归属不确定

## Instructions

### Step 1: TDesign MCP（强制开启，先于一切编码）

**前端开发必须开启 MCP 并使用，未连接成功不得动手**：以 `npx -y tdesign-mcp-server@latest` 连接。仓库根 `.mcp.json` 已内置该 server 配置；**配置改动需重启引擎或新开会话才加载**。

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

连接后内置 **4 个 tools**：

| Tool | 何时用 | 典型场景 |
|------|--------|----------|
| `get-component-docs` | 写/改任何 `t-*` **之前** | 核对 props/events/slots 后再生成代码，避免 API 用错 |
| `get-component-list` | 选型不确定 | 「该用哪个组件」、同类组件能力对比 |
| `get-component-dom` | 样式/DOM/选择器问题 | 查真实类名与 DOM 结构，修样式、迁移选择器 |
| `get-component-changelog` | 升级 `tdesign-vue-next` 前 | 破坏性变更核查、版本升级与代码迁移 |

标准顺序：**选型 `get-component-list` → 查文档 `get-component-docs` → 写码**；样式问题走 `get-component-dom`；升级前必查 `get-component-changelog`。

### Step 2: 组件分层

**动手前先盘点 `web/src/components/s-*` 目录**：有匹配的通用组件（如 `s-table`）就直接用，不要自建同类 UI；没有再按下表决定归属。

| 类型 | 位置 | 命名 | 条件 |
|------|------|------|------|
| 通用 | `web/src/components/s-<name>/` | 目录 `s-<name>`，组件 `S<Name>`（如 `STable` / `s-table`） | ≥3 处业务使用 |
| 业务 | `web/src/pages/<module>/components/` | 模块内 PascalCase 或 `index.vue` | 仅 1–2 业务依赖 |
| 原生 | 直接 `t-*` | 不包一层 | 先 MCP 查 docs |

现成通用组件：`s-table`（查询表格：搜索 / 工具栏 / 列显隐 / 分页 / 刷新 / 选择框 / `selection-change` / `#batch` 批量操作插槽）。

**抽 `s-*` 时必须回改**：第 3 处重复触发抽取后，把已存在的 3 处拷贝全部替换为新 `s-*` 的引用，再继续开发；禁止只在新代码里用、旧拷贝原样留着。

### Step 3: 文件命名（强制）

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

### Step 4: 高频任务设计

先定整体框架与页面布局，再按业务选任务模式。完整定义、流程与组件清单见 **`references/high-frequency-tasks.md`**（编码前先读）。

| 任务 | 模式选择要点 |
|------|----------------|
| **数据筛选与查询** | 范围明确、少改 → **A 查询后生效**（条件→查询→结果→清空）；反复改 → **B 立即生效**；项多 → **折叠条件**；个性化 → **条件模版** |
| **表格批量操作** | 批量为主/高频 → **A 所见即所得**（按钮常显，未选禁用）；次级 → **B 选择后触发** |
| **效果预览** | 配置多/有延迟 → **A 异步触发**（表单→预览按钮）；局部实时 → **B 同步预览** |
| **新手指引** | 必读聚焦 → **A 阻断式**（对话框）；辅助 → **B 非阻断式**（气泡/弹层）；反复看 → **主动触发入口** |
| **数据导入** | 轻量单文件 → **A 逐个导入**；多文件复杂 → **B 批量上传后导入** |
| **状态流转** | 信息复杂需核验 → **A 详情页流转**；摘要即可决策 → **B 列表内流转** |
| **任务引导** | 开始前准备 → **A 开始前**（步骤条/提示）；结果接续操作 → **B 结束后**（结果页引导） |

实现约定：

- 列表筛选优先落在 **`s-table` 搜索区**（keyword + `searchFields`，支持 `input` / `select` / `tree-select`；不要再在页面上自建筛选条）。
- 查询后生效 = 点「查询」才 `emit('search')`；立即生效 = 控件 `@change` 即触发。
- 批量：`s-table` 内置选择框（`:selectable="false"` 可关）；选中后工具栏行最左侧出现「已选 N 项」批量区，批量按钮放 `#batch` 插槽（作用域：`selected` / `selected-keys` / `clear`），选择变化事件 `selection-change`；次级批量可用底部操作条。
- 状态标签/步骤用 TDesign `t-tag` / `t-steps`，样式先查 MCP docs。

## Examples

**用户说「列表页加个状态筛选」** → 落在 `s-table` 的 `searchFields`（`select`），不自建筛选条；选「查询后生效」模式 → 查 `get-component-docs` 核对 props → 写码 → `WEB=1 ./script/check.sh` 绿。

**同一表格工具条结构第 3 次出现** → 停止复制，抽到 `web/src/components/s-<name>/`（组件 `S<Name>`），**并把已存在的 3 处全部改为引用该组件**，再继续写新代码。

## Troubleshooting

- **MCP 未开启 / 调用失败**：用 `npx -y tdesign-mcp-server@latest` 重连并确认配置已加载（改动需重启引擎或新开会话）；**MCP 必须开启使用，连接恢复前不得开始前端编码**，禁止凭记忆写 API。
- **`t-*` 属性不生效 / 报错**：先 `get-component-docs`（样式问题用 `get-component-dom`），再改代码。
- **弹窗表单过高过挤**：回到 Important 第 5 条，按 `t-form-item` 数量改用 `t-drawer`。

## 检查清单

- [ ] 开发前已盘点 `web/src/components/s-*`，有匹配即复用、未自建同类 UI
- [ ] 第 3 处重复已抽 `s-*`，且原有 3 处已全部改为引用（或 spec 写明暂不抽）
- [ ] 通用在 `components/s-*`，业务在模块 `components/`
- [ ] TDesign MCP 已开启；`t-*` 已查 `get-component-docs`
- [ ] 高频任务模式与 `references/high-frequency-tasks.md` 一致
- [ ] 弹出表单：`t-form-item` >5 用 `t-drawer`，≤5 用 `t-dialog`
- [ ] 命名四条；`WEB=1 ./script/check.sh` 绿
