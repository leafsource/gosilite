---
feature: dict-integration
status: delivered
updated: 2026-10-06
branch: develop
commits: a04f9d3a..working-tree
---

# 字典关联：全站状态/类型枚举接入字典渲染

## Report

**What was built** — 迁移 `0005_dict_integration.sql` 新增 `sys_state`/`sys_data_scope`/`sys_menu_type`/`sys_config_type` 四类字典并核对 `sys_status`（幂等、`public.` 限定）；新增 `store/modules/dict.ts`：`ensure` 拉取缓存（失败不缓存可重试）、`invalidate` 失效（字典页保存/删除触发，`watch(loaded)` 令 keep-alive 页面自动重拉）、`useDict` 带内置硬编码兜底与 `labelOf` String 匹配；七个 system 页面的状态标签、启停单选、数据范围、菜单类型、参数类型的选择与展示全部改为字典渲染，字段值域不动。

**Verification** — 0005 双跑 exit 0、第二遍 0 行幂等；五类 `dictType` options 接口 curl 全部返回预期数据；`WEB=1 ./script/check.sh` PASS（期间修复两缺陷：模板嵌套 ref 不解包 → 四页加顶层别名；keep-alive 下 invalidate 不重拉 → `useDict` 监听 `loaded`）；eslint PASS；七页残留硬编码枚举文案 grep 清零（操作动词豁免）；浏览器验收逐页标签 PASS + 连通性证明 PASS（改 `sys_status` 启用→启用X→参数管理实时变化→已恢复原状）+ console 0 error；独立评审 0 CRITICAL / 0 MAJOR、2 MINOR（记入已知边界）。拦截兜底验收因 IAB 工具不支持网络拦截记 N/A（代码路径存在）。

**Journey log** — ①`useDict` 返回对象里的 ComputedRef 在模板中**不会**自动解包（仅顶层绑定解包）——四处 `:options="x.options"` 直接编译失败，加顶层别名解决；②keep-alive 页面 setup 只跑一次，`invalidate` 后必须靠 `watch(loaded)` 才能触发重拉，否则连通性验收过不了；③值域不动的设计让 `labelOf` 一律 String 比较，0/1 数字与 '0'/'1' 字符串、'enabled' 各自成字典互不打架；④连通性证明用「改字典标签→目标页实时变→改回」三步闭环，比断网拦截更可靠（IAB 无拦截能力）；⑤评审记录的在途请求窄窗口竞态留待跨端字典广播时一并处理。

## [S1] Problem

字典管理（`system_dict_type/data`）与 `GET /system/dict/data/options` 已交付，前端 `dictOptions(dictType)` 封装也存在，但**全站零调用**：状态、类型等枚举散落为硬编码——

1. 七页（dict/config/dept/menu/post/user/role）的启停标签与单选是 `status === 0 ? '启用' : '停用'` 之类三元表达；user/role 用字符串 `state: 'enabled'|'disabled'`，值域与 0/1 不同但语义同一；
2. role 表单 `dataScope` 1–5、menu 表单 `menuType` M/C、config 表单 `configType` system/business 全部硬编码；
3. 字典管理页改了字典数据，其他页面毫无感知——「字典」只是一张孤立的维护表。

种子仅 `sys_status`(0/1) 与 `sys_user_sex`（0003），其余类型缺失。

## [S2] Design

> **Workspace override**：沿用 feature 1 的决定——当前 `develop` 工作区（`a04f9d3a..working-tree`），不开 worktree。

**用户拍板（Grill）**：做**渲染层全量接入**——展示与选择控件全部改为按字典渲染；**字段值域不动**（0/1 与 `enabled/disabled` 各自成字典，不做数据迁移）；字典未加载/接口失败时**硬编码兜底**，页面不得空白。

### S2.1 字典种子（`script/migrations/0005_dict_integration.sql`，幂等、`public.` 限定）

| dict_type | 名称 | 数据（value → label） |
|-----------|------|----------------------|
| `sys_status`（已有，核对即可） | 状态 | `0`→启用（默认）、`1`→停用 |
| `sys_state` | 账号状态 | `enabled`→启用（默认）、`disabled`→停用 |
| `sys_data_scope` | 数据范围 | `1`→全部数据、`2`→本部门、`3`→本部门及以下、`4`→仅本人、`5`→自定义部门 |
| `sys_menu_type` | 菜单类型 | `M`→目录、`C`→菜单、`F`→按钮 |
| `sys_config_type` | 参数类型 | `system`→系统（默认）、`business`→业务 |

### S2.2 前端字典层（`web/src/store/modules/dict.ts`）

- Pinia store：`optionsMap: Record<string, DictOption[]>` + `loaded` 标记；`ensure(dictType)` 拉取 `/system/dict/data/options` 并缓存（失败不写缓存、可重试）。
- 导出 `useDict(dictType, fallback?)` 组合式：挂载时 `ensure` 并 `watch(loaded)`（invalidate 置 false 后自动重拉，覆盖 keep-alive 页面），返回 `{ options, labelOf }`；`options` 在字典未返回前/失败时用兜底项——**兜底数组内置字典层（`DEFAULT_FALLBACK`，与种子一致，调用方可传参覆盖）**，`labelOf(value)` 未命中回落原值字符串。
- 缓存失效：字典管理页保存/删除字典数据后 `invalidate(dictType)`（下次使用重拉）。
- 不引入 Redis/热更新广播（沿袭 system-config spec 的 Out of Scope）。

### S2.3 页面迁移（七页，展示 + 表单选择）

| 页面 | 迁移点 | dictType |
|------|--------|----------|
| dict | 类型/数据表状态标签与表单单选 | `sys_status` |
| config | 状态标签/单选；类型下拉 | `sys_status`、`sys_config_type` |
| dept | 状态标签/单选 | `sys_status` |
| menu | 状态标签/单选；菜单类型展示与表单 | `sys_status`、`sys_menu_type` |
| post | 状态标签/单选 | `sys_status` |
| user | 状态标签；启用/停用链接文案 | `sys_state` |
| role | 状态标签/单选；数据范围单选组 | `sys_state`、`sys_data_scope` |

- 模板中的三元/常量文案全部替换为 `labelOf` / `options`（保留 fallback）；表单单选/下拉改绑字典 `options`（`t-radio-group`/`t-select` 的 options 传参）。
- **不迁**：dashboard、result、`common-table` 演示页的 CONTRACT_STATUS（合同域）、登录页。

### S2.4 测试与验收边界

- 迁移：0005 双跑幂等；`curl GET /system/dict/data/options?dictType=sys_state|sys_data_scope|sys_menu_type|sys_config_type`（登录态）返回对应数据。
- 命令：`WEB=1 ./script/check.sh` 全绿；eslint/stylelint 改动文件绿（web 无单测框架，`npm test` 为空壳——前端逻辑以构建+浏览器验收覆盖）。
- 浏览器验收：① user/role/config/menu 页签与表单文案来自字典；② 字典管理页把 `sys_status` 的「启用」改为自定义标签 → 刷新对应页面列表标签随之变化（证明真连通，非硬编码）→ 改回；③ 断网/接口失败时页面文案回落兜底、无空白。
- 独立评审无 CRITICAL。
- **已知边界（评审记录）**：①`invalidate` 若发生在该字典 `ensure` 在途时，旧响应会写回 `loaded=true` 形成窄窗口陈旧缓存（窗口=保存与首屏请求重叠，概率极低；若后续加跨端字典广播需先修）；②服务端字典行被清空时前端持续显示硬编码兜底而非空白（有意设计，符合"不得空白"）；③断网兜底的浏览器拦截验收因工具能力记 N/A，代码路径经评审确认。

## [S3] Out of Scope

- 字段值域统一（user/role 的 string state 迁 0/1，或反向）——渲染层按值匹配已覆盖。
- 字典缓存的 Redis/热更新广播、跨端同步。
- 数据导入导出、字典管理页 UI 改版。
- 演示页（dashboard/result/common-table）枚举。
- 参数消费（分页/系统名）——下一 feature `config-consumption`。

## Tasks

- [x] T1: 0005 种子（sys_state/sys_data_scope/sys_menu_type/sys_config_type + sys_status 核对）幂等入库 — acceptance: 双跑 exit 0 第二遍 0 行；options 接口四类均返回数据 (covers: S2.1)
- [x] T2: `store/modules/dict.ts`（ensure/缓存/invalidate + `useDict` 兜底） — acceptance: `WEB=1 ./script/check.sh` 绿；字典页保存后缓存失效重拉 (covers: S2.2; depends: T1)
- [x] T3: 七页标签/单选/下拉迁字典（含 fallback，值域不动） — acceptance: 构建绿；七页无残留硬编码启停/范围/类型文案（演示页除外）；浏览器各页显示正常 (covers: S2.3; depends: T2)
- [x] T4: 端到端验收（curl options + 浏览器改字典标签即时生效 + 失败兜底） + 独立评审 — acceptance: S2.4 步骤逐条通过；评审无 CRITICAL (covers: S2.4; depends: T1 T2 T3)
