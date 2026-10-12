import { request } from '@/utils/request';
import type { SchedulerJob, SchedulerJobSave, SchedulerRun } from './schedulerModel';
export const listSchedulerJobs=(params:{page:number;pageSize:number})=>request.get<{list:SchedulerJob[];total:number}>({url:'/system/scheduler/job/list',params});
export const schedulerKeys=()=>request.get<{list:string[]}>({url:'/system/scheduler/job/keys'});
export const saveSchedulerJob=(data:SchedulerJobSave)=>data.id?request.put({url:'/system/scheduler/job',data}):request.post({url:'/system/scheduler/job',data});
export const removeSchedulerJob=(id:string)=>request.delete({url:'/system/scheduler/job',data:{id},params:{id}});
export const toggleSchedulerJob=(id:string)=>request.put({url:'/system/scheduler/job/status',data:{id}});
export const runSchedulerJob=(id:string)=>request.post({url:'/system/scheduler/job/run',data:{id}});
export const listSchedulerRuns=(params:{jobId:string;page:number;pageSize:number})=>request.get<{list:SchedulerRun[];total:number}>({url:'/system/scheduler/run/list',params});
