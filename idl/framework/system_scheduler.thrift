namespace go framework.scheduler

struct SchedulerJobDTO {
  1: i64 id (go.tag = "json:\"id,string\"")
  2: string jobKey
  3: string jobName
  4: string cron
  6: i32 timeoutSec
  7: string paramsJson
  8: i32 retryCount
  9: i32 status
  10: string createdAt
  11: string updatedAt
}
struct SchedulerJobSaveRequest {
  1: optional i64 id (go.tag = "json:\"id,string\"")
  2: string jobKey
  3: string jobName
  4: string cron
  6: i32 timeoutSec
  7: string paramsJson
  8: i32 retryCount
  9: i32 status
}
struct SchedulerJobListRequest { 1: i32 page = 1, 2: i32 pageSize = 10 }
struct SchedulerJobListResponse { 1: list<SchedulerJobDTO> list, 2: i64 total }
struct SchedulerJobIDRequest { 1: i64 id (go.tag = "json:\"id,string\"") }
struct SchedulerJobRunDTO {
  1: i64 id (go.tag = "json:\"id,string\"")
  2: i64 jobId (go.tag = "json:\"jobId,string\"")
  3: string jobKey
  4: string trigger
  5: i32 retryCount
  6: string status
  7: string triggeredAt
  8: string finishedAt
  9: string output
  10: string error
}
struct SchedulerJobRunListRequest { 1: i64 jobId (go.tag = "json:\"jobId,string\""), 2: i32 page = 1, 3: i32 pageSize = 10 }
struct SchedulerJobRunListResponse { 1: list<SchedulerJobRunDTO> list, 2: i64 total }
struct SchedulerJobKeysResponse { 1: list<string> list }

service SystemSchedulerController {
  SchedulerJobListResponse JobList(1: SchedulerJobListRequest req)(api.get = "/system/scheduler/job/list")
  SchedulerJobDTO JobGet(1: SchedulerJobIDRequest req)(api.get = "/system/scheduler/job")
  SchedulerJobDTO JobCreate(1: SchedulerJobSaveRequest req)(api.post = "/system/scheduler/job")
  SchedulerJobDTO JobUpdate(1: SchedulerJobSaveRequest req)(api.put = "/system/scheduler/job")
  i32 JobRemove(1: SchedulerJobIDRequest req)(api.delete = "/system/scheduler/job")
  i32 JobChangeStatus(1: SchedulerJobIDRequest req)(api.put = "/system/scheduler/job/status")
  i32 JobRun(1: SchedulerJobIDRequest req)(api.post = "/system/scheduler/job/run")
  SchedulerJobRunListResponse RunList(1: SchedulerJobRunListRequest req)(api.get = "/system/scheduler/run/list")
  SchedulerJobKeysResponse JobKeys()(api.get = "/system/scheduler/job/keys")
}
