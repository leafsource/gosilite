-- 0015: 系统监控 · 在线用户
--   顶级菜单「系统监控」id=47（parent=0，排系统管理下方）
--   + 子菜单「在线用户」id=4701 + F 按钮 470101 强退 + system/admin 角色绑定。
--   无数据表：在线会话存 Redis（gosilite:auth:online_sessions）。
-- 幂等：重复执行不产生重复行。
-- spec: docs/compose/spec/monitor-online.md

BEGIN;

-- ========== 菜单：系统监控（顶级，系统管理 id=3 下方） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 47, 0, 'page:root:monitor', 'M', '系统监控', '/monitor', 3, 'precise-monitor', 'enabled', 'layout', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 47);

-- ========== C 在线用户 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4701, 47, 'system:online:list', 'C', '在线用户', 'online', 1, 'user-safety', 'enabled', 'monitor/online/index', 'show', 'system:online:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4701)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 47);

-- ========== F 强退 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 470101, 4701, 'system:online:force', 'F', '强退', '', 1, 'system:online:force', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 470101)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4701);

-- ========== 管理员角色绑定（仿 0011） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (47), (4701), (470101)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
