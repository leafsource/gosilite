<template>
  <div class="data-panel">
    <s-table
      row-key="sql"
      size="small"
      :columns="columns"
      :data="rows as any"
      :total="rows.length"
      :loading="loading"
      :selectable="false"
      :paginated="false"
      :show-search="false"
      @refresh="load"
    >
      <template #maxMs="{ row }">
        <span :class="maxClass((row as SqlRow).maxMs)">
          {{ (row as SqlRow).maxMs }}
        </span>
      </template>
      <template #errorCount="{ row }">
        <span :class="{ 'is-err': (row as SqlRow).errorCount > 0 }">
          {{ (row as SqlRow).errorCount }}
        </span>
      </template>
    </s-table>
  </div>
</template>

<script setup lang="ts">
import { InfoCircleIcon } from 'tdesign-icons-vue-next';
import { Tooltip as TTooltip } from 'tdesign-vue-next';
import { computed, h, onMounted, ref } from 'vue';

import { getSqlMonitor } from '@/api/system/monitor';
import type { SqlAggDTO, SqlMonitorResponse } from '@/api/system/monitorModel';
import STable from '@/components/s-table/index.vue';

type SqlRow = SqlAggDTO;

const loading = ref(false);
const data = ref<SqlMonitorResponse | null>(null);

/** 模板聚合已按 count 降序；慢 SQL 由 max 高亮 */
const rows = computed(() => data.value?.list || []);

const columns = computed(() => [
  { colKey: 'sql', title: 'SQL 模板', minWidth: 320, ellipsis: true },
  { colKey: 'count', title: '执行次数', width: 110 },
  { colKey: 'avgMs', title: '平均耗时(ms)', width: 120 },
  {
    colKey: 'maxMs',
    title: () =>
      h('span', { class: 'th-with-tip' }, [
        h('span', '最大耗时(ms)'),
        h(
          TTooltip,
          { content: `慢 SQL 阈值 ${data.value?.slowThresholdMs ?? 200} ms` },
          {
            default: () => h('span', { class: 'th-tip' }, [h(InfoCircleIcon, { size: '14px' })]),
          },
        ),
      ]),
    width: 140,
    cell: 'maxMs',
  },
  { colKey: 'totalMs', title: '总耗时(ms)', width: 120 },
  { colKey: 'errorCount', title: '错误数', width: 90, cell: 'errorCount' },
  { colKey: 'lastMs', title: '最近耗时(ms)', width: 120 },
] as any);

function threshold() {
  return data.value?.slowThresholdMs ?? 200;
}

function maxClass(ms: number) {
  if (ms >= threshold()) return 'max-slow';
  return '';
}

async function load() {
  loading.value = true;
  try {
    data.value = await getSqlMonitor();
  } finally {
    loading.value = false;
  }
}

defineExpose({ load });
onMounted(load);
</script>

<style scoped lang="less">
:deep(.th-with-tip) {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

:deep(.th-tip) {
  display: inline-flex;
  align-items: center;
  color: var(--td-text-color-placeholder);
  cursor: help;

  &:hover {
    color: var(--td-brand-color);
  }
}

:deep(.max-slow) {
  color: var(--td-error-color);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

:deep(.is-err) {
  color: var(--td-error-color);
}
</style>
