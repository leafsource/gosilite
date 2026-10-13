// OSS配置 DTO / 参数（对应 idl/framework/system_oss_config.thrift）
export interface OssConfigDTO {
  id: string;
  configKey: string;
  accessKey: string;
  secretKey: string;
  bucketName: string;
  prefix: string;
  endpoint: string;
  domain: string;
  /** Y/N */
  isHttps: string;
  region: string;
  /** 0=private 1=public 2=custom */
  accessPolicy: string;
  /** 是否默认：1=默认 0=否 */
  status: string;
  ext1: string;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface OssConfigListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
}

export interface OssConfigSaveParams {
  id?: string;
  configKey: string;
  accessKey: string;
  secretKey: string;
  bucketName: string;
  prefix?: string;
  endpoint: string;
  domain?: string;
  isHttps?: string;
  region?: string;
  accessPolicy?: string;
  remark?: string;
}

export interface OssConfigTestParams {
  accessKey: string;
  secretKey: string;
  bucketName: string;
  endpoint: string;
  region?: string;
}

export interface OssConfigTestResult {
  connected: boolean;
  message: string;
}
