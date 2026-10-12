-- 0025: durable scheduler active-execution fence
BEGIN;
ALTER TABLE public.system_job
  ADD COLUMN IF NOT EXISTS active_run_token VARCHAR(64) NOT NULL DEFAULT '';
COMMIT;
