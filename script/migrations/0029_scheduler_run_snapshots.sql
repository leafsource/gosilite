-- 0029: immutable scheduler run snapshots and cooperative stop state
BEGIN;

ALTER TABLE public.system_job
  ADD COLUMN IF NOT EXISTS config_version integer NOT NULL DEFAULT 1;

ALTER TABLE public.system_job_run
  ADD COLUMN IF NOT EXISTS job_name varchar(128) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS params_json text NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS timeout_sec integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS config_version integer NOT NULL DEFAULT 1;

UPDATE public.system_job_run r
SET job_name = j.job_name,
    params_json = j.params_json,
    timeout_sec = j.timeout_sec,
    config_version = j.config_version
FROM public.system_job j
WHERE r.job_id = j.id
  AND r.job_name = '';

COMMIT;
