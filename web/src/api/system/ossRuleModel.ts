// OSS处理规则 DTO / 参数（对应 idl/framework/system_oss_rule.thrift）
export interface OssRuleDTO {
  id: string;
  ruleName: string;
  domain: string;
  mimeType: string;
  rule: string;
  isOverwrite: string;
  isDefault: string;
  status: string;
  ruleSort: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface OssRuleListParams {
  page?: number;
  pageSize?: number;
  ruleName?: string;
  domain?: string;
  /** 缺省=全部；1=启用 0=停用 */
  status?: string;
  /** 缺省=全部；Y=默认 N=非默认 */
  isDefault?: string;
}

export interface OssRuleSaveParams {
  id?: string;
  ruleName: string;
  domain: string;
  mimeType: string;
  rule: string;
  isDefault: string;
  status: string;
  ruleSort?: number;
  remark?: string;
}
