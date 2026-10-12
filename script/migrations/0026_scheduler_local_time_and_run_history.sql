-- 0026: scheduler server-local cron and run-history terminology
BEGIN;

ALTER TABLE public.system_job DROP COLUMN IF EXISTS timezone;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'system_job_run' AND column_name = 'started_at'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'system_job_run' AND column_name = 'triggered_at'
  ) THEN
    ALTER TABLE public.system_job_run RENAME COLUMN started_at TO triggered_at;
  END IF;
END $$;

DROP INDEX IF EXISTS public.idx_system_job_run_job_time;
CREATE INDEX IF NOT EXISTS idx_system_job_run_job_time ON public.system_job_run(job_id, triggered_at DESC);

UPDATE public.system_menu
SET parent_id = 47, path = 'scheduler', component = 'system/scheduler/index', sort = 5
WHERE id = 51;

COMMIT;
