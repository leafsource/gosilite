<template>
  <t-radio-group
    :model-value="modelValue ?? undefined"
    :options="radioOptions"
    :disabled="disabled"
    :size="size"
    :variant="variant"
    @update:model-value="onUpdate"
    @change="onChange"
  />
</template>
<script setup lang="ts">
import type { RadioValue } from 'tdesign-vue-next';
import { computed } from 'vue';

import { useDict } from '@/store';

defineOptions({ name: 'SRadioDict' });

const props = withDefaults(
  defineProps<{
    modelValue?: RadioValue | null;
    /** 字典类型 */
    dictType: string;
    /** 选项值类型：auto=跟随 modelValue（number 则转 number，否则 string） */
    valueType?: 'auto' | 'string' | 'number';
    disabled?: boolean;
    size?: 'small' | 'medium' | 'large';
    variant?: 'outline' | 'primary-filled' | 'default-filled';
  }>(),
  { valueType: 'auto', disabled: false },
);

const emit = defineEmits<{
  (e: 'update:model-value', v: RadioValue): void;
  (e: 'change', v: RadioValue, ctx: unknown): void;
}>();

const { options } = useDict(props.dictType);

function coerce(v: string): RadioValue {
  if (props.valueType === 'number') return Number(v);
  if (props.valueType === 'string') return String(v);
  return typeof props.modelValue === 'number' ? Number(v) : String(v);
}

const radioOptions = computed(() => options.value.map((o) => ({ label: o.label, value: coerce(o.value) })));

function onUpdate(v: RadioValue) {
  emit('update:model-value', v);
}

function onChange(v: RadioValue, ctx: unknown) {
  emit('change', v, ctx);
}
</script>
