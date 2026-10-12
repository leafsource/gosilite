<template>
  <t-col v-if="visible" :span="span">
    <t-form-item :label="fieldConfig.label" :name="name" :rules="itemRules">
      <template v-if="fieldConfig.help" #help>
        <span v-html="fieldConfig.help" />
      </template>
      <t-input
        v-if="fieldConfig.component === 'input'"
        :model-value="(modelValue as string) ?? ''"
        v-bind="fieldConfig.inputProps"
        :placeholder="fieldConfig.placeholder ?? `请输入${fieldConfig.label}`"
        @update:model-value="emitUpdate"
      />
      <t-input-number
        v-else-if="fieldConfig.component === 'input-number'"
        :model-value="(modelValue as number) ?? undefined"
        v-bind="fieldConfig.inputNumberProps"
        :placeholder="fieldConfig.placeholder"
        style="width: 100%"
        @update:model-value="emitUpdate"
      />
      <t-switch
        v-else-if="fieldConfig.component === 'switch'"
        :model-value="(modelValue as boolean) ?? undefined"
        v-bind="fieldConfig.switchProps"
        @update:model-value="emitUpdate"
      />
      <t-select
        v-else-if="fieldConfig.component === 'select'"
        :model-value="modelValue"
        v-bind="fieldConfig.selectProps"
        :placeholder="fieldConfig.placeholder ?? `请选择${fieldConfig.label}`"
        clearable
        @update:model-value="emitUpdate"
      />
      <t-textarea
        v-else-if="fieldConfig.component === 'textarea'"
        :model-value="(modelValue as string) ?? ''"
        v-bind="fieldConfig.textareaProps"
        :placeholder="fieldConfig.placeholder ?? `请输入${fieldConfig.label}`"
        @update:model-value="emitUpdate"
      />
    </t-form-item>
  </t-col>
</template>
<script setup lang="ts">
import { computed } from 'vue';

import type { FieldConfig } from '@/api/model/fieldConfigModel';

defineOptions({ name: 'FormFieldRender' });

const props = defineProps<{
  /** 当前字段绑定值 */
  modelValue?: string | number | boolean | Array<string | number> | Record<string, any>;
  /** 表单字段名称（t-form data 上的取值路径） */
  name: string;
  /** 配置值对象（visible 依赖取值来源） */
  configValue: Record<string, any>;
  /** 表单字段配置 */
  fieldConfig: FieldConfig<any>;
}>();

const emit = defineEmits<{
  (e: 'update:model-value', v: any): void;
}>();

const visible = computed(() => !props.fieldConfig.visible || !!props.configValue[props.fieldConfig.visible]);

const span = computed(() => props.fieldConfig.span ?? 12);

const itemRules = computed(() => [
  { required: props.fieldConfig.required, message: `${props.fieldConfig.label}不能为空` },
  ...(props.fieldConfig.rules ?? []),
]);

function emitUpdate(v: any) {
  emit('update:model-value', v);
}
</script>
