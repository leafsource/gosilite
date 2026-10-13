<template>
  <template v-for="(item, key) in configs" :key="key">
    <form-field-render
      :model-value="getValue(String(key), item)"
      :config-value="modelValue"
      :name="itemName(String(key), item)"
      :field-config="item"
      @update:model-value="(v) => setValue(String(key), item, v)"
    />
  </template>
</template>
<script setup lang="ts">
import { watch } from 'vue';

import type { FieldConfig } from '@/api/model/fieldConfigModel';
import FormFieldRender from '@/components/field-config/FormFieldRender.vue';
import { getAndSetLastObject } from '@/components/field-config/util';

defineOptions({ name: 'FormFieldRenders' });

const props = withDefaults(
  defineProps<{
    /** 字段配置（supplierConfig） */
    configs?: Record<string, FieldConfig<any>>;
    /** 配置值绑定对象（v-model，对象引用就地变更） */
    modelValue?: Record<string, any>;
    /** 表单字段名称前缀 */
    name: string;
  }>(),
  { configs: () => ({}), modelValue: () => ({}) },
);

const emit = defineEmits<{
  (e: 'update:model-value', v: Record<string, any>): void;
}>();

/** 字段名称路径：优先 config.name（支持 a.b 嵌套），否则取 map key */
function nameList(key: string, item: FieldConfig<any>): string[] {
  return item.name?.split('.') ?? [key];
}

function itemName(key: string, item: FieldConfig<any>): string {
  const names = nameList(key, item);
  return `${props.name}[${names.join('][')}]`;
}

function getValue(key: string, item: FieldConfig<any>): any {
  const names = nameList(key, item);
  const target = props.modelValue ?? {};
  return getAndSetLastObject(target, names)[names[names.length - 1]];
}

function setValue(key: string, item: FieldConfig<any>, v: any) {
  const names = nameList(key, item);
  getAndSetLastObject(props.modelValue ?? {}, names)[names[names.length - 1]] = v;
}

/** 字段配置变更（切换供应商）时重置为默认值 */
watch(
  () => props.configs,
  () => {
    const data: Record<string, any> = {};
    Object.entries(props.configs ?? {}).forEach(([key, item]) => {
      const names = nameList(key, item);
      getAndSetLastObject(data, names)[names[names.length - 1]] = item.value;
    });
    emit('update:model-value', data);
  },
);
</script>
