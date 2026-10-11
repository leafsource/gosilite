<template>
  <div class="docs-page">
    <div v-if="loadFailed" class="docs-page__error">
      <t-empty description="开发文档加载失败，请确认后端已启动且已构建文档站">
        <t-button theme="primary" @click="reload">重试</t-button>
      </t-empty>
      <p class="docs-page__error-link">
        直接访问：
        <a :href="docsUrl" target="_blank">{{ docsUrl }}</a>
        （未构建时先执行 script/build-docs.sh）
      </p>
    </div>
    <iframe v-else :key="frameKey" class="docs-page__frame" :src="docsUrl" title="开发文档" @load="onLoad" />
  </div>
</template>
<script setup lang="ts">
// keep-alive 缓存名 = genRouteName('/tool/docs') = ToolDocs
defineOptions({ name: 'ToolDocs' });
import { ref } from 'vue';

/** 同源相对路径：dev 走 Vite 代理 /docs → 8888，生产由 Hertz 直接托管 */
const docsUrl = '/docs/';

const frameKey = ref(0);
const loadFailed = ref(false);

function reload() {
  loadFailed.value = false;
  frameKey.value += 1;
}

function onLoad() {
  try {
    const frame = document.querySelector('.docs-page__frame') as HTMLIFrameElement | null;
    const doc = frame?.contentDocument;
    if (doc && doc.body && !doc.body.innerHTML.trim()) {
      loadFailed.value = true;
    }
  } catch {
    // 跨域读不到视为正常
  }
}
</script>
<style scoped lang="less">
.docs-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--td-comp-size-xxxl, 56px) - 100px);
  min-height: 480px;
  overflow: hidden;
  background: var(--td-bg-color-container);
  border: none;
  border-radius: 0;
}

.docs-page__frame {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  border: 0;
  background: #fff;
}

.docs-page__error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
}

.docs-page__error-link {
  font-size: 12px;
  color: var(--td-text-color-secondary);

  a {
    color: var(--td-brand-color);
  }
}
</style>
