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
      keyword-placeholder="关键字"
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
        <t-button
          v-permission="'system:messageLog:clear'"
          theme="danger"
          variant="outline"
          @click="
            confirmKind = 'clear';
            confirmVisible = true;
          "
        >
          清空记录
        </t-button>
      </template>
      <template #batch="{ selected }">
        <t-button
          v-permission="'system:messageLog:remove'"
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
        <s-tag-dict dict-type="sys_message_type" :value="(row as MessageLogDTO).messageType" />
      </template>
      <template #templateMode="{ row }">
        <s-tag-dict dict-type="sys_message_template_mode" :value="(row as MessageLogDTO).templateMode" />
      </template>
      <template #supplierType="{ row }">
        {{ supplierLabel((row as MessageLogDTO).supplierType) }}
      </template>
      <template #isSuccess="{ row }">
        <t-tag :theme="(row as MessageLogDTO).isSuccess === 1 ? 'success' : 'danger'" variant="light">
          {{ (row as MessageLogDTO).isSuccess === 1 ? '成功' : '失败' }}
        </t-tag>
      </template>
      <template #costTime="{ row }"> {{ (row as MessageLogDTO).costTime }} 毫秒 </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:messageLog:query'" theme="primary" @click="view(row as MessageLogDTO)">
            详情
          </t-link>
          <t-link
            v-permission="'system:messageLog:remove'"
            theme="danger"
            @click="
              removeTargets = [row as MessageLogDTO];
              confirmKind = 'remove';
              confirmVisible = true;
            "
          >
            删除
          </t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 消息发送记录详情 -->
    <t-dialog v-model:visible="viewVisible" header="消息发送记录详情" width="min(950px, 90%)" :footer="false">
      <div v-if="viewLoading" class="message-log__loading">加载中…</div>
      <t-descriptions v-else :column="2" bordered size="small">
        <t-descriptions-item label="发送记录ID">{{ form.id }}</t-descriptions-item>
        <t-descriptions-item label="消息模板ID">{{ form.messageTemplateId }}</t-descriptions-item>
        <t-descriptions-item label="消息key">{{ form.messageKey }}</t-descriptions-item>
        <t-descriptions-item label="模板名称">{{ form.messageTemplateName }}</t-descriptions-item>
        <t-descriptions-item label="消息类型">
          <s-tag-dict dict-type="sys_message_type" :value="form.messageType" />
        </t-descriptions-item>
        <t-descriptions-item label="模板类型">
          <s-tag-dict dict-type="sys_message_template_mode" :value="form.templateMode" />
        </t-descriptions-item>
        <t-descriptions-item label="发送账号">{{ form.account }}</t-descriptions-item>
        <t-descriptions-item label="标题">{{ form.title }}</t-descriptions-item>
        <t-descriptions-item label="模板ID">{{ form.templateId }}</t-descriptions-item>
        <t-descriptions-item label="消息配置标题">{{ form.messageConfigTitle }}</t-descriptions-item>
        <t-descriptions-item label="平台标识" :span="2">{{ supplierLabel(form.supplierType) }}</t-descriptions-item>
        <t-descriptions-item label="发送内容" :span="2">{{ form.content }}</t-descriptions-item>
        <t-descriptions-item label="是否成功">
          <t-tag :theme="form.isSuccess === 1 ? 'success' : 'danger'" variant="light">
            {{ form.isSuccess === 1 ? '成功' : '失败' }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="返回主体消息" :span="2">{{ form.responseBody }}</t-descriptions-item>
        <t-descriptions-item label="记录时间">{{ form.logTime }}</t-descriptions-item>
        <t-descriptions-item label="消耗时间">{{ form.costTime ?? 0 }} 毫秒</t-descriptions-item>
      </t-descriptions>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      :header="confirmKind === 'clear' ? '清空记录' : '删除消息发送记录'"
      :body="confirmBody"
      :confirm-btn="{ content: '确认', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'SystemMessageLog' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { getMessageSuppliers } from '@/api/system/messageConfig';
import type { MessageSupplierOption } from '@/api/system/messageConfigModel';
import { clearMessageLogs, getMessageLog, listMessageLogs, removeMessageLog } from '@/api/system/messageLog';
import type { MessageLogDTO, MessageLogListParams } from '@/api/system/messageLogModel';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useDict } from '@/store';

// 是否成功：1 成功 0 失败（本仓无对应字典，硬编码选项）
const successOptions = [
  { label: '成功', value: 1 },
  { label: '失败', value: 0 },
];

const messageTypeDict = useDict('sys_message_type');

const loading = ref(false);
const saving = ref(false);
const list = ref<MessageLogDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');
const messageType = ref<string | undefined>();
const messageKey = ref<string | undefined>();
const isSuccess = ref<number | undefined>();

const viewVisible = ref(false);
const viewLoading = ref(false);
const form = ref<MessageLogDTO>({} as MessageLogDTO);

const confirmVisible = ref(false);
const confirmKind = ref<'remove' | 'clear'>('remove');
const removeTargets = ref<MessageLogDTO[]>([]);

const supplierLabelMap = ref<Record<string, string>>({});

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'messageKey', label: '消息key', type: 'input' },
  { name: 'messageType', label: '消息类型', type: 'select', options: messageTypeDict.options.value },
  { name: 'isSuccess', label: '是否成功', type: 'select', options: successOptions },
]);

const columns = [
  { colKey: 'messageKey', title: '消息key', minWidth: 140, ellipsis: true },
  { colKey: 'messageTemplateName', title: '模板名称', minWidth: 140, ellipsis: true },
  { colKey: 'messageType', title: '消息类型', width: 100, cell: 'messageType' },
  { colKey: 'templateMode', title: '模板类型', width: 120, cell: 'templateMode' },
  { colKey: 'account', title: '发送账号', width: 130, ellipsis: true },
  { colKey: 'title', title: '标题', minWidth: 130, ellipsis: true },
  { colKey: 'content', title: '发送内容', minWidth: 180, ellipsis: true },
  { colKey: 'supplierType', title: '平台标识', width: 110, cell: 'supplierType' },
  { colKey: 'isSuccess', title: '是否成功', width: 90, cell: 'isSuccess' },
  { colKey: 'costTime', title: '消耗', width: 90, cell: 'costTime' },
  { colKey: 'logTime', title: '记录时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 130, fixed: 'right', cell: 'ops' },
];

const confirmBody = computed(() => {
  if (confirmKind.value === 'clear') return '确认清空全部消息发送记录？删除后不可恢复。';
  if (removeTargets.value.length === 1) {
    return `确认删除消息发送记录「${removeTargets.value[0]?.messageTemplateName ?? ''}」？删除后不可恢复。`;
  }
  return `确认删除选中的 ${removeTargets.value.length} 条消息发送记录？删除后不可恢复。`;
});

function supplierLabel(code?: string): string {
  if (!code) return '—';
  return supplierLabelMap.value[code] ?? code;
}

function toNum(v: unknown): number | undefined {
  if (v === undefined || v === null || v === '') return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

async function load() {
  loading.value = true;
  try {
    const params: MessageLogListParams = {
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      messageType: messageType.value,
      messageKey: messageKey.value || undefined,
      isSuccess: isSuccess.value,
    };
    const res = await listMessageLogs(params);
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
  messageKey.value = String(params.messageKey ?? '');
  messageType.value =
    params.messageType === undefined || params.messageType === '' ? undefined : String(params.messageType);
  isSuccess.value = toNum(params.isSuccess);
  page.value = 1;
  load();
}

function onReset() {
  keyword.value = '';
  messageKey.value = '';
  messageType.value = undefined;
  isSuccess.value = undefined;
  page.value = 1;
  load();
}

async function view(row: MessageLogDTO) {
  viewLoading.value = true;
  viewVisible.value = true;
  try {
    form.value = await getMessageLog(row.id);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取详情失败');
  } finally {
    viewLoading.value = false;
  }
}

async function onConfirmDelete() {
  saving.value = true;
  try {
    if (confirmKind.value === 'clear') {
      await clearMessageLogs();
      MessagePlugin.success('已清空发送记录');
    } else {
      await Promise.all(removeTargets.value.map((row) => removeMessageLog(row.id)));
      MessagePlugin.success(`已删除 ${removeTargets.value.length} 条发送记录`);
    }
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
  removeTargets.value = selected as MessageLogDTO[];
  confirmKind.value = 'remove';
  confirmVisible.value = true;
}

async function loadSupplierLabels() {
  try {
    const [mail, sms] = await Promise.all([getMessageSuppliers('MAIL'), getMessageSuppliers('SMS')]);
    supplierLabelMap.value = [...(mail.list ?? []), ...(sms.list ?? [])].reduce<Record<string, string>>(
      (acc, item: MessageSupplierOption) => {
        acc[item.code] = item.label;
        return acc;
      },
      {},
    );
  } catch {
    supplierLabelMap.value = {};
  }
}

onMounted(() => {
  load();
  loadSupplierLabels();
});
</script>
<style scoped>
.message-log__loading {
  color: var(--td-text-color-secondary);
  padding: var(--td-comp-padding-condensed) 0;
}
</style>
