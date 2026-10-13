---
feature: notice-center
status: delivered
updated: 2026-10-08
branch: develop
commits: working-tree
---

# 通知中心：铃铛与「我的通知」对接通知公告（消费侧）

## Report

**What was built** — 消费侧闭环：迁移 `0012_notice_read.sql`（`system_notice_read` 唯一索引 + 用户索引，双跑幂等）+ codegen 登记出 model；`idl/framework/system_notice.thrift` 追加 4 个 login-only 端点（`/notice/inbox` 含 keyword 标题过滤、`/notice/detail` 服务端 upsert 自动已读、`/notice/unread-count`、`/notice/read` id 缺省=全部）经 hz 生成（新 mw 保持 nil，仅登录），`INoticeRepository` 扩 InboxPage/UnreadCount/GetPublishedByID/MarkRead(on conflict do nothing)/MarkAllRead(一次 INSERT…SELECT)，service 扩 Inbox/Detail/UnreadCount/Read（userID 取登录态，撤回复用 `ErrSystemNoticeNotFound`）；前端 `api/notice(.ts|Model.ts)`、`notification store` 全面去 mock（unread/recent/markOne/markAll/applyRead）、`Notice.vue` 铃铛接真数据（未读圆点、条目点击=详情自动已读、清空、查看全部→`/notice/list`）、新页 `pages/notice/index.vue`（s-table 只读 + 关键字 + 已读标签 + 详情对话框）+ 静态路由 `modules/notice.ts`（不进侧栏）；starter 遗留 `detail/secondary` 与 `useLocale` 对 mock 的依赖已解耦/移除。**流程**：本 feature 首次严格执行 skill 强制澄清门禁（三项拍板：全员可见 / 铃铛+独立页 / 打开详情自动已读）。

**Verification** — 0012 双跑幂等（索引齐全）；`go test ./application/framework/service/notice/` 13 用例 PASS（新增：Detail 自动已读、撤回 NotFound 不标已读、未登录、Read 缺省=MarkAll、单条已读、单条撤回 NotFound）；curl：unread-count 1→detail→0、inbox read 标志 false→true、read-all 200、detail missing 701000028、匿名 401、keyword=冒烟 total=2；`WEB=1 ./script/check.sh` **check OK**；改动文件 eslint exit 0；浏览器端到端（skyx，IAB）：角标真实 2 → 铃铛下拉 3 条（未读带「设为已读」、已读无按钮）→ 点开「临时调休提醒」详情正文可见且角标 2→1 →「清空」归零 →「查看全部」跳 `/notice/list`（3 行、已读标签、字典类型）→ 关键字「冒烟」筛出 2 行 → 管理端撤回一条后刷新用户侧只剩 2 行（消失生效）→ console error 0 → 冒烟数据清理（inbox 仅存用户自己的「gosilite框架介绍」）。

**Journey log** — ①thriftgo 对 IDL 字段 `read` 生成 Go 字段名 `Read_`（避开 thrift `Read(iprot)` 方法冲突），JSON 标签仍是 `read`，前端无感；②s-table 搜索区恒渲染关键字框，inbox 契约补了 `keyword` 标题过滤以免页面出现假搜索；③hz 对既有 controller 是「缺函数才追加骨架」，骨架引用未导入的 `consts`，替换为真实实现后二次 hz update 不再重复追加；④浏览器首提交必现「账号必填」是 fill 未进 Vue model 的已知怪癖，重填再提交即可；⑤keyword 加入后需重启后端才生效——曾误判为前端问题，curl 复核定位；⑥starter `detail/secondary` 演示页与 `useLocale` 依赖 mock store，已分别本地化 mock 与移除调用，避免类型断裂；⑦**铃铛详情框三合一 bug（用户报障）**：`t-dialog` 默认 `attach` 缺省=原位渲染，Notice.vue 位于 header → dialog+遮罩被塞进 `.t-head-menu`（`overflow:hidden`+`fixed`+`z-index:10`）导致内容被裁（「排版不对」）、遮罩被裁只剩顶条（表格裸露 hover 即「操作区亮」），叠加 popup 未随详情打开关闭（z-index 5500 > dialog 2500，浮层压在表格上）——修复：详情抽公共组件 `s-notice-detail`（第 3 处同构触发，三处回改）并 `attach="body"`、popup 改受控 `v-model:visible` 开详情即关、正文 CSS 变量更正为真实 token（`--td-comp-paddingLR/TB-m`、`--td-bg-color-secondarycontainer`，原名不存在致 `pad:0/bg:transparent`）；复验 elementFromPoint：dialog 中心命中自身、表格区顶层为 dialog 层、popup `getBoundingClientRect` 0×0，console 0 error；⑧同一次复验发现页面侧「查看」首点偶发不弹（detail 请求已发、600ms 未渲染），二次点击稳定复现正常，无 JS 错误——记录为待观察项，暂未复现。

## [S1] Problem

「通知公告」（feature: notice）只交付了**管理侧发布维护**，缺消费侧模型：
布局铃铛 `layouts/components/Notice.vue` 与 Header 未读角标仍消费
`store/modules/notification.ts` 的 mock 数据（固定 4 条、写死未读数），
没有「当前用户的站内通知 / 未读数 / 已读状态」能力。

**需求澄清（Gate，用户已拍板，见对话 2026-10-08）**：

1. 投递范围 = **全员可见**（已发布即全员，只加已读表，发布页不改）；
2. 前端范围 = **铃铛下拉 + 独立「我的通知」页**（「查看全部」跳该页）；
3. 已读行为 = **打开详情自动已读**（服务端在详情接口标记），保留「全部已读」按钮。

## [S2] Design

### S2.1 数据与迁移（`script/migrations/0012_notice_read.sql`，幂等、`public.` 限定）

`public.system_notice_read`：

| 列 | 类型 | 说明 |
|----|------|------|
| id | BIGSERIAL PK | |
| user_id | BIGINT NOT NULL | 读者 |
| notice_id | BIGINT NOT NULL | 关联 `system_notice.id` |
| read_at | TIMESTAMP NOT NULL DEFAULT NOW() | 首次已读时间 |

唯一 `uk_system_notice_read_user_notice(user_id, notice_id)`；索引
`idx_system_notice_read_user(user_id)`。未读 = 已发布数 − 本人已读数（不做未读行）。
撤回（status=1）即从用户侧消失；已读行保留不动。
codegen 登记 `system_notice_read`（`repository:false, service:false`，只出 model/query）。

### S2.2 后端（hz 生成，追加到 `idl/framework/system_notice.thrift` 同一 Register）

用户侧接口**均 login-only**（新 mw 函数保持 `nil`，登录由 `rootMw=permission.Default` 提供，
不占用 `system:notice:*` 管理权限码）：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/notice/inbox` | 分页（page/pageSize/keyword 标题过滤）；返回 `{list:[{id,title,noticeType,publisher,createdAt,read}], total, unread}`，列表**不含 content**，按 id DESC |
| GET | `/notice/detail?id=` | 含 content；**服务端 upsert 已读（on conflict do nothing）**；查 status=0，否则 `ErrSystemNoticeNotFound` |
| GET | `/notice/unread-count` | `{count}`（Header 角标用） |
| POST | `/notice/read` | `{id?}`：有 id 标单条，缺省=全部已读（只对已发布未读行）；返回影响行数 |

- 分层：扩 `INoticeRepository`（InboxPage / UnreadCount / GetPublishedByID /
  MarkRead / MarkAllRead）→ `service/notice`（`Inbox/Detail/UnreadCount/Read`，
  userID 取 `ctxutil.GetLoginUser`）→ handler `api/handler/framework/notice`
  追加函数；IDL 追加 struct 后 `hz update`，**生成的 4 个新 mw 保持 nil（仅登录）**。
- 复用错误码 `ErrSystemNoticeNotFound=701000028`；无新权限码、无菜单种子。

### S2.3 前端

- `web/src/api/notice.ts` + `noticeModel.ts`（用户侧 4 端点，envelope 对齐）。
- `store/modules/notification.ts` **去 mock 化**：state 改 `unread`、`recent`；
  actions：`loadUnread()`（布局挂载/刷新）、`loadRecent()`（铃铛下拉，pageSize=10）、
  `markAll()`、`openDetail()`（拉详情并刷新未读）；删除 `msgData/getMsgData` mock。
- `layouts/components/Notice.vue`：条目接 `recent`（未读圆点 + 标题/时间），
  点击条目 → 详情 `t-dialog`（自动已读，读后本条置已读、unread−1）；
  「全部已读」调 `markAll`；「查看全部」→ `/notice`；空态/加载态保留 locales 文案。
- 新页 `web/src/pages/notice/index.vue`（我的通知）：`s-table` 只读
  （标题/类型 `s-tag-dict`/发布人/时间/已读状态），行点击或「查看」打开同款详情
  对话框；静态路由 `router/modules/notice.ts`（path `/notice`，**不进侧栏**——
  入口 = 铃铛「查看全部」，假设：通知中心不占菜单位，如有偏差请纠正）。
- mock 的 `collected/quality/NotificationItem` 字段废弃；locales 保留可用文案。

### S2.4 测试与验收

- `service/notice` 单测扩：Inbox 未读计算委托、Detail 自动已读（fake repo 记录
  MarkRead 调用）、Read 缺省 id 走 MarkAll、未登录/找不到复用既有断言。
- 迁移 0012 双跑幂等；`WEB=1 ./script/check.sh` 全绿。
- 浏览器冒烟：skyx 登录 → Header 角标显真值 → 铃铛下拉显真实条目 →
  点开详情自动已读、角标 −1 →「全部已读」归零 →「查看全部」进我的通知页 →
  撤回一条（管理页）后用户侧消失 → console 无新增 error。

## [S3] Out of Scope

- 定向投递（部门/角色/用户 receiver 表）——用户拍板全员可见。
- 发布者侧已读回执/阅读统计报表。
- 邮件 / webhook / WebSocket 实时推送；轮询刷新（进入页面/手动刷新即拉取）。
- 收藏（mock `collected` 字段废弃）；通知类型页签筛选（可后补）。
- 定时发布、审批流。

## Tasks

- [x] T1: 0012 迁移（read 表 + 唯一/普通索引）+ codegen 登记生成 model — acceptance: 双跑幂等；`system_notice_read.gen.go` 生成 (covers: S2.1)
- [x] T2: IDL 追加 4 端点 + hz update + repo 扩展 + service + handler + mw 保持 nil — acceptance: curl 登录态 inbox/detail/unread-count/read 200；匿名 401；detail 自动已读落库 (covers: S2.2)
- [x] T3: service 单测 — acceptance: `go test ./application/framework/service/notice/` 绿，覆盖 Detail 自动已读/Read 全部已读 (covers: S2.4)
- [x] T4: 前端 api + notification store 去 mock + Notice.vue + 我的通知页/路由 — acceptance: 角标/下拉真数据；详情自动已读；查看全部跳转 (covers: S2.3)
- [x] T5: 验收 — acceptance: `WEB=1 ./script/check.sh` 绿；浏览器冒烟 S2.4 全过 (covers: S2.4)
