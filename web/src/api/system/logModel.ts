// 日志管理（操作日志 / 系统异常 / 登录日志）类型

export interface OperLogDTO {
  id: string;
  traceId: string;
  userName: string;
  /** sys_oper_type: 0其他 1新增 2修改 3删除 4查询 5导入 6导出 7上传 */
  operType: number;
  requestUri: string;
  httpMethod: string;
  module: string;
  /** sys_oper_status: 0 成功 1 失败 */
  operStatus: number;
  remark: string;
  errorMsg: string;
  costMs: number;
  clientIp: string;
  createdAt: string;
}

export interface OperLogDetailDTO {
  base: OperLogDTO;
  requestHeaders: string;
  requestBody: string;
  responseBody: string;
}

export interface ErrorLogDTO {
  id: string;
  traceId: string;
  userName: string;
  /** 1 业务异常 2 panic */
  errType: number;
  /** sys_error_level: 1 普通 2 严重 */
  errLevel: number;
  /** sys_handle_status: 0 待处理 1 已处理 */
  handleStatus: number;
  errCode: number;
  errMsg: string;
  requestUri: string;
  httpMethod: string;
  module: string;
  costMs: number;
  clientIp: string;
  createdAt: string;
  handledAt: string;
}

export interface ErrorLogDetailDTO {
  base: ErrorLogDTO;
  stack: string;
  requestHeaders: string;
  requestBody: string;
  handleRemark: string;
}

export interface LoginLogDTO {
  id: string;
  traceId: string;
  userName: string;
  userId: string;
  /** sys_login_status: 0 成功 1 失败 */
  status: number;
  failReason: string;
  clientIp: string;
  userAgent: string;
  browser: string;
  os: string;
  createdAt: string;
}

export interface OperLogListParams {
  page?: number;
  pageSize?: number;
  traceId?: string;
  userName?: string;
  operType?: number;
  execStatus?: number;
  method?: string;
  httpMethod?: string;
  startTime?: string;
  endTime?: string;
}

export interface ErrorLogListParams {
  page?: number;
  pageSize?: number;
  traceId?: string;
  userName?: string;
  keyword?: string;
  handleStatus?: number;
  errLevel?: number;
  httpMethod?: string;
  startTime?: string;
  endTime?: string;
}

export interface LoginLogListParams {
  page?: number;
  pageSize?: number;
  userName?: string;
  status?: number;
  clientIp?: string;
  startTime?: string;
  endTime?: string;
}

export interface LogCleanParams {
  ids?: string[];
  all?: boolean;
}
