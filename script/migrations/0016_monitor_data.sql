-- 0016: 系统监控 · 数据监控（Tab 页）
--   子菜单 4702 C 数据监控 component=monitor/data/index
--   权限 system:monitor:list（六个只读 API 共用）
-- 幂等：重复执行不产生重复行。
-- spec: docs/compose/spec/data-monitor.md

BEGIN;

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4702, 47, 'system:monitor:list', 'C', '数据监控', 'data', 2, 'chart-line', 'enabled', 'monitor/data/index', 'show', 'system:monitor:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4702)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 47);

INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (4702)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
