<template>
  <div v-if="!isRefreshing" class="page-container">
    <router-view v-if="!isFramePage" v-slot="{ Component }">
      <!-- v-if 必须在 keep-alive 外：内层 v-if 为假时会插入注释节点，触发 KeepAlive 单子节点告警 -->
      <template v-if="Component">
        <!-- 不用 out-in：与 keep-alive 组合切换子路由时易出现空白帧 -->
        <transition name="fade">
          <keep-alive :include="aliveViews">
            <component :is="Component" :key="route.path" />
          </keep-alive>
        </transition>
      </template>
    </router-view>
    <frame-page v-show="isFramePage" />
  </div>

  <t-loading v-else />
</template>
<script setup lang="ts">
import isBoolean from 'lodash/isBoolean';
import isUndefined from 'lodash/isUndefined';
import type { ComputedRef } from 'vue';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

import FramePage from '@/layouts/frame/index.vue';
import { useTabsRouterStore } from '@/store';

const route = useRoute();
const tabsRouterStore = useTabsRouterStore();

const aliveViews = computed(() =>
  tabsRouterStore.tabRouters
    .filter((r) => {
      const keepAliveConfig = r.meta?.keepAlive;
      const shouldKeep =
        isUndefined(keepAliveConfig) || (isBoolean(keepAliveConfig) && keepAliveConfig);
      return Boolean(r.isAlive) && shouldKeep && typeof r.name === 'string' && r.name !== '';
    })
    .map((r) => r.name as string),
) as ComputedRef<string[]>;

const isRefreshing = computed(() => tabsRouterStore.refreshing);

const isFramePage = computed(() => !!route.meta?.frameSrc);
</script>
<style lang="less" scoped>
.page-container {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  height: 100%;
}

.page-container > * {
  min-width: 0;
}

.fade-leave-active,
.fade-enter-active {
  grid-area: 1 / 1;
  transition: opacity @anim-duration-slow @anim-time-fn-easing;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
