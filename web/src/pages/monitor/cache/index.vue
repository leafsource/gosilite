<template>
  <div class="cache-monitor">
    <div class="cache-monitor__cols">
      <!-- 左：前缀 -->
      <section class="cache-monitor__panel">
        <header class="cache-monitor__head">
          <span class="cache-monitor__title">缓存前缀</span>
          <div class="cache-monitor__actions">
            <t-tooltip content="刷新">
              <t-button variant="text" shape="square" size="small" @click="loadPrefixes">
                <template #icon><refresh-icon /></template>
              </t-button>
            </t-tooltip>
          </div>
        </header>
        <div class="cache-monitor__body">
          <t-table
            row-key="prefix"
            size="small"
            :data="prefixes as any"
            :columns="prefixCols as any"
            :loading="prefixLoading"
            hover
            @row-click="onPrefixRow"
          >
            <template #remark="{ row }">
              <span v-if="(row as CachePrefixDTO).remark">{{ (row as CachePrefixDTO).remark }}</span>
              <span v-else class="cache-monitor__muted">—</span>
            </template>
            <template #ops="{ row }">
              <t-popconfirm
                content="清空该前缀下全部 key？高危操作请确认"
                theme="danger"
                @confirm="clearPrefix(row as CachePrefixDTO)"
              >
                <t-button
                  v-permission="'system:cachemonitor:delete'"
                  variant="text"
                  shape="square"
                  theme="danger"
                  size="small"
                >
                  <template #icon><delete-icon /></template>
                </t-button>
              </t-popconfirm>
            </template>
          </t-table>
        </div>
      </section>

      <!-- 中：键名 -->
      <section class="cache-monitor__panel">
        <header class="cache-monitor__head">
          <span class="cache-monitor__title">键名列表</span>
          <span class="cache-monitor__sub">{{ selectedPrefix || '请选择前缀' }}</span>
          <div class="cache-monitor__actions">
            <t-tooltip content="刷新">
              <t-button
                variant="text"
                shape="square"
                size="small"
                :disabled="!selectedPrefix"
                @click="loadKeys"
              >
                <template #icon><refresh-icon /></template>
              </t-button>
            </t-tooltip>
          </div>
        </header>
        <div class="cache-monitor__body">
          <t-table
            row-key="display"
            size="small"
            :data="keyRows as any"
            :columns="keyCols as any"
            :loading="keyLoading"
            hover
            @row-click="onKeyRow"
          >
            <template #name="{ row }">
              <span class="cache-monitor__mono">{{ (row as KeyRow).name }}</span>
            </template>
            <template #ops="{ row }">
              <t-popconfirm
                :content="(row as KeyRow).field ? '删除该 hash field？' : '删除该 key？'"
                theme="danger"
                @confirm="removeKey(row as KeyRow)"
              >
                <t-button
                  v-permission="'system:cachemonitor:delete'"
                  variant="text"
                  shape="square"
                  theme="danger"
                  size="small"
                >
                  <template #icon><delete-icon /></template>
                </t-button>
              </t-popconfirm>
            </template>
          </t-table>
        </div>
      </section>

      <!-- 右：内容 -->
      <section class="cache-monitor__panel cache-monitor__panel--value">
        <header class="cache-monitor__head">
          <span class="cache-monitor__title">缓存内容</span>
          <div class="cache-monitor__actions">
            <t-tooltip content="刷新">
              <t-button
                variant="text"
                shape="square"
                size="small"
                :disabled="!current"
                @click="reloadValue"
              >
                <template #icon><refresh-icon /></template>
              </t-button>
            </t-tooltip>
            <t-popconfirm
              v-if="current"
              :content="current.field ? '删除该 hash field？' : '删除该 key？'"
              theme="danger"
              @confirm="removeCurrent"
            >
              <t-button
                v-permission="'system:cachemonitor:delete'"
                variant="text"
                shape="square"
                theme="danger"
                size="small"
              >
                <template #icon><delete-icon /></template>
              </t-button>
            </t-popconfirm>
          </div>
        </header>
        <div class="cache-monitor__body cache-monitor__body--value">
          <t-form v-if="value" label-align="top" class="cache-monitor__form">
            <t-form-item label="键名">
              <t-input :model-value="value.displayKey || value.key" readonly />
            </t-form-item>
            <t-form-item label="完整 Key">
              <t-input :model-value="value.key" readonly />
            </t-form-item>
            <t-form-item label="类型">
              <t-input :model-value="value.type" readonly />
            </t-form-item>
            <t-form-item label="内容">
              <t-textarea
                :model-value="value.value"
                readonly
                :autosize="{ minRows: 12, maxRows: 24 }"
              />
            </t-form-item>
          </t-form>
          <t-empty v-else description="点击中间键名查看内容" />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">

defineOptions({ name: 'MonitorCache' });
import { DeleteIcon, RefreshIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  clearCachePrefix,
  deleteCacheKey,
  getCacheValue,
  listCacheKeys,
  listCachePrefixes,
} from '@/api/system/cacheMonitor';
import type {
  CacheKeyDTO,
  CachePrefixDTO,
  CacheValueResponse,
} from '@/api/system/cacheMonitorModel';

interface KeyRow {
  key: string;
  type: string;
  field: string;
  /** 中栏展示键名：hash → field，否则完整 key */
  name: string;
  display: string;
}

const prefixLoading = ref(false);
const keyLoading = ref(false);
const prefixes = ref<CachePrefixDTO[]>([]);
const keys = ref<CacheKeyDTO[]>([]);
const selectedPrefix = ref('');
const current = ref<KeyRow | null>(null);
const value = ref<CacheValueResponse | null>(null);

const keyRows = computed<KeyRow[]>(() =>
  keys.value.map((k) => {
    const field = k.field || '';
    return {
      key: k.key,
      type: k.type,
      field,
      name: field || k.key,
      display: `${k.key}::${field}`,
    };
  }),
);

const prefixCols = [
  { colKey: 'prefix', title: '前缀', minWidth: 180, ellipsis: true },
  { colKey: 'count', title: '数量', width: 64 },
  { colKey: 'remark', title: '备注', width: 120, cell: 'remark' },
  { colKey: 'ops', title: '操作', width: 56, cell: 'ops', align: 'center' as const },
];

const keyCols = [
  { colKey: 'name', title: '键名', minWidth: 200, ellipsis: true, cell: 'name' },
  { colKey: 'type', title: '类型', width: 72 },
  { colKey: 'ops', title: '操作', width: 56, cell: 'ops', align: 'center' as const },
];

async function loadPrefixes() {
  prefixLoading.value = true;
  try {
    const res = await listCachePrefixes();
    prefixes.value = res.list || [];
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    prefixLoading.value = false;
  }
}

async function loadKeys() {
  if (!selectedPrefix.value) return;
  keyLoading.value = true;
  try {
    const res = await listCacheKeys(selectedPrefix.value);
    keys.value = res.list || [];
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    keyLoading.value = false;
  }
}

async function loadValue(row: KeyRow) {
  current.value = row;
  try {
    value.value = await getCacheValue(row.key, row.field || undefined);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '读取失败');
    value.value = null;
  }
}

async function onPrefixRow(ctx: { row: unknown }) {
  const row = ctx.row as CachePrefixDTO;
  selectedPrefix.value = row.prefix;
  value.value = null;
  current.value = null;
  keys.value = [];
  await loadKeys();
}

async function onKeyRow(ctx: { row: unknown }) {
  await loadValue(ctx.row as KeyRow);
}

async function reloadValue() {
  if (current.value) await loadValue(current.value);
}

async function removeKey(row: KeyRow) {
  try {
    await deleteCacheKey(row.key, row.field || undefined);
    MessagePlugin.success('已删除');
    if (current.value && current.value.display === row.display) {
      value.value = null;
      current.value = null;
    }
    await Promise.all([loadKeys(), loadPrefixes()]);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  }
}

async function removeCurrent() {
  if (current.value) await removeKey(current.value);
}

async function clearPrefix(row: CachePrefixDTO) {
  try {
    const n = await clearCachePrefix(row.prefix);
    MessagePlugin.success(`已清理 ${n} 个 key`);
    if (selectedPrefix.value === row.prefix) {
      keys.value = [];
      value.value = null;
      current.value = null;
    }
    await loadPrefixes();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '清理失败');
  }
}

onMounted(loadPrefixes);
</script>

<style scoped lang="less">
/*
 * 高度链：卡片填满可视内容区（不触发外层 layout/body 滚动）；
 * 仅各栏 .cache-monitor__body（content）内部滚动，表头/工具条固定。
 */
.cache-monitor {
  display: flex;
  flex-direction: column;
  /* 头 xxxl + 页签约 40 + content padding 32 + 面包屑约 28 → 约 156px chrome */
  height: calc(100vh - var(--td-comp-size-xxxl, 56px) - 100px);
  min-height: 420px;
  max-height: calc(100vh - var(--td-comp-size-xxxl, 56px) - 72px);
  overflow: hidden;
}

.cache-monitor__cols {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) minmax(300px, 1.15fr) minmax(280px, 1.1fr);
  gap: 12px;
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  align-items: stretch;
  overflow: hidden;

  @media (max-width: 1200px) {
    grid-template-columns: 1fr;
    /* 窄屏允许整组上下滚动，仍不滚 body */
    overflow-y: auto;
  }
}

.cache-monitor__panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
  height: 100%;
  background: var(--td-bg-color-container);
  border: 1px solid #e7e9ed;
  border-radius: 8px;
  overflow: hidden; /* 卡片本身不滚 */
}

.cache-monitor__head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f1f3;
}

.cache-monitor__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.cache-monitor__sub {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cache-monitor__actions {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: auto;
}

/* 唯一滚动区：栏内 content */
.cache-monitor__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;

  :deep(.t-table) {
    border-top: none;
  }
}

.cache-monitor__body--value {
  overflow-y: auto;
}

.cache-monitor__form {
  padding: 12px;
}

.cache-monitor__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.cache-monitor__muted {
  color: var(--td-text-color-placeholder);
}
</style>
