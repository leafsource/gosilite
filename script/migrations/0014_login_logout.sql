-- 0014: 登录日志补充「登出」状态
--   sys_login_status 增加 2=登出（0 成功 1 失败 2 登出），
--   配合 operlog 中间件记录 POST /user/logout。
-- 幂等：重复执行不产生重复行。

BEGIN;

INSERT INTO public.system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_login_status', '登出', '2', 3, 0
WHERE NOT EXISTS (
    SELECT 1 FROM public.system_dict_data WHERE dict_type = 'sys_login_status' AND dict_value = '2'
);

COMMIT;
