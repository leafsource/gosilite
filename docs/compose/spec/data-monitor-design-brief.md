# 数据监控页 · 设计需求说明（交给设计师）

> 交付：桌面端中后台页面效果图（建议 Figma / Sketch / 高保真 PNG，1440 宽为基准）。
> 可参考系统已有「在线用户」列表页与 TDesign Starter 仪表盘的密度与语言；最终以本说明的字段与结构为准。

---

## 0. 产品背景（一句话）

企业级 Admin 的 **系统监控 → 数据监控** 单页多 Tab：用「卡片总览 + 列表」呈现数据库/进程/SQL/API 运行状态，**只读、无写操作**，强调信息密度与可扫读性，不是营销大屏。

## 1. 页面骨架

| 项 | 约定 |
|----|------|
| 菜单路径 | 系统监控 → 数据监控（与「在线用户」平级） |
| 容器 | 内容区白底卡片（content-card），顶部 **Tab 导航** |
| Tab 顺序 | **首页 → SQL 监控 → SQL 防火墙 → URI 监控**（共 4 个，无 Hertz 独立 Tab） |
| 交互 | 切 Tab 懒加载；每 Tab 可有「刷新」；无分页（列表一次拉全，服务端已限顶） |
| 语言 | 中文 |
| 断点 | 桌面优先（1440 / 1920）；1366 不崩即可，非强制移动 |

**请你输出：**  
1）整页框架图（Tab 切换态可叠 1 张示意）  
2）4 个 Tab 各 1 张完整高保真  
3）关键空态 / 异常态（可选：Redis 不可用、无慢 SQL）

---

## 2. Tab 1 · 首页（总览）

**目标：** 10 秒内看清「进程是否健康 + 数据库是否吃紧」。

### 2.1 分区 A · 运行时（Web 应用 / Go 进程）

| 卡片 | 主数值 | 辅助文案 / 二行 | 建议语义色 |
|------|--------|-----------------|------------|
| 运行时长 | 如 `3天 4时`（由秒格式化） | 进程启动至今 | 品牌蓝 |
| Goroutine | 如 `20` | `GOMAXPROCS {n}` | 绿 |
| 堆内存 | 如 `15 MB`（可读单位） | `sys 24 MB · idle 8 MB` | 橙 |
| GC 次数 | 如 `12` | 累计暂停 `{n} ms` | 紫 / 中性 |

可选补充（若版式有空位，非必须）：Go 版本 `go1.x.x`、`CPU 核数`、`PID`。

### 2.2 分区 B · 数据源

| 卡片 | 主数值 | 辅助文案 / 二行 | 建议语义色 |
|------|--------|-----------------|------------|
| 最大连接 | `100` | MaxOpenConnections | 品牌蓝 |
| 当前连接 | `1` | 使用中 `0` · 空闲 `1` | 绿 |
| 等待次数 | `0` | 累计等待 `{n} ms` | 0=中性，>0 橙/红 |
| 连接回收 | `{idle+lifetime 合计}` | idle 回收 n · lifetime 回收 n | 紫 / 中性 |

### 2.3 连接池条（分区 B 内整行）

| 字段 | 说明 |
|------|------|
| 标题 | 连接池占用 |
| 右侧数字 | `{pgInUse} / {pgMaxOpen}` |
| 进度条 | 百分比 = inUse/maxOpen；**≥70% 警告色，≥90% 错误色，否则品牌/成功色** |
| 底部状态条 | tag：`PG Open {n} · Idle {n}`；`Redis 正常 · {ms}ms` 或 `Redis 不可用`（红）；右侧灰字：Go 版本 · CPU · PID · `检查于 yyyy-MM-dd HH:mm:ss` |

### 2.4 首页设计提示

- 卡片建议统一：标题 + 主数 + 一行说明；图标/色条作层次，**不要做成大数字霓虹大屏**。
- 两区间用小节标题（运行时 / 数据源）+ 说明文字。
- 空数据：`—`，不要 0 全屏红。

---

## 3. Tab 2 · SQL 监控

**目标：** 哪类 SQL 跑得多、谁最慢。

### 3.1 页头

| 元素 | 文案/数据 |
|------|-----------|
| 标题 | SQL 模板聚合 |
| 说明 | 慢 SQL 阈值 **{slowThresholdMs} ms** · 内存有界统计，不落库 |
| 溢出标记 | 若 `limited=true` → 黄 tag「模板数已达上限」 |
| 操作 | 刷新 |

### 3.2 左表 · 执行统计（按 count 降序，最多约 100 行）

| 列 | 字段 | 说明 |
|----|------|------|
| SQL 模板 | `sql` | 参数已归一化为 `?`，等宽/可截断悬浮 |
| 次数 | `count` | |
| 平均(ms) | `avgMs` | |
| 最大(ms) | `maxMs` | 可对 max 高亮 |
| 总耗时(ms) | `totalMs` | |
| 错误 | `errorCount` | 0 灰，>0 红 |

### 3.3 右表 · 慢 SQL（近 50 条，新→旧）

| 列 | 字段 | 说明 |
|----|------|------|
| SQL | `sql` | |
| 耗时(ms) | `costMs` | ≥阈值；建议色阶（如 ≥500 红） |
| 结果 | `err` | 成功绿 / 失败红 |
| 时间 | `at` | Unix 秒 → `YYYY-MM-DD HH:mm:ss` |

### 3.4 版式建议

左右分栏（或上下，视宽度）；空态：「暂无 SQL 统计」「暂无慢 SQL」。

---

## 4. Tab 3 · SQL 防火墙

**目标：** 规则是否开着、命中了多少、最近命中了什么。**观察模式：不拦截业务**（文案需体现，避免误解为 WAF 已阻断）。

### 4.1 摘要区

| 卡片/标签 | 字段 | 说明 |
|-----------|------|------|
| 状态 | `enabled` | true→绿「已启用」；false→灰「已关闭」 |
| 命中次数 | `hitCount` | 环形缓冲内计数，溢出后为下限 |
| 规则 | `rules[]` | 固定：`multi_statement` / `dangerous_pattern` / `union_injection`（可展示中文别名+英文 code） |

页头说明建议：「观察模式：命中只计数并保留样本，不拦截业务查询 · 有界环形缓冲」

### 4.2 最近命中表（约 50 条环）

| 列 | 字段 | 说明 |
|----|------|------|
| 规则 | `rule` | 可 badge |
| SQL | `sql` | 归一化模板 |
| 时间 | `at` | Unix 秒 |

空态：「暂无命中记录」。

---

## 5. Tab 4 · URI 监控

**目标：** 哪些接口量大、哪些慢、错误多。path = **路由模板**（如 `/user/info`，无 query）。

### 5.1 摘要区（3 枚）

| 卡片 | 计算 | 说明 |
|------|------|------|
| 路由数 | `total` | 列表行数 |
| 累计请求 | Σ `count` | |
| 错误请求 | Σ `errorCount` | >0 用错误色底/描边 |

说明文案：按路由模板聚合 · 数据来自 Hertz 请求量/时延打点

### 5.2 路由明细表

| 列 | 字段 | 说明 |
|----|------|------|
| 方法 | `method` | GET/POST badge |
| 路径 | `path` | 等宽或弱等宽 |
| 请求量 | `count` | 可做小条形/仅排序 |
| 错误 | `errorCount` | 0 灰 tag，>0 红 tag |
| 平均耗时 | `avgMs` | 建议色阶：&lt;200 正常、200–499 橙、≥500 红 |

可选（若你愿意扩展，需后端补数）：`maxMs`、错误率百分比、时间趋势图。  
**当前后端未提供**：按时间序列的曲线；请先以**汇总表+KPI** 为主，勿设计依赖不存在字段的图表。

### 5.3 不要设计的 Tab

- ~~Hertz 监控~~（method×status 独立页）——已砍，信息并入 URI 概念。
- ~~在线用户~~——另一菜单，不在本页。

---

## 6. 全局视觉与组件约束

| 项 | 要求 |
|----|------|
| 组件语言 | 对齐 **TDesign Vue Next** 中后台（表格、Tag、Progress、Tabs、Card）；可自定义但需可落地 |
| 色板 | 中性底 + 单一品牌主色 + 语义色（成功/警告/错误）；避免彩虹卡 |
| 字体 | 系统 UI；数字 `tabular-nums` |
| 密度 | 中高密度（运维监控），一屏尽量多信息，忌大留白海报风 |
| 表格 | 斑马纹可选；列宽稳定；SQL 列必须可截断+tooltip |
| 交互 | 仅刷新与 Tab；无编辑、无批量、无导出（除非你主动提案并标注） |
| 空态/加载 | 每 Tab 有空态文案；骨架或 loading 可出示意 |

---

## 7. 数据字段速查（JSON 真名，供标注）

```
首页-运行时 ServerDTO:
  goVersion, goMaxProcs, goroutine, uptimeSec, heapAlloc, heapSys,
  heapIdle, heapInuse, sysBytes, gcNum, gcPauseTotalMs, numCPU, pid

首页-数据源 DatasourceDTO:
  pgMaxOpen, pgOpen, pgInUse, pgIdle, pgWaitCount, pgWaitDurationMs,
  pgMaxIdleClosed, pgMaxLifetimeClosed, redisUp, redisLatencyMs, checkedAt

SQL 监控:
  list[]: sql, count, avgMs, maxMs, totalMs, errorCount, lastMs
  slowList[]: sql, costMs, at, err
  slowThresholdMs, limited

防火墙:
  enabled, hitCount, rules[], recent[]{rule, sql, at}

URI:
  total
  list[]{path, method, count, errorCount, avgMs, maxMs}
```

---

## 8. 请回传给开发的内容

1. **4 张 Tab 高保真**（含 Tab 选中态）+ 1 张整页说明  
2. **标注**：间距、色值、字号、组件名（或「自定义 card」）  
3. **可选**：空态 / Redis down / 无慢 SQL  
4. 若有动效偏好（Tab 切换、数字滚动）请单独一页说明——实现时可裁剪  

开发侧收到效果图后按图实现 `web/src/pages/monitor/data/**`，字段名保持上表 JSON，不改接口契约（需加字段会另开 IDL）。
