-- 0008: 字典分组（dict-group）
--   新增 system_dict_group；system_dict_type 增加 group_id（0=未分组）。
--   分组是分类字段：dict_type 仍全局唯一，前端消费 options/useDict 行为不变。
-- 幂等：重复执行不产生重复行/列；全表 public. 限定。

BEGIN;

-- ========== 分组表 ==========
CREATE TABLE IF NOT EXISTS public.system_dict_group (
    id          BIGSERIAL PRIMARY KEY,
    group_name  VARCHAR(64) NOT NULL,
    order_num   INT NOT NULL DEFAULT 0,
    remark      VARCHAR(255) NOT NULL DEFAULT '',
    del_flag    SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_dict_group_name ON public.system_dict_group(group_name) WHERE del_flag = 0;

-- ========== 字典类型挂分组 ==========
ALTER TABLE public.system_dict_type ADD COLUMN IF NOT EXISTS group_id BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_system_dict_type_group ON public.system_dict_type(group_id);

-- ========== 种子：系统分组 + 存量 sys_* 归组 ==========
INSERT INTO public.system_dict_group (group_name, order_num, remark)
SELECT '系统', 1, '系统内置字典'
WHERE NOT EXISTS (SELECT 1 FROM public.system_dict_group WHERE group_name = '系统' AND del_flag = 0);

UPDATE public.system_dict_type t
SET group_id = g.id
FROM public.system_dict_group g
WHERE g.group_name = '系统' AND g.del_flag = 0
  AND t.group_id = 0 AND t.del_flag = 0 AND t.dict_type LIKE 'sys\_%';

COMMIT;
