-- 0017: 系统监控 · 缓存监控
--   C 4703 缓存监控 + F 删除/清理；权限组 system:cachemonitor:*
-- spec: docs/compose/spec/cache-monitor.md

BEGIN;

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4703, 47, 'system:cachemonitor:list', 'C', '缓存监控', 'cache', 3, 'database', 'enabled', 'monitor/cache/index', 'show', 'system:cachemonitor:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4703)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 47);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, perms, visible, status)
SELECT 470301, 4703, 'system:cachemonitor:delete', 'F', '删除', '', 1, 'delete', 'enabled', 'system:cachemonitor:delete', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 470301)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4703);

INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (4703), (470301)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
