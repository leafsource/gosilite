<template>
  <div class="s-table">
    <!-- 第 1 行：查询条件（查询后生效） -->
    <div class="s-table__search">
      <t-input
        v-model="keyword"
        class="s-table__keyword"
        type="search"
        :placeholder="keywordPlaceholder"
        clearable
        @enter="onSearch"
      />
      <template v-for="f in searchFields" :key="f.name">
        <t-select
          v-if="f.type === 'select'"
          :model-value="fieldModelValue(f.name)"
          :placeholder="f.placeholder ?? f.label"
          :options="f.options ?? []"
          clearable
          style="flex: 0 0 160px"
          @change="
            (v) => {
              extraParams[f.name] = v;
              onSearch();
            }
          "
        />
        <t-tree-select
          v-else-if="f.type === 'tree-select'"
          :model-value="fieldModelValue(f.name)"
          :data="f.treeData ?? []"
          :placeholder="f.placeholder ?? f.label"
          clearable
          style="flex: 0 0 200px"
          @change="
            (v) => {
              extraParams[f.name] = v;
              onSearch();
            }
          "
        />
        <t-input
          v-else
          :model-value="(extraParams[f.name] ?? '') as string"
          :placeholder="f.placeholder ?? f.label"
          clearable
          style="flex: 0 0 160px"
          @enter="onSearch"
          @change="
            (v) => {
              extraParams[f.name] = v;
            }
          "
        />
      </template>
      <div class="s-table__actions">
        <t-button theme="primary" @click="onSearch">查询</t-button>
        <t-button variant="outline" @click="onReset">重置</t-button>
      </div>
    </div>

    <!-- 第 2 行：批量操作（最左侧）+ 功能操作区（贴着 刷新 / 列设置） -->
    <div class="s-table__toolbar">
      <div v-if="$slots.batch && selectedRows.length" class="s-table__batch">
        <span class="s-table__batch-count">已选 {{ selectedRows.length }} 项</span>
        <t-link size="small" @click="clearSelection">取消选择</t-link>
        <slot name="batch" :selected="selectedRows" :selected-keys="selectedRowKeys" :clear="clearSelection" />
      </div>
      <div class="s-table__toolbar-right">
        <div class="s-table__toolbar-actions">
          <slot name="toolbar" />
        </div>
        <div class="s-table__toolbar-tools">
          <t-tooltip content="刷新">
            <t-button variant="text" shape="square" @click="emit('refresh')">
              <template #icon><refresh-icon /></template>
            </t-button>
          </t-tooltip>
          <t-popup v-model:visible="columnCtlVisible" trigger="click" placement="bottom-right">
            <span>
              <t-tooltip content="列设置">
                <t-button variant="text" shape="square">
                  <template #icon><setting-icon /></template>
                </t-button>
              </t-tooltip>
            </span>
            <template #content>
              <div class="s-table__column-panel">
                <div
                  v-for="k in columnOrder"
                  :key="k"
                  class="s-table__column-item"
                  :class="{
                    'is-dragging': dragKey === k,
                    'is-drop-target': dragOverKey === k && dragKey !== k,
                  }"
                  draggable="true"
                  @dragstart="onDragStart(k)"
                  @dragover.prevent="onDragOver(k)"
                  @drop.prevent="onDrop(k)"
                  @dragend="onDragEnd"
                >
                  <span class="s-table__column-grip"><move-icon size="16" /></span>
                  <t-checkbox :checked="columnVisible.includes(k)" @change="(v: any) => toggleColumn(k, Boolean(v))" />
                  <span class="s-table__column-name">{{ colTitle(k) }}</span>
                </div>
                <div v-if="!columnOrder.length" class="s-table__column-name" style="padding: 8px 12px">无列可配置</div>
              </div>
            </template>
          </t-popup>
        </div>
      </div>
    </div>

    <t-table
      :row-key="rowKey"
      :data="data as any"
      :columns="visibleColumns as any"
      :loading="loading"
      :pagination="pagination as TableProps['pagination']"
      :selected-row-keys="selectedRowKeys"
      hover
      @page-change="onPageChange"
      @select-change="onSelectChange"
    >
      <template v-for="(_, name) in $slots" #[name]="slotData">
        <slot :name="name" v-bind="slotData || {}" />
      </template>
    </t-table>
  </div>
</template>
<script setup lang="ts">
import './s-table.css';

import { MoveIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { computed, ref, watch } from 'vue';

import { useConfigStore } from '@/store';

import type { STableColumn, STableSearchField } from './sTable';

defineOptions({ name: 'STable' });

const props = withDefaults(
  defineProps<{
    rowKey?: string;
    columns: STableColumn[];
    data: Array<Record<string, unknown>>;
    total?: number;
    page?: number;
    pageSize?: number;
    loading?: boolean;
    keywordPlaceholder?: string;
    searchFields?: STableSearchField[];
    defaultVisibleKeys?: string[];
    selectable?: boolean;
  }>(),
  {
    rowKey: 'id',
    total: 0,
    page: 1,
    loading: false,
    keywordPlaceholder: '请输入关键字',
    searchFields: () => [],
    defaultVisibleKeys: undefined,
    selectable: true,
  },
);

const emit = defineEmits<{
  (e: 'search', params: Record<string, unknown>): void;
  (e: 'reset'): void;
  (e: 'refresh'): void;
  (e: 'page-change', page: number, pageSize: number): void;
  (e: 'selection-change', selected: Array<Record<string, unknown>>): void;
}>();

const configStore = useConfigStore();

const keyword = ref('');
const extraParams = ref<Record<string, unknown>>({});
const columnVisible = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const columnCtlVisible = ref(false);

const selectedRowKeys = ref<Array<string | number>>([]);
const selectedRows = ref<Array<Record<string, unknown>>>([]);

const dragKey = ref<string | null>(null);
const dragOverKey = ref<string | null>(null);

const colKey = (c: STableColumn): string => String((c as { colKey?: string }).colKey ?? '');

const colTitle = (k: string): string => {
  const hit = props.columns.find((c) => colKey(c) === k) as { title?: unknown } | undefined;
  return String(hit?.title || k);
};

// 模板中避免写 `as string | number`（`|` 会被 vue/no-deprecated-filter 误判为过滤器）
function fieldModelValue(name: string): string | number | undefined {
  return extraParams.value[name] as string | number | undefined;
}

function initVisible() {
  const keys = props.columns.map(colKey).filter(Boolean);
  columnOrder.value = props.defaultVisibleKeys?.length
    ? [...props.defaultVisibleKeys, ...keys.filter((k) => !props.defaultVisibleKeys!.includes(k))]
    : [...keys];
  columnVisible.value = props.defaultVisibleKeys?.length
    ? props.defaultVisibleKeys.filter((k) => keys.includes(k))
    : [...keys];
}

watch(
  () => props.columns,
  () => initVisible(),
  { immediate: true, deep: true },
);

const selectionColumn: STableColumn = { colKey: 'row-select', type: 'multiple', width: 64, fixed: 'left' };

const visibleColumns = computed(() => {
  const cols = columnOrder.value
    .map((k) => props.columns.find((c) => colKey(c) === k))
    .filter(
      (c): c is STableColumn =>
        Boolean(c) && (!colKey(c as STableColumn) || columnVisible.value.includes(colKey(c as STableColumn))),
    );
  return props.selectable ? [selectionColumn, ...cols] : cols;
});

function onSelectChange(keys: Array<string | number>, options?: { selectedRowData?: Array<Record<string, unknown>> }) {
  selectedRowKeys.value = keys;
  selectedRows.value = options?.selectedRowData ?? [];
  emit('selection-change', selectedRows.value);
}

function clearSelection() {
  if (!selectedRowKeys.value.length && !selectedRows.value.length) return;
  selectedRowKeys.value = [];
  selectedRows.value = [];
  emit('selection-change', []);
}

// 数据刷新/检索后，旧选择已不可靠，统一清空
watch(
  () => props.data,
  () => clearSelection(),
);

const pagination = computed(() => ({
  current: props.page,
  pageSize: props.pageSize ?? configStore.pageDefault,
  total: props.total,
  showJumper: true,
}));

function buildParams(): Record<string, unknown> {
  return { keyword: keyword.value || undefined, ...extraParams.value };
}

function onSearch() {
  emit('search', buildParams());
}

function onReset() {
  keyword.value = '';
  extraParams.value = {};
  emit('reset');
}

function onPageChange(ctx: { current?: number; pageSize?: number }) {
  emit('page-change', ctx.current ?? props.page, ctx.pageSize ?? props.pageSize ?? configStore.pageDefault);
}

function toggleColumn(k: string, on: boolean) {
  const set = new Set(columnVisible.value);
  if (on) set.add(k);
  else set.delete(k);
  columnVisible.value = columnOrder.value.filter((x) => set.has(x));
}

function onDragStart(k: string) {
  dragKey.value = k;
}

function onDragOver(k: string) {
  dragOverKey.value = k;
}

function onDrop(target: string) {
  const from = dragKey.value;
  dragKey.value = null;
  dragOverKey.value = null;
  if (!from || from === target) return;
  const order = [...columnOrder.value];
  const fi = order.indexOf(from);
  const ti = order.indexOf(target);
  if (fi < 0 || ti < 0) return;
  order.splice(fi, 1);
  order.splice(ti, 0, from);
  columnOrder.value = order;
  // 保持可见集跟随新顺序
  const set = new Set(columnVisible.value);
  columnVisible.value = order.filter((k) => set.has(k));
}

function onDragEnd() {
  dragKey.value = null;
  dragOverKey.value = null;
}

defineExpose({
  keyword,
  extraParams,
  buildParams,
  columnOrder,
  columnVisible,
  selectedRowKeys,
  selectedRows,
  clearSelection,
});
</script>
