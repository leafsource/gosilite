---
feature: user-center-storage
status: delivered
updated: 2026-10-10
branch: develop
commits: f9d98f4..5ee65c5
---

# 个人中心重做 · s-upload-image 通用组件 · 存储管理

## Report

**What was built** — 个人中心从 starter mock 页整页重做为真实资料页：左侧资料卡（`SUploadImage` 圆形头像 96px + hover 上传、用户名/账号/角色标签、改密入口复用 `ChangePasswordDialog`），右侧 `t-descriptions` 展示 `GET /user/profile` 真实字段（账号/用户名/邮箱/手机/`sys_state` 字典状态/创建时间/角色）；store `UserInfo` 增 `avatarURL`，上传成功后与 Header 头像同步。新增通用组件 `s-upload-image`（t-upload 封装：统一信封解析取 `data.webURI`、自动拼 `VITE_API_URL_PREFIX`、自带 `Bearer` token、大小校验、失败 toast 还原服务端 msg），头像与存储上传测试两处复用。存储管理落地为系统管理下完整对象管理：后端 4 接口（info 探活聚合 / list 分页+prefix+keyword / delete / upload ≤10MB，IDL→hz 生成、`system:storage:*` 权限、errno 701000032/33）、service 直连 infra `storage.Storage`、迁移 0020 菜单 48+4801/4802+角色绑定，前端概览卡（endpoint/bucket/连通/对象数/总大小）+ `s-table` 对象列表 + 删除强确认 + 上传测试对话框。顺带修复两处遗留：`/user/update/upload_avatar` 原挂 `page:root:page AND system:update` 任何账号恒 403（改 `permission.Default()`）；`GetProfile` handler 原未接 service（改返回资料全字段 VO）。

**Verification** — `./script/check.sh` 与 `WEB=1 ./script/check.sh` 全绿（go test/vet/build + vue-tsc/vite build，含新增 `storage/service_test.go` 9 用例）；curl 实测（登录态）：头像上传→webURI→GET 200 下载 PNG、`/user/profile` 全字段、`/system/storage/*` 探活/分页/前缀/关键词/上传/删除全闭环，并经 vite `/api` 代理复测上传（验证 t-upload 实际路径）；迁移 0019/0020 已在本地库执行，`/menu/routes` 下发 `/system/storage`；浏览器（IAB）实测个人中心与存储管理页真实数据渲染、上传对话框、console 0 error。独立评审发现 1 critical（t-upload 绕过 axios 无 `/api` 前缀）+ 2 major，修复后复审全部 PASS。

**Journey log** — ①t-upload 原生 XHR 不经 axios 的 `urlPrefix`，相对路径必须组件内拼 `VITE_API_URL_PREFIX`（否则打到 Vite 返回 index.html）——评审抓出，curl 直连后端测不出；②`_updateavatarMw` 遗留 `Roles("system:update")`（非角色码）导致头像上传从未可用，属既有 403 bug；③两并行代理在分页字段上分歧（snake vs camel），以 AGENTS「JSON camelCase」+ 现有 `PageQuery` 惯例裁决统一；④`GetProfile` 的 IDL 声明与实现早已不符（返回类型对不上），本 feature 顺带补齐；⑤IAB 后端拦不到 file chooser（三次超时），浏览器「点选文件」交互未走通，上传链路改由 API 层 + `/api` 代理路径全量覆盖，后续可用 Computer Use 手动补验。

## [S1] Problem

1. **个人中心是 starter 演示页**：手机号/邮箱/团队成员/图表全是 mock；`t-avatar` 只显示字母；后端 `POST /user/update/upload_avatar`（multipart → S3 `user_avatar/<uid>.<ext>` → 更新 `avatar_url` → 返回 `webURI`）与 `GET /user/info`（含 `avatarURL`）已就绪，但前端无封装——store 的 `UserInfo` 只有 `name/roles/permissions`，`avatarURL` 被丢弃，Header 头像位只有 icon。
2. **无通用图片上传组件**：`t-upload` 仅在 starter 演示页出现过一次；头像上传、存储上传测试都要各自拼 FormData。
3. **存储（RustFS/S3）无任何可视化**：作为系统级基础设施（头像、图片都落这里），连通性、用量、对象增删查全靠命令行；而 `infra/storage` 原语（`ListObjectsPaginated` / `HeadObject` / `DeleteObject` / `PutObject` / `GetObjectUrl`）与依赖注入（`AppDependencies.Storage`）均已就绪，菜单/权限/s-table 样板（0015 迁移、monitor 页）可直接复用。

## [S2] Design

> **Workspace override**：沿用本仓已交付 spec 惯例——在当前 `develop` 检出直接开发，不开 `.worktrees/`（真实主线为本地 `develop`）。

### S2.1 s-upload-image 通用组件

- 位置：`web/src/components/s-upload-image/index.vue`，组件 `SUploadImage`（用户点名要求通用组件，虽当前仅 2 处使用，特例记录于此）。
- 写码前必须 `scripts/tdesign_mcp.py call get-component-docs '{"framework":"vue-next","names":["upload"]}'` 核对 `t-upload` API，禁止凭记忆。
- 契约（以 MCP 文档微调实现为准）：
  - props：`action`（必填，上传端点，相对路径由组件拼 `VITE_API_URL_PREFIX`——t-upload 原生请求不经 axios）、`field`（multipart 字段名，默认 `file`；头像传 `avatar`）、`url`（受控当前图，配 `v-model:url`）、`accept`（默认 `image/*`）、`size`（预览尺寸，默认 `80px`；个人中心用 `96px`）、`disabled`、`maxSizeMb`（前端预校验，默认 5）
  - 事件：`update:url`、`success({ url, response })`、`error(err)`
  - 行为：选图 → 校验类型/大小（失败 toast，不发请求）→ `POST action`，FormData 字段名 `field` → 解析统一信封 `{code,msg,data}`，`code===200` 取 `data.webURI` 更新 `v-model:url` 并 emit `success`；否则 toast `msg` emit `error`
  - 组件只认 `webURI` 契约，不关心业务语义；上传中显示 loading，禁用重复提交

### S2.2 个人中心整页重做

- 删除全部 starter mock：`constants.ts` 的假 `userInfoList`/`teamMembers`、tabs 假图表、product 卡、假 greeting。
- 布局：左**资料卡**（`SUploadImage` 圆形头像 96px + hover 遮罩「更换头像」、用户名、账号、角色 `t-tag`）＋右**个人信息** `t-descriptions`（真实字段：账号/用户名/邮箱/手机/状态/创建时间，来自 `GET /user/profile`）＋**修改密码**入口（`t-dialog`，3 字段 ≤5 → dialog，走已有 `changePassword`）。
- `web/src/api/user.ts` 增 `uploadAvatar(file)`：POST `FormData`，字段名 `avatar`，端点 `/user/update/upload_avatar`。
- store：`UserInfo` 增 `avatarURL: string`（初始 `''`）；`getUserInfo()` 填充 `res.user.avatarURL`；上传成功后调 `store.getUserInfo()` 刷新。
- Header：`avatarURL` 非空时 `t-avatar` 显示真实头像，空则回退现有 `user-circle` icon。
- 文案补 `web/src/locales`（zh/en）。

### S2.3 存储管理（系统管理下新增）

**后端**

- IDL：新文件 `idl/framework/system_storage.thrift`（`namespace go framework.storage`），hz 生成路由/model，禁手写 Register：
  - `GetInfo()(api.get="/system/storage/info")` → `{endpoint, bucket, domain, connected, objectCount, totalSize, message}`
  - `GetObjects(req)(api.get="/system/storage/list")`，req `{page, pageSize, prefix, keyword}` → `{list:[{key, size, etag, lastModified, type}], total}`（分页 + 前缀 + key 模糊）
  - `DeleteObject(req)(api.delete="/system/storage/object")`，req `{key}` → 成功信封
  - `Upload()(api.post="/system/storage/upload")` multipart 字段 `file`（≤10MB，类型不限）→ `{webURI, key}`
- service：`application/framework/service/storage`，`type Service struct { store storage.Storage }` + `NewService`/`InitService(deps.Storage)`，装配在 `framework/module.go`（照 `monitor.InitService` 样例），路由注册照 `frameworkMonitorRouter.Register` 样例。
  - `Info`：`ListObjectsPaginated` 探活并遍历聚合 `objectCount`/`totalSize`（失败 `connected=false` + `message`，HTTP 200 业务信封）
  - `List`：分页聚合后按 `prefix` 截前缀、`keyword` 过滤 key；`Delete`：`DeleteObject`；`Upload`：随机 key（`uploads/<yyyyMMdd>/<uuid>.<ext>`）`PutObject` + `GetObjectUrl`
- 权限：`system:storage:list`（info/list）、`system:storage:delete`、`system:storage:upload`，挂 hz 生成的 `middleware.go`（照 monitor 的 `HPermissions` 样例）。
- 错误码：复用 `types/errno` Upload 段（104xxxxxx）与 System 段顺延（`701000032` 起），注册进 `code.Register`。
- 迁移 `script/migrations/0020_storage.sql`（幂等、`public.` 限定、照 0015 样板）：菜单 `48`（C，parent=3 系统管理，component `system/storage/index`，perms `system:storage:list`）＋F 按钮 `4801` delete、`4802` upload ＋ `system_role_menu` 绑定 system/admin 角色。

**前端**

- `web/src/api/system/storage.ts` ＋ `storageModel.ts`：`getStorageInfo` / `getObjectList` / `deleteStorageObject` / `uploadStorageObject`。
- `web/src/pages/system/storage/index.vue`：顶部概览卡（endpoint / bucket / domain / 连通 `t-tag` / 对象数 / 总大小（格式化）/ 刷新）＋ `s-table` 对象列表（key、大小、修改时间、操作列删除 `v-permission="'system:storage:delete'"`；搜索区 prefix + keyword「查询后生效」）＋「上传测试」按钮开 `t-dialog`，内嵌 `SUploadImage`（`action=/system/storage/upload` `field=file`），成功后刷新列表。
- 菜单由迁移下发，不写静态路由。

### S2.4 验收

- `WEB=1 ./script/check.sh` 全绿；`hz update` 生成物无手改。
- 上传实测（后端已起、容器在跑）：登录 → 个人中心换头像 → 返回 `webURI` 可 GET 到图、`system_user.avatar_url` 更新、页面与 Header 同步刷新；存储管理页上传测试对象出现在列表且可下载、删除后消失。
- 存储管理：侧栏「系统管理 → 存储管理」出现；概览显示真实 bucket 数据；列表分页/筛选生效。

## [S3] Out of Scope

图片裁剪与缩略图生成；目录管理（移动/复制/重命名）；多 bucket 切换；presigned 前端直传；imagex 图片处理接入；个人资料字段编辑（后端无 update profile 接口）；对象存储统计数据的缓存优化。

## Tasks

- [x] T1: `s-upload-image` 通用组件（MCP 查 t-upload docs 后实现，含类型/大小预校验与 envelope 解析）— acceptance: 组件可被 import，`cd web && npm run build` 绿 (covers: S2.1)
- [x] T2: store/api 基建——`UserInfo.avatarURL`、`uploadAvatar` 封装、Header 真实头像 — acceptance: `getUserInfo()` 后 store 含 `avatarURL`，build 绿 (covers: S2.2; depends: T1)
- [x] T3: 个人中心整页重做（资料卡 + SUploadImage 上传 + 真实 descriptions + 改密 dialog + locales）— acceptance: 页面无 mock；上传成功后页面与 Header 头像同步刷新 (covers: S2.2; depends: T2)
- [x] T4: 存储管理后端——IDL + hz update + errno — acceptance: 生成物有 storage 路由/model，无手写 Register，build 绿 (covers: S2.3; depends: T1)
- [x] T5: storage service/handler/装配（framework module Init + RegisterRoutes + middleware 权限）— acceptance: `./script/check.sh` 绿；4 个接口带登录态 curl 通 (covers: S2.3; depends: T4)
- [x] T6: 迁移 0020（菜单 48 + F 按钮 + 角色绑定）— acceptance: 库中菜单/绑定存在，侧栏出现「存储管理」 (covers: S2.3; depends: T5)
- [x] T7: 前端存储管理页（api + 概览卡 + s-table 列表 + 上传测试 dialog）— acceptance: 概览/列表/上传/删除可用，build 绿 (covers: S2.3; depends: T5)
- [x] T8: 上传实测 — acceptance: 头像与存储测试对象真实落 S3，`webURI` 可访问，列表/删除闭环 (covers: S2.4; depends: T3, T6, T7)
