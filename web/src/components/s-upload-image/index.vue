<template>
  <t-upload
    v-model:files="files"
    :action="resolvedAction"
    :name="field"
    :accept="accept"
    :headers="uploadHeaders"
    :size-limit="sizeLimit"
    :disabled="disabled"
    :format-response="formatResponse"
    theme="image"
    :show-image-file-name="false"
    auto-upload
    @success="handleSuccess"
    @fail="handleFail"
    @validate="handleValidate"
  >
    <template #trigger>
      <div
        class="s-upload-image"
        :class="{ 'is-disabled': disabled }"
        :style="{ width: size, height: size }"
      >
        <t-avatar :image="url || undefined" :size="size" shape="circle">
          <t-icon name="user" />
        </t-avatar>
        <div v-if="!disabled" class="s-upload-image__mask">
          <span class="s-upload-image__mask-text">{{ uploadText }}</span>
        </div>
      </div>
    </template>
    <!-- 单图 URL 语义：列表由外部 url 驱动，不渲染 t-upload 默认文件列表 -->
    <template #file-list-display></template>
  </t-upload>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { UploadProps } from 'tdesign-vue-next';

import { useUserStore } from '@/store/modules/user';

defineOptions({ name: 'SUploadImage' });

const props = withDefaults(
  defineProps<{
    /** 上传端点（统一信封，成功时 data.webURI 为图片地址） */
    action: string;
    /** multipart 字段名（头像接口为 avatar，通用上传为 file） */
    field?: string;
    /** 当前图片地址（v-model:url 双向绑定） */
    url?: string;
    /** 受控的 input accept，默认所有图片 */
    accept?: string;
    /** 预览尺寸（含遮罩区域），如 80px */
    size?: string;
    disabled?: boolean;
    /** 图片大小上限（MB），超出由 t-upload 拦截 */
    maxSizeMb?: number;
    /** 遮罩提示文案 */
    uploadText?: string;
  }>(),
  {
    field: 'file',
    url: '',
    accept: 'image/*',
    size: '80px',
    disabled: false,
    maxSizeMb: 5,
    uploadText: '更换头像',
  },
);

const emit = defineEmits<{
  (e: 'update:url', v: string): void;
  (e: 'success', payload: { url: string; response: unknown }): void;
  (e: 'error', error: unknown): void;
}>();

/** 上传需带会话 token（t-upload 走原生请求，不经 axios 拦截器） */
const userStore = useUserStore();
const uploadHeaders = computed<Record<string, string>>(() => {
  const headers: Record<string, string> = {};
  if (userStore.token) {
    headers.Authorization = `Bearer ${userStore.token}`;
  }
  return headers;
});

const sizeLimit = computed(() => ({
  size: props.maxSizeMb,
  unit: 'MB' as const,
  message: `图片大小不超过 {sizeLimit} MB`,
}));

/**
 * t-upload 走原生请求、不经 axios 的 urlPrefix：以 / 开头的相对路径需自行拼
 * VITE_API_URL_PREFIX（vite proxy 再剥离前缀转发后端），否则会打到前端返回 index.html。
 */
const resolvedAction = computed(() => {
  const prefix = (import.meta.env.VITE_API_URL_PREFIX || '').replace(/\/$/, '');
  const action = props.action;
  if (!prefix || !action.startsWith('/') || action.startsWith('//')) return action;
  if (action === prefix || action.startsWith(`${prefix}/`)) return action;
  return `${prefix}${action}`;
});

/** 单图替换语义：内部文件列表不驻留，成功后即清空 */
const files = ref<UploadProps['value']>([]);

function extractUrl(res: unknown): string | null {
  const r = res as { code?: number; data?: { webURI?: string } } | null;
  if (r && r.code === 200 && r.data?.webURI) return r.data.webURI;
  return null;
}

/** 信封 → t-upload 语义：code=200 取 webURI，否则给出错误文案 */
const formatResponse: UploadProps['formatResponse'] = (response) => {
  const url = extractUrl(response);
  if (url) return { url };
  const msg = (response as { msg?: string } | null)?.msg;
  return { error: msg || '上传失败' };
};

const handleSuccess: UploadProps['onSuccess'] = (ctx) => {
  const url = extractUrl(ctx.response) ?? ctx.file?.url ?? '';
  if (!url) return;
  emit('update:url', url);
  emit('success', { url, response: ctx.response });
  files.value = [];
};

const handleFail: UploadProps['onFail'] = (opts) => {
  files.value = [];
  // formatResponse 失败时 t-upload 透传的是 { error }（内含从信封提取的 msg），优先取它
  const resp = opts.response as { error?: string; msg?: string } | null;
  const msg =
    resp?.error ||
    resp?.msg ||
    `文件 ${opts.file?.name ?? ''} 上传失败`;
  MessagePlugin.error(msg);
  emit('error', opts);
};

const handleValidate: UploadProps['onValidate'] = (ctx) => {
  if (ctx.type === 'FILE_OVER_SIZE_LIMIT') {
    MessagePlugin.warning(`图片大小不超过 ${props.maxSizeMb} MB`);
  }
};
</script>
<style lang="less" scoped>
.s-upload-image {
  position: relative;
  display: inline-block;
  overflow: hidden;
  border-radius: 50%;
  cursor: pointer;

  &.is-disabled {
    cursor: not-allowed;
  }

  &__mask {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgb(0 0 0 / 45%);
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover &__mask {
    opacity: 1;
  }

  &__mask-text {
    color: #fff;
    font-size: var(--td-font-size-body-small, 12px);
    text-align: center;
    padding: 0 4px;
  }
}
</style>
