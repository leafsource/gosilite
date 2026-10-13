<template>
  <div class="content-card monitor-data">
    <div class="monitor-data__nav">
      <t-radio-group v-model="tab" variant="default-filled" @change="onTabChange">
        <t-radio-button value="overview">总览</t-radio-button>
        <t-radio-button value="sql">SQL 监控</t-radio-button>
        <t-radio-button value="firewall">SQL 防火墙</t-radio-button>
        <t-radio-button value="uri">URI 监控</t-radio-button>
      </t-radio-group>
    </div>

    <div class="monitor-data__body">
      <OverviewPanel v-if="tab === 'overview'" ref="overviewRef" />
      <SqlPanel v-else-if="tab === 'sql' && mounted.sql" ref="sqlRef" />
      <FirewallPanel v-else-if="tab === 'firewall' && mounted.firewall" ref="fwRef" />
      <UriPanel v-else-if="tab === 'uri' && mounted.uri" ref="uriRef" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';

import FirewallPanel from './components/FirewallPanel.vue';
import OverviewPanel from './components/OverviewPanel.vue';
import SqlPanel from './components/SqlPanel.vue';
import UriPanel from './components/UriPanel.vue';

defineOptions({
  name: 'MonitorData', });

type TabKey = 'overview' | 'sql' | 'firewall' | 'uri';

const tab = ref<TabKey>('overview');

/** 首次进入才挂载子面板，避免启动时并发 4 个请求 */
const mounted = reactive({
  overview: true,
  sql: false,
  firewall: false,
  uri: false,
});

const overviewRef = ref<InstanceType<typeof OverviewPanel> | null>(null);
const sqlRef = ref<InstanceType<typeof SqlPanel> | null>(null);
const fwRef = ref<InstanceType<typeof FirewallPanel> | null>(null);
const uriRef = ref<InstanceType<typeof UriPanel> | null>(null);

function onTabChange(value: unknown) {
  const key = String(value) as TabKey;
  if (key in mounted) {
    (mounted as Record<string, boolean>)[key] = true;
  }
}
</script>

<style scoped lang="less">
.monitor-data {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.monitor-data__nav {
  margin-bottom: 16px;
  padding-bottom: 4px;
}

.monitor-data__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;

  :deep(.overview-panel),
  :deep(.data-panel) {
    flex: 1 1 auto;
    min-height: 0;
  }
}
</style>
