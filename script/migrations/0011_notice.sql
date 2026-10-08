-- 0011: 通知公告（notice）
--   新增 system_notice 表；字典 sys_notice_type / sys_notice_status（挂“系统”分组）；
--   菜单「通知公告」id=45 + F 按钮 4501/4502/4503 + system/admin 角色绑定。
-- 幂等：重复执行不产生重复行/列；全表 public. 限定。
-- spec: docs/compose/spec/notice.md S2.1

BEGIN;

-- ========== 通知公告表 ==========
CREATE TABLE IF NOT EXISTS public.system_notice (
    id          BIGSERIAL PRIMARY KEY,
    title       VARCHAR(128) NOT NULL,
    notice_type SMALLINT NOT NULL DEFAULT 1,  -- sys_notice_type: 1 通知 2 公告
    content     TEXT NOT NULL DEFAULT '',
    status      SMALLINT NOT NULL DEFAULT 0,  -- sys_notice_status: 0 已发布 1 已撤回
    publisher   VARCHAR(64) NOT NULL DEFAULT '',
    del_flag    SMALLINT NOT NULL DEFAULT 0,  -- 0 正常 2 删除
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_notice IS '通知公告';
CREATE INDEX IF NOT EXISTS idx_system_notice_type ON public.system_notice(notice_type) WHERE del_flag = 0;
CREATE INDEX IF NOT EXISTS idx_system_notice_status ON public.system_notice(status) WHERE del_flag = 0;

-- ========== 字典：类型 / 状态（挂“系统”分组，仿 0010） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '通知公告类型', 'sys_notice_type', 8, '通知公告分类（1 通知 2 公告）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_notice_type' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_notice_type', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('通知', '1', 1, 1),
    ('公告', '2', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_notice_type');

INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '通知公告状态', 'sys_notice_status', 9, '发布状态（0 已发布 1 已撤回）',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_notice_status' AND del_flag = 0);

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_notice_status', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('已发布', '0', 1, 1),
    ('已撤回', '1', 2, 0)
) AS v(label, value, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_notice_status');

-- ========== 菜单：通知公告（父级系统管理 id=3；44 已被组织架构占用，用 45） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 45, 3, 'system:notice:list', 'C', '通知公告', 'notice', 9, 'notification', 'enabled', 'system/notice/index', 'show', 'system:notice:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 45);

-- ========== F 按钮（菜单行存在才挂） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4501, 45, 'system:notice:add', 'F', '通知新增', '', 1, 'system:notice:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4501)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 45);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4502, 45, 'system:notice:edit', 'F', '通知修改', '', 2, 'system:notice:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4502)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 45);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4503, 45, 'system:notice:remove', 'F', '通知删除', '', 3, 'system:notice:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4503)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 45);

-- ========== 管理员角色绑定（仿 0002；超管本就全量，此为完整性） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (45), (4501), (4502), (4503)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
