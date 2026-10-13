<template>
  <div class="content-card">
    <s-table
      row-key="id"
      :columns="columns as any"
      :data="list as any"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      keyword-placeholder="消息名称/消息key"
      @search="onSearch"
      @reset="onReset"
      @refresh="load"
      @page-change="
        (p, ps) => {
          page = p;
          pageSize = ps;
          load();
        }
      "
    >
      <template #toolbar>
        <t-button v-permission="'system:messageKey:add'" theme="primary" @click="open()">新增消息常量</t-button>
      </template>
      <template #batch="{ selected }">
        <t-button
          v-permission="'system:messageKey:remove'"
          size="small"
          variant="outline"
          theme="danger"
          :loading="saving"
          :disabled="saving"
          @click="batchRemove(selected)"
        >
          批量删除
        </t-button>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:messageKey:query'" theme="primary" @click="view(row as MessageKeyDTO)">
            详情
          </t-link>
          <t-link v-permission="'system:messageKey:edit'" theme="primary" @click="open(row as MessageKeyDTO)">
            编辑
          </t-link>
          <t-link
            v-permission="'system:messageKey:remove'"
            theme="danger"
            @click="
              removeTargets = [row as MessageKeyDTO];
              confirmVisible = true;
            "
          >
            删除
          </t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 添加或修改消息常量（3 字段 ≤5 → t-dialog） -->
    <s-form-dialog
      v-model:visible="dialogVisible"
      :title="form.id ? '编辑消息常量' : '新增消息常量'"
      :loading="saving"
      @confirm="onConfirm"
    >
      <t-form
        ref="formRef"
        :data="form"
        :rules="rules"
        label-width="90px"
        scroll-to-first-error="smooth"
        @submit="onSubmit"
      >
        <t-form-item label="消息名称" name="name">
          <t-input v-model="form.name" :maxlength="64" clearable placeholder="请输入消息名称" />
        </t-form-item>
        <t-form-item label="消息key" name="messageKey">
          <t-input v-model="form.messageKey" :maxlength="128" clearable placeholder="请输入消息key" />
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="form.remark" :maxlength="255" placeholder="请输入备注" />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 消息常量详情 -->
    <t-dialog v-model:visible="viewVisible" header="消息常量详情" width="700px" :footer="false">
      <div v-if="viewLoading" class="message-key__loading">加载中…</div>
      <t-descriptions v-else :column="2" bordered size="small">
        <t-descriptions-item label="主键">{{ form.id }}</t-descriptions-item>
        <t-descriptions-item label="消息名称">{{ form.name }}</t-descriptions-item>
        <t-descriptions-item label="消息key">{{ form.messageKey }}</t-descriptions-item>
        <t-descriptions-item label="备注" :span="2">{{ form.remark }}</t-descriptions-item>
        <t-descriptions-item label="更新时间">{{ form.updatedAt }}</t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ form.createdAt }}</t-descriptions-item>
      </t-descriptions>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除消息常量"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'SystemMessageKey' });
import type { FormInstanceFunctions, FormRule, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  createMessageKey,
  getMessageKey,
  listMessageKeys,
  removeMessageKey,
  updateMessageKey,
} from '@/api/system/messageKey';
import type { MessageKeyDTO, MessageKeySaveParams } from '@/api/system/messageKeyModel';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import STable from '@/components/s-table/index.vue';
import { useConfigStore } from '@/store';

interface KeyForm extends MessageKeySaveParams {
  createdAt?: string;
  updatedAt?: string;
}

const loading = ref(false);
const saving = ref(false);
const list = ref<MessageKeyDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');

const dialogVisible = ref(false);
const formRef = ref<FormInstanceFunctions>();
const form = ref<KeyForm>({ name: '', messageKey: '', remark: '' });

const viewVisible = ref(false);
const viewLoading = ref(false);

const confirmVisible = ref(false);
const removeTargets = ref<MessageKeyDTO[]>([]);

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '消息名称不能为空' }],
  messageKey: [{ required: true, message: '消息key不能为空' }],
};

const columns = [
  { colKey: 'name', title: '消息名称', minWidth: 160, ellipsis: true },
  { colKey: 'messageKey', title: '消息key', minWidth: 200, ellipsis: true },
  { colKey: 'remark', title: '备注', minWidth: 140, ellipsis: true },
  { colKey: 'updatedAt', title: '更新时间', width: 170 },
  { colKey: 'createdAt', title: '创建时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 150, fixed: 'right', cell: 'ops' },
];

const confirmBody = computed(() =>
  removeTargets.value.length === 1
    ? `确认删除消息常量「${removeTargets.value[0]?.name ?? ''}」？删除后不可恢复。`
    : `确认删除选中的 ${removeTargets.value.length} 条消息常量？删除后不可恢复。`,
);

async function load() {
  loading.value = true;
  try {
    const res = await listMessageKeys({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
    });
    list.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function onSearch(params: Record<string, unknown>) {
  keyword.value = String(params.keyword ?? '');
  page.value = 1;
  load();
}

function onReset() {
  keyword.value = '';
  page.value = 1;
  load();
}

/** 新增 / 编辑（先取详情再开弹窗） */
async function open(row?: MessageKeyDTO) {
  saving.value = true;
  try {
    if (row) {
      const data = await getMessageKey(row.id);
      form.value = { id: data.id, name: data.name, messageKey: data.messageKey, remark: data.remark };
    } else {
      form.value = { name: '', messageKey: '', remark: '' };
    }
    dialogVisible.value = true;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取详情失败');
  } finally {
    saving.value = false;
  }
}

async function view(row: MessageKeyDTO) {
  viewLoading.value = true;
  viewVisible.value = true;
  try {
    const data = await getMessageKey(row.id);
    form.value = {
      id: data.id,
      name: data.name,
      messageKey: data.messageKey,
      remark: data.remark,
      createdAt: data.createdAt,
      updatedAt: data.updatedAt,
    };
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取详情失败');
  } finally {
    viewLoading.value = false;
  }
}

function onConfirm() {
  formRef.value?.submit();
}

async function onSubmit({ validateResult, firstError }: SubmitContext) {
  if (validateResult !== true) {
    MessagePlugin.warning(firstError || '请检查表单');
    return;
  }
  saving.value = true;
  try {
    const payload: MessageKeySaveParams = {
      id: form.value.id,
      name: (form.value.name ?? '').trim(),
      messageKey: (form.value.messageKey ?? '').trim(),
      remark: form.value.remark,
    };
    if (payload.id) await updateMessageKey(payload);
    else await createMessageKey(payload);
    MessagePlugin.success('保存成功');
    dialogVisible.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function onConfirmDelete() {
  if (!removeTargets.value.length) return;
  saving.value = true;
  try {
    await Promise.all(removeTargets.value.map((row) => removeMessageKey(row.id)));
    MessagePlugin.success(`已删除 ${removeTargets.value.length} 条消息常量`);
    confirmVisible.value = false;
    removeTargets.value = [];
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  } finally {
    saving.value = false;
  }
}

function batchRemove(selected: unknown[]) {
  if (!selected.length) return;
  removeTargets.value = selected as MessageKeyDTO[];
  confirmVisible.value = true;
}

onMounted(load);
</script>
<style scoped>
.message-key__loading {
  color: var(--td-text-color-secondary);
  padding: var(--td-comp-padding-condensed) 0;
}
</style>
