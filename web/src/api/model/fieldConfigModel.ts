import type { FormRule, InputNumberProps, InputProps, SelectProps, SwitchProps, TextareaProps } from 'tdesign-vue-next';

export interface FieldOption {
  /** 当前选项是否为全选 */
  checkAll?: boolean;
  /** 是否禁用该选项 */
  disabled?: boolean;
  /** 选项标题，在选项过长时 hover 选项展示 */
  title?: string;
  /** 标签 */
  label: string;
  /** 值 */
  value: string | number;
}

export interface FieldOptionGroup {
  /** 是否显示分隔线，默认为 true */
  divider?: boolean;
  /** 分组别名 */
  label?: string;
  /** 分组名称 */
  group: string;
  /** 子选项 */
  children?: FieldOption[];
}

/** 字段基本配置对象（后端 FieldConfig JSON；name/span/placeholder 等缺省时按 map key 兜底） */
export interface FieldConfig<T extends string | number | boolean | Array<string | number> = string> {
  /** 字段默认值 */
  value?: T;
  /** 标签名称 */
  label: string;
  /** 自定义字段名称，默认为对象的属性名称。可以使用 . 作为嵌套对象的属性，例如：items.item */
  name?: string;
  /** 组件 */
  component: string;
  /** 帮助信息 */
  help?: string;
  /** 是否必填 */
  required: boolean;
  /** 占位符 */
  placeholder?: string;
  /** 占用栅格数 */
  span?: number;
  /** 可见性依赖字段。例如 visible='a'，则当 a 为真值时才展示该字段 */
  visible?: string;
  /** 其他校验规则 */
  rules?: FormRule[];
  /** 输入框组件属性 */
  inputProps?: InputProps;
  /** 数字输入框组件属性 */
  inputNumberProps?: InputNumberProps;
  /** 开关组件属性 */
  switchProps?: SwitchProps;
  /** 文本域组件属性 */
  textareaProps?: TextareaProps;
  /** 选择器组件属性 */
  selectProps?: SelectProps;
}

/**
 * 评估字段配置（pattern 字符串 → RegExp、validator 字符串 → 函数）
 * @param config 字段配置对象
 */
export function evalFieldConfig(config: FieldConfig<any>) {
  config.rules?.forEach((rule) => {
    if (rule.pattern) {
      rule.pattern = new RegExp(rule.pattern);
    }
    if (typeof rule.validator === 'string') {
      // 后端以字符串下发的自定义校验，编译为函数（当前后端不下发 validator，保留兼容）
      // eslint-disable-next-line no-new-func
      rule.validator = new Function(rule.validator) as FormRule['validator'];
    }
  });
}

/**
 * 评估字段配置对象
 * @param configs 字段配置对象
 */
export function evalFieldConfigs(configs: Record<string, FieldConfig<any>>) {
  Object.values(configs).forEach((config) => {
    if (config && typeof config === 'object') evalFieldConfig(config);
  });
}
