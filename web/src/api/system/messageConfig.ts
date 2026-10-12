import type { FieldConfig } from '@/api/model/fieldConfigModel';
import { evalFieldConfigs } from '@/api/model/fieldConfigModel';
import { request } from '@/utils/request';

import type {
  MessageConfigDTO,
  MessageConfigListParams,
  MessageConfigSaveParams,
  MessageFieldConfig,
  MessageSupplierOption,
} from './messageConfigModel';

// 查询消息配置列表
export function listMessageConfigs(params: MessageConfigListParams) {
  return request.get<{ list: MessageConfigDTO[]; total: number }>({ url: '/system/message/config/list', params });
}

// 查询消息配置详细
export function getMessageConfig(id: string) {
  return request.get<MessageConfigDTO>({ url: '/system/message/config', params: { id } });
}

// 新增消息配置
export function createMessageConfig(data: MessageConfigSaveParams) {
  return request.post({ url: '/system/message/config', data });
}

// 修改消息配置
export function updateMessageConfig(data: MessageConfigSaveParams) {
  return request.put({ url: '/system/message/config', data });
}

// 删除消息配置
export function removeMessageConfig(id: string) {
  return request.delete({ url: '/system/message/config', data: { id }, params: { id } });
}

// 按消息类型获取可用供应商
export function getMessageSuppliers(messageType: string) {
  return request.get<{ list: MessageSupplierOption[] }>({
    url: '/system/message/suppliers',
    params: { messageType },
  });
}

/**
 * 解析 field-configs JSON 并执行 evalFieldConfigs：
 * 兼容两种下发形态 —— ruoyi 形态 {supplier:{supplierConfig,messageType,templateMode}}
 * 与扁平形态 {supplier:{field:FieldConfig}}。
 */
function normalizeFieldConfigs(json: string): Record<string, MessageFieldConfig> {
  let parsed: Record<string, unknown> = {};
  try {
    parsed = JSON.parse(json || '{}') as Record<string, unknown>;
  } catch {
    parsed = {};
  }
  const result: Record<string, MessageFieldConfig> = {};
  Object.entries(parsed).forEach(([code, val]) => {
    const obj = (val ?? {}) as Record<string, any>;
    if (obj.supplierConfig && typeof obj.supplierConfig === 'object') {
      evalFieldConfigs(obj.supplierConfig as Record<string, FieldConfig<any>>);
      result[code] = {
        supplierConfig: obj.supplierConfig,
        messageType: typeof obj.messageType === 'string' ? obj.messageType : '',
        templateMode: obj.templateMode ?? { supportTemplateId: false, supportTemplateContent: false },
      };
    } else {
      const supplierConfig = obj as Record<string, FieldConfig<any>>;
      evalFieldConfigs(supplierConfig);
      result[code] = {
        supplierConfig,
        messageType: '',
        templateMode: { supportTemplateId: false, supportTemplateContent: false },
      };
    }
  });
  return result;
}

// 获取消息字段配置（supplier → 字段定义）
export function getMessageFieldConfigs() {
  return request
    .get<{ json: string }>({ url: '/system/message/field-configs' })
    .then((res) => normalizeFieldConfigs(res.json));
}
