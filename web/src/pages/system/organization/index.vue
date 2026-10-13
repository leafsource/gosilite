<template>
  <div class="org-page">
    <!-- 左：部门树 -->
    <div class="org-page__tree">
      <div class="org-page__tree-toolbar">
        <t-input v-model="filterText" placeholder="搜索部门" clearable>
          <template #suffix-icon>
            <search-icon />
          </template>
        </t-input>
        <span v-permission="'system:dept:add'">
          <t-tooltip content="新增部门">
            <t-button variant="outline" shape="square" @click="openCreate()">
              <add-icon />
            </t-button>
          </t-tooltip>
        </span>
      </div>
      <t-tree
        :data="filteredTreeData"
        :keys="{ value: 'id', label: 'label' }"
        hover
        activable
        :expanded="expandedKeys"
        :draggable="!filterText.trim()"
        :allow-drop="allowDrop"
        @expand="onDeptExpand"
        @active="onDeptActive"
        @drop="onTreeDrop"
      >
        <template #operations="{ node }">
          <t-dropdown trigger="click" :options="deptMenuOptions(node)" @click="(opt) => onDeptMenu(opt, node)">
            <t-button variant="text" shape="square" size="small">
              <more-icon />
            </t-button>
          </t-dropdown>
        </template>
      </t-tree>
    </div>

    <!-- 右：用户管理 -->
    <div class="org-page__main">
      <s-table
        row-key="id"
        :columns="columns as any"
        :data="list as any"
        :total="pagination.total"
        :page="pagination.current"
        :page-size="pagination.pageSize"
        :loading="loading"
        keyword-placeholder="账号/姓名"
        :search-fields="searchFields"
        @search="onSearch"
        @reset="onReset"
        @refresh="loadData"
        @page-change="onPageChange"
      >
        <template #toolbar>
          <t-button v-permission="'system:user:add'" theme="primary" @click="openCreateUser">新增用户</t-button>
        </template>
        <template #username="{ row }">
          <t-space>
            <span>{{ row.username }}</span>
            <t-tag v-if="isDeptLeader(row)" theme="primary" variant="light" size="small">负责人</t-tag>
          </t-space>
        </template>
        <template #deptName="{ row }">
          {{ row.deptName || '-' }}
        </template>
        <template #state="{ row }">
          <s-tag-dict v-if="row.state" dict-type="sys_state" :value="row.state" />
          <span v-else>-</span>
        </template>
        <template #loginSecurity="{ row }">
          <t-tag v-if="row.loginPermanentlyLocked" theme="danger" variant="light">永久锁定</t-tag>
          <t-tag v-else-if="row.loginFailureCount > 0" theme="warning" variant="light">
            失败 {{ row.loginFailureCount }} 次
          </t-tag>
          <span v-else>正常</span>
        </template>
        <template #op="{ row }">
          <t-dropdown
            v-if="userMenuVisible()"
            trigger="click"
            :options="userMenuOptions(row)"
            @click="(opt) => onUserMenu(opt, row)"
          >
            <t-button variant="text" shape="square" size="small">
              <more-icon />
            </t-button>
          </t-dropdown>
        </template>
      </s-table>

      <s-form-drawer
        v-model:visible="drawerVisible"
        :title="isEdit ? '编辑用户' : '新增用户'"
        size="large"
        :loading="saving"
        @confirm="onSaveUser"
      >
        <t-form :data="form" label-width="96px">
          <t-form-item label="账号" required>
            <t-input v-model="form.account" placeholder="登录账号" />
          </t-form-item>
          <t-form-item label="姓名" required>
            <t-input v-model="form.username" placeholder="用户姓名" />
          </t-form-item>
          <t-form-item :label="isEdit ? '重置密码' : '密码'">
            <t-input
              v-model="form.password"
              type="password"
              :placeholder="isEdit ? '重置密码将解除永久锁定；不修改请留空' : '请输入初始密码'"
              clearable
            />
          </t-form-item>
          <t-form-item label="手机号">
            <t-input v-model="form.mobilePhone" placeholder="手机号" />
          </t-form-item>
          <t-form-item label="邮箱">
            <t-input v-model="form.email" placeholder="邮箱" />
          </t-form-item>
          <t-form-item label="所属部门">
            <t-tree-select
              :model-value="form.deptId && form.deptId !== '0' ? form.deptId : undefined"
              :data="deptOptions"
              clearable
              placeholder="选择部门"
              style="width: 100%"
              @change="
                (v) => {
                  form.deptId = v == null || v === '' ? '0' : String(v);
                }
              "
            />
          </t-form-item>
          <t-form-item label="状态">
            <s-radio-dict v-model="form.state" dict-type="sys_state" />
          </t-form-item>
          <t-form-item label="角色">
            <t-select v-model="form.roleIds" multiple placeholder="选择角色" style="width: 100%">
              <t-option v-for="item in roleOptions" :key="item.value" :value="item.value" :label="item.label" />
            </t-select>
          </t-form-item>
          <t-form-item label="岗位">
            <t-select v-model="form.postIds" multiple placeholder="选择岗位" style="width: 100%">
              <t-option v-for="item in postOptions" :key="item.value" :value="item.value" :label="item.label" />
            </t-select>
          </t-form-item>
        </t-form>
      </s-form-drawer>

      <t-dialog
        v-model:visible="confirmVisible"
        header="删除用户"
        :body="confirmBody"
        :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
        @confirm="onConfirmDeleteUser"
      />
    </div>

    <!-- 部门：新增（仅名称；默认根部门、排序自动排最后） -->
    <s-form-dialog
      v-model:visible="dialogVisible"
      title="新增部门"
      :width="420"
      :loading="saving"
      @confirm="onSaveDept"
    >
      <t-form :data="form2" label-width="90px">
        <t-form-item label="部门名称" required>
          <t-input v-model="form2.deptName" placeholder="请输入部门名称" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 部门：修改名称 -->
    <s-form-dialog
      v-model:visible="renameDialog"
      title="修改名称"
      :width="420"
      :loading="saving"
      @confirm="onSaveRename"
    >
      <t-form :data="renameForm" label-width="90px">
        <t-form-item label="部门名称" required>
          <t-input v-model="renameForm.deptName" placeholder="请输入部门名称" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 部门：设置负责人 -->
    <s-form-dialog
      v-model:visible="leaderDialog"
      :title="leaderTarget ? `设置负责人：${leaderTarget.label}` : '设置负责人'"
      :width="420"
      :loading="saving"
      @confirm="onSaveLeader"
    >
      <t-form label-width="90px">
        <t-form-item label="负责人">
          <t-select
            v-model="leaderUserId"
            :options="leaderUserOptions"
            clearable
            filterable
            placeholder="选择负责人（留空清除）"
            style="width: 100%"
          />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 部门：删除确认 -->
    <t-dialog
      v-model:visible="deptConfirmVisible"
      header="删除部门"
      :body="deptConfirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDeleteDept"
    />
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemContacts' });
import { AddIcon, MoreIcon, SearchIcon } from 'tdesign-icons-vue-next';
import type { DropdownOption, PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { DeptNode, DeptSaveParams, DeptTreeItem } from '@/api/system/dept';
import { createDept, fetchDeptTree, removeDept, setDeptLeader, updateDept } from '@/api/system/dept';
import type { PostDTO } from '@/api/system/post';
import { fetchPostList } from '@/api/system/post';
import type { RoleDTO } from '@/api/system/role';
import { fetchRoleList } from '@/api/system/role';
import type { AdminUserDTO, AdminUserSaveParams } from '@/api/system/user';
import { createUser, fetchUserList, removeUser, updateUser, updateUserStatus } from '@/api/system/user';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { hasPerm } from '@/directives/permission';
import { useConfigStore, useDict } from '@/store';

const stateDict = useDict('sys_state');

const columns: PrimaryTableCol[] = [
  { title: '账号', colKey: 'account', width: 140 },
  { title: '姓名', colKey: 'username', width: 180, cell: 'username' },
  { title: '所属部门', colKey: 'deptName', width: 160, cell: 'deptName' },
  { title: '手机号', colKey: 'mobilePhone', width: 140 },
  { title: '邮箱', colKey: 'email', minWidth: 200 },
  { title: '状态', colKey: 'state', width: 100 },
  { title: '登录安全', colKey: 'loginSecurity', width: 110, cell: 'loginSecurity' },
  { title: '操作', colKey: 'op', width: 64, fixed: 'right' },
];

// ===== 用户列表状态 =====
const loading = ref(false);
const saving = ref(false);
const list = ref<AdminUserDTO[]>([]);
const query = ref<{ keyword: string; state: string | undefined }>({ keyword: '', state: undefined });
const pagination = ref({ current: 1, pageSize: useConfigStore().pageDefault, total: 0, showJumper: true });

const searchFields = computed<STableSearchField[]>(() => [
  {
    name: 'state',
    label: '状态',
    type: 'select',
    options: stateDict.options.value,
  },
]);

// ===== 部门树状态（左栏 = 用户筛选器） =====
const treeData = ref<DeptNode[]>([]);
const filterText = ref('');
const expandedKeys = ref<Array<string | number>>([]);
const selectedDeptId = ref<string | null>(null);

// 表单公共数据（用户抽屉 + 部门弹窗）
const deptOptions = ref<DeptTreeItem[]>([]);
const roleOptions = ref<{ label: string; value: string }[]>([]);
const postOptions = ref<{ label: string; value: string }[]>([]);

// 部门 → 负责人用户 id（用于列表“负责人”标记）
const leaderByDept = computed(() => {
  const map = new Map<string, string>();
  const walk = (nodes: DeptNode[]) => {
    nodes.forEach((n) => {
      if (n.leaderId && n.leaderId !== '0') map.set(n.id, n.leaderId);
      if (n.children) walk(n.children);
    });
  };
  walk(treeData.value);
  return map;
});

function isDeptLeader(row: AdminUserDTO): boolean {
  return !!row.deptId && row.deptId !== '0' && leaderByDept.value.get(row.deptId) === row.id;
}

function findNode(nodes: DeptNode[], id: string): DeptNode | null {
  for (const node of nodes) {
    if (node.id === id) return node;
    if (node.children) {
      const hit = findNode(node.children, id);
      if (hit) return hit;
    }
  }
  return null;
}

function toTreeItem(node: DeptNode): DeptTreeItem {
  return {
    value: node.id,
    label: node.label,
    children: (node.children || []).map(toTreeItem),
  };
}

function collectIds(nodes: DeptNode[], acc: string[] = []): string[] {
  nodes.forEach((node) => {
    acc.push(node.id);
    if (node.children) collectIds(node.children, acc);
  });
  return acc;
}

const filteredTreeData = computed<DeptNode[]>(() => {
  const keyword = filterText.value.trim();
  if (!keyword) return treeData.value;
  const walk = (nodes: DeptNode[]): DeptNode[] =>
    nodes
      .map((node) => ({ ...node, children: node.children ? walk(node.children) : node.children }))
      .filter((node) => node.label.includes(keyword) || (node.children ? node.children.length > 0 : false));
  return walk(treeData.value);
});

async function loadTree() {
  const res = await fetchDeptTree();
  treeData.value = res.list || [];
  expandedKeys.value = collectIds(treeData.value);
  // 用户抽屉「所属部门」下拉与树同源
  deptOptions.value = treeData.value.map(toTreeItem);
}

function onDeptExpand(value: Array<string | number>) {
  expandedKeys.value = value.map((item) => String(item));
}

function onDeptActive(value: Array<string | number>) {
  const v = value[0];
  selectedDeptId.value = v === undefined || v === '' ? null : String(v);
  pagination.value.current = 1;
  loadData();
}

// ===== 树节点 more 菜单 =====
function deptMenuOptions(node: { data: DeptNode }): DropdownOption[] {
  const data = node.data;
  const options: DropdownOption[] = [
    { content: `部门ID：${data.id}`, value: 'info', disabled: true },
    { content: `负责人：${data.leaderName || '未设置'}`, value: 'leaderInfo', disabled: true },
    { content: '', value: 'divider', divider: true },
  ];
  if (hasPerm('system:dept:add')) {
    options.push({ content: '添加子部门', value: 'addChild' });
  }
  if (hasPerm('system:dept:edit')) {
    options.push({ content: '修改名称', value: 'rename' });
    options.push({ content: '设置负责人', value: 'setLeader' });
  }
  if (hasPerm('system:dept:remove')) {
    options.push({ content: '删除', value: 'delete', theme: 'error' });
  }
  return options;
}

function onDeptMenu(option: DropdownOption, node: { data: DeptNode }) {
  const data = node.data;
  switch (String(option.value)) {
    case 'addChild':
      openCreate();
      break;
    case 'rename':
      openRename(data);
      break;
    case 'setLeader':
      openLeader(data);
      break;
    case 'delete':
      openDeleteDept(data);
      break;
    default:
      break;
  }
}

// ===== 树拖拽：排序（before/after）与修改上级（inside） =====
function collectSubtreeIds(node: DeptNode, acc: Set<string> = new Set()): Set<string> {
  acc.add(node.id);
  (node.children || []).forEach((c) => collectSubtreeIds(c, acc));
  return acc;
}

interface DropNodeLike {
  data: Record<string, unknown>;
}

function allowDrop(ctx: { dragNode: DropNodeLike; dropNode: DropNodeLike }): boolean {
  const drag = ctx.dragNode.data as unknown as DeptNode;
  const drop = ctx.dropNode.data as unknown as DeptNode;
  if (drag.id === drop.id) return false;
  // 拖入自己子树（含其前后位置）都会成环
  return !collectSubtreeIds(drag).has(drop.id);
}

// dropPosition: -1 拖到目标前 / 0 拖成目标子级（排末尾）/ 1 拖到目标后
async function onTreeDrop(ctx: { dragNode: DropNodeLike; dropNode: DropNodeLike; dropPosition: number }) {
  const drag = ctx.dragNode.data as unknown as DeptNode;
  const drop = ctx.dropNode.data as unknown as DeptNode;
  if (drag.id === drop.id) return;

  let newParentId: string;
  let refId: string | null = null;
  let placeAfter = false;
  if (ctx.dropPosition === 0) {
    newParentId = drop.id; // 成为目标子级 → 追加最后
  } else {
    newParentId = drop.parentId;
    refId = drop.id;
    placeAfter = ctx.dropPosition === 1;
  }

  const siblings = newParentId === '0' ? treeData.value : (findNode(treeData.value, newParentId)?.children ?? []);
  const order = siblings.filter((n) => n.id !== drag.id).map((n) => n.id);
  if (refId != null) {
    const idx = order.indexOf(refId);
    if (idx < 0) return;
    order.splice(placeAfter ? idx + 1 : idx, 0, drag.id);
  } else {
    order.push(drag.id);
  }

  saving.value = true;
  try {
    for (let i = 0; i < order.length; i += 1) {
      const id = order[i];
      const node = id === drag.id ? drag : findNode(treeData.value, id);
      if (!node) continue;
      const parentId = id === drag.id ? newParentId : node.parentId;
      if (parentId === node.parentId && node.orderNum === i + 1) continue;
      await updateDept({ id, parentId, deptName: node.label, orderNum: i + 1, status: node.status });
    }
    await loadTree();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 用户行 more 菜单（编辑 / 启停 / 删除） =====
function userMenuVisible(): boolean {
  return hasPerm('system:user:edit') || hasPerm('system:user:remove');
}

function userMenuOptions(row: AdminUserDTO): DropdownOption[] {
  const options: DropdownOption[] = [];
  if (hasPerm('system:user:edit')) {
    options.push({ content: '编辑', value: 'edit' });
    options.push({
      content: row.state === 'enabled' ? stateDict.labelOf('disabled') : stateDict.labelOf('enabled'),
      value: 'toggle',
    });
  }
  if (hasPerm('system:user:remove')) {
    options.push({ content: '删除', value: 'delete', theme: 'error' });
  }
  return options;
}

function onUserMenu(option: DropdownOption, row: AdminUserDTO) {
  switch (String(option.value)) {
    case 'edit':
      openEditUser(row);
      break;
    case 'toggle':
      toggleState(row);
      break;
    case 'delete':
      openDeleteUser(row);
      break;
    default:
      break;
  }
}

// ===== 用户列表 =====
async function loadData() {
  loading.value = true;
  try {
    const res = await fetchUserList({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      keyword: query.value.keyword || undefined,
      deptId: selectedDeptId.value || undefined,
      state: query.value.state || undefined,
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
  query.value = {
    keyword: String(params.keyword ?? ''),
    state: (params.state ?? undefined) as string | undefined,
  };
  pagination.value.current = 1;
  loadData();
}

function onReset() {
  query.value = { keyword: '', state: undefined };
  pagination.value.current = 1;
  loadData();
}

function onPageChange(page: number, pageSize: number) {
  pagination.value.current = page;
  pagination.value.pageSize = pageSize;
  loadData();
}

async function loadOptions() {
  try {
    const [roleRes, postRes] = await Promise.all([
      fetchRoleList({ page: 1, pageSize: 100 }),
      fetchPostList({ page: 1, pageSize: 100 }),
    ]);
    roleOptions.value = (roleRes.list || []).map((role: RoleDTO) => ({
      label: `${role.name}（${role.code}）`,
      value: role.id,
    }));
    postOptions.value = (postRes.list || []).map((post: PostDTO) => ({
      label: `${post.postName}（${post.postCode}）`,
      value: post.id,
    }));
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  }
}

// ===== 用户新增 / 编辑 =====
const INITIAL_FORM: AdminUserSaveParams = {
  account: '',
  username: '',
  password: '',
  mobilePhone: '',
  email: '',
  deptId: '0',
  state: 'enabled',
  roleIds: [],
  postIds: [],
};
const drawerVisible = ref(false);
const form = ref<AdminUserSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

function openCreateUser() {
  form.value = { ...INITIAL_FORM, roleIds: [], postIds: [] };
  drawerVisible.value = true;
}

function openEditUser(row: AdminUserDTO) {
  form.value = {
    id: row.id,
    account: row.account,
    username: row.username,
    password: '',
    mobilePhone: row.mobilePhone,
    email: row.email,
    deptId: row.deptId,
    state: row.state === 'disabled' ? 'disabled' : 'enabled',
    roleIds: [...(row.roleIds || [])],
    postIds: [...(row.postIds || [])],
  };
  drawerVisible.value = true;
}

async function onSaveUser() {
  if (!form.value.account.trim() || !form.value.username.trim()) {
    MessagePlugin.warning('账号/姓名不能为空');
    return;
  }
  saving.value = true;
  try {
    const payload: AdminUserSaveParams = {
      ...form.value,
      deptId: form.value.deptId || '0',
      roleIds: form.value.roleIds || [],
      postIds: form.value.postIds || [],
    };
    if (payload.id != null) {
      await updateUser(payload);
    } else {
      await createUser(payload);
    }
    MessagePlugin.success('保存成功');
    drawerVisible.value = false;
    await loadData();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

async function toggleState(row: AdminUserDTO) {
  const next = row.state === 'enabled' ? 'disabled' : 'enabled';
  try {
    await updateUserStatus({ id: row.id, state: next });
    MessagePlugin.success(next === 'enabled' ? '已启用' : '已停用');
    await loadData();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  }
}

// ===== 用户删除 =====
const confirmVisible = ref(false);
const deleteTarget = ref<AdminUserDTO | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除用户「${deleteTarget.value.username}（${deleteTarget.value.account}）」吗？` : '',
);

function openDeleteUser(row: AdminUserDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDeleteUser() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeUser(deleteTarget.value.id);
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

// ===== 部门新增（仅名称；默认根部门，排序自动排在根级最后） =====
const INITIAL_DEPT: DeptSaveParams = { parentId: '0', deptName: '', orderNum: 0, status: 0 };
const dialogVisible = ref(false);
const form2 = ref<DeptSaveParams>({ ...INITIAL_DEPT });

function openCreate() {
  form2.value = { ...INITIAL_DEPT };
  dialogVisible.value = true;
}

async function onSaveDept() {
  if (!form2.value.deptName.trim()) {
    MessagePlugin.warning('请输入部门名称');
    return;
  }
  const maxRootOrder = treeData.value.reduce((max, n) => Math.max(max, n.orderNum), 0);
  saving.value = true;
  try {
    await createDept({
      deptName: form2.value.deptName.trim(),
      parentId: '0',
      orderNum: maxRootOrder + 1,
      status: 0,
    });
    MessagePlugin.success('保存成功');
    dialogVisible.value = false;
    await loadTree();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 部门改名 =====
const renameDialog = ref(false);
const renameTarget = ref<DeptNode | null>(null);
const renameForm = ref<{ deptName: string }>({ deptName: '' });

function openRename(node: DeptNode) {
  renameTarget.value = node;
  renameForm.value = { deptName: node.label };
  renameDialog.value = true;
}

async function onSaveRename() {
  if (!renameTarget.value) return;
  if (!renameForm.value.deptName.trim()) {
    MessagePlugin.warning('请输入部门名称');
    return;
  }
  saving.value = true;
  try {
    await updateDept({
      id: renameTarget.value.id,
      parentId: renameTarget.value.parentId,
      deptName: renameForm.value.deptName.trim(),
      orderNum: renameTarget.value.orderNum,
      status: renameTarget.value.status,
    });
    MessagePlugin.success('已修改');
    renameDialog.value = false;
    await loadTree();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 设置负责人 =====
const leaderDialog = ref(false);
const leaderTarget = ref<DeptNode | null>(null);
const leaderUserId = ref<string | undefined>(undefined);
const leaderUserOptions = ref<{ label: string; value: string }[]>([]);

async function openLeader(node: DeptNode) {
  leaderTarget.value = node;
  leaderUserId.value = node.leaderId && node.leaderId !== '0' ? node.leaderId : undefined;
  leaderDialog.value = true;
  try {
    const res = await fetchUserList({ page: 1, pageSize: 200 });
    leaderUserOptions.value = (res.list || []).map((u: AdminUserDTO) => ({
      label: `${u.username}（${u.account}）`,
      value: u.id,
    }));
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  }
}

async function onSaveLeader() {
  if (!leaderTarget.value) return;
  saving.value = true;
  try {
    await setDeptLeader(leaderTarget.value.id, leaderUserId.value || '0');
    MessagePlugin.success(leaderUserId.value ? '负责人已设置' : '负责人已清除');
    leaderDialog.value = false;
    await loadTree();
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// ===== 部门删除 =====
const deptConfirmVisible = ref(false);
const deptDeleteTarget = ref<DeptNode | null>(null);
const deptConfirmBody = computed(() =>
  deptDeleteTarget.value ? `确定删除部门「${deptDeleteTarget.value.label}」吗？存在子部门或用户时将被拒绝。` : '',
);

function openDeleteDept(node: DeptNode) {
  deptDeleteTarget.value = node;
  deptConfirmVisible.value = true;
}

async function onConfirmDeleteDept() {
  if (!deptDeleteTarget.value) return;
  saving.value = true;
  try {
    const deletedId = deptDeleteTarget.value.id;
    await removeDept(deletedId);
    MessagePlugin.success('删除成功');
    // 删除的刚好是选中的部门 → 清掉选中，右表回到全部
    if (selectedDeptId.value === deletedId) {
      selectedDeptId.value = null;
    }
    deptConfirmVisible.value = false;
    deptDeleteTarget.value = null;
    await loadTree();
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
  loadTree();
  loadOptions();
  loadData();
});
</script>
<style lang="less" scoped>
.org-page {
  display: flex;
  gap: var(--td-comp-margin-xl);
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
  min-height: 480px;
}

.org-page__tree {
  width: 250px;
  flex-shrink: 0;
  border-right: 1px solid var(--td-border-level-1-color);
  padding-right: var(--td-comp-margin-xl);
}

.org-page__tree-toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-m);
}

.org-page__main {
  flex: 1;
  min-width: 0;
}
</style>
