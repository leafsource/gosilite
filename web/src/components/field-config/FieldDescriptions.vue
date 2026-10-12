<template>
  <t-descriptions :column="2" bordered size="small">
    <slot name="prefix" />
    <template v-for="(fieldConfig, key) in fieldConfigs" :key="key">
      <t-descriptions-item v-if="isVisible(fieldConfig)" :label="fieldConfig.label">
        <t-tag v-if="selectLabel(fieldConfig, getValue(fieldConfig, String(key)))" theme="primary" variant="light">
          {{ selectLabel(fieldConfig, getValue(fieldConfig, String(key))) }}
        </t-tag>
        <template v-else>{{ displayValue(fieldConfig, getValue(fieldConfig, String(key))) }}</template>
      </t-descriptions-item>
    </template>
    <slot name="suffix" />
  </t-descriptions>
</template>
<script setup lang="ts">
import type { FieldConfig, FieldOption, FieldOptionGroup } from '@/api/model/fieldConfigModel';
import { getAndSetLastObject } from '@/components/field-config/util';

defineOptions({ name: 'FieldDescriptions' });

const props = defineProps<{
  /** 配置值（JSON 字符串或对象） */
  configValue: string | Record<string, any>;
  /** 字段配置（supplierConfig） */
  fieldConfigs: Record<string, FieldConfig<any>>;
}>();

function effectValue(): Record<string, any> {
  if (typeof props.configValue === 'string') {
    try {
      return JSON.parse(props.configValue || '{}') ?? {};
    } catch {
      return {};
    }
  }
  return props.configValue ?? {};
}

function isVisible(fieldConfig: FieldConfig<any>): boolean {
  return !fieldConfig.visible || !!effectValue()[fieldConfig.visible];
}

/** 取值：优先 config.name（支持 a.b 嵌套），否则取 map key */
function getValue(fieldConfig: FieldConfig<any>, key: string): any {
  const names = fieldConfig.name?.split('.') ?? [key];
  return getAndSetLastObject(effectValue(), names)[names[names.length - 1]];
}

/** select 组件：按选项解析展示文案（未命中返回空，走原始值展示） */
function selectLabel(fieldConfig: FieldConfig<any>, value: any): string {
  if (fieldConfig.component !== 'select' || value === undefined || value === null || value === '') return '';
  const options = (fieldConfig.selectProps?.options ?? []) as Array<FieldOption | FieldOptionGroup>;
  const flat = options.flatMap(
    (option) => ('children' in option && option.children ? option.children : option) as FieldOption[],
  );
  const hit = flat.find((o) => String(o.value) === String(value));
  return hit?.label ?? '';
}

function displayValue(fieldConfig: FieldConfig<any>, value: any): string {
  if (value === undefined || value === null || value === '') return '';
  if (typeof value === 'boolean') return value ? '是' : '否';
  return String(value);
}
</script>
