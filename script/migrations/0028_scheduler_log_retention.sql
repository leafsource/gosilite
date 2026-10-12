-- 0028: scheduler execution-log retention
BEGIN;

CREATE INDEX IF NOT EXISTS idx_system_job_run_log_created_at ON public.system_job_run_log(created_at);

INSERT INTO public.system_job (job_key, job_name, cron, timeout_sec, params_json, retry_count, status)
SELECT 'system.scheduler.log_cleanup', '调度执行日志清理', '0 0 0 * * *', 300, '{}', 0, 1
WHERE NOT EXISTS (
  SELECT 1 FROM public.system_job WHERE job_key = 'system.scheduler.log_cleanup'
);

COMMIT;
