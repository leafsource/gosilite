<template>
  <div class="api-doc">
    <div v-if="loadFailed" class="api-doc__error">
      <t-empty description="Swagger 加载失败，请确认后端已启动">
        <t-button theme="primary" @click="reload">重试</t-button>
      </t-empty>
      <p class="api-doc__error-link">
        直接访问：
        <a :href="swaggerUrl" target="_blank">{{ swaggerUrl }}</a>
      </p>
    </div>
    <iframe v-else :key="frameKey" class="api-doc__frame" :src="swaggerUrl" title="Swagger UI" @load="onLoad" />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'ToolApi' });
import { ref } from 'vue';

/** 与 request 相同的 API 前缀（dev 走 Vite 代理 → 127.0.0.1:8888） */
const prefix = import.meta.env.VITE_API_URL_PREFIX || '/api';
const swaggerUrl = `${prefix}/swagger/index.html`;

const frameKey = ref(0);
const loadFailed = ref(false);

function reload() {
  loadFailed.value = false;
  frameKey.value += 1;
}

function onLoad() {
  try {
    const frame = document.querySelector('.api-doc__frame') as HTMLIFrameElement | null;
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
.api-doc {
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--td-comp-size-xxxl, 56px) - 100px);
  min-height: 480px;
  overflow: hidden;
  background: var(--td-bg-color-container);
  border: none;
  border-radius: 0;
}

.api-doc__frame {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  border: 0;
  background: #fff;
}

.api-doc__error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
}

.api-doc__error-link {
  font-size: 12px;
  color: var(--td-text-color-secondary);

  a {
    color: var(--td-brand-color);
  }
}
</style>
