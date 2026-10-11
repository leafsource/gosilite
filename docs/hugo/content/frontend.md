---
title: "前端"
weight: 4
---

gosilite 前端是 `web/` 下的 Vue 3 + TypeScript + Vite + TDesign 单页应用：业务侧栏与可访问路由**唯一来源**是后端菜单接口 `GET /menu/routes`，请求统一走 axios 封装并按 `{code,msg,data}` 信封解包，登录态用 JWT Bearer + 401 单飞刷新。开发时记住三条硬规则：写任何 `t-*` 组件前先用 TDesign MCP 查文档、同一 UI 第 3 次拷贝必须抽 `s-*` 通用组件、弹出表单字段 >5 个用 `t-drawer`。本文面向新读者，命令与目录均以仓库现状为准；后端契约见[后端](/docs/backend/)，启动步骤见[快速开始](/docs/quick-start/)。

## 一、技术栈与目录

### 1.1 技术栈

| 层 | 选型 | 说明（来源） |
|----|------|--------------|
| 框架 | Vue 3.5 + TypeScript 6 | `web/package.json` |
| 构建 | Vite 8（dev 端口 **3002**，`0.0.0.0`） | `web/vite.config.ts` |
| UI | TDesign `tdesign-vue-next` 1.20 + `tdesign-icons-vue-next` | 组件前缀 `t-*` |
| 状态 | Pinia 3 + `pinia-plugin-persistedstate`（登录态持久化） | `web/src/store/` |
| 路由 | vue-router 5（History 模式） | `web/src/router/` |
| 请求 | axios 1（封装类 `VAxios`） | `web/src/utils/request/` |
| 国际化 | vue-i18n 11 | `web/src/locales/` |
| 图表 | echarts 6 | 监控/看板页 |
| 样式 | Less + lightningcss | `web/src/style/` |
| 运行环境 | Node ≥ 22.12 | `package.json#engines` |

### 1.2 目录职责

`web/src/` 按「API → 页面 → 装配」划分：

| 目录 | 职责 |
|------|------|
| `api/` | 业务接口与类型：`user.ts`、`permission.ts`、`notice.ts`、`common/file.ts`、`system/*.ts`（dict、role、menu、oss、log、monitor…），类型与接口同放（`*Model.ts`） |
| `components/` | **通用组件**：`s-<name>/` 目录（组件名 `S<Name>`）+ 少量模板示例组件，清单见第五节 |
| `pages/` | 页面组件，被菜单 `component` 字段按相对路径引用（不带 `.vue`），见第六节 |
| `store/` | Pinia：`user`（会话）、`permission`（路由/权限）、`dict`（字典缓存）、`config`（系统参数）、`notification`（铃铛）、`tabs-router`（页签与 keep-alive）、`setting` |
| `router/` | 静态路由装配；`modules/` 仅保留 `homepage`、`result`、`notice`、`user` 四个非业务模块 |
| `utils/` | `request/`（axios 封装）、`route/`（菜单转路由、component 动态导入、404 兜底）、`date/color/file/charts` |
| `layouts/` | 登录后外壳：`index.vue` 主布局、`components/{SideNav,Header,Content,Breadcrumb,MenuContent,Notice}`、`setting.vue` 主题设置、`frame/` iframe 容器 |
| 其他 | `main.ts` 入口、`permission.ts` 全局路由守卫、`directives/permission`（`v-permission` 按权限码显隐）、`config/`、`constants/`、`hooks/`、`types/`、`assets/`、`style/` |

依赖方向（与后端一致，前端只经 HTTP 消费）：

```text
pages / layouts → api/*.ts → utils/request(axios, /api 前缀) → Vite proxy → Hertz :8888
                 ↘ store/(Pinia)  ↘ components/(s-* 与示例组件)
```

## 二、动态路由

**结论**：业务侧栏与可访问 path 唯一来源是 `GET /menu/routes`（数据库 `system_menu`，已按角色过滤）；`router/modules/` 不再新增业务表。菜单 `component` 决定加载哪个页面文件，路由 `name` 由前端按 path 自动生成，页面 `defineOptions({ name })` 必须与其一致，否则 keep-alive 失效。

### 2.1 注册流程（`src/permission.ts` 全局守卫）

| 步骤 | 动作 | 代码位置 |
|------|------|----------|
| 1 | 未登录 → 除白名单 `/login` 外一律跳 `/login?redirect=…` | `permission.ts` |
| 2 | 已登录 → 拉 `getUserInfo()`，再 `config.loadPrivate()` 拉私有参数（会话内一次） | `permission.ts` |
| 3 | `asyncRoutes` 为空时请求 `getMenuList()`（`/menu/routes`）并逐条 `router.addRoute()` | `store/modules/permission.ts` |
| 4 | 过滤静态独占 path（`/login` + 首页整树），其余后端菜单全部注册 | `reservedPaths()` |
| 5 | 注册后基于 `to` 重新 `next({ ...to, replace: true })`，否则刷新直达子路由会 404 | `permission.ts` |

菜单对象 → 路由对象由 `utils/route/index.ts` 的 `transformObjectToRoute` 完成：

- `component = layout` 的目录节点套一层 `layouts/index.vue`；无 `LAYOUT` 的业务节点包一层空 `path` 的壳，子节点承载真实 path；嵌套目录只渲染 `layouts/router-pass.vue`，避免双层侧栏。
- 页面组件按 `import.meta.glob('../../pages/**/*.vue')` 匹配 `component`（相对 `pages/`、无 `.vue`，如 `system/dict/index`）。**匹配不到 → 控制台 warn 并回退 404 页**（`utils/route/constant.ts` 的 `NOT_FOUND_COMPONENT`），所以菜单里的 `component` 必须与真实文件一致。
- 找不到的路由统一落到 `/:path(.*)*` → `/result/404`。

### 2.2 name 对齐与 keep-alive

`web/src/api/permission.ts` 的 `genRouteName(fullPath)` 把 path 转成 PascalCase 路由名（后端不下发 name）：

| 菜单 path（fullPath） | genRouteName | 页面 `defineOptions({ name })` |
|----------------------|--------------|-------------------------------|
| `/dashboard/base` | `DashboardBase` | `DashboardBase` |
| `/system/dict` | `SystemDict` | `SystemDict` |
| `/system/contacts` | `SystemContacts` | `SystemContacts` |
| `/system/storage/oss` | `SystemStorageOss` | `SystemStorageOss` |
| `/monitor/data` | `MonitorData` | `MonitorData` |
| `/tool/docs` | `ToolDocs` | `ToolDocs` |

keep-alive 机制：`layouts/components/Content.vue` 的 `<keep-alive :include="aliveViews">`，`aliveViews` 取自 `store/modules/tabs-router.ts` 当前页签的路由 `name`（`isAlive` 且 `meta.keepAlive !== false`，即**默认保活**，显式 `keepAlive: false` 才禁用）。因此：

- 页面 **`defineOptions({ name })` 必须等于 `genRouteName(fullPath)`**，且不能与静态路由重名（重名时 `dedupeNames` 会追加 `_2` 后缀，keep-alive 将失效）。

### 2.3 新页面三步

1. 建 `web/src/pages/<域>/<页>/index.vue`，写 `defineOptions({ name: genRouteName(fullPath) })`。
2. `system_menu` 迁移脚本里写 `component`（相对 `pages/`、无 `.vue`）+ 权限码 + 角色绑定，见 `script/migrations/`。
3. 不改 `router/modules/`；侧栏、面包屑、权限随 `/menu/routes` 自动生效。菜单配置入口即「系统管理 → 菜单管理」（`pages/system/menu`，`component` 输入框提示与上述规则一致）。

## 三、登录与 API 层

### 3.1 会话与请求信封

| 项 | 约定 |
|----|------|
| 登录 | `POST /user/login`（`store/modules/user.ts` 的 `login` → `api/user.ts`） |
| 会话字段（camelCase，禁止 `token`/`expire`） | `accessToken` / `tokenType` / `expiresIn` / `refreshToken` |
| 持久化 | Pinia persist，`key: 'user'`，只存 `token`、`refreshToken` |
| 请求头 | `Authorization: Bearer <accessToken>`（`utils/request/index.ts` 拦截器） |
| 响应信封 | `{code,msg,data}`；`code === 200` 才 resolve，且业务层拿到的已是有值的 `data`，否则 `throw new Error(msg)`（`transformRequestHook`） |
| 取消与重试 | 同 URL 请求可取消；网络类错误默认重试 3 次（间隔 1s），HTTP 错误透出后端 `msg` 不盲试 |

### 3.2 `VITE_API_URL_PREFIX=/api` 与 Vite 代理

`web/.env` / `.env.development` 中 `VITE_API_URL_PREFIX = /api`；开发模式 `VITE_IS_REQUEST_PROXY = false` 使 axios **不带 host**，全部相对路径请求交给 Vite 代理（`web/vite.config.ts`）：

| 代理规则 | 目标 | 行为 |
|----------|------|------|
| `/api` | `http://127.0.0.1:8888/` | **剥离 `/api` 前缀**后转发（本地 Hertz 路由无前缀） |
| `/docs` | `http://127.0.0.1:8888/` | 原样透传（Hugo 文档站由 Hertz 托管） |

即：浏览器请求 `/api/menu/routes` → 后端收到 `/menu/routes`。**自带的原生 XHR / `t-upload` 上传地址必须手动带 `/api` 前缀**，否则绕过代理直接打到 3002 端口会 404（见[常见问题](/docs/faq/)）。

### 3.3 401 刷新（单飞）

1. 响应拦截器捕获 401 → `userStore.refreshAccessToken()`（`utils/request/index.ts`），用 `refreshToken` 调 `POST /user/refresh` 换新 token 后**重放原请求一次**（`retriedAfterRefresh` 标记，不会循环）。
2. 并发 401 共享同一个 `refreshPromise`（单飞），避免 refresh token 轮换被重复使用（`store/modules/user.ts`）。
3. 刷新失败（过期/被强退）→ `clearSession()` 并跳 `/login?redirect=…`，抛友好文案。
4. 上传例外：`t-upload` 走原生 XHR 不经 axios 拦截器，页面上传前调用 `ensureFreshAccessToken()`（解析 JWT `exp`，距过期 <60s 才刷新）。

## 四、通用组件清单

**命名规则**：通用组件放 `web/src/components/s-<name>/`，导出组件名 `S<Name>`；文件命名——组件 `index.vue`（或 `STable.vue`）、接口 `camelCase.ts`（`sTable.ts`）、样式 `kebab-case.css`。仅 1–2 处业务使用的放 `pages/<模块>/components/`，不进全局。以下与 `ls web/src/components/` 一致（17 个目录 + 1 个单文件）：

| 组件 | 职责 | 何时用 |
|------|------|--------|
| `s-table` | 查询表格封装：关键字/`searchFields` 搜索区、工具栏、列显隐、分页、刷新、多选与 `#batch` 批量插槽，emit `search/reset/refresh/page-change/selection-change` | 所有列表页（现 17 处引用），**优先用它，不自建筛选条** |
| `s-form-dialog` | 表单弹窗壳（`visible`/`title`/`width` 默认 520、`confirm`/`update:visible`） | 弹出表单 **字段 ≤5** 个 `t-form-item` |
| `s-form-drawer` | 表单抽屉壳（`size` 默认 `large`，同上事件） | 弹出表单 **字段 >5** 个 |
| `s-upload-image` | 图片上传（`limit` 5、类型/大小校验、`mode: id\|url`、可选压缩，端点 `/common/file/upload`） | 头像、图片字段、OSS 文件管理 |
| `s-upload-file` | 文件上传（默认 `mode: id`，doc/xlsx/pdf 等类型白名单，同端点） | 附件类字段 |
| `s-select-dict` | 字典下拉：按 `dictType` 拉选项，`valueType: auto` 跟随值类型 | 表单里选字典枚举 |
| `s-radio-dict` | 字典单选组（`variant`/`size` 可调，同 `valueType` 策略） | 表单里字典单选 |
| `s-tag-dict` | 字典值渲染为彩色标签（primary/success/danger/warning 四主题） | 列表状态/类型列展示 |
| `s-select-icon` | TDesign 图标选择器（按图标 stem 选择，如 `user`、`folder`） | 菜单图标配置 |
| `s-image-preview` | 图片 + 悬停缩放 + 查看器预览（`src` 支持逗号分隔多图 http 地址） | 文件管理缩略图、需要点开大图处 |
| `s-notice-detail` | 公告详情弹窗（`visible` + `detail`） | 铃铛、公告列表查看正文 |
| `common-table` | 模板自带的合同查询示例表单（`list/filter`、`list/tree` 使用） | 仅模板示例页，业务页用 `s-table` |
| `color` | 色块容器（预置品牌色显示纯色，否则显示色环渐变） | 主题设置面板 `layouts/setting.vue` |
| `thumbnail` | 缩略图（`circle` 圆形 / `layout` 88×48） | 模板示例与设置面板 |
| `trend` | 涨跌趋势文案（`up`/`down` 箭头 + 文字，可反色） | 看板/详情数据指标 |
| `product-card` | 产品卡片（`t-card theme=poster2`，头像/状态标签/下拉操作） | 卡片列表示例（`list/card`、`dashboard/detail`） |
| `result` | 结果页容器（`type`: `403`/`404`/`500`/`ie`/`wifi`/`maintenance`，标题 + 提示 + 插槽） | `pages/result/*` 与表单提交结果提示 |
| `LanguageSwitcher` | 顶栏语言下拉（读 `web/src/locales`，切换 `changeLocale`） | 登录页与主布局头部 |

## 五、页面域（`web/src/pages/`）

`pages/` 下的目录即菜单 `component` 的根；实际暴露哪些页面由后端菜单决定，未挂菜单的目录不会出现在侧栏。

| 目录 | 职责 |
|------|------|
| `login/` | 登录页（含 `components/Header` 等），登录后跳 `redirect` 参数或首页 |
| `dashboard/` | 首页看板：`base`（主看板，`DashboardBase`）、`detail`（详情分析） |
| `system/` | 系统管理域：`dict` 字典、`config` 参数设置、`menu` 菜单、`role` 角色（含授权抽屉）、`post` 岗位、`organization` 组织架构（部门 + 用户合并页，`/system/contacts`）、`notice` 公告管理、`log/operlog` 操作日志、`log/loginlog` 登录日志、`oss` 文件管理、`ossConfig` OSS 配置、`ossRule` OSS 处理规则（`storage/` 为已删除菜单遗留的空目录，无页面） |
| `monitor/` | 系统监控：`online` 在线用户、`data` 数据监控（总览/SQL/防火墙/URI 四个 Tab）、`cache` 缓存监控；「系统接口」Swagger 页已迁至 `tool/api` |
| `tool/` | 系统工具：`api` 系统接口（iframe 打开 `{prefix}/swagger/index.html`）、`docs` 开发文档（iframe 打开 `/docs/`，dev 走 Vite 代理） |
| `notice/` | 通知公告列表（`NoticeList`，铃铛/公告入口落地页） |
| `user/` | 个人中心：头像上传、资料、角色标签、修改密码 |
| `result/` | 结果页：`403`/`404`/`500`/`success`/`fail`/`network-error`/`browser-incompatible`/`maintenance` |
| `detail/`、`form/`、`list/` | 模板自带示例页（详情/表单/列表），用于组件演示，一般不挂业务菜单 |

## 六、开发规范

1. **先查文档再写 `t-*`**：本环境引擎 MCP 挂载不可用，一律用脚本手动连接；未查到 API 不得写业务代码。

   ```bash
   MCP=.agents/skills/frontend-dev/scripts/tdesign_mcp.py
   python3 "$MCP" list-tools
   python3 "$MCP" call get-component-docs '{"framework":"vue-next","names":["table","select","form"]}'
   ```

   选型 `get-component-list` → 写码前 `get-component-docs`（固定 `framework: "vue-next"`）→ 样式 `get-component-dom` → 图标 `search-icon`（单数 icon）。细则见 `.agents/skills/frontend-dev/references/tdesign-mcp.md`。

2. **组件第 3 次拷贝必须抽 `s-*`**：动手前先盘点 `web/src/components/s-*` 有无现成组件；同一 UI 结构第 3 次出现时停止复制，抽出 `s-<name>/`，**并把已有 3 处全部改为引用**，之后一律复用。
3. **弹出表单容器看字段数**：按 `t-form-item` 计数（不含底部按钮行），`>5` 用 `t-drawer`（即 `s-form-drawer`），`≤5` 用 `t-dialog`（即 `s-form-dialog`）。
4. **命名四条**：组件 `PascalCase.vue`（`STable.vue`）、目录主文件 `index.ts`/`index.vue`、接口 `camelCase`（`sTable.ts`）、资源/样式 `kebab-case`（`s-table.css`）。
5. **契约对齐后端**：`api/*.ts` 字段 camelCase、信封 `{code,msg,data}`、会话四字段固定；类型放 `*Model.ts` 或 `types/`，禁止组件内联重复 DTO；文案走 `web/src/locales`，不硬编码散落中文；按钮显隐用 `v-permission`（权限码来自 `/user/info`）。
6. **列表筛选/批量用 `s-table`**：筛选落 `searchFields`（`input`/`select`/`tree-select`），批量用选择框 + `#batch` 插槽 + `selection-change`；高频交互模式见 `.agents/skills/frontend-dev/references/high-frequency-tasks.md`。

### 验收命令

```bash
cd web
npm run lint          # eslint，--max-warnings 0（全仓存在既有失败，保证本次触碰文件干净）
npm run lint:fix      # 自动修复
npm run stylelint     # 样式
npm run build         # vue-tsc --noEmit + vite build（类型 + 产物）

WEB=1 ./script/check.sh   # 仓库根：Go 测试/体检 + 前端构建，交付前必须全绿
```

登录链路冒烟：错误密码 → 统一 Toast；正确登录 → 主布局；过期 401 → 自动刷新重放，刷新失败回登录页。常见排错见[常见问题](/docs/faq/)。
