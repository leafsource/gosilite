# TDesign MCP 使用手册（frontend-dev）

> 契约来源：`tdesign-mcp-server@0.2.4` 实际注册 tool + 本目录 `scripts/tdesign_mcp.py` 实测。
> 本仓库前端固定：`framework = "vue-next"`。
>
> **环境事实**：MiMo 引擎侧 MCP 挂载在本环境**不可用**（新会话/重启引擎无效）。因此采用 **手动拉起 server + 脚本连接调用**。

## 1. 连接方式（主路径：脚本托管 server）

### 1.1 一键调用（推荐）

脚本会自己拉起 stdio server，完成 JSON-RPC 握手后调用工具，用完即退。

```bash
# 仓库根执行
SKILL=.agents/skills/frontend-dev/scripts/tdesign_mcp.py

# 1) 看工具（验证连接）
python3 "$SKILL" list-tools

# 2) 查组件文档（写 t-* 前强制）
python3 "$SKILL" call get-component-docs \
  '{"framework":"vue-next","names":["table","select","form"]}'

# 3) 查 DOM / 类名
python3 "$SKILL" call get-component-dom \
  '{"framework":"vue-next","names":["dialog"]}'

# 4) 选型
python3 "$SKILL" call get-component-list '{"framework":"vue-next"}'

# 5) 图标名
python3 "$SKILL" call search-icon '{"keywords":["download","export"],"limit":8}'

# 6) 升级前 changelog
python3 "$SKILL" call get-component-changelog \
  '{"framework":"vue-next","names":["table","form"],"version":"1.10.0"}'
```

启动顺序（脚本内）：

1. `$TDESIGN_MCP_CMD`（自定义命令，空格分隔）
2. 已安装包入口 `node .../tdesign-mcp-server/dist/stdio.js`（优先，免 npx 网络）
3. `npx -y tdesign-mcp-server@latest`（兜底）

### 1.2 重要约束

| 做 | 不做 |
|----|------|
| 用 `scripts/tdesign_mcp.py` 自拉自管子进程 | `sudo npx ...` 挂在终端（管道不在脚本手里，无法被调用） |
| 失败时看 stderr / 脚本报错 | 假装引擎 MCP 已挂载 |
| 一次任务可反复调用脚本 | 凭记忆写 `t-*` API |
| 调不通就停在编码前 | 用 shell 裸拼 JSON-RPC 长脚本（用现成脚本） |

**协议**：Newline-delimited JSON-RPC 2.0 over stdio（`initialize` → `notifications/initialized` → `tools/list` / `tools/call`）。详见 §7。

### 1.3 若会话工具列表里真的有 MCP 工具

直接调用原生 `get-component-*` / `search-icon`，参数与下表一致，**不必**再跑脚本。本环境通常没有，有则优先原生工具。

---

## 2. 工具契约（准确参数）

`names` 为 TDesign 组件名（`table`、`form-item`、`select`…），支持批量数组。

### 2.1 `get-component-list`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `framework` | string | 是 | 本仓库 `"vue-next"` |

```json
{ "framework": "vue-next" }
```

### 2.2 `get-component-docs`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `names` | string[] | 是 | 组件名，批量 |
| `framework` | string | 是 | `"vue-next"` |

**何时用**：写/改任何 `t-*` **之前**（强制）。

```json
{
  "framework": "vue-next",
  "names": ["table", "form", "select"]
}
```

### 2.3 `get-component-dom`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `names` | string[] | 是 | 组件名，批量 |
| `framework` | string | 是 | `"vue-next"` |

**何时用**：自定义 CSS、覆盖主题、类名/DOM 层级。

```json
{
  "framework": "vue-next",
  "names": ["dialog", "message"]
}
```

### 2.4 `get-component-changelog`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `names` | string[] | 是 | 组件名，批量 |
| `framework` | string | 是 | `"vue-next"` |
| `version` | string | 否 | 返回 ≥ 该版本日志 |

```json
{
  "framework": "vue-next",
  "names": ["table", "form"],
  "version": "1.10.0"
}
```

### 2.5 `search-icon`（单数，不是 `search-icons`）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keywords` | string[] | 是 | 英文关键词，≥1；多词优先 AND，无结果降级 OR |
| `limit` | int > 0 | 否 | 默认 `20` |

```json
{
  "keywords": ["search", "filter"],
  "limit": 5
}
```

实测返回示例：`["filter","search","filter-1","filter-2","filter-3"]`。

### 2.6 `framework` 枚举

```text
react | react-chat | vue | vue-next | vue-next-chat
mobile-react | mobile-vue | miniprogram | miniprogram-chat | uniapp | uniapp-chat
```

**本仓库 web/ 一律 `vue-next`。**

---

## 3. 标准调用顺序

```text
选型不定     → get-component-list
写/改 t-*   → get-component-docs（强制）
样式/DOM    → get-component-dom
图标        → search-icon
升级依赖    → get-component-changelog
```

可一次 `names` 批量多组件，减少往返。

---

## 4. 真实场景调用示例

### 场景 A：列表页加「状态筛选」

```bash
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-docs \
  '{"framework":"vue-next","names":["select","option","form"]}'
```

落在 `s-table` 的 `searchFields`，不自建筛选条。

### 场景 B：表单 dialog → drawer

```bash
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-docs \
  '{"framework":"vue-next","names":["dialog","drawer","form"]}'
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-dom \
  '{"framework":"vue-next","names":["drawer"]}'
```

### 场景 C：自定义 t-tag 样式

```bash
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-docs \
  '{"framework":"vue-next","names":["tag","steps"]}'
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-dom \
  '{"framework":"vue-next","names":["tag"]}'
```

### 场景 D：图标选型

```bash
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call search-icon \
  '{"keywords":["download","export"],"limit":8}'
```

### 场景 E：升级 tdesign-vue-next

```bash
python3 .agents/skills/frontend-dev/scripts/tdesign_mcp.py call get-component-changelog \
  '{"framework":"vue-next","names":["table","form","dialog","select"],"version":"<当前版本>"}'
```

---

## 5. 与本仓库约定的衔接

| 约定 | MCP 用法 |
|------|----------|
| 通用 `s-*` 优先 | 先盘点 `web/src/components/s-*`；原生 `t-*` 仍须 docs |
| 第 3 处重复抽 `s-*` | 抽之前 docs 确认组合能力 |
| 弹窗表单 >5 字段用 drawer | docs 核对 `t-drawer` / `t-dialog` |
| 列表筛选进 `s-table` | docs 核对 `t-select` / `t-form` 事件 |
| 交付 `WEB=1 ./script/check.sh` | MCP 不替代构建验收 |

---

## 6. 故障排查

| 现象 | 原因 | 处理 |
|------|------|------|
| 引擎/会话工具列表无 MCP | **本环境挂载不可用** | 不要指望重启；用 `scripts/tdesign_mcp.py` |
| `sudo npx` 在终端挂着却调不通 | stdio 管道在 sudo/TTY，客户端接不上 | 杀掉；改用脚本自拉子进程 |
| 脚本 timeout | npx 网络慢 / 包未缓存 | 先保证已安装；或设 `TDESIGN_MCP_CMD` 指向 `node <entry>` |
| `initialize failed` | Node 环境异常 | 换用户级 node；看 stderr |
| 返回「组件不存在」 | `names` 错或 framework 不对 | `get-component-list` 核对；本仓库 `vue-next` |
| 工具名搜不到 | 写成 `search-icons` | 实际是 **`search-icon`** |
| 想凭记忆写 t-* | — | **禁止** |

环境变量：

```bash
export TDESIGN_MCP_CMD="node /path/to/tdesign-mcp-server/dist/stdio.js"
export TDESIGN_MCP_TIMEOUT=120
```

---

## 7. 协议备忘（脚本已封装，一般不必手写）

```text
stdin  → {"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}\n
stdout ← {"jsonrpc":"2.0","id":1,"result":{...}}\n
stdin  → {"jsonrpc":"2.0","method":"notifications/initialized"}\n
stdin  → {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"...","arguments":{...}}}\n
```

- 帧：一行一个 JSON，`\n` 结尾
- banner 在 **stderr**（如 `TDesign MCP Stdio Server is running`）
- `tools/call` 的 `result.content[].type === "text"` 里是 JSON 字符串

---

## 8. 快速自检

- [ ] `python3 scripts/tdesign_mcp.py list-tools` 能列出 5 个工具
- [ ] 调用参数带 `"framework": "vue-next"`
- [ ] 写 `t-*` 前已 `get-component-docs`
- [ ] 样式问题已 `get-component-dom`
- [ ] 图标用 `search-icon`（单数）
- [ ] 未依赖引擎 MCP 挂载 / 未 `sudo` 挂空进程
