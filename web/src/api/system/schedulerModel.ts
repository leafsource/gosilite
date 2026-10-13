export interface SchedulerJob { id: string; jobKey: string; jobName: string; cron: string; timeoutSec: number; paramsJson: string; retryCount: number; status: number; maxInstances: number; createdAt: string; updatedAt: string; }
export type SchedulerJobSave = Omit<SchedulerJob, 'id' | 'createdAt' | 'updatedAt'> & { id?: string };
export interface SchedulerRun { id: string; jobId: string; jobKey: string; jobName: string; trigger: string; retryCount: number; status: string; triggeredAt: string; finishedAt: string; output: string; error: string; paramsJson: string; timeoutSec: number; configVersion: number; }
export interface SchedulerRunLog { id: string; runId: string; jobKey: string; level: string; stage: string; message: string; createdAt: string; }
