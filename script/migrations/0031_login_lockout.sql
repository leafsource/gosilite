-- 0031: 登录失败限制与账号锁定
BEGIN;

ALTER TABLE public.system_user
  ADD COLUMN IF NOT EXISTS login_failure_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS login_permanently_locked BOOLEAN NOT NULL DEFAULT FALSE;

INSERT INTO public.system_config (config_name, config_key, config_value, config_type, remark)
SELECT '登录最大失败次数', 'sys.login.max_failures', '5', 'system', '15分钟内达到该次数后临时锁定'
WHERE NOT EXISTS (SELECT 1 FROM public.system_config WHERE config_key = 'sys.login.max_failures' AND del_flag = 0);

INSERT INTO public.system_config (config_name, config_key, config_value, config_type, remark)
SELECT '登录失败统计分钟', 'sys.login.failure_window_minutes', '15', 'system', '登录失败次数的统计窗口'
WHERE NOT EXISTS (SELECT 1 FROM public.system_config WHERE config_key = 'sys.login.failure_window_minutes' AND del_flag = 0);

INSERT INTO public.system_config (config_name, config_key, config_value, config_type, remark)
SELECT '登录临时锁定分钟', 'sys.login.lock_minutes', '30', 'system', '达到失败阈值后的临时锁定时长'
WHERE NOT EXISTS (SELECT 1 FROM public.system_config WHERE config_key = 'sys.login.lock_minutes' AND del_flag = 0);

COMMIT;
