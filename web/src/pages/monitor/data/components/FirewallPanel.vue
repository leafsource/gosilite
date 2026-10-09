<template>
  <div class="data-panel">
    <!-- 三指标一行 -->
    <div class="fw-stats">
      <div class="fw-card">
        <div class="fw-card__label">状态</div>
        <div class="fw-card__value--num" :class="{ 'is-on': data?.enabled }">
          {{ data?.enabled ? '已启用' : '已关闭' }}
        </div>
        <div class="fw-card__sub">观察模式 · 不拦截</div>
      </div>

      <div class="fw-card">
        <div class="fw-card__label">命中次数</div>
        <div class="fw-card__value fw-card__value--num">{{ data?.hitCount ?? 0 }}</div>
        <div class="fw-card__sub">有界环形缓冲内的计数</div>
      </div>

      <div class="fw-card">
        <div class="fw-card__label">规则</div>
        <div class="fw-card__value fw-card__value--tags">
          <t-tag v-for="r in data?.rules || []" :key="r" size="small" variant="outline">
            {{ r }}
          </t-tag>
        </div>
        <div class="fw-card__sub">命中只记样本，不阻断业务</div>
      </div>
    </div>

    <!-- 最近命中 -->
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
    />
  </div>
</template>

<script setup lang="ts">
import dayjs from 'dayjs';
import { computed, onMounted, ref } from 'vue';

import { getFirewall } from '@/api/system/monitor';
import type { FirewallResponse, FwHitDTO } from '@/api/system/monitorModel';
import STable from '@/components/s-table/index.vue';

const loading = ref(false);
const data = ref<FirewallResponse | null>(null);

const rows = computed(() => data.value?.recent || []);

const columns = [
  { colKey: 'rule', title: '规则', width: 180 },
  { colKey: 'sql', title: 'SQL', minWidth: 320, ellipsis: true },
  {
    colKey: 'at',
    title: '时间',
    width: 180,
    cell: (h: (...args: any[]) => any, { row }: { row: FwHitDTO }) =>
      h('span', dayjs(row.at * 1000).format('YYYY-MM-DD HH:mm:ss')),
  },
] as any;

async function load() {
  loading.value = true;
  try {
    data.value = await getFirewall();
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
}

.fw-card__label {
  font-size: 13px;
  color: #86909c;
  margin-bottom: 10px;
}

.fw-card__value {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 32px;
}

.fw-card__value--num {
  font-size: 28px;
  font-weight: 600;
  line-height: 1.15;
  color: var(--td-text-color-primary);
  font-variant-numeric: tabular-nums;

  &.is-on {
    color: var(--td-success-color);
  }
}

.fw-card__value--tags {
  gap: 6px;
}

.fw-card__sub {
  margin-top: 10px;
  font-size: 12px;
  color: #a9aeb8;
}
</style>
