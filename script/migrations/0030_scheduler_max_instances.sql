-- 0030: per-job concurrent execution capacity
BEGIN;
ALTER TABLE public.system_job
  ADD COLUMN IF NOT EXISTS max_instances INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS active_run_count INTEGER NOT NULL DEFAULT 0;
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_system_job_max_instances') THEN
    ALTER TABLE public.system_job ADD CONSTRAINT ck_system_job_max_instances CHECK (max_instances > 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_system_job_active_run_count') THEN
    ALTER TABLE public.system_job ADD CONSTRAINT ck_system_job_active_run_count CHECK (active_run_count >= 0);
  END IF;
END $$;
COMMIT;
