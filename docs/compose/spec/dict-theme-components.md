---
feature: dict-theme-components
status: designed
updated: 2026-10-07
branch: develop
commits:
---

# 字典主题色与统一字典组件

## Report

## [S1] Problem

1. 字典管理页类型表的「数据项」入口改名「列表」保留，同时**点击「字典类型」单元格**也能打开同一抽屉——两条路径功能相同，行尾与行内都能直达数据列表。
2. 字典数据只有 label/value，没有视觉属性；业务想用字典值直接驱动标签颜色（启用=绿、危险=红等），目前各页只能手写三元表达式 `row.status === 0 ? 'success' : 'danger'`。数据项列表中希望**标签列直接渲染主题结果**（默认 `default`=不着色），所见即所得。
3. 全站字典消费是散装的：六个页面各自维护 `statusOptions/stateOptions/...` computed、手写 `t-tag` 三元主题、手写 `t-radio-group :options`，同一语义重复 N 遍，新增一个字典就要再拷贝一次。

## [S2] Design

> **Workspace override**：沿用本仓已交付 4 个 spec 的惯例——在当前 `develop` 检出直接开发，不开 `.worktrees/`（origin/main 为空初始提交，真实主线为本地 `develop`）。
>
> **用户拍板（Grill）**：①主题与自定义颜色**两字段并存、自定义优先**；②三个新组件**顺带全量回改存量页面**；③ TDesign MCP 用户已在命令行开启，前端编码前实测连接，若本会话工具仍未注入则降级用 `node_modules/tdesign-vue-next` 权威 `.d.ts` 并记入 journey。

### S2.1 数据模型（迁移 `script/migrations/0009_dict_data_theme.sql`，幂等、`public.` 限定）

`system_dict_data` 新增两列（沿 0008 的 `ADD COLUMN IF NOT EXISTS` 风格，BEGIN/COMMIT）：

| 列 | 类型 | 默认 | 语义 |
|----|------|------|------|
| `theme` | VARCHAR(16) NOT NULL | `'default'` | 主题令牌：`info`（纯文本，不渲染标签）/ `default`（中性标签）/ `primary` / `success` / `danger` / `warning` / `''`（未勾选，伴随自定义色使用，渲染回落 default） |
| `color` | VARCHAR(32) NOT NULL | `''` | 自定义色：`#RGB` 或 `#RRGGBB`，`''`=未自定义 |

- **主题与颜色互斥（用户拍板修订）**：UI 上填了自定义色 → 主题单选全部取消勾选（`theme=''`）；勾选任一主题 → 颜色清空（`color=''`）。**服务端同样强制**：`color != '' && theme != ''` → 参数错误。
- **渲染优先级**：`color` 非空 → 自定义色标签；`theme='info'` → 纯文本（不包 t-tag）；其余令牌 → 对应主题标签；`theme=''` 且无色 → 中性 default 标签。
- **ADD COLUMN 带 DEFAULT**：存量行自动得 `'default'`。
- **种子回填（保持改前视觉，不改 label/value；守卫 `theme='default'`，不覆盖用户已改值）**：
  - `sys_status`：`0`→`success`，`1`→`danger`（现页签绿/红）。
  - `sys_state`：`enabled`→`success`，`disabled`→`default`（现状 `enabled?success:default`）。
  - `sys_menu_type`：按 menu 页 `menuTypeTheme()` 现状回填；`sys_data_scope` 全部回填 `primary`（role 页数据范围标签现状恒 primary）；其余字典保持 `default`。
  - 全部不回填 `color`。
- 重复执行幂等：列 IF NOT EXISTS；回填 UPDATE 带守卫。

### S2.2 后端契约（先 IDL 后生成）

- `idl/framework/system_dict.thrift`：
  - `DictDataDTO` += `9: string theme`、`10: string color`；
  - `DictDataSaveRequest` += `optional string theme`、`optional string color`（缺省空串）；
  - `DictOptionDTO` += `3: string theme`、`4: string color`（options 接口同步携带，供组件着色）。
- `hz update -I idl -idl idl/framework/system_dict.thrift -customize_package idl/package.yaml` 重新生成 `api/model` / `api/router`（生成物不手改）。
- 迁移后 `go run ./script/codegen` 重生成 `dal/model|query`（`*.gen.go` 不手改）。
- `api/handler/framework/syscfg/convert.go`：`toDictDataDTO` / `fromDictDataSaveRequest` / `toDictOptionDTOs` 三处映射 theme/color。
- **Service 校验**（`application/framework/service/dict` 的 Create/Update）：`theme` 必须 `∈ {'', info, default, primary, success, danger, warning}`；`color` 必须 `''` 或 `#`+3/6 位 hex（大小写不敏感）；**互斥**：`color` 非空时 `theme` 必须为 `''`（否则参数错误）。非法值 `errorx` 参数错误回包，Handler 不吞错。Options 接口不过滤 theme/color，仍按现有 status/排序规则返回。
- 转换集中在 handler 包 `convert.go`；Repository 只透传模型字段。

### S2.3 字典管理页改造（`web/src/pages/system/dict/index.vue`）

- **类型表**：操作列保留三入口，「数据项」**改名「列表」**（宽度 180 不变），编辑/删除不动；`dictType` 单元格改为 `t-link`，点击调用原 `onSelectType(row)` 打开数据项抽屉——与「列表」入口功能完全相同。行内其余区域不整行绑定点击，避免误触。
- **数据项表**：**「标签」列以行内 `theme`/`color` 渲染**——`theme='info'` 显示纯文本，其余渲染 `t-tag`（自定义色优先、default 中性），不新增独立方块颜色列。注意：**不用 `s-tag-dict` 按值回查**——`options` 仓库查询仅含 `status=0` 的行，停用行无法解析回查会回落裸值；`s-tag-dict` 留给消费侧页面（S2.5）。值/排序/状态/操作列不变。
- **数据项表单**：新增「主题」后 `t-form-item` 共 6（标签/值/排序/状态/默认/主题+颜色合一行），按 >5 规则从 `s-form-dialog` 迁到 `s-form-drawer`。
  - **主题与颜色在同一表单项（同一列）**，因二者互斥：主题单选组 = `info`（纯文本）/ `default`（中性）/ `primary` / `success` / `danger` / `warning`，label 内嵌色块预览；旁边 `t-color-picker`（可清空，hex）。
  - **互斥联动**：选择任一主题 → 清空 `color`；填入自定义色 → 主题单选全部取消勾选（`theme=''`）；清空颜色 → 保持未勾选，渲染回落中性 default。
- 状态类标签/单选在本页直接改用新组件（见 S2.5/dict 行），删除本地 `statusOptions` computed。

### S2.4 三个通用组件（`web/src/components/s-<name>/index.vue`，沿现有直接 import 模式，不全局注册）

共享解析逻辑：`useDict(dictType)` → 按 `String(value)` 命中 option 得 `label/theme/color`；未命中回落 `labelOf` 原值文案 + 默认主题。`DictOption` 类型扩展 `theme?: string; color?: string`。

**`STagDict`（s-tag-dict）** 纯展示标签：

| prop | 类型 | 默认 | 说明 |
|------|------|------|------|
| `dictType` | string | 必填 | 字典类型 |
| `value` | string \| number | — | 字典值；null/undefined/'' 渲染空 |
| `variant` | `'light' \| 'light-outline' \| 'outline'` | `'light'` | 透传 t-tag |
| `size` | `'small' \| 'middle' \| 'large'` | `'middle'` | 透传 t-tag |

渲染：`color` 非空 → t-tag 自定义色（背景着色，前景/边框按 TDesign `color` 语义）；`theme='info'` → **纯文本 span（不渲染 t-tag）**；`theme` 为 `primary/success/danger/warning` → `:theme`；`default`/`''`/未命中 → t-tag `theme="default"`（中性）。不 emit。

**`SRadioDict`（s-radio-dict）** 字典单选：

| prop | 类型 | 默认 | 说明 |
|------|------|------|------|
| `modelValue` | string \| number | — | v-model |
| `dictType` | string | 必填 | |
| `valueType` | `'auto' \| 'string' \| 'number'` | `'auto'` | `auto` = modelValue 是 number 则选项值转 number，否则保持 string（对接现有 status/dataScope 数字字段） |
| `disabled` / `size` / `variant` | — | — | 透传 t-radio-group |

options 直接取 `useDict().options`（经 valueType 规整）；emit `update:modelValue` + `change`。

**`SSelectDict`（s-select-dict）** 字典下拉：

| prop | 类型 | 默认 | 说明 |
|------|------|------|------|
| `modelValue` | string \| number | — | v-model |
| `dictType` | string | 必填 | |
| `valueType` | 同上 | `'auto'` | 同上 |
| `placeholder` | string | `'请选择'` | |
| `clearable` | boolean | `true` | |
| `disabled` | boolean | `false` | |

emit `update:modelValue` + `change`；options 同源同规则。

组件命名/目录遵循 frontend-dev：`components/s-tag-dict/index.vue` 等，类型内联在组件或就近 `s*Dict.ts`，不重复建 DTO。

### S2.5 存量页面全量回改（用户拍板范围）

以「标签 → `s-tag-dict`、选择 → `s-radio-dict`/`s-select-dict`、手写 options computed 全删」为准则，逐页：

| 页面 | 回改点 | dictType / 备注 |
|------|--------|-----------------|
| dict | 类型表+数据项表状态标签、类型表单+数据项表单状态单选 | `sys_status`（valueType=auto，表单字段为 number） |
| config | 状态标签/单选；**类型列标签 + 类型下拉** | `sys_status`、`sys_config_type` |
| organization（部门） | state 标签/单选 | `sys_state` |
| menu | 状态标签/单选；菜单类型标签/单选 | `sys_status`、`sys_menu_type` |
| post | 状态标签/单选 | `sys_status` |
| role | state 标签/单选；数据范围单选 | `sys_state`、`sys_data_scope` |

- 删除各页 `statusOptions/stateOptions/dataScopeOptions/menuTypeOptions/configTypeOptions` 及 labelOf 三元展示；`useDict().labelOf` 仅保留**非标签**场景（链接文案等）。
- `menuTypeTheme()` 等映射函数删除，视觉由 S2.1 回填的 theme 承接（回填以改前视觉为准则）。
- **不迁**：`sys_menu_type` 之外的展示型三元（menu 的 `visible`、dict 的 `isDefault` 是/否、演示页 dashboard/result/common-table、登录页）——非字典语义，保持现状。
- 验收基准：改前后各页标签视觉一致（成功绿/危险红/中性灰不变），仅来源换成组件。

### S2.6 测试与验收

- 迁移：0009 双跑幂等（第二遍 0 变更）；回填后 `SELECT dict_type,dict_value,theme FROM system_dict_data WHERE theme<>''` 符合 S2.1。
- 契约：登录态 `curl GET /system/dict/data/options?dictType=sys_status` 返回项含 `theme/color`；`POST /system/dict/data` 带 `theme='nope'` 或 `color='red'` 返回参数错误。
- 命令：`./script/check.sh` 与 `WEB=1 ./script/check.sh` 全绿；eslint/stylelint 改动文件绿（web 无单测框架，以构建+浏览器验收覆盖）。
- 浏览器验收：① 类型表操作列为「列表/编辑/删除」、点「列表」与点「字典类型」单元格均开抽屉；② 数据项表单在 drawer、主题+取色器**互斥联动**（填色取消全部勾选、勾主题清空颜色）、保存后「标签」列即时生效（info=纯文本、自定义色覆盖、default 中性）；③ 系统各页状态/类型标签与改前视觉一致，来源为新组件；④ config 页类型列显示 label 标签、类型下拉（s-select-dict）选择保存生效；⑤ console 0 error。
- 独立评审无 CRITICAL。
- **TDesign MCP**：前端编码前实测 `get-component-docs`；若会话工具仍未注入 → 以 `node_modules/tdesign-vue-next` `.d.ts` 为 API 权威降级，记入 journey。

## [S3] Out of Scope

- 字段值域统一（0/1 vs enabled/disabled 各自成字典，沿 dict-integration 既定）。
- 字典缓存 Redis/热更新广播、跨端同步。
- `cssClass` 字段语义改造或删除（保持现状不动）。
- 演示页（dashboard/result/common-table/form/detail）、登录页枚举。
- s-* 组件全局注册（沿现有直接 import 模式）。
- s-table `searchFields` 筛选区接 s-select-dict（当前筛选均为 keyword/分组，无字典筛选场景）。

## Tasks

- [ ] T1: 0009 迁移（theme/color 两列 + 种子主题回填）幂等入库 — acceptance: 双跑 exit 0 第二遍 0 变更；回填行与 S2.1 清单一致 (covers: S2.1)
- [ ] T2: IDL 扩展 → hz update → gorm gen → convert 三处映射 → Service theme/color 校验（含 info 令牌与主题×颜色互斥） — acceptance: curl options 带 theme/color；非法 theme/color/互斥并存被参数错误拒绝；`./script/check.sh` 绿 (covers: S2.2; depends: T1)
- [ ] T3: 前端 DictOption 扩展 + s-tag-dict / s-radio-dict / s-select-dict 三组件（s-tag-dict 支持 info 纯文本） — acceptance: `WEB=1 ./script/check.sh` 绿；三组件 props/emit 对齐 S2.4；MCP 实测或 .d.ts 降级有记录 (covers: S2.4)
- [ ] T4: 字典页改造（操作列数据项改名列表、dictType 可点同功能、标签列 info/主题/自定义色渲染、数据项表单迁 drawer + 主题与颜色同行互斥联动、本页状态控件换组件） — acceptance: 浏览器验收 ①②⑤ 通过 (covers: S2.3; depends: T2 T3)
- [ ] T5: 存量六页全量回改接三组件，删手写 options computed — acceptance: 浏览器验收 ③④ 通过；各页无 `statusOptions/stateOptions/dataScopeOptions/menuTypeOptions` 残留 (covers: S2.5; depends: T3)
- [ ] T6: 端到端验收（迁移/契约/命令/浏览器） + 独立评审 — acceptance: S2.6 逐条通过；评审 0 CRITICAL (covers: S2.6; depends: T1 T2 T3 T4 T5)
