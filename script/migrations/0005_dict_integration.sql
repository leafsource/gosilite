-- ============================================================================
-- 0005: 字典关联种子（spec: docs/compose/spec/dict-integration.md S2.1）
--   sys_state / sys_data_scope / sys_menu_type / sys_config_type 四类新字典；
--   sys_status 为 0003 已建，此处仅幂等核对补缺。
-- 幂等：重复执行不产生重复行；全表 public. 限定（规避容器 psql 裸表名解析异常）
-- ============================================================================

BEGIN;

-- ========== 核对：sys_status（0003 已建，缺失才补） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '状态', 'sys_status', 1, '通用状态'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_status');

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_status', '启用', '0', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_status' AND dict_value = '0');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_status', '停用', '1', 2
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_status' AND dict_value = '1');

-- ========== sys_state：账号状态（user/role 的 enabled/disabled 值域） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '账号状态', 'sys_state', 3, '启用/停用（值域 enabled/disabled）'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_state');

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_state', '启用', 'enabled', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_state' AND dict_value = 'enabled');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_state', '停用', 'disabled', 2
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_state' AND dict_value = 'disabled');

-- ========== sys_data_scope：数据范围（role.data_scope 1-5） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '数据范围', 'sys_data_scope', 4, '角色数据权限范围'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_data_scope');

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_data_scope', '全部数据', '1', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_data_scope' AND dict_value = '1');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_data_scope', '本部门', '2', 2
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_data_scope' AND dict_value = '2');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_data_scope', '本部门及以下', '3', 3
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_data_scope' AND dict_value = '3');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_data_scope', '仅本人', '4', 4
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_data_scope' AND dict_value = '4');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_data_scope', '自定义部门', '5', 5
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_data_scope' AND dict_value = '5');

-- ========== sys_menu_type：菜单类型（M/C/F） ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '菜单类型', 'sys_menu_type', 5, '目录/菜单/按钮'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_menu_type');

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_menu_type', '目录', 'M', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_menu_type' AND dict_value = 'M');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_menu_type', '菜单', 'C', 2
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_menu_type' AND dict_value = 'C');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_menu_type', '按钮', 'F', 3
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_menu_type' AND dict_value = 'F');

-- ========== sys_config_type：参数类型 ==========
INSERT INTO public.system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '参数类型', 'sys_config_type', 6, '系统/业务参数'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_type WHERE dict_type = 'sys_config_type');

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_config_type', '系统', 'system', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_config_type' AND dict_value = 'system');
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_config_type', '业务', 'business', 2
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_config_type' AND dict_value = 'business');

COMMIT;
