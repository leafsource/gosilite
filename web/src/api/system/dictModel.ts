// 字典 API 类型（接口定义文件 camelCase）
export interface DictTypeDTO {
  id: string;
  dictName: string;
  dictType: string;
  status: number;
  orderNum: number;
  remark: string;
  groupId: string;
}

export interface DictDataDTO {
  id: string;
  dictType: string;
  dictLabel: string;
  dictValue: string;
  cssClass: string;
  listOrder: number;
  status: number;
  isDefault: number;
  /** 主题令牌 default/primary/success/danger/warning（default=不着色） */
  theme: string;
  /** 自定义色 #RGB/#RRGGBB，非空时优先于 theme */
  color: string;
}

export interface DictGroupDTO {
  id: string;
  groupName: string;
  orderNum: number;
  remark: string;
}

export type DictTypeSaveParams = Omit<DictTypeDTO, 'id'> & { id?: string };
export type DictDataSaveParams = Omit<DictDataDTO, 'id'> & { id?: string };
export type DictGroupSaveParams = Omit<DictGroupDTO, 'id'> & { id?: string };

export interface DictOption {
  label: string;
  value: string;
  theme?: string;
  color?: string;
}
