export interface SchedulerJob { id: string; jobKey: string; jobName: string; cron: string; timezone: string; timeoutSec: number; paramsJson: string; retryCount: number; status: number; createdAt: string; updatedAt: string; }
export type SchedulerJobSave = Omit<SchedulerJob, 'id' | 'createdAt' | 'updatedAt'> & { id?: string };
export interface SchedulerRun { id: string; jobId: string; jobKey: string; trigger: string; attempt: number; status: string; startedAt: string; finishedAt: string; output: string; error: string; }
