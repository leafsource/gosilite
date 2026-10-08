-- 0010: 菜单按钮操作类型（menu-action-type）
--   system_menu 新增 action_type 列（仅 F 按钮使用，供后续操作审计按动作归类）；
--   新增字典 sys_action_type（添加/修改/删除/下载/上传 + 常用动作 + 其他）；
--   存量 F 按钮按 perms 动作后缀回填，兜底 other。
-- 幂等：重复执行不产生重复列/行；全表 public. 限定。

BEGIN;

-- ========== 列 ==========
ALTER TABLE public.system_menu ADD COLUMN IF NOT EXISTS action_type VARCHAR(32) NOT NULL DEFAULT '';
COMMENT ON COLUMN public.system_menu.action_type IS '按钮操作类型（sys_action_type），仅 menu_type=F 使用';

-- ========== 字典类型（挂“系统”分组，与既有 sys_* 一致） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark, group_id)
SELECT '按钮操作类型', 'sys_action_type', 7, '按钮动作分类（操作审计用），F 菜单必选',
       (SELECT g.id FROM public.system_dict_group g WHERE g.group_name = '系统' AND g.del_flag = 0)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_action_type' AND del_flag = 0);

-- ========== 字典数据（other 固定排最后） ==========
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_action_type', v.label, v.value, v.ord, v.def
FROM (VALUES
    ('add',     '添加', 1,  1),
    ('edit',    '修改', 2,  0),
    ('delete',  '删除', 3,  0),
    ('download','下载', 4,  0),
    ('upload',  '上传', 5,  0),
    ('query',   '查询', 6,  0),
    ('import',  '导入', 7,  0),
    ('export',  '导出', 8,  0),
    ('approve', '审批', 9,  0),
    ('assign',  '分配', 10, 0),
    ('other',   '其他', 99, 0)
) AS v(value, label, ord, def)
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_action_type');

-- ========== 存量 F 按钮回填（按 perms 动作后缀推断，已有值不覆盖） ==========
UPDATE public.system_menu SET action_type = CASE
    WHEN perms LIKE '%:add'      THEN 'add'
    WHEN perms LIKE '%:edit' OR perms LIKE '%:update' THEN 'edit'
    WHEN perms LIKE '%:remove' OR perms LIKE '%:delete' THEN 'delete'
    WHEN perms LIKE '%:list' OR perms LIKE '%:query' OR perms LIKE '%:view' THEN 'query'
    WHEN perms LIKE '%:export'   THEN 'export'
    WHEN perms LIKE '%:import'   THEN 'import'
    WHEN perms LIKE '%:download' THEN 'download'
    WHEN perms LIKE '%:upload'   THEN 'upload'
    WHEN perms LIKE '%:approve' OR perms LIKE '%:audit' THEN 'approve'
    WHEN perms LIKE '%:assign'   THEN 'assign'
    ELSE 'other'
END
WHERE menu_type = 'F' AND action_type = '';

COMMIT;
