-- 0022: 系统工具菜单
--   顶级菜单「系统工具」id=49（parent=0，系统监控 id=47 下方）
--   + 子菜单「系统接口」4704 由 47 改挂 49（保留 id/权限码/角色绑定，component 迁 tool/api/index）
--   + 子菜单「开发文档」4901（iframe 打开 Hertz 托管的 Hugo 文档站 /docs/）
--   + system/admin 角色绑定 49/4901。
-- 幂等：重复执行不产生重复行、不重复改挂。
-- spec: docs/compose/spec/tool-menu-devdocs.md

BEGIN;

-- ========== 菜单：系统工具（顶级，系统监控下方） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 49, 0, 'page:root:tool', 'M', '系统工具', '/tool', 4, 'tools', 'enabled', 'layout', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 49);

-- ========== 系统接口 4704：系统监控 → 系统工具（id/权限/绑定不动，仅改挂与组件） ==========
UPDATE public.system_menu
SET parent_id = 49, sort = 1, component = 'tool/api/index'
WHERE id = 4704 AND parent_id = 47;

-- ========== C 开发文档 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 4901, 49, 'page:root:tool:docs', 'C', '开发文档', 'docs', 2, 'book-open', 'enabled', 'tool/docs/index', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4901)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 49);

-- ========== 管理员角色绑定（仿 0015；4704 绑定已在库中，无需重建） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (49), (4901)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
