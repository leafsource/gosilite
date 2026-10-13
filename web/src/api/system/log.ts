import { request } from '@/utils/request';

import type {
  ErrorLogDetailDTO,
  ErrorLogDTO,
  ErrorLogListParams,
  LogCleanParams,
  LoginLogDTO,
  LoginLogListParams,
  OperLogDetailDTO,
  OperLogDTO,
  OperLogListParams,
} from './logModel';

// ===== 操作日志 =====

export function listOperLogs(params: OperLogListParams) {
  return request.get<{ list: OperLogDTO[]; total: number }>({ url: '/system/operlog/list', params });
}

export function getOperLog(id: string) {
  return request.get<OperLogDetailDTO>({ url: '/system/operlog', params: { id } });
}

export function cleanOperLogs(data: LogCleanParams) {
  return request.post<number>({ url: '/system/operlog/clean', data });
}

// ===== 系统异常日志 =====

export function listErrorLogs(params: ErrorLogListParams) {
  return request.get<{ list: ErrorLogDTO[]; total: number }>({ url: '/system/errorlog/list', params });
}

export function getErrorLog(id: string) {
  return request.get<ErrorLogDetailDTO>({ url: '/system/errorlog', params: { id } });
}

export function handleErrorLog(data: { id: string; handleRemark?: string }) {
  return request.post<number>({ url: '/system/errorlog/handle', data });
}

export function cleanErrorLogs(data: LogCleanParams) {
  return request.post<number>({ url: '/system/errorlog/clean', data });
}

// ===== 登录日志 =====

export function listLoginLogs(params: LoginLogListParams) {
  return request.get<{ list: LoginLogDTO[]; total: number }>({ url: '/system/loginlog/list', params });
}
