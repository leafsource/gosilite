-- 0024: 系统管理 · 定时调度
BEGIN;
CREATE TABLE IF NOT EXISTS public.system_job (
  id BIGSERIAL PRIMARY KEY,
  job_key VARCHAR(128) NOT NULL,
  job_name VARCHAR(255) NOT NULL,
  cron VARCHAR(128) NOT NULL,
  timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
  timeout_sec INTEGER NOT NULL DEFAULT 60,
  params_json TEXT NOT NULL DEFAULT '{}',
  retry_count INTEGER NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT uk_system_job_key UNIQUE (job_key),
  CONSTRAINT ck_system_job_timeout CHECK (timeout_sec > 0),
  CONSTRAINT ck_system_job_retry CHECK (retry_count >= 0),
  CONSTRAINT ck_system_job_status CHECK (status IN (0, 1))
);
CREATE TABLE IF NOT EXISTS public.system_job_run (
  id BIGSERIAL PRIMARY KEY,
  job_id BIGINT NOT NULL,
  job_key VARCHAR(128) NOT NULL,
  trigger_type VARCHAR(20) NOT NULL,
  attempt INTEGER NOT NULL,
  status VARCHAR(32) NOT NULL,
  started_at TIMESTAMP NOT NULL,
  finished_at TIMESTAMP,
  output TEXT NOT NULL DEFAULT '',
  error_message TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_system_job_run_job_time ON public.system_job_run(job_id, started_at DESC);

INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 51, 3, 'system:scheduler:list', 'C', '定时任务', '/system/scheduler', 13, 'time', 'enabled', 'system/scheduler/index', 'show', 'system:scheduler:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 51);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT v.id, 51, v.code, 'F', v.title, '', v.sort, v.code, 'enabled', 'hide', 0
FROM (VALUES (5101::bigint, 'system:scheduler:list'::text, '任务查询'::text, 1::int), (5102::bigint, 'system:scheduler:add'::text, '任务新增'::text, 2::int), (5103::bigint, 'system:scheduler:edit'::text, '任务修改'::text, 3::int), (5104::bigint, 'system:scheduler:remove'::text, '任务删除'::text, 4::int), (5105::bigint, 'system:scheduler:run'::text, '任务运行'::text, 5::int)) AS v(id, code, title, sort)
WHERE EXISTS (SELECT 1 FROM public.system_menu WHERE id = 51) AND NOT EXISTS (SELECT 1 FROM public.system_menu m WHERE m.id = v.id);
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id FROM public.system_role r CROSS JOIN (VALUES (51), (5101), (5102), (5103), (5104), (5105)) AS m(id)
WHERE r.code IN ('system', 'admin') AND EXISTS (SELECT 1 FROM public.system_menu sm WHERE sm.id=m.id) AND NOT EXISTS (SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id=r.id AND rm.menu_id=m.id);
COMMIT;
