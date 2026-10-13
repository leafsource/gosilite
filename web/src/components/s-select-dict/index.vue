<template>
  <t-select
    :model-value="modelValue ?? undefined"
    :options="selectOptions"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled"
    @update:model-value="onUpdate"
    @change="onChange"
  />
</template>
<script setup lang="ts">
import type { SelectValue } from 'tdesign-vue-next';
import { computed } from 'vue';

import { useDict } from '@/store';

defineOptions({ name: 'SSelectDict' });

const props = withDefaults(
  defineProps<{
    modelValue?: SelectValue | null;
    /** 字典类型 */
    dictType: string;
    /** 选项值类型：auto=跟随 modelValue（number 则转 number，否则 string） */
    valueType?: 'auto' | 'string' | 'number';
    placeholder?: string;
    clearable?: boolean;
    disabled?: boolean;
  }>(),
  { valueType: 'auto', placeholder: '请选择', clearable: true, disabled: false },
);

const emit = defineEmits<{
  (e: 'update:model-value', v: SelectValue): void;
  (e: 'change', v: SelectValue, ctx: unknown): void;
}>();

const { options } = useDict(props.dictType);

function coerce(v: string): string | number {
  if (props.valueType === 'number') return Number(v);
  if (props.valueType === 'string') return String(v);
  return typeof props.modelValue === 'number' ? Number(v) : String(v);
}

const selectOptions = computed(() => options.value.map((o) => ({ label: o.label, value: coerce(o.value) })));

function onUpdate(v: SelectValue) {
  emit('update:model-value', v);
}

function onChange(v: SelectValue, ctx: unknown) {
  emit('change', v, ctx);
}
</script>
