<template>
  <div class="user-page">
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
        <t-button v-permission="'system:user:add'" theme="primary" @click="openCreate">新增用户</t-button>
      </template>
      <template #state="{ row }">
        <t-tag :theme="row.state === 'enabled' ? 'success' : 'default'" variant="light">
          {{ stateDict.labelOf(row.state) || '-' }}
        </t-tag>
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:user:edit'" theme="primary" @click="openEdit(row)">编辑</t-link>
          <t-link
            v-permission="'system:user:edit'"
            :theme="row.state === 'enabled' ? 'warning' : 'success'"
            @click="toggleState(row)"
          >
            {{ row.state === 'enabled' ? stateDict.labelOf('disabled') : stateDict.labelOf('enabled') }}
          </t-link>
          <t-link v-permission="'system:user:remove'" theme="danger" @click="openDelete(row)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <t-drawer
      v-model:visible="drawerVisible"
      :header="isEdit ? '编辑用户' : '新增用户'"
      size="large"
      :footer="false"
      destroy-on-close
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
            :placeholder="isEdit ? '不修改请留空' : '请输入初始密码'"
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
            v-model="form.deptId"
            :data="deptOptions"
            clearable
            placeholder="选择部门"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="form.state" :options="stateOptions" />
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
        <t-form-item class="user-page__form-actions">
          <t-button variant="outline" @click="drawerVisible = false">取消</t-button>
          <t-button theme="primary" :loading="saving" @click="onSave">保存</t-button>
        </t-form-item>
      </t-form>
    </t-drawer>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除用户"
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
import type { PostDTO } from '@/api/system/post';
import { fetchPostList } from '@/api/system/post';
import type { RoleDTO } from '@/api/system/role';
import { fetchRoleList } from '@/api/system/role';
import type { AdminUserDTO, AdminUserSaveParams } from '@/api/system/user';
import { createUser, fetchUserList, removeUser, updateUser, updateUserStatus } from '@/api/system/user';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import { useConfigStore, useDict } from '@/store';

const stateDict = useDict('sys_state');
const stateOptions = stateDict.options;

const columns: PrimaryTableCol[] = [
  { title: '账号', colKey: 'account', width: 140 },
  { title: '姓名', colKey: 'username', width: 140 },
  { title: '所属部门', colKey: 'deptName', width: 160 },
  { title: '手机号', colKey: 'mobilePhone', width: 140 },
  { title: '邮箱', colKey: 'email', minWidth: 200 },
  { title: '状态', colKey: 'state', width: 100 },
  { title: '操作', colKey: 'op', width: 200, fixed: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const list = ref<AdminUserDTO[]>([]);
const query = ref<{ keyword: string; deptId: number | undefined; state: string | undefined }>({
  keyword: '',
  deptId: undefined,
  state: undefined,
});
const pagination = ref({ current: 1, pageSize: useConfigStore().pageDefault, total: 0, showJumper: true });

// 部门树（筛选 + 表单共用）
const deptTree = ref<DeptNode[]>([]);
const deptOptions = ref<DeptTreeItem[]>([]);

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'deptId', label: '所属部门', type: 'tree-select', treeData: deptOptions.value },
  {
    name: 'state',
    label: '状态',
    type: 'select',
    options: stateDict.options.value,
  },
]);

const roleOptions = ref<{ label: string; value: number }[]>([]);
const postOptions = ref<{ label: string; value: number }[]>([]);

function toTreeItem(node: DeptNode): DeptTreeItem {
  return {
    value: node.id,
    label: node.label,
    children: (node.children || []).map(toTreeItem),
  };
}

async function loadOptions() {
  try {
    const [deptRes, roleRes, postRes] = await Promise.all([
      fetchDeptTree(),
      fetchRoleList({ page: 1, pageSize: 100 }),
      fetchPostList({ page: 1, pageSize: 100 }),
    ]);
    deptTree.value = deptRes.list || [];
    deptOptions.value = deptTree.value.map(toTreeItem);
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

async function loadData() {
  loading.value = true;
  try {
    const res = await fetchUserList({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      keyword: query.value.keyword || undefined,
      deptId: query.value.deptId || undefined,
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
    deptId: (params.deptId ?? undefined) as number | undefined,
    state: (params.state ?? undefined) as string | undefined,
  };
  pagination.value.current = 1;
  loadData();
}

function onReset() {
  query.value = { keyword: '', deptId: undefined, state: undefined };
  pagination.value.current = 1;
  loadData();
}

function onPageChange(page: number, pageSize: number) {
  pagination.value.current = page;
  pagination.value.pageSize = pageSize;
  loadData();
}

// ===== 新增 / 编辑 =====
const INITIAL_FORM: AdminUserSaveParams = {
  account: '',
  username: '',
  password: '',
  mobilePhone: '',
  email: '',
  deptId: 0,
  state: 'enabled',
  roleIds: [],
  postIds: [],
};
const drawerVisible = ref(false);
const form = ref<AdminUserSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

function openCreate() {
  form.value = { ...INITIAL_FORM, roleIds: [], postIds: [] };
  drawerVisible.value = true;
}

function openEdit(row: AdminUserDTO) {
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

async function onSave() {
  if (!form.value.account.trim() || !form.value.username.trim()) {
    MessagePlugin.warning('账号/姓名不能为空');
    return;
  }
  saving.value = true;
  try {
    const payload: AdminUserSaveParams = {
      ...form.value,
      deptId: form.value.deptId || 0,
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

// ===== 启用 / 停用 =====
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

// ===== 删除 =====
const confirmVisible = ref(false);
const deleteTarget = ref<AdminUserDTO | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除用户「${deleteTarget.value.username}（${deleteTarget.value.account}）」吗？` : '',
);

function openDelete(row: AdminUserDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
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

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

onMounted(() => {
  loadOptions();
  loadData();
});
</script>
<style lang="less" scoped>
.user-page {
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
}

.user-page__form-actions {
  :deep(.t-form__controls) {
    justify-content: flex-end;
  }
}
</style>
