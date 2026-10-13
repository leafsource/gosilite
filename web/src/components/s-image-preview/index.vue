<template>
  <!-- 完全对齐 ruoyi-tdesign image-preview：t-image-viewer(trigger) 包裹 t-image(hover 遮罩+点击放大) -->
  <t-image-viewer
    :key="props.src"
    :default-index="0"
    :mode="mode"
    :viewer-scale="viewerScale"
    :images="realPreviewSrcList"
    :z-index="zIndex"
    :title="title"
  >
    <template #trigger="{ open }">
      <t-image
        :key="realSrc"
        :src="realSrc"
        :style="{
          display: 'inline-block',
          width: realWidth,
          height: realHeight,
        }"
        :alt="alt"
        :fit="fit"
        :shape="shape"
        :position="position"
        :referrerpolicy="referrerpolicy"
        class="s-image-preview"
        :class="{ s_image_animation: animation }"
        :lazy="lazy"
        :gallery="realPreviewSrcList.length > 1"
        overlay-trigger="hover"
        @mouseenter="hover = true"
        @mouseleave="hover = false"
      >
        <!-- 插槽始终注册（v-if 与 v-slot 同用会导致插槽不生效），条件放插槽内部 -->
        <template #overlayContent>
          <div v-if="preview" class="s-image-preview__overlay" @click.stop="open()">
            <t-tag style="background: transparent; color: #fff">
              <slot name="previewSlot">
                <t-icon name="browse" size="16" />
                <span class="s-image-preview__text">{{ previewText }}</span>
              </slot>
            </t-tag>
          </div>
        </template>
        <template #error>
          <t-icon name="image-error" size="24" style="color: var(--td-text-color-placeholder)" />
        </template>
      </t-image>
    </template>
  </t-image-viewer>
</template>

<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import type { TdImageProps, TdImageViewerProps } from 'tdesign-vue-next';

defineOptions({ name: 'SImagePreview' });

const props = withDefaults(
  defineProps<{
    /** 显示地址（支持逗号分隔多图 http URL；id 模式未接入——本仓无 listByIds，富文本场景接入时补） */
    src?: string;
    /** 是否可预览 */
    preview?: boolean;
    /** 预览地址，为空则使用 src */
    previewSrc?: string;
    /** 预览文本（previewSlot 插槽会覆盖它） */
    previewText?: string;
    width?: number | string;
    height?: number | string;
    lazy?: boolean;
    animation?: boolean;
    zIndex?: number;
    alt?: string;
    fit?: TdImageProps['fit'];
    shape?: TdImageProps['shape'];
    position?: TdImageProps['position'];
    referrerpolicy?: TdImageProps['referrerpolicy'];
    mode?: TdImageViewerProps['mode'];
    viewerScale?: TdImageViewerProps['viewerScale'];
    title?: string;
  }>(),
  {
    src: '',
    preview: true,
    previewSrc: '',
    previewText: '预览',
    width: '',
    height: '',
    lazy: true,
    animation: true,
    zIndex: 4100,
    alt: undefined,
    fit: 'cover',
    shape: 'round',
    position: 'center',
    referrerpolicy: 'strict-origin-when-cross-origin',
    mode: 'modal',
    viewerScale: undefined,
    title: undefined,
  },
);

const realSrc = ref('');
const realPreviewSrcList = ref<string[]>([]);
const hover = ref(false);
const scale = computed(() => (hover.value ? 1.1 : 1));

const splitHttp = /,(?=http)/;

watchEffect(() => {
  const src = props.src;
  const previewSrc = props.previewSrc || props.src;
  // 仅 http 模式（逗号分隔多图）；id 集中查询模式未接入（见 props.src 注释）
  realSrc.value = src ? src.split(splitHttp)[0] : '';
  realPreviewSrcList.value = previewSrc ? previewSrc.split(splitHttp) : [];
});

const realWidth = computed(() => (typeof props.width === 'string' ? props.width : `${props.width}px`));
const realHeight = computed(() => (typeof props.height === 'string' ? props.height : `${props.height}px`));
</script>

<style lang="less" scoped>
.s_image_animation:deep {
  img,
  .s-image-preview__overlay {
    transition: all 0.3s;
    transform: scale(v-bind(scale));
  }
}

.s-image-preview {
  cursor: unset;

  :deep(img) {
    vertical-align: top;
  }

  .s-image-preview__overlay {
    background: rgb(0 0 0 / 40%);
    color: #fff;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: inherit;
    cursor: pointer;
    container: overlay-container / size;

    .s-image-preview__text {
      display: inline;
    }
  }

  @container overlay-container (width <= 50px) {
    .s-image-preview__text {
      display: none;
    }

    .t-tag .t-icon {
      margin-right: 0;
    }
  }
}

:deep(.t-image__overlay-content) {
  border-radius: inherit;
}
</style>
