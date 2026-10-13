---
name: frontend-dev
description: Use for ANY frontend development in this repo (web/ Vue3 + TypeScript + TDesign) — new pages, features, components, bug fixes, styles, or UI changes. Loads on any mention of 前端开发, web 页面, 新增/修改 .vue, Vue3 组件, 页面/组件开发, 样式调整. Covers s-* shared components (survey web/src/components/s-* before building; extract on 3rd UI repeat and refactor all existing copies), file naming, TDesign MCP via scripts/tdesign_mcp.py (manual stdio connect; get-component-docs before any t-*; framework vue-next; get-component-list/dom/changelog; search-icon), and high-frequency task UX (列表筛选/查询, 批量操作, 表单弹窗/抽屉选择, 效果预览, 新手引导, 数据导入, 状态流转, 任务引导). Also triggers on t-* widgets, 通用组件/s-组件/s-table/pro-table, TDesign MCP 手动连接/调用示例/tdesign_mcp.py, tdesign-mcp-server, search-icon.
---

# 前端开发（frontend-dev）

中后台前端（Vue3 + TDesign）：组件分层（s-* / 业务）、文件命名、TDesign MCP、高频任务交互、验收闭环。

## Important

1. **第 3 处重复 UI → 抽 `s-*`，禁止继续复制**：同一 UI 结构第 3 次出现时，必须抽出 `s-*` 通用组件，**并把已存在的 3 处全部改为引用该 `s-*`**，之后一律复用；不得只抽新的、留下旧拷贝。
2. **开发新功能前先盘点通用组件**：先浏览 `web/src/components/s-*`（现成 `s-table` 等），有匹配就优先使用，禁止自建同类 UI。
3. **前端开发必须通过 TDesign MCP 查文档**（Step 1）：用 `scripts/tdesign_mcp.py` 调用；写任何 `t-*` 之前先 `get-component-docs`（`framework: "vue-next"`）核对 API，禁止凭记忆。细则见 `references/tdesign-mcp.md`。
4. 交付前 `WEB=1 ./script/check.sh`。
5. **弹出表单容器由字段数决定**：按 `t-form-item` 计数（不含底部按钮行），**字段 >5 → `t-drawer`**，**字段 ≤5 → `t-dialog`**；禁止把多字段表单塞进对话框拉宽拉高。
6. **契约与后端对齐**：`web/src/api/*.ts` 字段名 camelCase，envelope `{code,msg,data}`；会话字段 `accessToken`/`tokenType`/`expiresIn`/`refreshToken`，禁止 `token`/`expire`。

## 何时自动加载

满足任一即用（与 AGENTS「Skills 自动使用」一致）：

- 新增/修改 `web/**/*.vue`、`components/**`
- 使用或排查 `t-*`
- 同一 UI 结构将出现第 3 次拷贝
- 设计列表筛选、批量、导入、引导、状态流转等高频任务
- 组件命名/目录归属不确定
- 前端类型/envelope/登录态与后端联调

改 Go API 时同时加载 `.agents/skills/gosilite-dev/SKILL.md`。

## Instructions

### Step 1: TDesign MCP（强制，先于一切 `t-*` 编码）

完整契约、场景与排障见 **`references/tdesign-mcp.md`**。本环境**引擎 MCP 挂载不可用**（重启/新会话无效），必须走 **脚本手动连接**。

**未查到 API 不得写 `t-*` 业务代码。** 连接与调用：

```bash
MCP=.agents/skills/frontend-dev/scripts/tdesign_mcp.py
python3 "$MCP" list-tools
python3 "$MCP" call get-component-docs '{"framework":"vue-next","names":["table","select","form"]}'
```

| Tool | 参数（必填加粗） | 何时用 |
|------|------------------|--------|
| `get-component-list` | **`framework`** | 选型、「有没有现成组件」 |
| `get-component-docs` | **`names[]` + `framework`** | 写/改任何 `t-*` **之前**（强制） |
| `get-component-dom` | **`names[]` + `framework`** | 自定义 CSS / 类名 / DOM |
| `get-component-changelog` | **`names[]` + `framework`**，可选 `version` | 升级 `tdesign-vue-next` 前 |
| `search-icon` | **`keywords[]`**，可选 `limit`（默认 20） | 图标完整名（**单数 icon**） |

- 本仓库固定 `"framework": "vue-next"`；`names` 可批量。
- 脚本自拉 stdio server（newline JSON-RPC），用完退出。**不要 `sudo npx` 长期挂在终端**（管道接不上）。
- 若会话工具列表恰好有原生 `get-component-*`，可直接调原生工具，参数同上；没有就一律跑脚本。
- 标准顺序：**选型 `get-component-list` → `get-component-docs` → 写码**；样式 `get-component-dom`；图标 `search-icon`；升级前 changelog。

### Step 2: 组件分层

**动手前先盘点 `web/src/components/s-*`**，有匹配就直接用。

| 类型 | 位置 | 命名 | 条件 |
|------|------|------|------|
| 通用 | `web/src/components/s-<name>/` | 目录 `s-<name>`，组件 `S<Name>` | ≥3 处业务使用 |
| 业务 | `web/src/pages/<module>/components/` | PascalCase 或 `index.vue` | 仅 1–2 业务依赖 |
| 原生 | 直接 `t-*` | 不包一层 | 先 MCP 查 docs |

现成通用组件：`s-table`（查询表格：搜索 / 工具栏 / 列显隐 / 分页 / 刷新 / 选择框 / `selection-change` / `#batch` 批量插槽）。

**抽 `s-*` 时必须回改**：第 3 处触发后，把已有 3 处拷贝全部替换为引用，再继续开发。

### Step 3: 文件命名（强制）

1. 单文件组件/类：`PascalCase.vue`（`STable.vue`）
2. 目录主文件：`index.ts` / `index.vue`
3. 接口定义：`camelCase`（`sTable.ts`、`dictModel.ts`）
4. 资源/样式：`kebab-case`（`s-table.css`）

```text
web/src/components/s-table/index.vue
web/src/components/s-table/sTable.ts
web/src/components/s-table/s-table.css
web/src/api/system/dict.ts
web/src/api/system/dictModel.ts
```

### Step 4: 请求、类型与登录态

- 请求封装改 `web/src/utils/request`，业务只写 `web/src/api/<feature>.ts`。
- 类型放 `api` 旁（`*Model.ts`）或 `web/src/types`；禁止组件内联重复 DTO。
- 登录/刷新集中在 `store/modules/user.ts`（并发 401 单飞 refresh）；页面不直接碰 token 存储。
- 路由权限由 `store/modules/permission` 控制；业务页 `web/src/pages/<feature>/`。
- 文案走 `web/src/locales`；避免硬编码多处中文。

### Step 5: 高频任务设计

先定整体框架与页面布局，再按业务选任务模式。完整定义见 **`references/high-frequency-tasks.md`**（编码前先读）。

| 任务 | 模式选择要点 |
|------|----------------|
| **数据筛选与查询** | 范围明确、少改 → **A 查询后生效**；反复改 → **B 立即生效**；项多 → **折叠条件**；个性化 → **条件模版** |
| **表格批量操作** | 批量为主/高频 → **A 所见即所得**；次级 → **B 选择后触发** |
| **效果预览** | 配置多/有延迟 → **A 异步触发**；局部实时 → **B 同步预览** |
| **新手指引** | 必读聚焦 → **A 阻断式**；辅助 → **B 非阻断式**；反复看 → **主动触发入口** |
| **数据导入** | 轻量单文件 → **A 逐个导入**；多文件复杂 → **B 批量上传后导入** |
| **状态流转** | 信息复杂需核验 → **A 详情页流转**；摘要即可决策 → **B 列表内流转** |
| **任务引导** | 开始前准备 → **A 开始前**；结果接续操作 → **B 结束后** |

实现约定：

- 列表筛选优先落在 **`s-table` 搜索区**（keyword + `searchFields`，支持 `input` / `select` / `tree-select`）；不要自建筛选条。
- 查询后生效 = 点「查询」才 `emit('search')`；立即生效 = 控件 `@change` 即触发。
- 批量：`s-table` 内置选择框（`:selectable="false"` 可关）；「已选 N 项」+ `#batch` 插槽（`selected` / `selected-keys` / `clear`）+ `selection-change`。
- 状态标签/步骤用 `t-tag` / `t-steps`，样式先查 MCP docs。

### Step 6: 验收

```bash
WEB=1 ./script/check.sh   # 含 go 侧；或 cd web && npm run build（vue-tsc + vite）
```

登录链路冒烟：错误密码 → 统一 Toast；正确登录 → 布局；401 自动 refresh 或跳登录。

## Examples

**用户说「列表页加个状态筛选」** → 落在 `s-table` 的 `searchFields`（`select`），不自建筛选条；选「查询后生效」→ `python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-docs '{"framework":"vue-next","names":["select"]}'` 核对 props → 写码 → `WEB=1 ./script/check.sh` 绿。更多示例见 `references/tdesign-mcp.md` §4。

**同一表格工具条结构第 3 次出现** → 停止复制，抽 `s-<name>`，**已有 3 处全部改为引用**，再继续写新代码。

**表单弹窗字段变多** → 数 `t-form-item`（不含按钮行）；>5 迁 `t-drawer`，≤5 留 `t-dialog`。

## Troubleshooting

- **引擎/会话里没有 MCP 工具**：本环境**不要靠重启/新会话**；用 `scripts/tdesign_mcp.py` 手动连接。**禁止 `sudo npx` 挂空进程**。详见 `references/tdesign-mcp.md` §6。
- **脚本 timeout / 连不上**：优先已安装包入口（脚本自动找）；或设 `TDESIGN_MCP_CMD="node <...>/dist/stdio.js"`；失败看 stderr。
- **工具名搜不到 / 调不通**：图标是 **`search-icon`**（不是 `search-icons`）；组件类必须 `"framework": "vue-next"`。
- **`t-*` 属性不生效 / 报错**：先 `get-component-docs`（样式用 `get-component-dom`），再改代码。
- **弹窗表单过高过挤**：按 Important 第 5 条改 `t-drawer`。
- **与后端字段对不上**：核对 IDL camelCase 与 envelope；会话四字段；不要本地改名绕过。

## 检查清单

- [ ] 开发前已盘点 `web/src/components/s-*`，有匹配即复用
- [ ] 第 3 处重复已抽 `s-*`，原有拷贝已全部改为引用（或 spec 写明暂不抽）
- [ ] 通用在 `components/s-*`，业务在模块 `components/`
- [ ] 已用 `scripts/tdesign_mcp.py` 连接并 `list-tools` 成功
- [ ] 写 `t-*` 前已 `get-component-docs`（`framework: "vue-next"`）；样式问题已查 `get-component-dom`
- [ ] 命名四条；类型不内联重复；文案走 locales
- [ ] 请求/envelope/会话字段与后端一致
- [ ] 高频任务模式与 `references/high-frequency-tasks.md` 一致
- [ ] 弹出表单：`t-form-item` >5 用 `t-drawer`，≤5 用 `t-dialog`
- [ ] `WEB=1 ./script/check.sh` 绿
