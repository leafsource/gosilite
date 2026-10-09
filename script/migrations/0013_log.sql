-- 0013: 日志管理（operlog / errorlog / loginlog）
--   三张日志表 + 5 个字典（挂“系统”分组）；
--   菜单「日志管理」id=46(M) → 4601 操作日志(C) / 4602 登录日志(C)
--   + F 按钮 460101 清空、460102 标记已处理 + system/admin 角色绑定。
-- 幂等：重复执行不产生重复行/列；全表 public. 限定。
-- spec: docs/compose/spec/operlog.md

BEGIN;

-- ========== 操作日志 ==========
CREATE TABLE IF NOT EXISTS public.system_oper_log (
    id              BIGSERIAL PRIMARY KEY,
    trace_id        VARCHAR(64) NOT NULL DEFAULT '',
    user_id         BIGINT NOT NULL DEFAULT 0,
    user_name       VARCHAR(64) NOT NULL DEFAULT '',
    oper_type       SMALLINT NOT NULL DEFAULT 0,   -- sys_oper_type: 0其他 1新增 2修改 3删除 4查询 5导入 6导出 7上传
    request_uri     VARCHAR(255) NOT NULL DEFAULT '',
    http_method     VARCHAR(16) NOT NULL DEFAULT '',
    module          VARCHAR(64) NOT NULL DEFAULT '',
    oper_status     SMALLINT NOT NULL DEFAULT 0,   -- sys_oper_status: 0 成功 1 失败
    request_headers TEXT NOT NULL DEFAULT '',
    request_body    TEXT NOT NULL DEFAULT '',
    response_body   TEXT NOT NULL DEFAULT '',
    error_msg       TEXT NOT NULL DEFAULT '',
    remark          VARCHAR(512) NOT NULL DEFAULT '',
    cost_ms         INTEGER NOT NULL DEFAULT 0,
    client_ip       VARCHAR(64) NOT NULL DEFAULT '',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_oper_log IS '操作日志';
CREATE INDEX IF NOT EXISTS idx_oper_log_created ON public.system_oper_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_oper_log_trace ON public.system_oper_log(trace_id);
CREATE INDEX IF NOT EXISTS idx_oper_log_user ON public.system_oper_log(user_id);
CREATE INDEX IF NOT EXISTS idx_oper_log_type ON public.system_oper_log(oper_type);

-- ========== 系统异常日志 ==========
CREATE TABLE IF NOT EXISTS public.system_error_log (
    id              BIGSERIAL PRIMARY KEY,
    trace_id        VARCHAR(64) NOT NULL DEFAULT '',
    user_id         BIGINT NOT NULL DEFAULT 0,
    user_name       VARCHAR(64) NOT NULL DEFAULT '',
    err_type        SMALLINT NOT NULL DEFAULT 1,   -- 1 业务异常 2 panic
    err_level       SMALLINT NOT NULL DEFAULT 1,   -- sys_error_level: 1 普通 2 严重
    handle_status   SMALLINT NOT NULL DEFAULT 0,   -- sys_handle_status: 0 待处理 1 已处理
    err_code        INTEGER NOT NULL DEFAULT 0,
    err_msg         TEXT NOT NULL DEFAULT '',
    stack           TEXT NOT NULL DEFAULT '',
    request_uri     VARCHAR(255) NOT NULL DEFAULT '',
    http_method     VARCHAR(16) NOT NULL DEFAULT '',
    module          VARCHAR(64) NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL DEFAULT '',
    request_body    TEXT NOT NULL DEFAULT '',
    cost_ms         INTEGER NOT NULL DEFAULT 0,
    client_ip       VARCHAR(64) NOT NULL DEFAULT '',
    handled_at      TIMESTAMP,
    handle_remark   VARCHAR(512) NOT NULL DEFAULT '',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_error_log IS '系统异常日志';
CREATE INDEX IF NOT EXISTS idx_error_log_created ON public.system_error_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_error_log_trace ON public.system_error_log(trace_id);
CREATE INDEX IF NOT EXISTS idx_error_log_handle ON public.system_error_log(handle_status);
CREATE INDEX IF NOT EXISTS idx_error_log_level ON public.system_error_log(err_level);

-- ========== 登录日志 ==========
CREATE TABLE IF NOT EXISTS public.system_login_log (
    id          BIGSERIAL PRIMARY KEY,
    trace_id    VARCHAR(64) NOT NULL DEFAULT '',
    user_id     BIGINT NOT NULL DEFAULT 0,
    user_name   VARCHAR(64) NOT NULL DEFAULT '',
    status      SMALLINT NOT NULL DEFAULT 0,       -- sys_login_status: 0 成功 1 失败
    fail_reason VARCHAR(255) NOT NULL DEFAULT '',
    client_ip   VARCHAR(64) NOT NULL DEFAULT '',
    user_agent  VARCHAR(512) NOT NULL DEFAULT '',
    browser     VARCHAR(64) NOT NULL DEFAULT '',
    os          VARCHAR(64) NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_login_log IS '登录日志';
CREATE INDEX IF NOT EXISTS idx_login_log_created ON public.system_login_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_log_name ON public.system_login_log(user_name);
CREATE INDEX IF NOT EXISTS idx_login_log_status ON public.system_login_log(status);

-- ========== 字典（挂“系统”分组，仿 0011） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '操作日志类型', 'sys_oper_type', 10, '操作分类（0其他 1新增 2修改 3删除 4查询 5导入 6导出 7上传）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_oper_type' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_oper_type', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('其他', '0', 1, 1),
    ('新增', '1', 2, 0),
    ('修改', '2', 3, 0),
    ('删除', '3', 4, 0),
    ('查询', '4', 5, 0),
    ('导入', '5', 6, 0),
    ('导出', '6', 7, 0),
    ('上传', '7', 8, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_oper_type');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '操作执行状态', 'sys_oper_status', 11, '执行状态（0 成功 1 失败）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_oper_status' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_oper_status', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('成功', '0', 1, 1),
    ('失败', '1', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_oper_status');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '异常级别', 'sys_error_level', 12, '异常严重级别（1 普通 2 严重）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_error_level' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_error_level', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('普通', '1', 1, 1),
    ('严重', '2', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_error_level');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '异常处理状态', 'sys_handle_status', 13, '异常处理状态（0 待处理 1 已处理）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_handle_status' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_handle_status', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('待处理', '0', 1, 1),
    ('已处理', '1', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_handle_status');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '登录状态', 'sys_login_status', 14, '登录结果（0 成功 1 失败）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_login_status' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_login_status', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('成功', '0', 1, 1),
    ('失败', '1', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_login_status');

-- ========== 菜单：日志管理（父级系统管理 id=3；sort 取末位自增） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 46, 3, 'page:root:log', 'M', '日志管理', 'log',
       (SELECT COALESCE(MAX(sort), 0) + 1 FROM public.system_menu WHERE parent_id = 3),
       'log', 'enabled', 'layout', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 46);

-- 4601 操作日志（页内两 tab：操作日志 / 系统异常信息）
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4601, 46, 'system:operlog:list', 'C', '操作日志', 'operlog', 1, 'log', 'enabled', 'system/log/operlog/index', 'show', 'system:operlog:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4601)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 46);

-- 4602 登录日志
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4602, 46, 'system:loginlog:list', 'C', '登录日志', 'loginlog', 2, 'login', 'enabled', 'system/log/loginlog/index', 'show', 'system:loginlog:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4602)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 46);

-- ========== F 按钮（挂 4601） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 460101, 4601, 'system:operlog:clean', 'F', '清空日志', '', 1, 'system:operlog:clean', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 460101)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4601);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 460102, 4601, 'system:operlog:handle', 'F', '异常处理', '', 2, 'system:operlog:handle', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 460102)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4601);

-- ========== 管理员角色绑定（仿 0011） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (46), (4601), (4602), (460101), (460102)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
