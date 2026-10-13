# Feature Spec: 日志管理（操作日志 / 系统异常 / 登录日志）

- status: delivered
- created: 2026-02-06
- delivered: 2026-02-07

## Problem

系统缺少可查询的操作留痕与异常审计能力：正常操作无记录、系统异常（panic / 业务失败）无统一入口、登录成败只散落在 hlog 文件里。需要在「系统管理」末位新增**日志管理**：

- **操作日志**页（两 tab）：① 操作日志——正常操作记录与查询；② 系统异常信息——异常记录与查询，均可查看详情。
- **登录日志**页：登录记录查询，含登录异常（失败原因）。

参考案例（用户提供截图）：操作日志列表（traceId/操作人/类型/执行方法/状态/耗时/时间 + 清理）、详情（请求 Header/Body + 返回响应）、异常列表（异常信息/级别/处理状态）。

## 澄清拍板（Gate 已过）

| 决策点 | 拍板 |
|--------|------|
| 采集范围 | **C：全量记录（含查询类 GET），但 GET/查询不存请求响应报文**；写请求（POST/PUT/PATCH/DELETE）存报文（截断 16KB） |
| 写入方式 | 有界异步队列 + 批量 flush（详见 Design），业务请求路径零阻塞 |
| 清理能力 | 做：按选中删除 + 清空全部（独立权限按钮）；登录日志只读不提供清理 |
| 登录日志字段 | 用户名 / IP / 时间 / 状态 / 失败原因 / User-Agent + 浏览器·OS 简析 |
| 异常处理状态 | 待处理/已处理 + 「标记已处理」接口（含处理备注） |

## Design

### 数据模型（迁移 0013_log.sql）

三张表（图1/图2 筛选字段差异大，不合并）：

1. `system_oper_log`：trace_id, user_id, user_name, oper_type(dict sys_oper_type), request_uri, http_method, module, oper_status(0成功/1失败), request_headers, request_body, response_body, error_msg, remark(=响应 msg), cost_ms, client_ip, created_at。
2. `system_error_log`：trace_id, user_id, user_name, err_type(1业务异常/2panic), err_level(dict sys_error_level 1普通/2严重), handle_status(dict sys_handle_status 0待处理/1已处理), err_code, err_msg, stack, request_headers, request_body, request_uri, http_method, module, cost_ms, client_ip, handled_at, handle_remark, created_at。
3. `system_login_log`：trace_id, user_name, user_id, status(dict sys_login_status 0成功/1失败), fail_reason, client_ip, user_agent, browser, os, created_at。

字典挂「系统」分组：`sys_oper_type`（0其他/1新增/2修改/3删除/4查询/5导入/6导出/7上传）、`sys_oper_status`（0成功/1失败）、`sys_error_level`、`sys_handle_status`、`sys_login_status`。

菜单：46 `M` 日志管理（parent=3，sort=末位自增）→ 4601 `C` 操作日志（component `system/log/operlog/index`，perms `system:operlog:list`）、4602 `C` 登录日志（component `system/log/loginlog/index`，perms `system:loginlog:list`）；按钮 460101 `system:operlog:clean`（清空）、460102 `system:operlog:handle`（标记已处理）。system/admin 角色绑定。

### 契约（IDL `idl/framework/system_log.thrift`，namespace `framework.log` → hz 生成）

```
GET  /system/operlog/list   traceId,userName,operType,execStatus,method,httpMethod,startTime,endTime,page,pageSize → {list,total}（列表不查大字段）
GET  /system/operlog        id → 详情（含 headers/body/response 全量）
POST /system/operlog/clean  {ids?:list<i64>, all?:bool} → i32 影响行数
GET  /system/errorlog/list  traceId,userName,keyword(异常信息),handleStatus,errLevel,httpMethod,startTime,endTime,page,pageSize
GET  /system/errorlog       id → 详情（含 stack / 报文）
POST /system/errorlog/handle {id, handleRemark} → i32（标记已处理）
POST /system/errorlog/clean  {ids?, all?} → i32
GET  /system/loginlog/list  userName,status,clientIp,startTime,endTime,page,pageSize
```

管理路由挂 `permission`：oper/error 列表详情挂 `system:operlog:list`，clean 挂 `system:operlog:clean`，handle 挂 `system:operlog:handle`，login 列表挂 `system:loginlog:list`。

### 采集链路（性能核心）

```
requestid → operlog 中间件 → recovery → cors → 业务
```

- **operlog 注册在 recovery 之前（外层）**：panic 由内层 recovery 捕获并写 500 响应；recovery handler 将 panic 信息 `c.Set` 到请求上下文，operlog 收尾时读取 → 异常日志（err_type=panic，带 stack）。无需中间件自行 recover。
- **热路径成本**：`time.Now` ×2、traceId 读 header（复用 requestid 生成的 `x-request-id`）、响应信封 bytes 级提取 `"code"`/`"msg"`（不做整包 unmarshal）、写请求才 copy 报文（`min(len,16KB)` 截断）、登录请求与 GET 不存 request body。目标：除报文 copy 外 <50µs/请求。
- **用户信息**：`ctxutil.GetLoginUser(ctx)`（JWT Authorizator 已解析，零额外开销）。
- **脱敏**：header 序列化时 `Authorization` 打码；`/user/login` body 永不落库；body 内 `"password"` 值字节级掩码。
- **排除**：`/ping`、`/swagger`、OPTIONS 预检不记录。

### 异步批量写（防日志量拖垮基建）

- 三条**有界非阻塞**队列：oper cap 8192 / error cap 2048 / login cap 1024；满则丢弃 + atomic 计数 + 限频 hlog 告警——**绝不阻塞业务请求**。
- 每队列一个 worker：按 **batch≥500 条或 200ms 窗口**（先到为准）GORM `CreateInBatches` 批量插入；失败重试 1 次后整批丢弃 + hlog.Error。
- 优雅退出：进程 shutdown 时 flush 剩余队列（5s 超时）。
- 存储保护：详情大字段与列表分离（列表 SELECT 不含 headers/body/stack）；列表按 created_at DESC + trace_id/user_id/created_at/handle_status 索引。
- 清理接口支撑人工保留期管理（自动清理不在本期范围）。

### 前端

- `web/src/pages/system/log/operlog/index.vue`：两 tab（操作日志 / 系统异常信息），顶部筛选对齐参考图（traceId/操作人/操作类型/执行方法/执行状态/日志时间…异常 tab 对应 异常信息/处理状态/异常级别）；耗时彩色徽标；行操作「查看」→ 详情抽屉（头部元信息 + 请求参数 Header/Body 表 + 返回响应 JSON，异常 tab 展示异常信息/堆栈/处理状态）；顶部「清空日志」（二次确认，按当前 tab 清理）、异常 tab「标记已处理」。
- `web/src/pages/system/log/loginlog/index.vue`：列表 + 筛选（用户名/状态/IP/时间）。
- 静态路由：`/system/log/operlog`、`/system/log/loginlog`（system.ts 内嵌 log 分组）。
- API：`web/src/api/system/log.ts`。

## Out of Scope

- 日志自动清理/定时任务、日志导出、登录日志清理。
- OpenTelemetry/Prometheus 接入、hlog 换第三方 logger（Hertz 框架扩展按需后续引入）。
- 左侧筛选面板布局（对齐本项目顶部筛选惯例）。

## Tasks

- [x] T1 spec + 迁移 0013_log.sql（表/字典/菜单/权限）
- [x] T2 IDL + hz 生成
- [x] T3 operlog 中间件 + 有界异步批量写 worker
- [x] T4 Repository/Service/Handler + 单测
- [x] T5 前端两页面 + 菜单路由
- [x] T6 WEB=1 check.sh 全绿 + 浏览器验收

## 交付后调整（用户反馈，已合入）

- 操作日志页 tab 内嵌 s-table 高度链修复：`.t-tab-panel` auto 高把 `.s-table{height:100%}` flex 链压塌（t-table 折 0、分页条错位盖住数据行、"查看"点击被截走）；改为 content-card → t-tabs → content → panel 全链确定高度 flex 列（scoped deep）。
- 登录日志列补全：新增「浏览器」「操作系统」独立列；「失败原因」改为「操作信息」——成功显示「登录成功」，失败显示具体失败原因。
- 详情抽屉标题带 `#id`（对齐参考案例头部）。
- 第二轮反馈（2026-02-07）：tab 内容区整体 `padding: 12px`；tab 改名「系统异常信息」→「错误日志」（抽屉头「错误详情」、清空确认同步）；堆栈 `white-space: pre` + 横向滚动（不折行，便于对照帧）；错误详情堆栈区恒显（业务异常无栈时给排错指引占位）。
- 登出记入登录日志：迁移 0014 给 `sys_login_status` 加 2=登出；中间件识别 `POST /user/logout`（status=2，用户名取账号口径）；`_logoutMw` 挂 `permission.Default()` 使 Authorizator 解析身份（前端 logout 已有失败容错，401 不阻塞本地清理）。
- 真实错误验收：临时 `/debug/panic-demo` 触发真实越界 panic → 500 + 异常表 errType=2、51 行堆栈含崩溃帧 `nums[5]`，浏览器详情确认；验收后临时路由已删除，`WEB=1 ./script/check.sh` 复验全绿。
- 第三轮反馈（2026-02-07）：三个日志页时间范围**默认今天**（`todayRange()`；s-table 新增 `defaultParams`，初始化与重置均恢复默认，防止无时间条件全量扫描拖垮系统）；错误 tab 按钮「标记处理」→「处理」；**查询类请求也记录报文**（推翻原拍板 C 的 GET 不存限制，压缩响应除外），落库 URI 改为原始请求行（path + query，分类仍用纯 path）。

## Report

**验收（2026-02-07）**

- `WEB=1 ./script/check.sh` → check OK（go test/vet/build + vue-tsc + vite build）；新增单测全绿（classify/peekEnvelope/maskSecrets/operType/UA/skip/登录名提取、服务层分页归一化/时间范围/not-found/clean 透传、sink 批量 flush 与有界丢弃不阻塞）。
- API 端到端（curl）：登录失败/成功 → 登录日志两行（成功 userId=51、失败含原因）；查询 → 操作日志成功行；不存在 id → 操作失败行 + 异常行（errCode 701000028）；详情含 Header（Authorization 打码）与报文截断；标记处理 0→1 并写 handled_at/handle_remark；clean ids 影响行数正确；无 token → 401。
- 浏览器（IAB）：侧栏 系统管理→日志管理→操作日志/登录日志；高度修复后表格完整可见（tabs content 128→558px，t-table 0→430px），「查看」详情抽屉正常（标题 #id + 元信息 + Header/Body/响应三段）；异常 tab 详情含错误码/状态/Header、「标记处理」点击生效；登录日志新列渲染正确（登录成功 / 具体失败原因 / 浏览器）。
- 过程记录：IAB 截图管线两次返回无关参考图，验收以 DOM 结构与 computed 布局为准。

**性能要点（按用户要求的极限优化）**

- 请求热路径：2×time.Now + 响应信封字节扫描（不 unmarshal）+ 仅写请求 copy 报文（16KB 截断）；登录/GET 不落 request body；header 序列化 8KB 上限，Authorization/Cookie 打码。
- 有界非阻塞：oper 8192 / error 2048 / login 1024，满即丢 + 计数 + 60s 限频告警，绝不反压业务（单测覆盖满仓不阻塞）。
- 批量写：200ms 窗口或 500 条先到、CreateInBatches(100)、失败重试 1 次整批丢弃；进程退出 OnShutdown 5s 排空。
- 列表查询不 SELECT 大字段（headers/body/stack 详情单查）；created_at DESC + trace/user/status/handle 索引。
