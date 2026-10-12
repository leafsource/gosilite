-- 0023: 系统管理 · 消息管理（复刻 ruoyi-message）
--   四表 system_message_{config,key,template,log}；
--   字典 sys_message_type / sys_message_template_mode（挂“系统”分组）；
--   菜单「消息管理」id=50（系统管理 id=3 下，sort=12）+ C 5001-5004 + F 按钮 + system/admin 绑定。
--   status 语义对齐 ruoyi：1 正常 0 停用；is_success：1 成功 0 失败。
-- 幂等：重复执行不产生重复行/列；全表 public. 限定。
-- spec: docs/compose/spec/system-message.md S2.2

BEGIN;

-- ========== 消息配置表 ==========
CREATE TABLE IF NOT EXISTS public.system_message_config (
    id            BIGSERIAL PRIMARY KEY,
    title         VARCHAR(255) NOT NULL,
    message_type  VARCHAR(20)  NOT NULL,              -- MAIL / SMS
    supplier_type VARCHAR(20)  NOT NULL,              -- 供应商标识（MAIL / BAIDU…）
    config_json   TEXT,                               -- 通道凭据/参数 JSON
    status        SMALLINT     NOT NULL DEFAULT 1,    -- 1 正常 0 停用
    remark        VARCHAR(500) NOT NULL DEFAULT '',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_message_config IS '消息配置表';
CREATE INDEX IF NOT EXISTS idx_sys_message_config_type ON public.system_message_config(message_type, status);

-- ========== 消息常量表 ==========
CREATE TABLE IF NOT EXISTS public.system_message_key (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    message_key VARCHAR(50) NOT NULL,
    remark     VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_message_key IS '消息常量表';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_message_key ON public.system_message_key(message_key);

-- ========== 消息模板表 ==========
CREATE TABLE IF NOT EXISTS public.system_message_template (
    id                BIGSERIAL PRIMARY KEY,
    template_name     VARCHAR(255) NOT NULL,
    message_config_id BIGINT       NOT NULL,
    message_key_id    BIGINT       NOT NULL,
    message_key       VARCHAR(50)  NOT NULL,
    message_type      VARCHAR(20)  NOT NULL,          -- MAIL / SMS
    template_mode     VARCHAR(20)  NOT NULL,          -- TEMPLATE_ID / TEMPLATE_CONTENT
    title             VARCHAR(255) NOT NULL DEFAULT '',
    signature         VARCHAR(100) NOT NULL DEFAULT '',
    template_id       VARCHAR(100) NOT NULL DEFAULT '',
    content           TEXT,
    vars_json         TEXT,                           -- 输入变量定义
    status            SMALLINT     NOT NULL DEFAULT 1,
    remark            VARCHAR(500) NOT NULL DEFAULT '',
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP    NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_message_template IS '消息模板表';
CREATE INDEX IF NOT EXISTS idx_sys_message_template_key_id ON public.system_message_template(message_key_id);
CREATE INDEX IF NOT EXISTS idx_sys_message_template_key ON public.system_message_template(message_key, message_type, status);

-- ========== 消息发送记录表 ==========
CREATE TABLE IF NOT EXISTS public.system_message_log (
    id                   BIGSERIAL PRIMARY KEY,
    message_template_id  BIGINT,
    message_key          VARCHAR(50)  NOT NULL,
    message_template_name VARCHAR(255) NOT NULL DEFAULT '',
    message_type         VARCHAR(20)  NOT NULL,
    template_mode        VARCHAR(20)  NOT NULL DEFAULT '',
    account              VARCHAR(255) NOT NULL,
    title                VARCHAR(255) NOT NULL DEFAULT '',
    template_id          VARCHAR(100) NOT NULL DEFAULT '',
    content              TEXT,
    message_config_title VARCHAR(255) NOT NULL DEFAULT '',
    supplier_type        VARCHAR(20)  NOT NULL DEFAULT '',
    is_success           SMALLINT,                     -- 1 成功 0 失败
    response_body        VARCHAR(1000),
    log_time             TIMESTAMP    NOT NULL,
    cost_time            BIGINT                        -- 毫秒
);
COMMENT ON TABLE public.system_message_log IS '消息发送记录表';
CREATE INDEX IF NOT EXISTS idx_sys_message_log_template_id ON public.system_message_log(message_template_id);
CREATE INDEX IF NOT EXISTS idx_sys_message_log_time ON public.system_message_log(log_time DESC);

-- ========== 字典：消息类型 / 模板模式（挂“系统”分组，仿 0011） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '消息类型', 'sys_message_type', 20, '消息通道类型（MAIL 邮件 SMS 短信）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_message_type' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_message_type', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('邮件', 'MAIL', 1, 1),
    ('短信', 'SMS', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_message_type');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '消息模板模式', 'sys_message_template_mode', 21, '模板ID模式 / 模板内容模式',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_message_template_mode' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_message_template_mode', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('模板ID', 'TEMPLATE_ID', 1, 0),
    ('模板内容', 'TEMPLATE_CONTENT', 2, 1)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_message_template_mode');

-- ========== 菜单：消息管理（M，系统管理 id=3 下，sort 取现有子菜单最大值+1=12） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 50, 3, 'page:root:message', 'M', '消息管理', 'message', 12, 'chat', 'enabled', 'layout', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 50);

-- ========== C 四页 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 5001, 50, 'system:messageConfig:list', 'C', '消息配置', 'config', 1, 'tools', 'enabled', 'system/message/config/index', 'show', 'system:messageConfig:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 5001)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 50);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 5002, 50, 'system:messageKey:list', 'C', '消息常量', 'key', 2, 'root-list', 'enabled', 'system/message/key/index', 'show', 'system:messageKey:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 5002)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 50);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 5003, 50, 'system:messageTemplate:list', 'C', '消息模板', 'template', 3, 'file', 'enabled', 'system/message/template/index', 'show', 'system:messageTemplate:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 5003)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 50);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 5004, 50, 'system:messageLog:list', 'C', '发送记录', 'log', 4, 'history', 'enabled', 'system/message/log/index', 'show', 'system:messageLog:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 5004)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 50);

-- ========== F 按钮（query/add/edit/remove；记录页 query/clear） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT v.id, v.parent, v.code, 'F', v.title, '', v.sort, v.code, 'enabled', 'hide', 0
FROM (VALUES
    (500101::bigint, 5001::bigint, 'system:messageConfig:query'::text, '配置查询'::text, 1::int),
    (500102::bigint, 5001::bigint, 'system:messageConfig:add'::text,    '配置新增'::text, 2::int),
    (500103::bigint, 5001::bigint, 'system:messageConfig:edit'::text,   '配置修改'::text, 3::int),
    (500104::bigint, 5001::bigint, 'system:messageConfig:remove'::text, '配置删除'::text, 4::int),
    (500201::bigint, 5002::bigint, 'system:messageKey:query'::text, '常量查询'::text, 1::int),
    (500202::bigint, 5002::bigint, 'system:messageKey:add'::text,    '常量新增'::text, 2::int),
    (500203::bigint, 5002::bigint, 'system:messageKey:edit'::text,   '常量修改'::text, 3::int),
    (500204::bigint, 5002::bigint, 'system:messageKey:remove'::text, '常量删除'::text, 4::int),
    (500301::bigint, 5003::bigint, 'system:messageTemplate:query'::text, '模板查询'::text, 1::int),
    (500302::bigint, 5003::bigint, 'system:messageTemplate:add'::text,    '模板新增'::text, 2::int),
    (500303::bigint, 5003::bigint, 'system:messageTemplate:edit'::text,   '模板修改'::text, 3::int),
    (500304::bigint, 5003::bigint, 'system:messageTemplate:remove'::text, '模板删除'::text, 4::int),
    (500401::bigint, 5004::bigint, 'system:messageLog:query'::text, '记录查询'::text, 1::int),
    (500402::bigint, 5004::bigint, 'system:messageLog:clear'::text, '记录清空'::text, 2::int),
    (500403::bigint, 5004::bigint, 'system:messageLog:remove'::text, '记录删除'::text, 3::int)
) AS v(id, parent, code, title, sort)
WHERE EXISTS (SELECT 1 FROM public.system_menu m WHERE m.id = v.parent)
  AND NOT EXISTS (SELECT 1 FROM public.system_menu m2 WHERE m2.id = v.id);

-- ========== system/admin 角色绑定（仿 0015/0022） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES
    (50), (5001), (5002), (5003), (5004),
    (500101), (500102), (500103), (500104),
    (500201), (500202), (500203), (500204),
    (500301), (500302), (500303), (500304),
    (500401), (500402), (500403)
) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND EXISTS (SELECT 1 FROM public.system_menu sm WHERE sm.id = m.id)
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
