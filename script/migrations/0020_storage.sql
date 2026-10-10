-- 0020: 存储管理（RustFS/S3 可视化）
--   菜单「存储管理」id=48（C，parent=3 系统管理，component system/storage/index）
--   + F 按钮 4801 删除、4802 上传 + system/admin 角色绑定。
--   无数据表：对象数据实时来自 infra/storage 列表接口。
-- 幂等：重复执行不产生重复行；全表 public. 限定。
-- spec: docs/compose/spec/user-center-storage.md

BEGIN;

-- ========== C 存储管理（父级系统管理 id=3；sort 取末位自增） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 48, 3, 'system:storage:list', 'C', '存储管理', 'storage',
       (SELECT COALESCE(MAX(sort), 0) + 1 FROM public.system_menu WHERE parent_id = 3),
       'folder', 'enabled', 'system/storage/index', 'show', 'system:storage:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 48);

-- ========== F 删除对象 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4801, 48, 'system:storage:delete', 'F', '删除对象', '', 1, 'system:storage:delete', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4801)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 48);

-- ========== F 上传对象 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4802, 48, 'system:storage:upload', 'F', '上传对象', '', 2, 'system:storage:upload', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 4802)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 48);

-- ========== 管理员角色绑定（照 0015） ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES (48), (4801), (4802)) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
