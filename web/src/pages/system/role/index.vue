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
          <t-link v-permission="'system:role:assign'" theme="primary" @click="openAuthorize(row)">授权</t-link>
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
        <t-form-item label="状态">
          <t-radio-group v-model="form.state" :options="stateOptions" />
        </t-form-item>
        <t-form-item label="排序">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 授权（菜单权限 + 数据权限） -->
    <authorize-drawer v-model:visible="authVisible" :role="authRole" />

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

import type { RoleDTO, RoleSaveParams } from '@/api/system/role';
import { createRole, fetchRoleList, removeRole, updateRole } from '@/api/system/role';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import STable from '@/components/s-table/index.vue';
import { useConfigStore, useDict } from '@/store';

import AuthorizeDrawer from './components/AuthorizeDrawer.vue';

const stateDict = useDict('sys_state');
const stateOptions = stateDict.options;
const dataScopeDict = useDict('sys_data_scope');

const columns: PrimaryTableCol[] = [
  { title: '角色名称', colKey: 'name', width: 160 },
  { title: '角色编码', colKey: 'code', width: 140 },
  { title: '数据范围', colKey: 'dataScope', width: 140 },
  { title: '状态', colKey: 'state', width: 100 },
  { title: '描述', colKey: 'description', minWidth: 200 },
  { title: '排序', colKey: 'orderNum', width: 80 },
  { title: '操作', colKey: 'op', width: 220, fixed: 'right' },
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

// ===== 授权（菜单权限 + 数据部门） =====
const authVisible = ref(false);
const authRole = ref<RoleDTO | null>(null);

function openAuthorize(row: RoleDTO) {
  authRole.value = row;
  authVisible.value = true;
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
</style>
