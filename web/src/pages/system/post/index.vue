<template>
  <div class="post-page">
    <s-table
      row-key="id"
      :columns="columns as any"
      :data="list as any"
      :total="pagination.total"
      :page="pagination.current"
      :page-size="pagination.pageSize"
      :loading="loading"
      keyword-placeholder="岗位名称/编码"
      @search="onSearch"
      @reset="onReset"
      @refresh="loadData"
      @page-change="onPageChange"
    >
      <template #toolbar>
        <t-button v-permission="'system:post:add'" theme="primary" @click="openCreate">新增岗位</t-button>
      </template>
      <template #status="{ row }">
        <s-tag-dict dict-type="sys_status" :value="row.status" />
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:post:edit'" theme="primary" @click="openEdit(row as PostDTO)">编辑</t-link>
          <t-link v-permission="'system:post:remove'" theme="danger" @click="openDelete(row as PostDTO)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <s-form-dialog
      v-model:visible="dialogVisible"
      :title="isEdit ? '编辑岗位' : '新增岗位'"
      :width="520"
      :loading="saving"
      @confirm="onSave"
    >
      <t-form :data="form" label-width="90px">
        <t-form-item label="岗位编码" required>
          <t-input v-model="form.postCode" placeholder="如 CEO" />
        </t-form-item>
        <t-form-item label="岗位名称" required>
          <t-input v-model="form.postName" placeholder="请输入岗位名称" />
        </t-form-item>
        <t-form-item label="排序">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
        <t-form-item label="状态">
          <s-radio-dict v-model="form.status" dict-type="sys_status" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除岗位"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemPost' });
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { PostDTO, PostSaveParams } from '@/api/system/post';
import { createPost, fetchPostList, removePost, updatePost } from '@/api/system/post';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore } from '@/store';

const columns: PrimaryTableCol[] = [
  { title: '岗位编码', colKey: 'postCode', width: 160 },
  { title: '岗位名称', colKey: 'postName', minWidth: 200 },
  { title: '排序', colKey: 'orderNum', width: 100 },
  { title: '状态', colKey: 'status', width: 100 },
  { title: '操作', colKey: 'op', width: 160, fixed: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const list = ref<PostDTO[]>([]);
const query = ref({ keyword: '' });
const pagination = ref({ current: 1, pageSize: useConfigStore().pageDefault, total: 0, showJumper: true });

async function loadData() {
  loading.value = true;
  try {
    const res = await fetchPostList({
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

const INITIAL_FORM: PostSaveParams = { postCode: '', postName: '', orderNum: 0, status: 0 };
const dialogVisible = ref(false);
const form = ref<PostSaveParams>({ ...INITIAL_FORM });
const isEdit = computed(() => form.value.id != null);

function openCreate() {
  form.value = { ...INITIAL_FORM };
  dialogVisible.value = true;
}

function openEdit(row: PostDTO) {
  form.value = {
    id: row.id,
    postCode: row.postCode,
    postName: row.postName,
    orderNum: row.orderNum,
    status: row.status,
  };
  dialogVisible.value = true;
}

async function onSave() {
  if (!form.value.postCode.trim() || !form.value.postName.trim()) {
    MessagePlugin.warning('岗位编码/名称不能为空');
    return;
  }
  saving.value = true;
  try {
    if (form.value.id != null) {
      await updatePost({ ...form.value });
    } else {
      await createPost({ ...form.value });
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

const confirmVisible = ref(false);
const deleteTarget = ref<PostDTO | null>(null);
const confirmBody = computed(() => (deleteTarget.value ? `确定删除岗位「${deleteTarget.value.postName}」吗？` : ''));

function openDelete(row: PostDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removePost(deleteTarget.value.id);
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
.post-page {
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
}
</style>
