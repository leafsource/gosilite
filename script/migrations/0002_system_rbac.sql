-- 0002_system_rbac.sql
-- 系统 RBAC：部门 / 岗位 / 数据范围 / 菜单按钮权限增强
-- 幂等：可重复执行

BEGIN;

-- ========== 部门 ==========
CREATE TABLE IF NOT EXISTS system_dept (
    id          BIGSERIAL PRIMARY KEY,
    parent_id   BIGINT      NOT NULL DEFAULT 0,
    ancestors   VARCHAR(500) NOT NULL DEFAULT '0',
    dept_name   VARCHAR(64) NOT NULL,
    order_num   INT         NOT NULL DEFAULT 0,
    status      SMALLINT    NOT NULL DEFAULT 0,  -- 0 启用 1 停用
    del_flag    SMALLINT    NOT NULL DEFAULT 0,  -- 0 正常 2 删除
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE system_dept IS '部门';
CREATE INDEX IF NOT EXISTS idx_system_dept_parent ON system_dept(parent_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_dept_name_parent ON system_dept(dept_name, parent_id) WHERE del_flag = 0;

-- ========== 岗位 ==========
CREATE TABLE IF NOT EXISTS system_post (
    id          BIGSERIAL PRIMARY KEY,
    post_code   VARCHAR(64) NOT NULL,
    post_name   VARCHAR(64) NOT NULL,
    order_num   INT         NOT NULL DEFAULT 0,
    status      SMALLINT    NOT NULL DEFAULT 0,
    del_flag    SMALLINT    NOT NULL DEFAULT 0,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE system_post IS '岗位';
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_post_code ON system_post(post_code) WHERE del_flag = 0;

-- ========== 用户增强 ==========
ALTER TABLE public.system_user ADD COLUMN IF NOT EXISTS dept_id BIGINT NOT NULL DEFAULT 0;
COMMENT ON COLUMN public.system_user.dept_id IS '所属部门';

-- ========== 用户-岗位 ==========
CREATE TABLE IF NOT EXISTS system_user_post (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    post_id     BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_user_post ON system_user_post(user_id, post_id);

-- ========== 角色增强：数据范围 ==========
ALTER TABLE public.system_role ADD COLUMN IF NOT EXISTS data_scope SMALLINT NOT NULL DEFAULT 1;
COMMENT ON COLUMN public.system_role.data_scope IS '数据范围 1全部 2本部门 3本部门及以下 4仅本人 5自定义';
ALTER TABLE public.system_role ADD COLUMN IF NOT EXISTS order_num INT NOT NULL DEFAULT 0;

-- ========== 角色-菜单 ==========
CREATE TABLE IF NOT EXISTS system_role_menu (
    id          BIGSERIAL PRIMARY KEY,
    role_id     BIGINT NOT NULL,
    menu_id     BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_role_menu ON system_role_menu(role_id, menu_id);

-- ========== 角色-部门（自定义数据范围） ==========
CREATE TABLE IF NOT EXISTS system_role_dept (
    id          BIGSERIAL PRIMARY KEY,
    role_id     BIGINT NOT NULL,
    dept_id     BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_role_dept ON system_role_dept(role_id, dept_id);

-- ========== 菜单增强（主流字段别名） ==========
ALTER TABLE public.system_menu ADD COLUMN IF NOT EXISTS perms VARCHAR(255) NOT NULL DEFAULT '';
COMMENT ON COLUMN public.system_menu.perms IS '权限标识 e.g. system:user:list';
ALTER TABLE public.system_menu ADD COLUMN IF NOT EXISTS status SMALLINT NOT NULL DEFAULT 0;
-- menu_type 统一：M 目录 / C 菜单 / F 按钮；旧 parent/menu 迁移
UPDATE system_menu SET menu_type = 'M' WHERE menu_type = 'parent';
UPDATE system_menu SET menu_type = 'C' WHERE menu_type = 'menu';
UPDATE system_menu SET perms = permission_code WHERE perms = '' AND permission_code IS NOT NULL AND permission_code <> '';

-- 根父节点统一为 0（旧数据 -1）
UPDATE system_menu SET parent_id = 0 WHERE parent_id = -1;
UPDATE system_dept SET parent_id = 0 WHERE parent_id < 0;

-- ========== 种子：部门 / 岗位 / 角色数据范围 ==========
INSERT INTO system_dept (id, parent_id, ancestors, dept_name, order_num)
SELECT 100, 0, '0', '总公司', 1
WHERE NOT EXISTS (SELECT 1 FROM system_dept WHERE id = 100);

INSERT INTO system_dept (id, parent_id, ancestors, dept_name, order_num)
SELECT 101, 100, '0,100', '研发部门', 2
WHERE NOT EXISTS (SELECT 1 FROM system_dept WHERE id = 101);

INSERT INTO system_dept (id, parent_id, ancestors, dept_name, order_num)
SELECT 102, 100, '0,100', '市场部门', 3
WHERE NOT EXISTS (SELECT 1 FROM system_dept WHERE id = 102);

INSERT INTO system_post (id, post_code, post_name, order_num)
SELECT 1, 'CEO', '董事长', 1 WHERE NOT EXISTS (SELECT 1 FROM system_post WHERE id = 1);
INSERT INTO system_post (id, post_code, post_name, order_num)
SELECT 2, 'CTO', '技术负责人', 2 WHERE NOT EXISTS (SELECT 1 FROM system_post WHERE id = 2);
INSERT INTO system_post (id, post_code, post_name, order_num)
SELECT 3, 'PM', '项目经理', 3 WHERE NOT EXISTS (SELECT 1 FROM system_post WHERE id = 3);

UPDATE public.system_user SET dept_id = 101 WHERE dept_id = 0 AND id = 51;

-- 管理员角色给全部数据范围
UPDATE system_role SET data_scope = 1 WHERE code IN ('system', 'admin') AND data_scope <> 1;

-- ========== 种子：系统管理按钮权限（挂在已有系统菜单下） ==========
-- 确保系统管理目录 parent_id=3 下有 user/dept/post/role/menu 子菜单（若缺）
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 40, 3, 'system:dept:list', 'C', '部门管理', 'dept', 5, 'organization', 'enabled', 'system/dept/index', 'show', 'system:dept:list', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 40);

INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 41, 3, 'system:post:list', 'C', '岗位管理', 'post', 6, 'work', 'enabled', 'system/post/index', 'show', 'system:post:list', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 41);

INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 42, 3, 'system:role:list', 'C', '角色管理', 'role', 7, 'user-queue', 'enabled', 'system/role/index', 'show', 'system:role:list', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 42);

-- 用户管理已有 id=12，补 perms
UPDATE system_menu SET perms = 'system:user:list', menu_type = 'C' WHERE id = 12 AND (perms = '' OR perms IS NULL);
-- 菜单管理已有 id=10，补 perms
UPDATE system_menu SET perms = 'system:menu:list', menu_type = 'C' WHERE id = 10 AND (perms = '' OR perms IS NULL);

-- 按钮权限样例
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4001, 40, 'system:dept:add', 'F', '部门新增', '', 1, 'system:dept:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4001);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4002, 40, 'system:dept:edit', 'F', '部门修改', '', 2, 'system:dept:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4002);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4003, 40, 'system:dept:remove', 'F', '部门删除', '', 3, 'system:dept:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4003);

INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4101, 41, 'system:post:add', 'F', '岗位新增', '', 1, 'system:post:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4101);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4102, 41, 'system:post:edit', 'F', '岗位修改', '', 2, 'system:post:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4102);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4103, 41, 'system:post:remove', 'F', '岗位删除', '', 3, 'system:post:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4103);

INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4201, 42, 'system:role:add', 'F', '角色新增', '', 1, 'system:role:add', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4201);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4202, 42, 'system:role:edit', 'F', '角色修改', '', 2, 'system:role:edit', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4202);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4203, 42, 'system:role:remove', 'F', '角色删除', '', 3, 'system:role:remove', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4203);
INSERT INTO system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT 4204, 42, 'system:role:assign', 'F', '分配菜单', '', 4, 'system:role:assign', 'enabled', 'hide', 0
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE id = 4204);

-- 管理员角色绑定新菜单
INSERT INTO system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM system_role r
CROSS JOIN system_menu m
WHERE r.code IN ('system', 'admin')
  AND m.id IN (40, 41, 42, 4001, 4002, 4003, 4101, 4102, 4103, 4201, 4202, 4203, 4204, 10, 12)
  AND NOT EXISTS (
    SELECT 1 FROM system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
