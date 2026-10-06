-- ============================================================================
-- 0006: 参数消费种子（spec: docs/compose/spec/config-consumption.md S2.1）
--   1) sys.app.name 参数类型转为 global（匿名可读品牌参数）
--   2) sys_config_type 字典补「全局/global」行（类型下拉自动出现）
-- 幂等：重复执行不产生重复/覆盖；全表 public. 限定
-- ============================================================================

BEGIN;

-- sys.app.name → 全局类型（匿名端点 /system/config/global 提供）
UPDATE public.system_config
SET config_type = 'global'
WHERE config_key = 'sys.app.name'
  AND config_type IS DISTINCT FROM 'global';

-- sys_config_type 字典补数据行
INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_config_type', '全局', 'global', 3, 0
WHERE NOT EXISTS (
  SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_config_type' AND dict_value = 'global'
);

COMMIT;
