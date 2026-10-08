-- 0009: 字典数据主题色（dict-theme-components）
--   system_dict_data 新增 theme/color 两列：
--   theme = info(纯文本)/default(中性标签)/primary/success/danger/warning/''(未勾选，伴随自定义色)；
--   color = 自定义 hex（#RGB/#RRGGBB，与 theme 互斥）。
--   列默认 ''：GORM 对零值字段省略插入、回落列默认——默认必须是 ''（= 未勾选），
--   若为 'default' 会把颜色模式的 theme='' 覆盖成 'default'，破坏互斥。
--   存量行经 ADD COLUMN 自动补齐；回填保持改前页面视觉，
--   守卫 theme IN ('','default')，不覆盖用户已改值。
-- 幂等：重复执行不产生重复列、不重复回填；全表 public. 限定。

BEGIN;

ALTER TABLE public.system_dict_data ADD COLUMN IF NOT EXISTS theme VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE public.system_dict_data ADD COLUMN IF NOT EXISTS color VARCHAR(32) NOT NULL DEFAULT '';
-- 已存在列的环境（默认曾为 'default'）同步为 ''
ALTER TABLE public.system_dict_data ALTER COLUMN theme SET DEFAULT '';

-- 状态字典：启用=success / 停用=danger（dict/config/post/menu 页现状）
UPDATE public.system_dict_data SET theme = 'success'
 WHERE dict_type = 'sys_status' AND dict_value = '0' AND theme IN ('', 'default');
UPDATE public.system_dict_data SET theme = 'danger'
 WHERE dict_type = 'sys_status' AND dict_value = '1' AND theme IN ('', 'default');

-- 账号状态：enabled=success / disabled=default（organization/role 页现状 enabled?success:default）
UPDATE public.system_dict_data SET theme = 'success'
 WHERE dict_type = 'sys_state' AND dict_value = 'enabled' AND theme IN ('', 'default');

-- 菜单类型：目录=primary / 按钮=warning / 菜单=default（menu 页 menuTypeTheme 现状）
UPDATE public.system_dict_data SET theme = 'primary'
 WHERE dict_type = 'sys_menu_type' AND dict_value = 'M' AND theme IN ('', 'default');
UPDATE public.system_dict_data SET theme = 'warning'
 WHERE dict_type = 'sys_menu_type' AND dict_value = 'F' AND theme IN ('', 'default');

-- 数据范围：全部条目=primary（role 页数据范围标签现状恒 primary）
UPDATE public.system_dict_data SET theme = 'primary'
 WHERE dict_type = 'sys_data_scope' AND theme IN ('', 'default');

COMMIT;
