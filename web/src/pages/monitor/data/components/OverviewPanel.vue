<template>
  <div class="overview-panel" v-loading="loading">
    <!-- ===== 运行时 ===== -->
    <section class="ov-section">
      <h3 class="ov-section__title">运行时（Web 应用 / Go 进程）</h3>
      <div class="ov-grid">
        <div v-for="item in runtimeCards" :key="item.key" class="ov-card">
          <div class="ov-card__label">{{ item.label }}</div>
          <div class="ov-card__value" :style="{ color: item.color }">{{ item.value }}</div>
          <div class="ov-card__sub">{{ item.sub }}</div>
        </div>
      </div>
    </section>

    <!-- ===== 数据源 ===== -->
    <section class="ov-section">
      <h3 class="ov-section__title">数据源</h3>
      <div class="ov-grid">
        <div v-for="item in dsCards" :key="item.key" class="ov-card">
          <div class="ov-card__label">{{ item.label }}</div>
          <div class="ov-card__value" :style="{ color: item.color }">{{ item.value }}</div>
          <div class="ov-card__sub">{{ item.sub }}</div>
        </div>
      </div>

      <!-- 连接池占用 -->
      <div v-if="datasource" class="ov-pool">
        <div class="ov-pool__head">
          <span class="ov-pool__title">连接池占用</span>
          <span class="ov-pool__ratio">{{ datasource.pgInUse }} / {{ datasource.pgMaxOpen }}</span>
        </div>
        <t-progress
          :percentage="poolPercent"
          :status="poolStatus"
          :label="false"
          class="ov-pool__bar"
        />
        <div class="ov-pool__foot">
          <div class="ov-pool__tags">
            <span class="ov-tag ov-tag--blue">
              PG Open {{ datasource.pgOpen }} · Idle {{ datasource.pgIdle }}
            </span>
            <span
              class="ov-tag"
              :class="datasource.redisUp ? 'ov-tag--green' : 'ov-tag--red'"
            >
              Redis {{ datasource.redisUp ? `正常 · ${datasource.redisLatencyMs}ms` : '不可用' }}
            </span>
          </div>
          <div class="ov-pool__meta">
            <template v-if="server">
              <span>{{ server.goVersion }}</span>
              <span class="ov-dot">·</span>
              <span>{{ server.numCPU }} 核</span>
              <span class="ov-dot">·</span>
              <span>PID {{ server.pid }}</span>
              <span class="ov-dot">·</span>
            </template>
            <span>检查于 {{ datasource.checkedAt }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 存储（OSS 配置逐条探活，spec system-oss 修订） ===== -->
    <section class="ov-section">
      <h3 class="ov-section__title">存储（对象存储配置）</h3>
      <div v-if="storageList.length" class="ov-grid">
        <div v-for="item in storageList" :key="item.configKey" class="ov-card">
          <div class="ov-card__label">
            {{ item.configKey }}
            <t-tag v-if="item.isDefault" theme="success" size="small" variant="light">默认</t-tag>
          </div>
          <div class="ov-card__value" :style="{ color: item.connected ? C.teal : '#f54a45' }">
            {{ item.connected ? '连通' : '异常' }}
          </div>
          <div class="ov-card__sub" :title="item.endpoint">
            {{ item.endpoint }} · {{ item.bucketName || '—' }}
          </div>
          <div v-if="item.message && !item.connected" class="ov-storage__error" :title="item.message">
            {{ item.message }}
          </div>
        </div>
      </div>
      <div v-else class="ov-storage__empty">暂无存储配置</div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { getDatasource, getServer, getStorageOverview } from '@/api/system/monitor';
import type { DatasourceDTO, ServerDTO, StorageOverviewItem } from '@/api/system/monitorModel';

const loading = ref(false);
const datasource = ref<DatasourceDTO | null>(null);
const server = ref<ServerDTO | null>(null);
const storageList = ref<StorageOverviewItem[]>([]);

/** 参考图色板 */
const C = {
  blue: '#2b6de8',
  teal: '#05a87a',
  orange: '#e8611a',
  purple: '#7c5cfc',
  dark: '#1d2129',
} as const;

function fmtBytes(n: number) {
  if (n === undefined || n === null) return '—';
  const units = ['B', 'KB', 'MB', 'GB'];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}

function fmtUptime(sec: number) {
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}天 ${h}时`;
  if (h > 0) return `${h}时 ${m}分`;
  return `${m}分`;
}

const runtimeCards = computed(() => {
  const s = server.value;
  return [
    {
      key: 'uptime',
      label: '运行时长',
      value: s ? fmtUptime(s.uptimeSec) : '—',
      sub: '进程启动至今',
      color: C.blue,
    },
    {
      key: 'goroutine',
      label: 'Goroutine',
      value: s ? String(s.goroutine) : '—',
      sub: s ? `GOMAXPROCS ${s.goMaxProcs}` : '',
      color: C.teal,
    },
    {
      key: 'heap',
      label: '堆内存',
      value: s ? fmtBytes(s.heapAlloc) : '—',
      sub: s ? `sys ${fmtBytes(s.sysBytes)} · idle ${fmtBytes(s.heapIdle)}` : '',
      color: C.orange,
    },
    {
      key: 'gc',
      label: 'GC 次数',
      value: s ? String(s.gcNum) : '—',
      sub: s ? `累计暂停 ${s.gcPauseTotalMs} ms` : '',
      color: C.purple,
    },
  ];
});

const dsCards = computed(() => {
  const d = datasource.value;
  return [
    {
      key: 'max',
      label: '最大连接',
      value: d ? String(d.pgMaxOpen) : '—',
      sub: 'MaxOpenConnections',
      color: C.blue,
    },
    {
      key: 'open',
      label: '当前连接',
      value: d ? String(d.pgOpen) : '—',
      sub: d ? `使用中 ${d.pgInUse} · 空闲 ${d.pgIdle}` : '',
      color: C.teal,
    },
    {
      key: 'wait',
      label: '等待次数',
      value: d ? String(d.pgWaitCount) : '—',
      sub: d ? `累计等待 ${d.pgWaitDurationMs} ms` : '',
      color: d && d.pgWaitCount > 0 ? C.orange : C.dark,
    },
    {
      key: 'closed',
      label: '连接回收',
      value: d ? String(d.pgMaxIdleClosed + d.pgMaxLifetimeClosed) : '—',
      sub: d ? `idle 回收 ${d.pgMaxIdleClosed} · lifetime 回收 ${d.pgMaxLifetimeClosed}` : '',
      color: C.dark,
    },
  ];
});

const poolPercent = computed(() => {
  const d = datasource.value;
  if (!d || !d.pgMaxOpen) return 0;
  return Math.min(100, Math.round((d.pgInUse / d.pgMaxOpen) * 100));
});

const poolStatus = computed(() => {
  if (poolPercent.value >= 90) return 'error' as const;
  if (poolPercent.value >= 70) return 'warning' as const;
  return undefined;
});

async function load() {
  loading.value = true;
  try {
    const [d, s, st] = await Promise.all([
      getDatasource(),
      getServer(),
      getStorageOverview().catch(() => ({ list: [] as StorageOverviewItem[] })),
    ]);
    datasource.value = d;
    server.value = s;
    storageList.value = st.list || [];
  } finally {
    loading.value = false;
  }
}

defineExpose({ load });
onMounted(load);
</script>

<style scoped lang="less">
.overview-panel {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.ov-section__title {
  margin: 0 0 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  letter-spacing: 0.01em;
}

.ov-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;

  @media (max-width: 1100px) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  @media (max-width: 640px) {
    grid-template-columns: 1fr;
  }
}

.ov-card {
  padding: 16px 18px 14px;
  background: #fff;
  border: 1px solid #e7e9ed;
  border-radius: 8px;
}

.ov-card__label {
  font-size: 13px;
  color: #86909c;
  line-height: 1.4;
  margin-bottom: 10px;
}

.ov-card__value {
  font-size: 28px;
  font-weight: 600;
  line-height: 1.15;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
}

.ov-card__sub {
  margin-top: 10px;
  font-size: 12px;
  color: #a9aeb8;
  line-height: 1.4;
  min-height: 17px;
}

/* ===== 连接池 ===== */
.ov-pool {
  margin-top: 14px;
  padding: 18px 20px 16px;
  background: #fff;
  border: 1px solid #e7e9ed;
  border-radius: 8px;
}

.ov-pool__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.ov-pool__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.ov-pool__ratio {
  font-size: 14px;
  font-weight: 600;
  color: #2b6de8;
  font-variant-numeric: tabular-nums;
}

.ov-pool__bar {
  margin-bottom: 14px;
}

.ov-pool__foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
}

.ov-pool__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ov-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  font-size: 12px;
  line-height: 1.6;
  border-radius: 4px;

  &--blue {
    color: #2b6de8;
    background: #e8f1ff;
  }

  &--green {
    color: #05a87a;
    background: #e8faf3;
  }

  &--red {
    color: #f53f3f;
    background: #ffece8;
  }
}

.ov-pool__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #a9aeb8;
}

.ov-dot {
  color: #c9cdd4;
}
</style>

<style scoped>
.ov-storage__error {
  overflow: hidden;
  margin-top: 4px;
  color: var(--td-error-color);
  font-size: 11px;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ov-storage__empty {
  padding: 12px 0;
  color: var(--td-text-color-placeholder);
  font-size: 13px;
}
</style>
