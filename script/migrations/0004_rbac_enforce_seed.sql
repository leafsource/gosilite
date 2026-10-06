-- ============================================================================
-- 0004: RBAC 落地种子（spec: docs/compose/spec/system-rbac-enforce.md S2.2/S2.4）
--   1) 字典/参数菜单与遗留组件修正
--   2) 按钮级 perms 行（user/menu/dict/config）
--   3) zc1001 绑定普通角色 system_default
--   4) 普通角色菜单子集（仅列表，不授予按钮——避免经 user:add/role:assign 提权，
--      同时用于验收 v-permission 按钮隐藏）
-- 幂等：重复执行不产生重复行
-- ============================================================================

BEGIN;

-- ========== 1. 字典/参数菜单 ==========

-- 字典管理：dump 遗留 component=result/500/index，修正为真实页面并补 perms
UPDATE public.system_menu
SET component = 'system/dict/index',
    perms     = 'system:dict:list'
WHERE id = 9
  AND (component IS DISTINCT FROM 'system/dict/index' OR perms IS DISTINCT FROM 'system:dict:list');

-- 用户管理：dump 遗留 component=framework/user/index，修正为真实页面
UPDATE public.system_menu
SET component = 'system/user/index'
WHERE id = 12 AND component IS DISTINCT FROM 'system/user/index';

-- 0002 的回填 bug：0002 先把 permission_code(page:*) 拷进 perms，再按 perms='' 补真码时
-- 条件已不成立，导致 id=12/10 的 perms 停留在 page:* —— 强制修正
UPDATE public.system_menu SET perms = 'system:user:list'
WHERE id = 12 AND perms IS DISTINCT FROM 'system:user:list';
UPDATE public.system_menu SET perms = 'system:menu:list'
WHERE id = 10 AND perms IS DISTINCT FROM 'system:menu:list';
-- 根目录行不留 perms（避免 page:* 垃圾权限码进入 GetInfo）
UPDATE public.system_menu SET perms = ''
WHERE id IN (1, 2, 3) AND perms IS DISTINCT FROM '';

-- 参数管理菜单（父级：系统管理 id=3；当前库中完全缺失）
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 43, 3, 'system:config:list', 'C', '参数管理', 'config', 8, 'setting', 'enabled', 'system/config/index', 'show', 'system:config:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 43);

-- ========== 2. 按钮权限行（F） ==========

-- 用户管理按钮
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1201, 12, 'system:user:add', 'F', '用户新增', '', 1, 'system:user:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1201);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1202, 12, 'system:user:edit', 'F', '用户修改', '', 2, 'system:user:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1202);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1203, 12, 'system:user:remove', 'F', '用户删除', '', 3, 'system:user:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1203);

-- 菜单管理按钮
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1001, 10, 'system:menu:add', 'F', '菜单新增', '', 1, 'system:menu:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1001);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1002, 10, 'system:menu:edit', 'F', '菜单修改', '', 2, 'system:menu:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1002);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 1003, 10, 'system:menu:remove', 'F', '菜单删除', '', 3, 'system:menu:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 1003);

-- 字典管理按钮（类型与数据共用 dict 域）
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 9001, 9, 'system:dict:add', 'F', '字典新增', '', 1, 'system:dict:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 9001);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 9002, 9, 'system:dict:edit', 'F', '字典修改', '', 2, 'system:dict:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 9002);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 9003, 9, 'system:dict:remove', 'F', '字典删除', '', 3, 'system:dict:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 9003);

-- 参数管理按钮（菜单行存在才挂）
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4301, 43, 'system:config:add', 'F', '参数新增', '', 1, 'system:config:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4301)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 43);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4302, 43, 'system:config:edit', 'F', '参数修改', '', 2, 'system:config:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4302)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 43);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4303, 43, 'system:config:remove', 'F', '参数删除', '', 3, 'system:config:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4303)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 43);

-- ========== 3. zc1001 绑定普通角色 ==========

INSERT INTO public.system_user_role (id, user_id, role_id, type, is_remove)
SELECT 99000001, u.id, r.id, 'built-n', 0
FROM public.system_user u
JOIN public.system_role r ON r.code = 'system_default'
WHERE u.account = 'zc1001'
  AND NOT EXISTS (
    SELECT 1 FROM public.system_user_role ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id AND ur.is_remove = 0
  );

-- ========== 4. 普通角色菜单子集（仪表盘 + 系统管理(用户/岗位/角色)，仅列表） ==========
-- 菜单 id：1 仪表盘目录、2 概览、3 系统管理目录、12 用户、41 岗位、42 角色
-- 不含按钮行：zc1001 无新增/编辑/删除/分配权限（验收按钮隐藏 + 防提权）

INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (1), (2), (3), (12), (41), (42)) AS m(id)
WHERE r.code = 'system_default'
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
