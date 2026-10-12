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
      keyword-placeholder="标题/备注"
      :search-fields="searchFields"
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
        <t-button v-permission="'system:messageConfig:add'" theme="primary" @click="open()">新增消息配置</t-button>
      </template>
      <template #batch="{ selected }">
        <t-button
          v-permission="'system:messageConfig:remove'"
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
      <template #messageType="{ row }">
        <s-tag-dict dict-type="sys_message_type" :value="(row as MessageConfigDTO).messageType" />
      </template>
      <template #supplierType="{ row }">
        {{ supplierLabel((row as MessageConfigDTO).supplierType) }}
      </template>
      <template #status="{ row }">
        <t-tag :theme="(row as MessageConfigDTO).status === 1 ? 'success' : 'default'" variant="light">
          {{ (row as MessageConfigDTO).status === 1 ? '正常' : '停用' }}
        </t-tag>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:messageConfig:query'" theme="primary" @click="view(row as MessageConfigDTO)">
            详情
          </t-link>
          <t-link v-permission="'system:messageConfig:edit'" theme="primary" @click="open(row as MessageConfigDTO)">
            编辑
          </t-link>
          <t-link
            v-permission="'system:messageConfig:remove'"
            theme="danger"
            @click="
              removeTargets = [row as MessageConfigDTO];
              confirmVisible = true;
            "
          >
            删除
          </t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 添加或修改消息配置（字段 >5 → s-form-drawer；动态字段 FormFieldRenders） -->
    <s-form-drawer
      v-model:visible="drawerVisible"
      :title="form.id ? '编辑消息配置' : '新增消息配置'"
      :loading="saving"
      @confirm="onConfirm"
    >
      <t-form
        ref="formRef"
        :data="form"
        :rules="rules"
        label-width="96px"
        scroll-to-first-error="smooth"
        @submit="onSubmit"
      >
        <t-form-item label="标题" name="title">
          <t-input v-model="form.title" :maxlength="64" clearable placeholder="请输入标题" />
        </t-form-item>
        <t-form-item label="消息类型" name="messageType">
          <t-select
            v-model="form.messageType"
            placeholder="请选择消息类型"
            clearable
            filterable
            @change="onMessageTypeChange"
          >
            <t-option
              v-for="item in messageTypeDict.options.value"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="支持平台" name="supplierType">
          <t-select v-model="form.supplierType" placeholder="请选择支持平台" clearable filterable>
            <t-option v-for="item in supplierOptions" :key="item.code" :label="item.label" :value="item.code" />
          </t-select>
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-select v-model="form.status" placeholder="请选择状态" :options="statusOptions" />
        </t-form-item>
        <form-field-renders v-model="form.configObject" name="configObject" :configs="supplierFieldConfigs" />
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="form.remark" :maxlength="255" placeholder="请输入备注" />
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <!-- 消息配置详情 -->
    <t-dialog v-model:visible="viewVisible" header="消息配置详情" width="min(850px, 100%)" :footer="false">
      <div v-if="viewLoading" class="message-config__loading">加载中…</div>
      <field-descriptions v-else :config-value="form.configJson || '{}'" :field-configs="supplierFieldConfigs">
        <template #prefix>
          <t-descriptions-item label="消息配置ID">{{ form.id }}</t-descriptions-item>
          <t-descriptions-item label="标题">{{ form.title }}</t-descriptions-item>
          <t-descriptions-item label="消息类型">
            <s-tag-dict dict-type="sys_message_type" :value="form.messageType" />
          </t-descriptions-item>
          <t-descriptions-item label="支持平台标识">{{ supplierLabel(form.supplierType) }}</t-descriptions-item>
        </template>
        <template #suffix>
          <t-descriptions-item label="状态">
            <t-tag :theme="form.status === 1 ? 'success' : 'default'" variant="light">
              {{ form.status === 1 ? '正常' : '停用' }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="备注" :span="2">{{ form.remark }}</t-descriptions-item>
          <t-descriptions-item label="更新时间">{{ form.updatedAt }}</t-descriptions-item>
          <t-descriptions-item label="创建时间">{{ form.createdAt }}</t-descriptions-item>
        </template>
      </field-descriptions>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除消息配置"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'SystemMessageConfig' });
import type { FormInstanceFunctions, FormRule, SelectValue, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  createMessageConfig,
  getMessageConfig,
  getMessageFieldConfigs,
  getMessageSuppliers,
  listMessageConfigs,
  removeMessageConfig,
  updateMessageConfig,
} from '@/api/system/messageConfig';
import type { MessageConfigDTO, MessageFieldConfig, MessageSupplierOption } from '@/api/system/messageConfigModel';
import FieldDescriptions from '@/components/field-config/FieldDescriptions.vue';
import FormFieldRenders from '@/components/field-config/FormFieldRenders.vue';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useDict } from '@/store';

// 状态语义与本仓 0023 迁移一致：1 正常 0 停用（本仓无对应字典，前端硬编码选项）
const statusOptions = [
  { label: '正常', value: 1 },
  { label: '停用', value: 0 },
];

interface ConfigForm {
  id?: string;
  title: string;
  messageType: string;
  supplierType: string;
  configJson?: string;
  configObject: Record<string, any>;
  status: number;
  remark: string;
  createdAt?: string;
  updatedAt?: string;
}

const messageTypeDict = useDict('sys_message_type');

const loading = ref(false);
const saving = ref(false);
const list = ref<MessageConfigDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');
const messageType = ref<string | undefined>();
const status = ref<number | undefined>();

const drawerVisible = ref(false);
const formRef = ref<FormInstanceFunctions>();
const form = ref<ConfigForm>({ title: '', messageType: '', supplierType: '', configObject: {}, status: 1, remark: '' });

const viewVisible = ref(false);
const viewLoading = ref(false);

const confirmVisible = ref(false);
const removeTargets = ref<MessageConfigDTO[]>([]);

const supplierOptions = ref<MessageSupplierOption[]>([]);
const supplierLabelMap = ref<Record<string, string>>({});
const messageFieldConfigs = ref<Record<string, MessageFieldConfig>>({});

const rules: Record<string, FormRule[]> = {
  title: [{ required: true, message: '标题不能为空' }],
  messageType: [{ required: true, message: '消息类型不能为空' }],
  supplierType: [{ required: true, message: '支持平台不能为空' }],
  status: [{ required: true, message: '状态不能为空' }],
};

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'messageType', label: '消息类型', type: 'select', options: messageTypeDict.options.value },
  { name: 'status', label: '状态', type: 'select', options: statusOptions },
]);

const columns = [
  { colKey: 'title', title: '标题', minWidth: 160, ellipsis: true },
  { colKey: 'messageType', title: '消息类型', width: 100, cell: 'messageType' },
  { colKey: 'supplierType', title: '支持平台标识', width: 130, cell: 'supplierType' },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'remark', title: '备注', minWidth: 140, ellipsis: true },
  { colKey: 'updatedAt', title: '更新时间', width: 170 },
  { colKey: 'createdAt', title: '创建时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 150, fixed: 'right', cell: 'ops' },
];

/** 当前表单/详情供应商的动态字段配置 */
const supplierFieldConfigs = computed(() => {
  if (!form.value.supplierType) return {};
  return messageFieldConfigs.value[form.value.supplierType]?.supplierConfig ?? {};
});

const confirmBody = computed(() =>
  removeTargets.value.length === 1
    ? `确认删除消息配置「${removeTargets.value[0]?.title ?? ''}」？删除后不可恢复。`
    : `确认删除选中的 ${removeTargets.value.length} 条消息配置？删除后不可恢复。`,
);

function supplierLabel(code?: string): string {
  if (!code) return '—';
  return supplierLabelMap.value[code] ?? code;
}

function toNum(v: unknown): number | undefined {
  if (v === undefined || v === null || v === '') return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

function safeParseConfig(json?: string): Record<string, any> {
  if (!json) return {};
  try {
    const parsed = JSON.parse(json);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

async function loadSuppliers(messageTypeValue?: string): Promise<MessageSupplierOption[]> {
  if (!messageTypeValue) return [];
  try {
    const res = await getMessageSuppliers(messageTypeValue);
    return res.list ?? [];
  } catch {
    return [];
  }
}

async function load() {
  loading.value = true;
  try {
    const res = await listMessageConfigs({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      messageType: messageType.value,
      status: status.value,
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
  messageType.value =
    params.messageType === undefined || params.messageType === '' ? undefined : String(params.messageType);
  status.value = toNum(params.status);
  page.value = 1;
  load();
}

function onReset() {
  keyword.value = '';
  messageType.value = undefined;
  status.value = undefined;
  page.value = 1;
  load();
}

async function onMessageTypeChange(v: SelectValue) {
  form.value.supplierType = '';
  supplierOptions.value = await loadSuppliers(v ? String(v) : undefined);
}

/** 新增 / 编辑（先取详情再开抽屉，避免动态字段默认值覆盖已存配置） */
async function open(row?: MessageConfigDTO) {
  saving.value = true;
  try {
    if (row) {
      const data = await getMessageConfig(row.id);
      form.value = {
        id: data.id,
        title: data.title,
        messageType: data.messageType,
        supplierType: data.supplierType,
        configJson: data.configJson,
        configObject: safeParseConfig(data.configJson),
        status: data.status,
        remark: data.remark,
        createdAt: data.createdAt,
        updatedAt: data.updatedAt,
      };
      supplierOptions.value = await loadSuppliers(data.messageType);
    } else {
      form.value = { title: '', messageType: '', supplierType: '', configObject: {}, status: 1, remark: '' };
      supplierOptions.value = [];
    }
    drawerVisible.value = true;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取详情失败');
  } finally {
    saving.value = false;
  }
}

async function view(row: MessageConfigDTO) {
  viewLoading.value = true;
  viewVisible.value = true;
  try {
    const data = await getMessageConfig(row.id);
    form.value = {
      id: data.id,
      title: data.title,
      messageType: data.messageType,
      supplierType: data.supplierType,
      configJson: data.configJson,
      configObject: safeParseConfig(data.configJson),
      status: data.status,
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
    const payload = {
      id: form.value.id,
      title: (form.value.title ?? '').trim(),
      messageType: form.value.messageType,
      supplierType: form.value.supplierType,
      configJson: JSON.stringify(form.value.configObject ?? {}),
      status: Number(form.value.status),
      remark: form.value.remark,
    };
    if (payload.id) await updateMessageConfig(payload);
    else await createMessageConfig(payload);
    MessagePlugin.success('保存成功');
    drawerVisible.value = false;
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
    await Promise.all(removeTargets.value.map((row) => removeMessageConfig(row.id)));
    MessagePlugin.success(`已删除 ${removeTargets.value.length} 条消息配置`);
    confirmVisible.value = false;
    removeTargets.value = [];
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  } finally {
    saving.value = false;
  }
}

async function batchRemove(selected: unknown[]) {
  if (!selected.length) return;
  removeTargets.value = selected as MessageConfigDTO[];
  confirmVisible.value = true;
}

async function initAux() {
  try {
    messageFieldConfigs.value = await getMessageFieldConfigs();
  } catch {
    messageFieldConfigs.value = {};
  }
  // 平台标识展示用：拉齐 MAIL / SMS 两类型供应商 label
  const [mail, sms] = await Promise.all([loadSuppliers('MAIL'), loadSuppliers('SMS')]);
  supplierLabelMap.value = [...mail, ...sms].reduce<Record<string, string>>((acc, item) => {
    acc[item.code] = item.label;
    return acc;
  }, {});
}

onMounted(() => {
  load();
  initAux();
});
</script>
<style scoped>
.message-config__loading {
  color: var(--td-text-color-secondary);
  padding: var(--td-comp-padding-condensed) 0;
}
</style>
