---
feature: system-oss
status: in-progress
updated: 2026-10-11
branch: develop
commits: 
---

# 对象存储（OSS配置 / 文件管理 / OSS处理规则）

## Report

## [S1] Problem

既有「存储管理」（`/system/storage`，迁移 0020）只有 4 个接口 + 单页：配置写死在 `conf.yaml` 不可运行时变更，对象列表只扫桶、无文件台账（文件名/原名/大小/类型/上传人/锁定全无记录），也没有任何图片处理规则能力。源码项目 ruoyi-tdesign 的对象存储模块字段更全、设计更完整（DB 多配置 + 默认切换、文件台账、URL 处理规则），按用户要求移植其三件套：

1. **OSS配置** — 系统文件上传、下载服务配置（多条配置并存、设默认、运行时生效）。
2. **文件管理** — 系统文件上传、下载等管理（台账 CRUD + 上传/下载/锁定，吸收既有桶浏览能力）。
3. **OSS处理规则** — 例如图片压缩配置（URL 拼接规则，按域名+mime 匹配）。

## [S2] Design

> **Workspace override**：沿用本仓已交付 spec 惯例——在当前 `develop` 检出直接开发，不开 `.worktrees/`。
>
> **源码基线**：`/Users/leafsource/tmp-project/ruoyi-tdesign`（旧代 OSS 三件套：`SysOssConfigServiceImpl` / `SysOssServiceImpl` / `SysOssRuleServiceImpl` + `OssFactory` + 菜单 1500/118/1521）。源码 1.4.0 后另有新代 `sys_file` 体系，**不在本次范围**。
>
> **Grill 已定决策**：①3 菜单 + 吸收既有存储管理（不保留独立菜单，不做「我的文件/OSS分类」）；②DB 多配置 + S3 兼容客户端（不抄阿里/七牛等非 S3 SDK，字段保留源码全集）；③规则语义照抄源码 = URL 层拼接（不做服务端真实压缩）。

### S2.1 数据表（迁移 `script/migrations/0021_oss.sql`，幂等）

三表均去租户/部门字段（本仓无多租户），审计字段用本仓实际惯例 **`created_at/updated_at` + `del_flag`**（0 软删 2 删除；源码的 update_by/create_time 形制不引入），`system_oss` 额外 `create_by varchar(64)` 记上传人；`CREATE TABLE IF NOT EXISTS` + 菜单 `WHERE NOT EXISTS`。

**`system_oss_config` 对象存储配置**（照抄 `sys_oss_config` 字段全集）：

| 列 | 类型 | 说明 |
|---|---|---|
| id | bigserial PK | |
| config_key | varchar(64) NOT NULL UNIQUE | 配置 key（如 local/minio/aliyun-s3） |
| access_key / secret_key | varchar(255) | 明文存库（与 conf.yaml 一致），接口仅 `system:ossConfig:*` 可读 |
| bucket_name | varchar(255) NOT NULL | |
| prefix | varchar(255) DEFAULT '' | 对象 key 前缀 |
| endpoint | varchar(255) NOT NULL | 访问站点 |
| domain | varchar(255) DEFAULT '' | 自定义域名（拼 URL 用，空则用 endpoint） |
| is_https | char(1) DEFAULT 'N' | Y/N |
| region | varchar(64) DEFAULT '' | |
| access_policy | char(1) DEFAULT '1' | 0=private 1=public 2=custom（**2026-10-11 修订：local 种子为 '0'**——RustFS 桶实际私有，标 public 会导致接口返回的 plain URL 直接 AccessDenied；private 时列表/详情走预签名） |
| status | char(1) DEFAULT '0' | 是否默认：1=默认（全表至多一条 1），0=否 |
| ext1 / remark | varchar | 扩展/备注 |
| 审计 4 字段 | | |

**`system_oss` 文件台账**（照抄 `sys_oss`，去掉 `tenant_id/create_dept/oss_category_id/user_type/ext1`——不做分类与「我的文件」；**保留 `is_lock` 与 `service`**）：

| 列 | 类型 | 说明 |
|---|---|---|
| id | bigserial PK | |
| file_name | varchar(255) NOT NULL | 对象 key（存储路径） |
| original_name | varchar(255) NOT NULL | 上传时原始文件名 |
| file_suffix | varchar(32) NOT NULL | 小写后缀（含点或不含点与现 storage 一致，实现时统一为不带点小写） |
| url | varchar(500) NOT NULL | 上传完成时按 domain/is_https 拼的访问 URL |
| size | bigint | 字节 |
| content_type | varchar(128) | |
| is_lock | boolean NOT NULL DEFAULT false | 锁定；锁定文件禁止删除（源码语义） |
| service | varchar(64) NOT NULL | 上传时使用的 config_key；下载/删除按此找客户端 |
| 审计 4 字段 | | create_by=上传人 |

索引：`(create_by, create_time)`、`(original_name)`。

**`system_oss_rule` 处理规则**（照抄 `sys_oss_rule`）：

| 列 | 类型 | 说明 |
|---|---|---|
| id | bigserial PK | |
| rule_name | varchar(255) NOT NULL | 规则名（如 80x80，输出字段 `url_80x80`） |
| domain | varchar(255) NOT NULL | 匹配域名（URL contains 即命中） |
| mime_type | varchar(255) NOT NULL | 逗号分隔；命中 = contentType 匹配 或 文件后缀匹配 |
| rule | text NOT NULL | 规则串，`#{#var}` 占位：domain/path/fullPath/filename/url/fullUrl/query |
| is_overwrite | char(1) NOT NULL DEFAULT 'N' | Y=覆盖主 url 字段；N=生成变体字段 |
| is_default | char(1) NOT NULL DEFAULT 'N' | 不指定规则时使用的默认规则 |
| status | char(1) NOT NULL DEFAULT '1' | 启用 1 / 停用 0 |
| rule_sort | int NOT NULL DEFAULT 0 | 顺序 |
| remark | varchar(500) | |
| 审计 4 字段 | | |

唯一约束：`(domain, rule_name)`（源码 checkRepeat）。

**种子**：
- `system_oss_config`：插入 1 条 `config_key='local'`、**`status='0'`（模板、非默认）**，endpoint/bucket/密钥与本地 dev conf 一致（`http://127.0.0.1:19000` / `skyx` / `gosilite`·`gosilite-secret`，开箱可用；2026-10-11 由空密钥修订），`access_policy='0'`（RustFS 桶实际私有）。不插默认行 ⇒ 运行时回退 conf.yaml，存量环境升级零破坏。
- `system_oss_rule`（2026-10-11 修订：预置常用规则，推翻原「不插种子」）：**阿里 `aliyuncs.com` / 腾讯 `myqcloud.com` / 华为 `myhuaweicloud.com` 各 2 条**（180x180 缩略、800x800 展示，用宽域名 contains 匹配、全区域通用），加 **本地 `127.0.0.1` 示例 1 条（status='0' 停用）**——原生 S3/RustFS 不处理图片参数、带未知 query 的 GET 会被拒绝（AccessDenied），本地条目仅作调试模板，对接 CDN/图片网关后可启用。全部 `is_overwrite='N'`：只产出变体字段，绝不改写主 URL（主 URL 是 private 桶预签名，拼 query 会破坏签名）。

### S2.2 菜单与权限（同迁移）

菜单 id 取 **60 段**（已核对迁移内 40–48/51 被占、60–63 空闲；实现时再以 `system_menu` 实表核对一次）：

```text
3 系统管理
└─ 60 对象存储 (M, path='storage', icon='cloud')          ← 复用 48 释放出的 path
   ├─ 61 OSS配置    (C, path='ossConfig', component='system/ossConfig/index', perm system:ossConfig:list)
   ├─ 62 文件管理    (C, path='oss',      component='system/oss/index',      perm system:oss:list)
   └─ 63 OSS处理规则 (C, path='ossRule',   component='system/ossRule/index',  perm system:ossRule:list)
```

F 按钮：`6101-6104` config add/edit/remove/test；`6201-6205` oss upload/download/edit/remove/lock；`6301-6304` rule add/edit/remove/refresh。权限码：`system:ossConfig:{list,add,edit,remove,test}`、`system:oss:{list,upload,download,edit,remove,lock}`、`system:ossRule:{list,add,edit,remove,refresh}`。（修订：原 6206 桶对象管理权限随桶视图移除）

同迁移：**删除旧菜单 48/4801/4802 及其角色绑定**（吸收存储管理），再插入上述新菜单 + `system/admin` 角色绑定（照 0011/0020 写法）。

前端路由名推导：`/system/storage/ossConfig` → `SystemStorageOssConfig`，页面 `defineOptions({ name })` 必须一致。

### S2.3 动态存储客户端（吸收配置运行时）

- 新包 `application/infra/storage/dynamic`：
  - `Controller` 持 `resolver` + 缓存的当前 `storage.Storage` 客户端 + `sync.RWMutex`。
  - `Current() storage.Storage`：无缓存时调 resolver 构建；resolver 返回 `nil`/错误 → **回退 conf 构造的客户端**（保底，头像上传等既有功能永不断）。
  - `Invalidate()`：清缓存，下次 `Current()` 重建。
  - `Controller` 自身实现 `storage.Storage`（委托 `Current()`），实现失效对消费方透明。
- resolver（在 `application/infra/init.go` 装配）：gorm 查 `system_oss_config WHERE status='1' AND del_flag…` 取唯一默认行 → `s3.New(ctx, ak, sk, bucket, endpoint, region)`（domain/is_https 只影响 URL 拼接，不进 s3.New；前缀由上层 key 生成时拼）。无默认行 → 返回 nil 走 conf 回退。
- 装配：`AppDependencies.Storage` 字段类型保持 `storage.Storage`，赋值改为 `dynamic.Controller`；`AppDependencies` 增加 `StorageCtl *dynamic.Controller` 供配置服务调 `Invalidate()`。既有消费方（avatar、monitor 探活等）零改动。
- 注意 `s3.New` 内含 `CheckAndCreateBucket`，重建时会确保桶存在（与现状一致，可接受）。

### S2.4 OSS配置 服务

`application/framework/service/ossconfig`（具体 `Service` + `NewService` + `InitService(deps)`，仓储 `dal/repository/oss_config_repository.go` 接口 + 裸 `*gorm.DB`，照 `config_repository.go` 样板）。

IDL `idl/framework/system_oss_config.thrift`（hz 生成路由，禁手写 Register）：

| 方法 | 路由 | 权限 | 行为 |
|---|---|---|---|
| List | GET `/system/oss/config/list` | list | 分页 + configKey/endpoint keyword |
| Get | GET `/system/oss/config/{id}` | list | 详情（含完整 secretKey） |
| Create | POST `/system/oss/config` | add | config_key 唯一校验；写库后 `Invalidate()` |
| Update | PUT `/system/oss/config` | edit | 同上；不允许改 config_key（源码允许？——**不允许**，台账 service 字段引用它） |
| Remove | DELETE `/system/oss/config/{ids}` | remove | **拒绝删除 status='1' 的默认配置**（替代源码 SYSTEM_DATA_IDS 内置保护）；删后 `Invalidate()` |
| ChangeStatus | PUT `/system/oss/config/status` | edit | 事务：全部置 0 → 目标置 1 → `Invalidate()`（设默认即时生效） |
| Test | POST `/system/oss/config/test` | test | 用请求体配置临时 `s3.New` + `ListObjectsPaginated(limit=1)` 探活，返回 `{connected, message}`；不写库 |

Swagger 注解按 AGENTS §三 全量补齐；错误码 errno 段在 7010 段内现查后申请（实现时以 `types/errno` 实际为准）。

### S2.5 文件管理 服务

`application/framework/service/ossfile` + 仓储 `oss_file_repository.go`。IDL `idl/framework/system_oss.thrift`。

**台账接口**（列表类挂 list，动作类按下表）：

| 方法 | 路由 | 权限 | 行为 |
|---|---|---|---|
| List | GET `/system/oss/list` | list | 分页；筛选 originalName/fileSuffix/service/createBy/时间范围；返回行经 `matchingURL`：默认配置 `access_policy=0(private)` 时生成**1h 预签名 URL**（源码 120s 页面停留几分钟图即挂，2026-10-11 放宽），否则原 url；另附 `ruleUrls map[string]string`（见 S2.6）。**已知组合约束**：overwrite 规则会向已签名 URL 拼接 query 破坏签名（AccessDenied），private 桶勿配 overwrite 规则 |
| Upload | POST `/system/oss/upload` | upload | multipart 字段 `file`，≤20MB；key = `<默认配置prefix><uploads/yyyyMMdd/uuid.ext>`（沿用现 storage key 生成规则）；写桶（`Current()`）+ 插台账（service=默认配置 config_key）+ 返回 VO（含 ruleUrls） |
| Get | GET `/system/oss/{id}` | list | 详情 + matchingURL |
| Download | GET `/system/oss/download/{id}` | download | 按台账 `service` 找客户端（配置缺失→业务错「存储配置不存在」）；attachment 头 + 流式 `GetObject`，`Content-Disposition` 用 original_name |
| Update | PUT `/system/oss` | edit | 可改 `original_name`、`is_lock`（源码 updateByBo 语义，无分类字段） |
| Remove | DELETE `/system/oss/{ids}` | remove | 任一 `is_lock=true` → 拒绝（「加锁文件必须解锁后才能删除」）；先删台账，成功后按各行 `service` 删桶对象（删对象失败记日志不回滚台账，与源码 removeRealOss 一致） |


`is_lock` 解锁动作走 Update（改 is_lock），故 F 权限码 `system:oss:lock` 挂在 Update 路由的权限里 = edit OR lock？——**实现取简**：Update 挂 `system:oss:edit`，`lock` 按钮在前端用 `system:oss:edit` 控制，迁移里 6205 按钮权限码仍写 `system:oss:lock` 供后续细分（记录此简化）。

### S2.6 OSS处理规则 服务 + URL 消费

`application/framework/service/ossrule` + 仓储。IDL `idl/framework/system_oss_rule.thrift`。

接口：List（分页+筛选）/ Get / Create / Update / Remove / `PUT /system/oss/rule/overwrite`（设覆盖）/ `DELETE /system/oss/rule/cache`（刷新缓存）。语义照抄源码：

- `checkRepeat`：同 domain+rule_name 拒绝。
- Update 时从 N→Y 改默认 → 强制 `is_overwrite=N`（源码 updateByBo）。
- `changeOverwrite`：目标为 Y 且该规则是默认 → 同 domain 其他默认规则 `is_overwrite` 置 N。
- 任何变更/刷新 → 清进程内规则缓存（`sync.RWMutex` 持 `[]Rule` 快照；**不引入 Redis**，单进程刷新即生效，刷新按钮保留源码交互）。

**消费**（核心，`ResolveURLs(originalURL, contentType, fileName, ruleNames, join, useDefault) map[string]string`）：

1. 取启用规则快照（`status=1`，按 rule_sort,create_time 升序）。
2. 按源码 `SysOssRuleServiceImpl.getUrls` 过滤：`originalURL contains rule.domain` 且（contentType 含任一 mime_type 分量 **或** 文件名后缀匹配）；无指定 ruleNames 或 useDefault → 仅取 `is_default=Y`。
3. 命中集合中 `is_overwrite=Y` 的第一条 → 渲染其 rule 替换主 url；其余 `is_overwrite=N` → 生成 `join+ruleName` 变体键（join 默认 `_`，即 `url_80x80`）。
4. 模板渲染：解析 URL 得 domain/path/fullPath/filename/url/fullUrl/query 七变量，`#{#var}` 简单占位替换（**不引 SpringEL**，只支持字面变量；不支持嵌套表达式——记录为与源码的已知差异）。
5. 多 URL（逗号分隔含 http）拆分逻辑照抄 `splitUrl`。

**接线**：文件管理 List/Get/Upload 返回 VO 时调 `ResolveURLs`，结果放 `ruleUrls` 字段（`map[ruleName]processedURL`，overwrite 命中则直接改写 VO 的 `url`）。不实现源码 `@OssRule` 注解序列化层——Gosilite 无此机制，以 VO 字段替代（已知差异）。

### S2.7 前端（web/，写码前必须 `tdesign_mcp.py get-component-docs` 核对 t-table/t-upload/t-drawer 等）

| 页面 | 文件 | 结构 |
|---|---|---|
| OSS配置 | `pages/system/ossConfig/index.vue`（name `SystemStorageOssConfig`） | 查询栏（configKey/endpoint）+ `s-table`（configKey、endpoint、bucket、accessPolicy tag、是否默认 tag、remark、更新时间；操作：编辑/设默认/测试/删除）+ 新增编辑 **t-drawer**（字段 >5：configKey、endpoint、domain、accessKey、secretKey、bucket、prefix、region、isHttps、accessPolicy、remark）+ 测试对话框（t-dialog，临时表单 → test 接口） |
| 文件管理 | `pages/system/oss/index.vue`（name `SystemStorageOss`） | **单页台账（修订：移除概览卡与 t-tabs 桶对象 tab）**：筛选（原名/后缀/上传人/时间）+ s-table（originalName、fileSuffix、size、contentType、url 缩略预览+复制、service、isLock tag、createBy、createTime；操作：下载/锁定解锁/删除强确认）+ 上传 t-dialog（t-upload，走 `/system/oss/upload`，action 需组件内拼 `VITE_API_URL_PREFIX`）。存储概览迁至数据监控（见 S2.10） |
| OSS处理规则 | `pages/system/ossRule/index.vue`（name `SystemStorageOssRule`） | 查询栏（ruleName/domain/isDefault/status）+ s-table（ruleName、domain、mimeType、rule 源码截断、覆盖 switch、默认 switch、状态 switch、ruleSort、操作：编辑/删除）+ 新增编辑 **t-drawer**（8 字段，带源码 help 文案：「例如：80x80，则字段名称将输出字段名_80x80」等）+ 工具栏「刷新缓存」按钮 |

API 层：`web/src/api/system/ossConfig.ts` / `oss.ts` / `ossRule.ts` + `model/*Model.ts`，信封 `{code,msg,data}` 对齐，`code===200` 判定。

**删除**：`pages/system/storage/index.vue`、`api/system/storage.ts`（被吸收）。`s-upload-image` 组件不动（action 为 prop；个人中心头像走 `/user/update/upload_avatar` 不受影响）。

### S2.8 旧存储管理拆除

- 后端：`idl/framework/system_storage.thrift`、`api/handler/framework/storage/`、`api/router/framework/storage/`（Register+middleware）、`application/framework/service/storage/`（含 `service_test.go`）全部删除；`api/router/register.go` 同步去引用；`framework/module.go` 去 `storage.InitService` 改接新服务。
- `application/framework/service/storage` 里可复用的逻辑（key 生成、10MB→20MB 上限、探活聚合、分页前缀过滤）**迁移**到 `ossfile`，对应测试场景在 `ossfile/service_test.go` 重建覆盖（分页/筛选/锁定拦截/上传 key/删除）。
- 迁移 0021 删菜单 48/4801/4802（S2.2）。旧迁移文件 0020 不回改（历史记录）。

### S2.9 测试与验收边界

- 单测（`go test`）：`ossconfig`（唯一性/默认切换/拒删默认）、`ossrule`（重复校验/默认切换关 overwrite/changeOverwrite 联动/ResolveURLs 的域名+mime+default+overwrite+模板/缓存刷新）、`ossfile`（上传 key+台账/锁定拒删/分页筛选/matchingURL private 分支）、`dynamic`（回退 conf/Invalidate 重建）。
- 集成验收：`./script/check.sh` + `WEB=1 ./script/check.sh` 全绿；curl 登录态跑 config CRUD→设默认→test、rule CRUD→resolve 出变体、oss 上传→列表→下载→锁定→拒删→解锁→删除全链路；浏览器（IAB）三页冒烟 + console 0 error。

### S2.10 存储概览（迁至数据监控，2026-10-11 修订）

- **动因（用户反馈）**：文件管理页不应携带概览卡；概览应为**多配置**（DB 多配置并存后单卡无意义），归属数据监控。
- **后端**：`system_monitor.thrift` 新增 `StorageOverviewResponse StorageOverview()(api.get="/system/monitor/storage")`；monitor service `StorageOverview()` 查 `system_oss_config` 全部行，逐条 `s3.New + List(limit=1)` 探活，返回 `{configKey,endpoint,bucketName,domain,isDefault,connected,message}`；失败单条 `connected=false` 不中断。权限 `system:monitor:list`。
- **前端**：`OverviewPanel.vue` 新增「存储（对象存储配置）」section，多配置卡片（默认 tag、连通/异常、endpoint·bucket、失败 message），与运行时/数据源 section 并列；`load` 并行拉取（403 静默降级为空列表）。
- **移除**：文件管理页概览卡与 `t-tabs`；后端 `BucketInfo/BucketList/BucketDelete` 三接口（IDL/handler/middleware/service）；前端 `api/system/oss.ts` 桶接口；权限按钮 6206。

### S2.11 统一文件服务 common/file（2026-10-11 修订 · 用户要求收口）

**动因**：头像直写桶不入台账、文件管理自建上传/下载 handler——每加一个附件场景就要重写 multipart handler。统一收口后业务侧只剩「拿到 id/url 之后干什么」。路由采 `/common/file/*`（三层结构与本仓 `/system/oss/*` 同构，给 `/common` 域留扩展位）。

**接口**（新 IDL `idl/framework/common_file.thrift`，独立 Register，`rootMw = permission.Default()` 登录即可）：

| 方法 | 路由 | 权限 | 行为 |
|---|---|---|---|
| Upload | POST `/common/file/upload` | 登录 | multipart 字段 `file` **唯一参数**（不收自定义 key，防路径混乱），≤20MB；key=`uploads/<yyyyMMdd>/<uuid>.<ext>`；写桶 + **自动写 system_oss 台账**（create_by=登录用户）；返回 `{id, url, key, originalName, size}`。业务实现直接复用 `ossfile.Service.Upload`（零新业务逻辑，纯路由收口） |
| Download | GET `/common/file/download` | 登录 | 按台账 id 流式下载（复用 `ossfile.Service.Download`），attachment 头 |

**头像两步链路**（防客户端塞任意 url）：
1. `s-upload-image` 上传至 `/common/file/upload` → `{id, ...}`；
2. `PUT /user/profile/avatar {fileId}`（user.thrift 新增 `UpdateProfileAvatar`）：服务端按 id 查台账取 **fileName（key）** 写入 `user.avatar_url`（保持相对 uri 语义，`GetProfile` 现有 GetResourceURL 转换零改动）。

**删除接口与权限退役**：
- `POST /system/oss/upload`、`GET /system/oss/download`（IDL/handler/middleware 函数同步删）；
- `POST /user/update/upload_avatar`（`UpdateAvatar` 下线，imageOSS 直写路径废弃）；
- 迁移 0021：F 按钮 6201（upload）/6202（download）删除，`system:oss:upload/download` 权限码退役（库同步 DELETE）；文件管理页上传/下载受 common 登录态控制，列表/编辑/删除仍为管理权限 `system:oss:list/edit/remove`。

**前端契约**：
- `s-upload-image`：action 默认改 `/common/file/upload`（仍自动拼 `VITE_API_URL_PREFIX` + Bearer），响应契约改信封 `{code:200, data:{id,url}}`，`success` emit `{id, url, response}`；
- 个人中心：success → `PUT /user/profile/avatar {fileId}` → 刷新 getUserInfo；
- 文件管理页：上传 dialog action 与下载 URL 改指 common；`api/system/oss.ts` 删 `uploadOssFile/ossFileDownloadUrl`，新增 `api/common/file.ts`。

**token 预刷新（2026-10-11 修 401）**：t-upload 原生 XHR 绕过 axios 的 401→refresh→重试拦截器，access token 过期时上传直接 401（页面其他请求却静默刷新正常）。修复：user store 新增 `ensureFreshAccessToken()`（解析 JWT exp，距过期 <60s 才走单飞刷新）；`s-upload-image`/`s-upload-file` 的 trigger 与文件管理页 `openUpload()` 均先 await 它再触发。

**预览组件 `s-image-preview`（2026-10-11，用户点名封装特例）**：`web/src/components/s-image-preview/`——基于 `t-image-viewer` 的 **trigger 插槽模式**（trigger 即自定义缩略图，不用 `t-image`：其无 `preview` 属性），缩略图 + 纯 CSS hover 遮罩（zoom-in 放大镜）+ 点击 `open()` 放大。props：`src/images/index/size/alt`（images 支持图集翻页）。文件管理预览列使用；头像墙/富文本图片等场景可直接复用。文件管理页预览列同步：去掉复制URL、服务商筛选改下拉（options 来自 `/system/oss/config/list`）。

**附件组件 `s-upload-file`（2026-10-11）**：与 s-upload-image 同构的通用附件组件 `web/src/components/s-upload-file/`（组件名 `SUploadFile`）：`v-model` 值 `{id,url,name,size}`；未选中渲染「选择附件」按钮（同样走 ensureFresh + common 契约 + formatResponse 解信封），已选中渲染文件名（下载链接）+ 大小 + 移除（移除只清本地选择，远程文件由业务删除接口决定）。供未来表单附件场景直接复用。

**已知取舍**：头像每次上传生成新 uuid key，旧文件残留（量级可忽略，不做覆盖 key 的参数魔法）；common upload/download 为 login-only（台账记录上传人可审计，20MB 上限；越权细节后续可加本人校验）。

## [S3] Out of Scope

- 源码新代 `sys_file`/`sys_storage_config` 体系（1.4.0 的 storageConfig/myFile/file）、`ruoyi-common-storage` 多厂商 adapter。
- 「我的文件 / OSS分类」（ossCategory、my/list、move、userType、分类树）。
- 非 S3 厂商 SDK（阿里 OSS 原生/七牛/腾讯 COS 原生协议）——S3 兼容端全覆盖。
- 源码 `@OssRule` Jackson 序列化注解层、SpringEL 模板、Redis 二级缓存、Sa-Token/租户/部门字段、规则导出、`listByIds`/`listByUrls`（无富文本消费方）。
- ~~头像上传写台账（保持直写桶）~~（2026-10-11 修订推翻：头像改走 common/file/upload 自动入台账，见 S2.11）。
- secretKey 加密存储/脱敏回显、对象版本控制、断点续传（`sys_file_part`）。

## Tasks

- [ ] T1: 迁移 0021_oss.sql — acceptance: 幂等执行通过，三表+菜单 60-63+按钮+角色绑定存在，旧菜单 48/4801/4802 已删，种子 config 模板 status=0 (covers: S2.1, S2.2)
- [ ] T2: IDL 三文件 + hz update — acceptance: `api/model`、`api/router` 生成物出现，`./script/check.sh` 编译绿 (covers: S2.4, S2.5, S2.6; depends: T1)
- [ ] T3: dynamic 客户端包 + init.go 装配 — acceptance: 单测证明无默认配置回退 conf、Invalidate 后重建；既有 avatar 路径不回归 (covers: S2.3; depends: T1)
- [ ] T4: ossconfig 仓储+服务+Handler+swagger — acceptance: curl 完成 CRUD/唯一校验/设默认即时切换/Test 接口返回连通结果；单测覆盖拒删默认 (covers: S2.4; depends: T2, T3)
- [ ] T5: ossrule 仓储+服务+Handler — acceptance: curl CRUD/changeOverwrite 联动/刷新缓存生效；ResolveURLs 单测含域名+mime+default+overwrite+`#{#url}` 渲染 (covers: S2.6; depends: T2)
- [ ] T6: ossfile 仓储+服务+Handler（台账+matchingURL+ruleUrls）— acceptance: curl 上传→列表→下载→锁定拒删→解锁删除全链路；service_test 覆盖锁定拦截/key 生成/URL 装饰等场景（修订：桶视图三接口随 S2.10 移除，等价能力由 monitor 存储概览+台账承担）(covers: S2.5, S2.6; depends: T2, T3, T5)
- [ ] T7: 拆除旧 storage（IDL/handler/router/service/web/api/注册引用）— acceptance: 全仓 grep 无 `system/storage` 运行时引用（迁移历史除外）；`./script/check.sh` 绿 (covers: S2.8; depends: T4, T6)
- [ ] T8: 前端三页 + api + 删除旧页 — acceptance: `WEB=1 ./script/check.sh` 绿；defineOptions name 与菜单推导一致；浏览器三页冒烟 console 0 error (covers: S2.7; depends: T4, T5, T6, T7)
- [ ] T9: 验收 — acceptance: `./script/check.sh` 与 `WEB=1` 全绿；curl 全链路记录；IAB 浏览器三页+上传下载实测 (covers: S2.9; depends: T8)
