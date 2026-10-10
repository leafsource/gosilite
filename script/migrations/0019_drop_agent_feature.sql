-- 0019: 移除 agent（single_agent_draft / single_agent_publish）遗留
--   代码侧已删除：application/agent、api/{handler,model,router}/agent、idl/agent；
--   本迁移清理存量库：AI/AI编辑器 菜单(13/20)、角色绑定与两张 agent 表。
--   后续 agent 功能按详细设计方案重新立项。
-- 幂等：重复执行不报错（DELETE/DROP IF EXISTS）；全表 public. 限定。

BEGIN;

DELETE FROM public.system_role_menu WHERE menu_id IN (13, 20);
DELETE FROM public.system_role_permission WHERE permission_id IN (13, 20);
DELETE FROM public.system_menu WHERE id IN (13, 20);

DROP TABLE IF EXISTS public.single_agent_publish;
DROP TABLE IF EXISTS public.single_agent_draft;

COMMIT;
