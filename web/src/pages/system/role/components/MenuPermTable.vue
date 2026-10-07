<template>
  <div class="menu-perm-table">
    <div class="menu-perm-table__toolbar">
      <t-input
        v-model="keyword"
        class="menu-perm-table__search"
        placeholder="搜索菜单 / 权限名称 / 权限码"
        clearable
        :prefix="renderSearchIcon"
      />
      <t-button variant="outline" size="small" @click="checkAll">全选</t-button>
      <t-button variant="outline" size="small" theme="default" @click="clearAll">清空</t-button>
      <div class="menu-perm-table__spacer" />
      <t-button variant="text" size="small" @click="expandAll">展开全部</t-button>
      <t-button variant="text" size="small" @click="foldAll">收起</t-button>
      <span class="menu-perm-table__stats">
        已选 <b>{{ selectedRowKeys.length }}</b
        >/{{ totalRows }} 菜单 · <b>{{ permChecked.size }}</b
        >/{{ totalPerms }} 权限
      </span>
    </div>

    <div class="menu-perm-table__table">
      <t-enhanced-table
        :key="tableKey"
        row-key="id"
        :columns="columns"
        :data="filteredRows"
        :selected-row-keys="selectedRowKeys"
        :expanded-tree-nodes="expandedTreeNodes"
        :tree="treeConfig"
        max-height="100%"
        hover
        @select-change="onSelectChange"
        @expanded-tree-nodes-change="onExpandedTreeNodesChange"
      >
        <template #menuName="{ row }">
          <div class="menu-cell" :class="`menu-cell--${row.menuType.toLowerCase()}`">
            <t-icon v-if="row.icon" :name="row.icon" class="menu-cell__icon menu-cell__icon--menu" />
            <folder-icon v-else-if="row.menuType === 'M'" class="menu-cell__icon menu-cell__icon--dir" />
            <file-icon v-else class="menu-cell__icon" />
            <span class="menu-cell__name" :title="row.menuName">{{ row.menuName }}</span>
          </div>
        </template>
        <template #menuType="{ row }">
          <t-tag :theme="typeTheme(row.menuType)" variant="light">{{ typeLabel(row.menuType) }}</t-tag>
        </template>
        <template #perms="{ row }">
          <div v-if="row.perms.length" class="perm-list">
            <t-checkbox
              v-for="p in row.perms"
              :key="p.id"
              class="perm-chip"
              :checked="permChecked.has(p.id)"
              @change="(checked: boolean) => onPermChange(row, p, checked)"
            >
              <span class="perm-chip__name">{{ p.name }}</span>
            </t-checkbox>
          </div>
          <span v-else class="perm-empty">—</span>
        </template>
        <template #empty>
          <div class="menu-perm-table__empty">无匹配菜单</div>
        </template>
      </t-enhanced-table>
    </div>
  </div>
</template>
<script setup lang="ts">
import { FileIcon, FolderIcon, SearchIcon } from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { computed, h, ref, watch } from 'vue';

import type { RoleAuthMenuNode, RoleId } from '@/api/system/role';

interface PermItem {
  id: string;
  rawId: RoleId;
  name: string;
  code: string;
}

interface AuthorizeMenuRow {
  id: string;
  rawId: RoleId;
  menuName: string;
  menuType: string;
  icon: string;
  children: AuthorizeMenuRow[];
  perms: PermItem[];
}

interface RowEntry {
  row: AuthorizeMenuRow;
  parentId: string | null;
}

const props = defineProps<{ menus?: RoleAuthMenuNode[] }>();
const emit = defineEmits<{ (e: 'stats', payload: { menus: number; perms: number }): void }>();

/** 收集抽屉保存所需的菜单 id：选中行 + 勾选权限（祖先闭包由服务端补全） */
function collect(): RoleId[] {
  const ids: RoleId[] = [];
  for (const id of selectedRowKeys.value) {
    const entry = index.value.get(id);
    if (entry) ids.push(entry.row.rawId);
  }
  for (const pid of permChecked.value) {
    const entry = permIndex.value.get(pid);
    if (entry) ids.push(entry.rawId);
  }
  return ids;
}

defineExpose({ collect });

// ===== 模型 =====
const rows = ref<AuthorizeMenuRow[]>([]);
const index = ref(new Map<string, RowEntry>());
const permIndex = ref(new Map<string, PermItem>());
const selectedRowKeys = ref<string[]>([]);
const permChecked = ref(new Set<string>());
const expandedTreeNodes = ref<string[]>([]);
const keyword = ref('');
const tableKey = ref(0);

// checkStrictly 必须显式 false：TDesign 仅在此字面量下执行父子级联（useTreeSelect.handleSelect）
const treeConfig = { childrenKey: 'children', treeNodeColumnIndex: 1, checkStrictly: false };

const columns: PrimaryTableCol[] = [
  { colKey: 'row-selection', type: 'multiple', width: 48 },
  { title: '菜单名称', colKey: 'menuName', minWidth: 260 },
  { title: '菜单类型', colKey: 'menuType', width: 110 },
  { title: '权限标识', colKey: 'perms', minWidth: 320 },
];

/** 后端已组装好 F 提升与 checked 归一化，这里只做状态渲染映射 */
function buildModel(menus: RoleAuthMenuNode[]) {
  const nextIndex = new Map<string, RowEntry>();
  const nextPermIndex = new Map<string, PermItem>();
  const roots: AuthorizeMenuRow[] = [];
  const selected = new Set<string>();
  const checkedPerms = new Set<string>();

  function walk(nodes: RoleAuthMenuNode[] | undefined, parentRow: AuthorizeMenuRow | null) {
    for (const node of nodes || []) {
      const row: AuthorizeMenuRow = {
        id: String(node.id),
        rawId: node.id,
        menuName: node.menuName,
        menuType: node.menuType,
        icon: node.icon || '',
        children: [],
        perms: [],
      };
      nextIndex.set(row.id, { row, parentId: parentRow ? parentRow.id : null });
      if (node.checked) selected.add(row.id);
      for (const p of node.perms || []) {
        const perm: PermItem = { id: String(p.id), rawId: p.id, name: p.menuName, code: p.perms };
        row.perms.push(perm);
        nextPermIndex.set(perm.id, perm);
        if (p.checked) checkedPerms.add(perm.id);
      }
      if (parentRow) parentRow.children.push(row);
      else roots.push(row);
      walk(node.children, row);
    }
  }
  walk(menus, null);

  rows.value = roots;
  index.value = nextIndex;
  permIndex.value = nextPermIndex;
  selectedRowKeys.value = [...selected];
  permChecked.value = checkedPerms;
  expandedTreeNodes.value = flatRows.value.map((r) => r.id);
  tableKey.value += 1;
  emitStats();
}

const flatRows = computed(() => {
  const out: AuthorizeMenuRow[] = [];
  const walk = (list: AuthorizeMenuRow[]) => {
    for (const r of list) {
      out.push(r);
      walk(r.children);
    }
  };
  walk(rows.value);
  return out;
});

const totalRows = computed(() => flatRows.value.length);
const totalPerms = computed(() => permIndex.value.size);

// ===== 过滤 =====
const filteredRows = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return rows.value;
  const keep = new Set<string>();
  const markSubtree = (row: AuthorizeMenuRow) => {
    keep.add(row.id);
    row.children.forEach(markSubtree);
  };
  const markAncestors = (id: string) => {
    let cur: string | null = id;
    while (cur && !keep.has(cur)) {
      keep.add(cur);
      cur = index.value.get(cur)?.parentId ?? null;
    }
  };
  for (const row of flatRows.value) {
    const hitName = row.menuName.toLowerCase().includes(kw);
    const hitPerm = row.perms.some(
      (p) => p.name.toLowerCase().includes(kw) || (p.code && p.code.toLowerCase().includes(kw)),
    );
    if (hitName || hitPerm) {
      markSubtree(row);
      markAncestors(row.id);
    }
  }
  const filter = (list: AuthorizeMenuRow[]): AuthorizeMenuRow[] =>
    list.filter((r) => keep.has(r.id)).map((r) => ({ ...r, children: filter(r.children) }));
  return filter(rows.value);
});

// 过滤或数据变化后重放全展开，避免收起的行在搜索结果里不可见
watch(filteredRows, (val) => {
  expandedTreeNodes.value = collectIds(val);
});

// ===== 展开 =====
function collectIds(list: AuthorizeMenuRow[], acc: string[] = []): string[] {
  for (const r of list) {
    acc.push(r.id);
    collectIds(r.children, acc);
  }
  return acc;
}

function onExpandedTreeNodesChange(keys: Array<string | number>) {
  expandedTreeNodes.value = keys.map((k) => String(k));
}

function expandAll() {
  expandedTreeNodes.value = collectIds(filteredRows.value);
}

function foldAll() {
  expandedTreeNodes.value = [];
}

// ===== 行勾选联动 =====
function subtreeRows(row: AuthorizeMenuRow, acc: AuthorizeMenuRow[] = []): AuthorizeMenuRow[] {
  acc.push(row);
  row.children.forEach((c) => subtreeRows(c, acc));
  return acc;
}

function subtreeHasSelected(row: AuthorizeMenuRow, selected: Set<string>): boolean {
  return subtreeRows(row).some((r) => selected.has(r.id));
}

function onSelectChange(keys: Array<string | number>) {
  const next = new Set(keys.map((k) => String(k)));
  // 全树闭包：可见勾选行的全部子孙行补进选中（保证过滤视图下也落到全量状态）
  for (const row of flatRows.value) {
    if (next.has(row.id)) subtreeRows(row).forEach((r) => next.add(r.id));
  }
  const prev = new Set(selectedRowKeys.value);
  const newlyOn: AuthorizeMenuRow[] = [];
  const newlyOff: AuthorizeMenuRow[] = [];
  for (const row of flatRows.value) {
    const on = next.has(row.id);
    if (on && !prev.has(row.id)) newlyOn.push(row);
    if (!on && prev.has(row.id)) newlyOff.push(row);
  }
  for (const row of newlyOn) {
    subtreeRows(row).forEach((r) => r.perms.forEach((p) => permChecked.value.add(p.id)));
  }
  for (const row of newlyOff) {
    if (!subtreeHasSelected(row, next)) {
      subtreeRows(row).forEach((r) => r.perms.forEach((p) => permChecked.value.delete(p.id)));
    }
  }
  selectedRowKeys.value = [...next];
  emitStats();
}

// ===== 权限勾选联动 =====
function onPermChange(row: AuthorizeMenuRow, perm: PermItem, checked: boolean) {
  if (checked) {
    permChecked.value.add(perm.id);
    if (!selectedRowKeys.value.includes(row.id)) {
      selectedRowKeys.value = [...selectedRowKeys.value, row.id];
    }
  } else {
    permChecked.value.delete(perm.id);
  }
  emitStats();
}

// ===== 批量 =====
function checkAll() {
  selectedRowKeys.value = flatRows.value.map((r) => r.id);
  permChecked.value = new Set(permIndex.value.keys());
  emitStats();
}

function clearAll() {
  selectedRowKeys.value = [];
  permChecked.value = new Set();
  emitStats();
}

// ===== 统计 =====
function emitStats() {
  emit('stats', { menus: selectedRowKeys.value.length, perms: permChecked.value.size });
}

watch(
  () => props.menus,
  (menus) => {
    if (!menus) return;
    buildModel(menus);
  },
  { immediate: true },
);

// ===== 展示辅助 =====
const menuTypeDictLabels = new Map([
  ['M', '目录'],
  ['C', '菜单'],
  ['F', '按钮'],
]);

function typeLabel(type: string): string {
  return menuTypeDictLabels.get(type) || type;
}

function typeTheme(type: string): 'primary' | 'default' | 'warning' {
  if (type === 'M') return 'primary';
  if (type === 'F') return 'warning';
  return 'default';
}

function renderSearchIcon() {
  return h(SearchIcon, { size: '16px' });
}
</script>
<style lang="less" scoped>
.menu-perm-table {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.menu-perm-table__toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-m);
}

.menu-perm-table__search {
  width: 280px;
}

.menu-perm-table__spacer {
  flex: 1;
}

.menu-perm-table__table {
  flex: 1;
  min-height: 0;

  // 吃满容器高度，令 max-height=100% 可解析 → 触发 TDesign 固定表头（表体内滚）
  :deep(.t-table) {
    height: 100%;
  }
}

.menu-perm-table__stats {
  color: var(--td-text-color-secondary);
  font-size: 12px;
  white-space: nowrap;

  b {
    color: var(--td-brand-color);
    font-weight: 600;
  }
}

.menu-perm-table__empty {
  padding: 32px 0;
  color: var(--td-text-color-placeholder);
  text-align: center;
}

.menu-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;

  &__icon {
    flex: none;
    color: var(--td-text-color-placeholder);

    &--dir,
    &--menu {
      color: var(--td-brand-color);
    }
  }

  &__name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.perm-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 4px 16px;
}

.perm-chip {
  align-items: center;
  padding: 2px 0;
}

.perm-chip__name {
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.perm-empty {
  color: var(--td-text-color-placeholder);
}
</style>
