<template>
  <t-dialog
    :visible="visible"
    :header="title"
    :width="width"
    :destroy-on-close="destroyOnClose"
    :confirm-btn="{ content: confirmText, loading }"
    :cancel-btn="cancelText"
    @update:visible="(v) => emit('update:visible', v)"
    @confirm="emit('confirm')"
  >
    <slot />
  </t-dialog>
</template>
<script setup lang="ts">
defineOptions({ name: 'SFormDialog' });

withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    width?: number | string;
    loading?: boolean;
    confirmText?: string;
    cancelText?: string;
    destroyOnClose?: boolean;
  }>(),
  {
    width: 520,
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
