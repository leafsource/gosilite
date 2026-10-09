// 在线用户（系统监控）

export interface OnlineDTO {
  /** 行标识 = access token 摘要 */
  id: string;
  userId: string;
  /** 登录名称（账号） */
  userName: string;
  deptName: string;
  clientIp: string;
  browser: string;
  os: string;
  loginAt: string;
}

export interface OnlineListParams {
  page?: number;
  pageSize?: number;
  userName?: string;
}
