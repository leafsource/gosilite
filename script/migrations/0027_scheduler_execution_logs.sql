-- 0027: scheduler immutable execution logs
BEGIN;

CREATE TABLE IF NOT EXISTS public.system_job_run_log (
  id BIGSERIAL PRIMARY KEY,
  run_id BIGINT NOT NULL,
  job_key VARCHAR(128) NOT NULL,
  level VARCHAR(16) NOT NULL,
  stage VARCHAR(32) NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_system_job_run_log_run_id ON public.system_job_run_log(run_id, id);

COMMIT;
