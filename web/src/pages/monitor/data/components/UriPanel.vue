<template>
  <div class="data-panel">
    <!-- 三指标一行 -->
    <div class="fw-stats">
      <div class="fw-card">
        <div class="fw-card__label">路由数</div>
        <div class="fw-card__value fw-card__value--num">{{ data?.total ?? 0 }}</div>
        <div class="fw-card__sub">按路由模板聚合，无 query</div>
      </div>

      <div class="fw-card">
        <div class="fw-card__label">累计请求</div>
        <div class="fw-card__value fw-card__value--num">{{ totalReq }}</div>
        <div class="fw-card__sub">全部路由请求量之和</div>
      </div>

      <div class="fw-card" :class="{ 'fw-card--danger': totalErr > 0 }">
        <div class="fw-card__label">错误请求</div>
        <div class="fw-card__value fw-card__value--num" :class="{ 'is-err': totalErr > 0 }">
          {{ totalErr }}
        </div>
        <div class="fw-card__sub">状态码 ≥ 400 的请求</div>
      </div>
    </div>

    <!-- 路由明细 -->
    <s-table
      row-key="path"
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
      <template #errorCount="{ row }">
        <t-tag
          :theme="(row as UriStatDTO).errorCount > 0 ? 'danger' : 'default'"
          size="small"
          variant="light"
        >
          {{ (row as UriStatDTO).errorCount }}
        </t-tag>
      </template>
      <template #avgMs="{ row }">
        <span :class="avgClass((row as UriStatDTO).avgMs)">
          {{ (row as UriStatDTO).avgMs }} ms
        </span>
      </template>
    </s-table>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { getUriList } from '@/api/system/monitor';
import type { UriListResponse, UriStatDTO } from '@/api/system/monitorModel';
import STable from '@/components/s-table/index.vue';

const loading = ref(false);
const data = ref<UriListResponse | null>(null);

const rows = computed(() => data.value?.list || []);

const columns = [
  { colKey: 'method', title: '方法', width: 90 },
  { colKey: 'path', title: '路径', minWidth: 260, ellipsis: true },
  { colKey: 'count', title: '请求量', width: 110 },
  { colKey: 'errorCount', title: '错误', width: 90, cell: 'errorCount' },
  { colKey: 'avgMs', title: '平均耗时', width: 120, cell: 'avgMs' },
] as any;

const totalReq = computed(() => rows.value.reduce((s, r) => s + r.count, 0));
const totalErr = computed(() => rows.value.reduce((s, r) => s + r.errorCount, 0));

function avgClass(ms: number) {
  if (ms >= 500) return 'latency latency--bad';
  if (ms >= 200) return 'latency latency--warn';
  return 'latency';
}

async function load() {
  loading.value = true;
  try {
    data.value = await getUriList();
  } finally {
    loading.value = false;
  }
}

defineExpose({ load });
onMounted(load);
</script>

<style scoped lang="less">
.data-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.fw-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;

  @media (max-width: 900px) {
    grid-template-columns: 1fr;
  }
}

.fw-card {
  padding: 16px 18px 14px;
  background: var(--td-bg-color-container);
  border: 1px solid #e7e9ed;
  border-radius: 8px;

  &--danger {
    border-color: #fde0e0;
    background: #fff8f7;
  }
}

.fw-card__label {
  font-size: 13px;
  color: #86909c;
  margin-bottom: 10px;
}

.fw-card__value--num {
  font-size: 28px;
  font-weight: 600;
  line-height: 1.15;
  color: var(--td-text-color-primary);
  font-variant-numeric: tabular-nums;

  &.is-err {
    color: var(--td-error-color);
  }
}

.fw-card__sub {
  margin-top: 10px;
  font-size: 12px;
  color: #a9aeb8;
}

.latency {
  font-variant-numeric: tabular-nums;

  &--warn {
    color: var(--td-warning-color);
  }

  &--bad {
    color: var(--td-error-color);
    font-weight: 600;
  }
}
</style>
