<template>
  <t-dialog
    v-model:visible="visibleProxy"
    :header="detail?.title || '通知详情'"
    :footer="false"
    width="560px"
    attach="body"
    destroy-on-close
  >
    <div v-if="detail" class="s-notice-detail">
      <div class="s-notice-detail__meta">
        <s-tag-dict
          v-if="detail.noticeType !== undefined && detail.noticeType !== null"
          dict-type="sys_notice_type"
          :value="detail.noticeType"
        />
        <span v-if="detail.publisher" class="s-notice-detail__sub">{{ detail.publisher }}</span>
        <span v-if="detail.createdAt" class="s-notice-detail__sub">{{ detail.createdAt }}</span>
      </div>
      <div class="s-notice-detail__content">{{ detail.content }}</div>
    </div>
  </t-dialog>
</template>
<script setup lang="ts">
import { computed } from 'vue';

import STagDict from '@/components/s-tag-dict/index.vue';

defineOptions({ name: 'SNoticeDetail' });

const props = defineProps<{
  visible: boolean;
  detail: NoticeDetailData | null;
}>();

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
}>();

/** 通知详情展示所需字段（管理 NoticeDTO / 用户 NoticeDetail 均结构兼容） */
interface NoticeDetailData {
  title?: string;
  noticeType?: number | null;
  content?: string;
  publisher?: string;
  createdAt?: string;
}

const visibleProxy = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
});
</script>
<style lang="less" scoped>
.s-notice-detail {
  &__meta {
    display: flex;
    align-items: center;
    gap: var(--td-comp-margin-s, 8px);
    margin-bottom: var(--td-comp-margin-m, 12px);
  }

  &__sub {
    font-size: var(--td-font-size-body-small, 12px);
    color: var(--td-text-color-secondary);
  }

  &__content {
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.8;
    color: var(--td-text-color-primary);
    background: var(--td-bg-color-secondarycontainer, #f3f3f3);
    border-radius: var(--td-radius-default, 3px);
    padding: var(--td-comp-paddingTB-m, 12px) var(--td-comp-paddingLR-m, 12px);
    max-height: 50vh;
    overflow: auto;
  }
}
</style>
