-- 0018: 系统监控 · 系统接口（内嵌 Swagger）
-- spec: docs/compose/spec/api-swagger-page.md

BEGIN;

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4704, 47, 'system:apimonitor:list', 'C', '系统接口', 'api', 4, 'file-code', 'enabled', 'monitor/api/index', 'show', 'system:apimonitor:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4704)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 47);

INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (4704)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
