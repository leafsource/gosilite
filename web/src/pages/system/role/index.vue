<template>
  <div class="role-page">
    <s-table
      row-key="id"
      :columns="columns as any"
      :data="list as any"
      :total="pagination.total"
      :page="pagination.current"
      :page-size="pagination.pageSize"
      :loading="loading"
      keyword-placeholder="角色名称 / 编码"
      @search="onSearch"
      @reset="onReset"
      @refresh="loadData"
      @page-change="onPageChange"
    >
      <template #toolbar>
        <t-button v-permission="'system:role:add'" theme="primary" @click="openCreate">新增角色</t-button>
      </template>
      <template #dataScope="{ row }">
        <t-tag theme="primary" variant="light">{{ dataScopeDict.labelOf(row.dataScope) }}</t-tag>
      </template>
      <template #state="{ row }">
        <t-tag :theme="row.state === 'enabled' ? 'success' : 'default'" variant="light">
          {{ stateDict.labelOf(row.state) || '-' }}
        </t-tag>
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:role:edit'" theme="primary" @click="openEdit(row)">编辑</t-link>
          <t-link v-permission="'system:role:assign'" theme="primary" @click="openMenuAssign(row)">分配菜单</t-link>
          <t-link v-permission="'system:role:assign'" theme="primary" @click="openDeptAssign(row)">数据部门</t-link>
          <t-link v-permission="'system:role:remove'" theme="danger" @click="openDelete(row)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 新增 / 编辑角色 -->
    <s-form-dialog
      v-model:visible="dialogVisible"
      :title="isEdit ? '编辑角色' : '新增角色'"
      :width="600"
      :loading="saving"
      @confirm="onSave"
    >
      <t-form :data="form" label-width="96px">
        <t-form-item label="角色名称" required>
          <t-input v-model="form.name" placeholder="请输入角色名称" />
        </t-form-item>
        <t-form-item label="角色编码" required>
          <t-input v-model="form.code" placeholder="如 admin / ops" />
        </t-form-item>
        <t-form-item label="描述">
          <t-textarea v-model="form.description" placeholder="角色描述" :autosize="{ minRows: 2, maxRows: 4 }" />
        </t-form-item>
        <t-form-item label="数据范围">
          <t-radio-group v-model="form.dataScope" :options="dataScopeOptions" />
        </t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="form.state" :options="stateOptions" />
        </t-form-item>
        <t-form-item label="排序">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 分配菜单 -->
    <t-drawer
      v-model:visible="menuDrawerVisible"
      :header="menuDrawerRole ? `分配菜单：${menuDrawerRole.name}` : '分配菜单'"
      size="large"
      :footer="false"
      destroy-on-close
    >
      <div class="role-page__drawer-toolbar">
        <t-input v-model="menuFilter" placeholder="搜索菜单" clearable />
        <t-button variant="outline" size="small" @click="menuCheckedIds = collectAllIds(menuTreeData)"> 全选 </t-button>
        <t-button variant="outline" size="small" @click="menuCheckedIds = []">清空</t-button>
      </div>
      <t-tree
        :data="filteredMenuTree"
        checkable
        value-mode="all"
        hover
        :expanded="menuExpanded"
        :value="menuCheckedIds"
        @expand="onMenuExpand"
        @change="onMenuCheck"
      />
      <div class="role-page__drawer-actions">
        <t-button variant="outline" @click="menuDrawerVisible = false">取消</t-button>
        <t-button theme="primary" :loading="menuSaving" @click="onSaveMenuAssign">保存</t-button>
      </div>
    </t-drawer>

    <!-- 分配数据部门 -->
    <t-drawer
      v-model:visible="deptDrawerVisible"
      :header="deptDrawerRole ? `数据部门：${deptDrawerRole.name}` : '数据部门'"
      size="large"
      :footer="false"
      destroy-on-close
    >
      <div class="role-page__drawer-toolbar">
        <t-input v-model="deptFilter" placeholder="搜索部门" clearable />
        <t-button variant="outline" size="small" @click="deptCheckedIds = collectAllIds(deptTreeData)"> 全选 </t-button>
        <t-button variant="outline" size="small" @click="deptCheckedIds = []">清空</t-button>
      </div>
      <t-tree
        :data="filteredDeptTree"
        checkable
        value-mode="all"
        hover
        :expanded="deptExpanded"
        :value="deptCheckedIds"
        @expand="onDeptExpand"
        @change="onDeptCheck"
      />
      <div class="role-page__drawer-actions">
        <t-button variant="outline" @click="deptDrawerVisible = false">取消</t-button>
        <t-button theme="primary" :loading="deptSaving" @click="onSaveDeptAssign">保存</t-button>
      </div>
    </t-drawer>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除角色"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { DeptNode, DeptTreeItem } from '@/api/system/dept';
import { fetchDeptTree } from '@/api/system/dept';
import type { MenuNode, MenuTreeItem } from '@/api/system/menu';
import { fetchMenuTree } from '@/api/system/menu';
import type { RoleDTO, RoleSaveParams } from '@/api/system/role';
import {
  assignRoleDept,
  assignRoleMenu,
  createRole,
  fetchRoleDepts,
  fetchRoleList,
  fetchRoleMenus,
  removeRole,
  updateRole,
} from '@/api/system/role';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import STable from '@/components/s-table/index.vue';
import { useConfigStore, useDict } from '@/store';

const stateDict = useDict('sys_state');
const stateOptions = stateDict.options;
const dataScopeDict = useDict('sys_data_scope');
const dataScopeOptions = computed(() =>
  dataScopeDict.options.value.map((o) => ({ label: o.label, value: Number(o.value) })),
);

interface TreeNode {
  value: string;
  label: string;
  children?: TreeNode[];
}

const columns: PrimaryTableCol[] = [
  { title: '角色名称', colKey: 'name', width: 160 },
  { title: '角色编码', colKey: 'code', width: 140 },
  { title: '数据范围', colKey: 'dataScope', width: 140 },
  { title: '状态', colKey: 'state', width: 100 },
  { title: '描述', colKey: 'description', minWidth: 200 },
  { title: '排序', colKey: 'orderNum', width: 80 },
  { title: '操作', colKey: 'op', width: 300, fixed: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const list = ref<RoleDTO[]>([]);
const query = ref({ keyword: '' });
const pagination = ref({ current: 1, pageSize: useConfigStore().pageDefault, total: 0, showJumper: true });

async function loadData() {
  loading.value = true;
  try {
    const res = await fetchRoleList({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      keyword: query.value.keyword || undefined,
    });
    list.value = res.list || [];
    pagination.value.total = res.total || 0;
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

function onSearch(params: Record<string, unknown>) {
  query.value.keyword = String(params.keyword ?? '');
  pagination.value.current = 1;
  loadData();
}

function onReset() {
  query.value.keyword = '';
  pagination.value.current = 1;
  loadData();
}

function onPageChange(page: number, pageSize: number) {
  pagination.value.current = page;
  pagination.value.pageSize = pageSize;
  loadData();
}

// ===== 新增 / 编辑 =====
const INITIAL_FORM: RoleSaveParams = {
  name: '',
  code: '',
  description: '',
  state: 'enabled',
  dataScope: 1,
  orderNum: 0,
};
const dialogVisible = ref(false);
const form = ref<RoleSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

function openCreate() {
  form.value = { ...INITIAL_FORM };
  dialogVisible.value = true;
}

function openEdit(row: RoleDTO) {
  form.value = {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    state: row.state === 'disabled' ? 'disabled' : 'enabled',
    dataScope: row.dataScope || 1,
    orderNum: row.orderNum,
  };
  dialogVisible.value = true;
}

async function onSave() {
  if (!form.value.name.trim() || !form.value.code.trim()) {
    MessagePlugin.warning('角色名称/编码不能为空');
    return;
  }
  saving.value = true;
  try {
    const payload: RoleSaveParams = { ...form.value, dataScope: form.value.dataScope || 1 };
    if (payload.id != null) {
      await updateRole(payload);
    } else {
      await createRole(payload);
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
const deleteTarget = ref<RoleDTO | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除角色「${deleteTarget.value.name}」吗？内置角色将被拒绝。` : '',
);

function openDelete(row: RoleDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeRole(deleteTarget.value.id);
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

// ===== 通用树工具 =====
function collectIds(nodes: TreeNode[], acc: string[] = []): string[] {
  nodes.forEach((node) => {
    acc.push(node.value);
    if (node.children) collectIds(node.children, acc);
  });
  return acc;
}

function collectAllIds(nodes: TreeNode[]): string[] {
  return collectIds(nodes, []);
}

function filterTree(nodes: TreeNode[], keyword: string): TreeNode[] {
  if (!keyword) return nodes;
  const lower = keyword.toLowerCase();
  const walk = (items: TreeNode[]): TreeNode[] =>
    items
      .map((item) => ({ ...item, children: item.children ? walk(item.children) : item.children }))
      .filter((item) => item.label.toLowerCase().includes(lower) || (item.children ? item.children.length > 0 : false));
  return walk(nodes);
}

// ===== 分配菜单 =====
const menuDrawerVisible = ref(false);
const menuDrawerRole = ref<RoleDTO | null>(null);
const menuSaving = ref(false);
const menuTreeData = ref<MenuTreeItem[]>([]);
const menuCheckedIds = ref<string[]>([]);
const menuExpanded = ref<string[]>([]);
const menuFilter = ref('');
const menuLoaded = ref(false);

const filteredMenuTree = computed(() => filterTree(menuTreeData.value, menuFilter.value.trim()));

function menuNodesToTree(nodes: MenuNode[]): MenuTreeItem[] {
  return nodes.map((node) => ({
    value: node.id,
    label: node.label,
    menuType: node.menuType,
    children: node.children && node.children.length ? menuNodesToTree(node.children) : undefined,
  }));
}

function onMenuExpand(value: Array<string | number>) {
  menuExpanded.value = value.map((item) => String(item));
}

function onMenuCheck(value: Array<string | number>) {
  menuCheckedIds.value = value.map((item) => String(item));
}

async function openMenuAssign(row: RoleDTO) {
  menuDrawerRole.value = row;
  menuCheckedIds.value = [];
  menuFilter.value = '';
  menuDrawerVisible.value = true;
  try {
    if (!menuLoaded.value) {
      const treeRes = await fetchMenuTree();
      menuTreeData.value = menuNodesToTree(treeRes.list || []);
      menuExpanded.value = collectIds(menuTreeData.value);
      menuLoaded.value = true;
    }
    const checkedRes = await fetchRoleMenus(row.id);
    menuCheckedIds.value = checkedRes.ids || [];
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  }
}

async function onSaveMenuAssign() {
  if (!menuDrawerRole.value) return;
  menuSaving.value = true;
  try {
    await assignRoleMenu({ roleId: menuDrawerRole.value.id, menuIds: menuCheckedIds.value });
    MessagePlugin.success('菜单分配成功');
    menuDrawerVisible.value = false;
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    menuSaving.value = false;
  }
}

// ===== 分配数据部门 =====
const deptDrawerVisible = ref(false);
const deptDrawerRole = ref<RoleDTO | null>(null);
const deptSaving = ref(false);
const deptTreeData = ref<DeptTreeItem[]>([]);
const deptCheckedIds = ref<string[]>([]);
const deptExpanded = ref<string[]>([]);
const deptFilter = ref('');
const deptLoaded = ref(false);

const filteredDeptTree = computed(() => filterTree(deptTreeData.value, deptFilter.value.trim()));

function deptNodesToTree(nodes: DeptNode[]): DeptTreeItem[] {
  return nodes.map((node) => ({
    value: node.id,
    label: node.label,
    children: node.children && node.children.length ? deptNodesToTree(node.children) : undefined,
  }));
}

function onDeptExpand(value: Array<string | number>) {
  deptExpanded.value = value.map((item) => String(item));
}

function onDeptCheck(value: Array<string | number>) {
  deptCheckedIds.value = value.map((item) => String(item));
}

async function openDeptAssign(row: RoleDTO) {
  deptDrawerRole.value = row;
  deptCheckedIds.value = [];
  deptFilter.value = '';
  deptDrawerVisible.value = true;
  try {
    if (!deptLoaded.value) {
      const treeRes = await fetchDeptTree();
      deptTreeData.value = deptNodesToTree(treeRes.list || []);
      deptExpanded.value = collectIds(deptTreeData.value);
      deptLoaded.value = true;
    }
    const checkedRes = await fetchRoleDepts(row.id);
    deptCheckedIds.value = checkedRes.ids || [];
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  }
}

async function onSaveDeptAssign() {
  if (!deptDrawerRole.value) return;
  deptSaving.value = true;
  try {
    await assignRoleDept({ roleId: deptDrawerRole.value.id, deptIds: deptCheckedIds.value });
    MessagePlugin.success('数据部门分配成功');
    deptDrawerVisible.value = false;
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    deptSaving.value = false;
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
.role-page {
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
}

.role-page__drawer-toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-m);
}

.role-page__drawer-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--td-comp-margin-s);
  margin-top: var(--td-comp-margin-xl);
  padding-top: var(--td-comp-margin-m);
  border-top: 1px solid var(--td-border-level-1-color);
}
</style>
