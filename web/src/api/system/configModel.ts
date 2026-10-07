// 参数配置类型
export type ConfigDTO = {
  id: string;
  configName: string;
  configKey: string;
  configValue: string;
  configType: string;
  remark: string;
  status: number;
};

export type ConfigSaveParams = Omit<ConfigDTO, 'id'> & { id?: string };
