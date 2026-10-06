<template>
  <div class="menu-page">
    <div class="menu-page__toolbar">
      <t-button v-permission="'system:menu:add'" theme="primary" @click="openCreate()">新增菜单</t-button>
      <t-input v-model="filterText" placeholder="菜单名称 / 权限标识" clearable style="width: 240px" />
    </div>

    <!-- 树形表格需要 EnhancedTable（t-table 不含 tree 能力） -->
    <t-enhanced-table
      row-key="id"
      :data="filteredRows"
      :columns="columns"
      :loading="loading"
      :tree="treeConfig"
      hover
      vertical-align="middle"
    >
      <template #menuType="{ row }">
        <t-tag :theme="menuTypeTheme(row.menuType)" variant="light">{{ menuTypeLabel(row.menuType) }}</t-tag>
      </template>
      <template #visible="{ row }">
        <t-tag :theme="row.visible ? 'success' : 'default'" variant="light">
          {{ row.visible ? '显示' : '隐藏' }}
        </t-tag>
      </template>
      <template #status="{ row }">
        <t-tag :theme="row.status === 0 ? 'success' : 'danger'" variant="light">
          {{ statusDict.labelOf(row.status) }}
        </t-tag>
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:menu:add'" theme="primary" @click="openCreate(row.id)">新增子菜单</t-link>
          <t-link v-permission="'system:menu:edit'" theme="primary" @click="openEdit(row)">编辑</t-link>
          <t-link v-permission="'system:menu:remove'" theme="danger" @click="openDelete(row)">删除</t-link>
        </t-space>
      </template>
    </t-enhanced-table>

    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑菜单' : '新增菜单'"
      :width="640"
      :footer="false"
      destroy-on-close
    >
      <t-form :data="form" label-width="90px">
        <t-form-item label="上级菜单">
          <t-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            clearable
            placeholder="选择上级菜单（根目录留空）"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="菜单类型">
          <t-radio-group v-model="form.menuType" :options="menuTypeOptions" />
        </t-form-item>
        <t-form-item label="菜单名称" required>
          <t-input v-model="form.menuName" placeholder="请输入菜单名称" />
        </t-form-item>
        <t-form-item label="路由路径">
          <t-input v-model="form.path" placeholder="如 /system/user（按钮可留空）" />
        </t-form-item>
        <t-form-item label="组件路径">
          <t-input v-model="form.component" placeholder="如 system/user/index（目录/按钮可留空）" />
        </t-form-item>
        <t-form-item label="权限标识">
          <t-input v-model="form.perms" placeholder="如 system:user:list" />
        </t-form-item>
        <t-form-item label="图标">
          <t-input v-model="form.icon" placeholder="图标名称，如 organization" />
        </t-form-item>
        <t-form-item label="排序">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
        <t-form-item label="是否可见">
          <t-switch v-model="form.visible">
            <template #label>{{ form.visible ? '显示' : '隐藏' }}</template>
          </t-switch>
        </t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="form.status" :options="statusOptions" />
        </t-form-item>
        <t-form-item class="menu-page__form-actions">
          <t-button variant="outline" @click="dialogVisible = false">取消</t-button>
          <t-button theme="primary" :loading="saving" @click="onSave">保存</t-button>
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除菜单"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { EnhancedTable as TEnhancedTable, MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { MenuNode, MenuSaveParams, MenuTreeItem } from '@/api/system/menu';
import { createMenu, fetchMenuTree, removeMenu, updateMenu } from '@/api/system/menu';
import { useDict } from '@/store';

const statusDict = useDict('sys_status');
const statusOptions = computed(() => statusDict.options.value.map((o) => ({ label: o.label, value: Number(o.value) })));
const menuTypeDict = useDict('sys_menu_type');
const menuTypeOptions = menuTypeDict.options;

const columns: PrimaryTableCol[] = [
  { title: '菜单名称', colKey: 'label', minWidth: 220 },
  { title: '类型', colKey: 'menuType', width: 90 },
  { title: '权限标识', colKey: 'perms', width: 180 },
  { title: '路径', colKey: 'path', width: 160 },
  { title: '组件', colKey: 'component', width: 180 },
  { title: '排序', colKey: 'orderNum', width: 80 },
  { title: '可见', colKey: 'visible', width: 80 },
  { title: '状态', colKey: 'status', width: 80 },
  { title: '操作', colKey: 'op', width: 220, fixed: 'right' },
];

const treeConfig = {
  childrenKey: 'children',
  treeNodeColumnIndex: 0,
  defaultExpandAll: true,
};

const loading = ref(false);
const saving = ref(false);
const rows = ref<MenuNode[]>([]);
const filterText = ref('');

function normalize(nodes: MenuNode[]): MenuNode[] {
  return nodes.map((node) => ({
    ...node,
    children: node.children && node.children.length ? normalize(node.children) : undefined,
  }));
}

const filteredRows = computed<MenuNode[]>(() => {
  const keyword = filterText.value.trim().toLowerCase();
  if (!keyword) return rows.value;
  const walk = (nodes: MenuNode[]): MenuNode[] =>
    nodes
      .map((node) => ({ ...node, children: node.children ? walk(node.children) : node.children }))
      .filter(
        (node) =>
          node.label.toLowerCase().includes(keyword) ||
          (node.perms || '').toLowerCase().includes(keyword) ||
          (node.children ? node.children.length > 0 : false),
      );
  return walk(rows.value);
});

function menuTypeLabel(type: string): string {
  return menuTypeDict.labelOf(type) || '-';
}

function menuTypeTheme(type: string): 'primary' | 'default' | 'warning' {
  if (type === 'M') return 'primary';
  if (type === 'F') return 'warning';
  return 'default';
}

async function loadData() {
  loading.value = true;
  try {
    const res = await fetchMenuTree();
    rows.value = normalize(res.list || []);
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

// ===== 新增 / 编辑 =====
const INITIAL_FORM: MenuSaveParams = {
  parentId: 0,
  menuName: '',
  menuType: 'C',
  path: '',
  component: '',
  perms: '',
  icon: '',
  orderNum: 0,
  visible: true,
  status: 0,
};
const dialogVisible = ref(false);
const form = ref<MenuSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

const parentOptions = computed<MenuTreeItem[]>(() => {
  const source = toTreeItems(rows.value);
  if (form.value.id == null) return source;
  return pruneNode(source, form.value.id);
});

function toTreeItems(nodes: MenuNode[]): MenuTreeItem[] {
  return nodes.map((node) => ({
    value: node.id,
    label: node.label,
    menuType: node.menuType,
    children: node.children && node.children.length ? toTreeItems(node.children) : undefined,
  }));
}

function pruneNode(items: MenuTreeItem[], excludeId: number): MenuTreeItem[] {
  return items
    .filter((item) => item.value !== excludeId)
    .map((item) => ({ ...item, children: item.children ? pruneNode(item.children, excludeId) : undefined }));
}

function openCreate(parentId?: number) {
  form.value = { ...INITIAL_FORM, parentId: parentId ?? 0 };
  dialogVisible.value = true;
}

function openEdit(row: MenuNode) {
  form.value = {
    id: row.id,
    parentId: row.parentId,
    menuName: row.label,
    menuType: row.menuType,
    path: row.path,
    component: row.component,
    perms: row.perms,
    icon: row.icon,
    orderNum: row.orderNum,
    visible: row.visible,
    status: row.status,
  };
  dialogVisible.value = true;
}

async function onSave() {
  if (!form.value.menuName.trim()) {
    MessagePlugin.warning('请输入菜单名称');
    return;
  }
  saving.value = true;
  try {
    const payload: MenuSaveParams = { ...form.value, parentId: form.value.parentId || 0 };
    if (payload.id != null) {
      await updateMenu(payload);
    } else {
      await createMenu(payload);
    }
    MessagePlugin.success('保存成功');
    dialogVisible.value = false;
    await loadData();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 删除 =====
const confirmVisible = ref(false);
const deleteTarget = ref<MenuNode | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除菜单「${deleteTarget.value.label}」吗？存在子菜单时将被拒绝。` : '',
);

function openDelete(row: MenuNode) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeMenu(deleteTarget.value.id);
    MessagePlugin.success('删除成功');
    confirmVisible.value = false;
    deleteTarget.value = null;
    await loadData();
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
  loadData();
});
</script>
<style lang="less" scoped>
.menu-page {
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
}

.menu-page__toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-m);
  margin-bottom: var(--td-comp-margin-m);
}

.menu-page__form-actions {
  :deep(.t-form__controls) {
    justify-content: flex-end;
  }
}
</style>
