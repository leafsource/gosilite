<template>
  <span v-if="mode === 'text'" class="s-tag-dict__text">{{ text }}</span>
  <t-tag v-else-if="mode === 'tag'" :theme="tagTheme" :color="tagColor || undefined" :variant="variant" :size="size">
    {{ text }}
  </t-tag>
</template>
<script setup lang="ts">
import { computed } from 'vue';

import { useDict } from '@/store';

defineOptions({ name: 'STagDict' });

const props = withDefaults(
  defineProps<{
    /** 字典类型 */
    dictType: string;
    /** 字典值；null/undefined/'' 渲染空 */
    value?: string | number | null;
    variant?: 'dark' | 'light' | 'outline' | 'light-outline';
    size?: 'small' | 'medium' | 'large';
  }>(),
  { variant: 'light', size: 'medium' },
);

const { options, labelOf } = useDict(props.dictType);

const THEMES = ['primary', 'success', 'danger', 'warning'];

const option = computed(() => {
  if (props.value === undefined || props.value === null || props.value === '') return undefined;
  return options.value.find((o) => String(o.value) === String(props.value));
});

const text = computed(() => {
  if (props.value === undefined || props.value === null || props.value === '') return '';
  return option.value ? option.value.label : labelOf(props.value);
});

/** info=纯文本不渲染标签；空值不渲染 */
const mode = computed<'none' | 'text' | 'tag'>(() => {
  if (text.value === '') return 'none';
  return (option.value?.theme ?? '') === 'info' ? 'text' : 'tag';
});

/** 主题令牌；default/空/未知一律中性 */
const tagTheme = computed(() => {
  const t = option.value?.theme ?? '';
  return THEMES.includes(t) ? (t as 'primary' | 'success' | 'danger' | 'warning') : 'default';
});

/** 自定义色优先于主题色 */
const tagColor = computed(() => option.value?.color ?? '');
</script>
