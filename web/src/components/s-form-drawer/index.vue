<template>
  <t-drawer
    :visible="visible"
    :header="title"
    :size="size"
    :footer="true"
    :destroy-on-close="destroyOnClose"
    :confirm-btn="{ content: confirmText, loading }"
    :cancel-btn="cancelText"
    @update:visible="(v) => emit('update:visible', v)"
    @confirm="emit('confirm')"
  >
    <slot />
  </t-drawer>
</template>
<script setup lang="ts">
defineOptions({ name: 'SFormDrawer' });

withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    size?: string;
    loading?: boolean;
    confirmText?: string;
    cancelText?: string;
    destroyOnClose?: boolean;
  }>(),
  {
    size: 'large',
    loading: false,
    confirmText: '保存',
    cancelText: '取消',
    destroyOnClose: true,
  },
);

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'confirm'): void;
}>();
</script>
<style scoped>
/* 与 s-form-dialog 统一：按钮右对齐、取消在左、保存在右 */
:deep(.t-drawer__footer) {
  display: flex;
  flex-direction: row-reverse;
  justify-content: flex-start;
  gap: var(--td-comp-margin-s);
}
</style>
