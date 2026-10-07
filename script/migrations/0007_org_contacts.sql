-- ============================================================================
-- 0007: 组织架构合并（spec 外需求：部门+用户合并为组织架构，手写决定）
--   1) system_dept 增加负责人字段 leader_user_id
--   2) 菜单：新增「组织架构」（/system/contacts），删除「用户管理」「部门管理」
--      - 两个旧菜单的按钮行（user/dept 的 add/edit/remove）迁到新菜单下，权限码不变
--      - 原持有 12/40 的角色自动补授 44，再清理旧授权与旧菜单行（权限链不断）
-- 幂等：重复执行不产生重复/误删；全表 public. 限定
-- ============================================================================

BEGIN;

-- ========== 1. 负责人字段 ==========
ALTER TABLE public.system_dept
  ADD COLUMN IF NOT EXISTS leader_user_id BIGINT NOT NULL DEFAULT 0;

-- ========== 2. 组织架构菜单（path 与前端路由 /system/contacts 对齐） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 44, 3, 'system:org:list', 'C', '组织架构', 'contacts', 5, 'organization', 'enabled', 'system/organization/index', 'show', 'system:user:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 44);

-- ========== 3. 按钮行迁到组织架构下（ids 不变 → 既有 role_menu 授权继续有效） ==========
UPDATE public.system_menu
SET parent_id = 44
WHERE id IN (1201, 1202, 1203, 4001, 4002, 4003)
  AND parent_id <> 44;

-- ========== 4. 原持有旧菜单的角色补授新菜单 ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT DISTINCT rm.role_id, 44
FROM public.system_role_menu rm
WHERE rm.menu_id IN (12, 40)
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu x WHERE x.role_id = rm.role_id AND x.menu_id = 44
  );

-- ========== 5. 清理旧授权与旧菜单 ==========
DELETE FROM public.system_role_menu WHERE menu_id IN (12, 40);
DELETE FROM public.system_menu WHERE id IN (12, 40);

COMMIT;
