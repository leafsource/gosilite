import type { FieldConfig } from '@/api/model/fieldConfigModel';

/** 消息配置 DTO（idl/framework/system_message.thrift MessageConfigDTO） */
export interface MessageConfigDTO {
  /** 主键（字符串契约） */
  id: string;
  title: string;
  /** MAIL / SMS */
  messageType: string;
  /** 供应商标识：MAIL / BAIDU */
  supplierType: string;
  /** 通道凭据与参数 JSON 字符串 */
  configJson: string;
  /** 1 正常 0 停用 */
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageConfigListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  messageType?: string;
  status?: number;
}

export interface MessageConfigSaveParams {
  id?: string;
  title: string;
  messageType: string;
  supplierType: string;
  configJson: string;
  status: number;
  remark?: string;
}

/** 供应商选项（GET /system/message/suppliers） */
export interface MessageSupplierOption {
  code: string;
  label: string;
}

/** 消息模板支持类型 */
export interface TemplateMode {
  /** 支持模板 id */
  supportTemplateId: boolean;
  /** 支持模板内容 */
  supportTemplateContent: boolean;
}

export type SupplierFieldConfig = Record<string, FieldConfig<any>>;

/** supplier → 动态表单字段定义（GET /system/message/field-configs 解析后） */
export interface MessageFieldConfig {
  /** 消息配置动态字段 */
  supplierConfig: SupplierFieldConfig;
  /** 消息类型 */
  messageType: string;
  /** 消息模板支持类型 */
  templateMode: TemplateMode;
}
