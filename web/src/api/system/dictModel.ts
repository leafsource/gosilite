// 字典 API 类型（接口定义文件 camelCase）
export type DictTypeDTO = {
  id: number;
  dictName: string;
  dictType: string;
  status: number;
  orderNum: number;
  remark: string;
};

export type DictDataDTO = {
  id: number;
  dictType: string;
  dictLabel: string;
  dictValue: string;
  cssClass: string;
  listOrder: number;
  status: number;
  isDefault: number;
};

export type DictTypeSaveParams = Omit<DictTypeDTO, 'id'> & { id?: number };
export type DictDataSaveParams = Omit<DictDataDTO, 'id'> & { id?: number };

export type DictOption = { label: string; value: string };
