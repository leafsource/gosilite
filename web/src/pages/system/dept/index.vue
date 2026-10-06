<template>
  <div class="dept-page">
    <div class="dept-page__tree">
      <div class="dept-page__tree-toolbar">
        <t-input v-model="filterText" placeholder="搜索部门" clearable>
          <template #suffix-icon>
            <search-icon />
          </template>
        </t-input>
        <t-link theme="primary" @click="clearSelection">全部部门</t-link>
      </div>
      <t-tree
        :data="filteredTreeData"
        hover
        expand-on-click-node
        activable
        :expanded="expandedKeys"
        @expand="onExpand"
        @active="onTreeActive"
      />
    </div>

    <div class="dept-page__main">
      <div class="dept-page__toolbar">
        <span v-if="selectedId" class="dept-page__hint">
          当前：{{ selectedLabel }}
          <t-link theme="primary" @click="clearSelection">返回全部</t-link>
        </span>
      </div>

      <s-table
        row-key="id"
        :columns="columns as any"
        :data="displayList as any"
        :total="pagination.total"
        :page="pagination.current"
        :page-size="pagination.pageSize"
        :loading="loading"
        keyword-placeholder="部门名称"
        @search="onTableSearch"
        @reset="onTableReset"
        @refresh="loadAll"
        @page-change="onPageChange"
      >
        <template #toolbar>
          <t-button v-permission="'system:dept:add'" theme="primary" @click="openCreate()">新增部门</t-button>
        </template>
        <template #parentName="{ row }">
          {{ parentNameOf(row) }}
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 0 ? 'success' : 'danger'" variant="light">
            {{ statusDict.labelOf(row.status) }}
          </t-tag>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link v-permission="'system:dept:add'" theme="primary" @click="openCreate(row.id)">新增子部门</t-link>
            <t-link v-permission="'system:dept:edit'" theme="primary" @click="openEdit(row)">编辑</t-link>
            <t-link v-permission="'system:dept:remove'" theme="danger" @click="openDelete(row)">删除</t-link>
          </t-space>
        </template>
      </s-table>
    </div>

    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑部门' : '新增部门'"
      :width="560"
      :footer="false"
      destroy-on-close
    >
      <t-form :data="form" label-width="90px">
        <t-form-item label="上级部门">
          <t-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            clearable
            placeholder="选择上级部门（根部门留空）"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="部门名称" required>
          <t-input v-model="form.deptName" placeholder="请输入部门名称" />
        </t-form-item>
        <t-form-item label="排序">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="form.status" :options="statusOptions" />
        </t-form-item>
        <t-form-item class="dept-page__form-actions">
          <t-button variant="outline" @click="dialogVisible = false">取消</t-button>
          <t-button theme="primary" :loading="saving" @click="onSave">保存</t-button>
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除部门"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
import { SearchIcon } from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { DeptDTO, DeptNode, DeptSaveParams, DeptTreeItem } from '@/api/system/dept';
import { createDept, fetchDeptList, fetchDeptTree, removeDept, updateDept } from '@/api/system/dept';
import STable from '@/components/s-table/index.vue';
import { useConfigStore, useDict } from '@/store';

const statusDict = useDict('sys_status');
const statusOptions = computed(() => statusDict.options.value.map((o) => ({ label: o.label, value: Number(o.value) })));

type DeptRow = DeptDTO;

const columns: PrimaryTableCol[] = [
  { title: '部门名称', colKey: 'deptName', minWidth: 200 },
  { title: '上级部门', colKey: 'parentName', width: 160 },
  { title: '排序', colKey: 'orderNum', width: 100 },
  { title: '状态', colKey: 'status', width: 100 },
  { title: '操作', colKey: 'op', width: 240, fixed: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const treeData = ref<DeptNode[]>([]);
const list = ref<DeptRow[]>([]);
const selectedId = ref<number | null>(null);
const filterText = ref('');
const tableKeyword = ref('');
const pagination = ref({ current: 1, pageSize: useConfigStore().pageDefault, total: 0, showJumper: true });

const expandedKeys = ref<number[]>([]);

function onExpand(value: Array<string | number>) {
  expandedKeys.value = value.map((item) => Number(item));
}

function collectIds(nodes: DeptNode[], acc: number[] = []): number[] {
  nodes.forEach((node) => {
    acc.push(node.id);
    if (node.children) collectIds(node.children, acc);
  });
  return acc;
}

const toTreeItem = (node: DeptNode): DeptTreeItem => ({
  value: node.id,
  label: node.label,
  children: (node.children || []).map(toTreeItem),
});

const filteredTreeData = computed<DeptNode[]>(() => {
  const keyword = filterText.value.trim();
  if (!keyword) return treeData.value;
  const walk = (nodes: DeptNode[]): DeptNode[] =>
    nodes
      .map((node) => ({ ...node, children: node.children ? walk(node.children) : node.children }))
      .filter((node) => node.label.includes(keyword) || (node.children ? node.children.length > 0 : false));
  return walk(treeData.value);
});

const selectedLabel = computed(() => {
  if (selectedId.value == null) return '';
  const found = findNode(treeData.value, selectedId.value);
  return found ? found.label : '';
});

function findNode(nodes: DeptNode[], id: number): DeptNode | null {
  for (const node of nodes) {
    if (node.id === id) return node;
    if (node.children) {
      const hit = findNode(node.children, id);
      if (hit) return hit;
    }
  }
  return null;
}

const displayList = computed(() => {
  let rows = list.value;
  if (selectedId.value != null) {
    const selected = String(selectedId.value);
    rows = rows.filter((row) => row.id === selectedId.value || row.ancestors.split(',').includes(selected));
  }
  const kw = tableKeyword.value.trim();
  if (kw) {
    rows = rows.filter((row) => row.deptName.includes(kw));
  }
  return rows;
});

const nameMap = computed(() => {
  const map: Record<number, string> = {};
  list.value.forEach((row) => {
    map[row.id] = row.deptName;
  });
  return map;
});

function parentNameOf(row: DeptRow): string {
  if (row.parentId === 0) return '根部门';
  return nameMap.value[row.parentId] || String(row.parentId);
}

function onPageChange(page: number, pageSize: number) {
  pagination.value.current = page;
  pagination.value.pageSize = pageSize;
  loadList();
}

function onTableSearch(params: Record<string, unknown>) {
  tableKeyword.value = String(params.keyword ?? '');
  pagination.value.current = 1;
  loadList();
}

function onTableReset() {
  tableKeyword.value = '';
  pagination.value.current = 1;
  loadList();
}

function onTreeActive(value: Array<string | number>) {
  selectedId.value = typeof value[0] === 'number' ? value[0] : Number(value[0] ?? 0) || null;
  pagination.value.current = 1;
}

function clearSelection() {
  selectedId.value = null;
  pagination.value.current = 1;
}

async function loadTree() {
  const res = await fetchDeptTree();
  treeData.value = res.list || [];
  expandedKeys.value = collectIds(treeData.value);
}

async function loadList() {
  loading.value = true;
  try {
    const res = await fetchDeptList();
    list.value = res.list || [];
    pagination.value.total = displayList.value.length;
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

async function loadAll() {
  loading.value = true;
  try {
    await Promise.all([loadTree(), loadList()]);
    pagination.value.total = displayList.value.length;
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

// ===== 新增 / 编辑 =====
const INITIAL_FORM: DeptSaveParams = { parentId: 0, deptName: '', orderNum: 0, status: 0 };
const dialogVisible = ref(false);
const form = ref<DeptSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

const parentOptions = computed<DeptTreeItem[]>(() => {
  const source = treeData.value.map(toTreeItem);
  if (form.value.id == null) return source;
  return pruneNode(source, form.value.id);
});

function pruneNode(items: DeptTreeItem[], excludeId: number): DeptTreeItem[] {
  return items
    .filter((item) => item.value !== excludeId)
    .map((item) => ({ ...item, children: item.children ? pruneNode(item.children, excludeId) : undefined }));
}

function openCreate(parentId?: number) {
  form.value = { ...INITIAL_FORM, parentId: parentId ?? 0 };
  dialogVisible.value = true;
}

function openEdit(row: DeptRow) {
  form.value = {
    id: row.id,
    parentId: row.parentId,
    deptName: row.deptName,
    orderNum: row.orderNum,
    status: row.status,
  };
  dialogVisible.value = true;
}

async function onSave() {
  if (!form.value.deptName.trim()) {
    MessagePlugin.warning('请输入部门名称');
    return;
  }
  saving.value = true;
  try {
    if (form.value.id != null) {
      await updateDept({ ...form.value });
    } else {
      await createDept({ ...form.value });
    }
    MessagePlugin.success('保存成功');
    dialogVisible.value = false;
    await loadAll();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 删除 =====
const confirmVisible = ref(false);
const deleteTarget = ref<DeptRow | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除部门「${deleteTarget.value.deptName}」吗？存在子部门或用户时将被拒绝。` : '',
);

function openDelete(row: DeptRow) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeDept(deleteTarget.value.id);
    MessagePlugin.success('删除成功');
    confirmVisible.value = false;
    deleteTarget.value = null;
    await loadAll();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

onMounted(() => {
  loadAll();
});
</script>
<style lang="less" scoped>
.dept-page {
  display: flex;
  gap: var(--td-comp-margin-xl);
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
  min-height: 480px;
}

.dept-page__tree {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--td-border-level-1-color);
  padding-right: var(--td-comp-margin-xl);
}

.dept-page__tree-toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-m);
}

.dept-page__main {
  flex: 1;
  min-width: 0;
}

.dept-page__toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-m);
  margin-bottom: var(--td-comp-margin-m);
}

.dept-page__hint {
  color: var(--td-text-color-secondary);
}

.dept-page__form-actions {
  :deep(.t-form__controls) {
    justify-content: flex-end;
  }
}
</style>
