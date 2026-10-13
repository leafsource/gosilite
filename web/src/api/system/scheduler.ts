import { request } from '@/utils/request';
import { openSSE } from '@/utils/request/sse';
import { useUserStore } from '@/store';

import type { SchedulerJob, SchedulerJobSave, SchedulerRun, SchedulerRunLog } from './schedulerModel';

export const listSchedulerJobs = (params: { page: number; pageSize: number }) =>
  request.get<{ list: SchedulerJob[]; total: number }>({ url: '/system/scheduler/job/list', params });
export const schedulerKeys = () => request.get<{ list: string[] }>({ url: '/system/scheduler/job/keys' });
export const saveSchedulerJob = (data: SchedulerJobSave) =>
  data.id ? request.put({ url: '/system/scheduler/job', data }) : request.post({ url: '/system/scheduler/job', data });
export const removeSchedulerJob = (id: string) => request.delete({ url: '/system/scheduler/job', data: { id }, params: { id } });
export const toggleSchedulerJob = (id: string) => request.put({ url: '/system/scheduler/job/status', data: { id } });
export const runSchedulerJob = (id: string) => request.post({ url: '/system/scheduler/job/run', data: { id } });
export const retrySchedulerRun = (id: string) => request.post({ url: '/system/scheduler/run/retry', data: { id } });
export const stopSchedulerRun = (id: string) => request.post({ url: '/system/scheduler/run/stop', data: { id } });
export const listSchedulerRuns = (params: { jobId?: string; page: number; pageSize: number }) =>
  request.get<{ list: SchedulerRun[]; total: number }>({ url: '/system/scheduler/run/list', params });
export const listSchedulerRunLogs = (params: { runId: string; afterId?: string; page?: number; pageSize?: number }) =>
  request.get<{ list: SchedulerRunLog[]; total: number }>({ url: '/system/scheduler/run/log/list', params });

export function streamSchedulerRunLogs(
  runId: string,
  onLog: (log: SchedulerRunLog) => void,
  options: { afterId?: string; onError?: (error: Error) => void; onClose?: () => void } = {},
) {
  return openSSE({
    url: '/system/scheduler/run/log/stream',
    params: { runId, page: 1, pageSize: 200 },
    lastEventId: options.afterId,
    getToken: () => useUserStore().ensureFreshAccessToken(),
    onMessage: (event) => {
      if (event.event !== 'log') return;
      onLog(JSON.parse(event.data) as SchedulerRunLog);
    },
    onError: options.onError,
    onClose: options.onClose,
  });
}
