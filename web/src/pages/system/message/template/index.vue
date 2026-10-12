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
      keyword-placeholder="模板名称/消息key"
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
        <t-button v-permission="'system:messageTemplate:add'" theme="primary" @click="open()">
          新增消息模板
        </t-button>
      </template>
      <template #batch="{ selected }">
        <t-button
          v-permission="'system:messageTemplate:remove'"
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
        <s-tag-dict dict-type="sys_message_type" :value="(row as MessageTemplateDTO).messageType" />
      </template>
      <template #templateMode="{ row }">
        <s-tag-dict dict-type="sys_message_template_mode" :value="(row as MessageTemplateDTO).templateMode" />
      </template>
      <template #status="{ row }">
        <t-tag :theme="(row as MessageTemplateDTO).status === 1 ? 'success' : 'default'" variant="light">
          {{ (row as MessageTemplateDTO).status === 1 ? '正常' : '停用' }}
        </t-tag>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:messageTemplate:edit'" theme="primary" @click="handleTest(row as MessageTemplateDTO)">
            测试
          </t-link>
          <t-link v-permission="'system:messageTemplate:query'" theme="primary" @click="view(row as MessageTemplateDTO)">
            详情
          </t-link>
          <t-link v-permission="'system:messageTemplate:edit'" theme="primary" @click="open(row as MessageTemplateDTO)">
            编辑
          </t-link>
          <t-link
            v-permission="'system:messageTemplate:remove'"
            theme="danger"
            @click="
              removeTargets = [row as MessageTemplateDTO];
              confirmVisible = true;
            "
          >
            删除
          </t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 添加或修改消息模板（字段 >5 → s-form-drawer） -->
    <s-form-drawer
      v-model:visible="drawerVisible"
      :title="form.id ? '编辑消息模板' : '新增消息模板'"
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
        <t-form-item label="模板名称" name="templateName">
          <t-input v-model="form.templateName" :maxlength="64" clearable placeholder="请输入模板名称" />
        </t-form-item>
        <t-form-item label="消息key" name="messageKeyId">
          <t-select v-model="form.messageKeyId" placeholder="请选择消息key" clearable filterable>
            <t-option v-for="item in messageKeys" :key="item.id" :label="item.name" :value="item.id" />
          </t-select>
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
        <t-form-item v-if="form.messageType" label="消息配置" name="messageConfigId">
          <t-select
            v-model="form.messageConfigId"
            placeholder="请选择消息配置"
            clearable
            filterable
            @change="onMessageConfigChange"
          >
            <t-option
              v-for="item in messageConfigs"
              :key="item.id"
              :label="item.title"
              :value="item.id"
            />
          </t-select>
        </t-form-item>
        <t-form-item v-if="form.messageType === 'SMS'" label="签名" name="signature">
          <t-input v-model="form.signature" :maxlength="64" clearable placeholder="请输入签名" />
        </t-form-item>
        <t-form-item v-if="form.messageType === 'MAIL'" label="标题" name="title" help="支持变量模式：${title}">
          <t-input v-model="form.title" clearable placeholder="请输入标题" @change="handleVarsChange" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-select v-model="form.status" placeholder="请选择状态" :options="statusOptions" />
        </t-form-item>
        <t-form-item v-if="form.messageConfigId" label="模板类型" name="templateMode">
          <t-select v-model="form.templateMode" placeholder="请选择模板类型" clearable :options="templateModeOptions" />
        </t-form-item>
        <t-form-item v-if="form.templateMode === 'TEMPLATE_ID'" label="模板ID" name="templateId">
          <t-input v-model="form.templateId" :maxlength="64" clearable placeholder="请输入模板ID" />
        </t-form-item>
        <t-form-item label="内容" name="content">
          <template #help>
            <b>变量格式</b>：${name}；例如，尊敬的 ${name}，您的快递已飞奔在路上，将今天 ${time} 送达您的手里，请留意查收。
            <p v-if="form.templateMode === 'TEMPLATE_ID'" class="message-template__help-warn">
              <b>模版内容以实际平台配置的内容为准</b>
            </p>
          </template>
          <t-textarea
            v-model="form.content"
            :autosize="{ minRows: 4, maxRows: 10 }"
            placeholder="请输入内容"
            @change="handleVarsChange"
          />
        </t-form-item>
        <t-form-item v-if="form.varsList.length > 0" label="变量属性">
          <t-space direction="vertical" size="small">
            <t-form-item
              v-for="(item, index) in form.varsList"
              :key="item.key"
              :label="item.key"
              :name="`varsList[${index}].value`"
              label-align="left"
              :label-width="`${maxVarsLabelWidth * 12}px`"
              :rules="[{ required: true, message: '输入变量不能为空' }]"
            >
              <template v-if="index === form.varsList.length - 1" #help>
                <b>变量格式</b>：${name}、${name:这里可以填写默认值}<br />
                变量属性根据内容自动生成，调用时将值注入到 Map 中，后台按值组合填入内容或模板变量中发送
              </template>
              <t-input v-model="form.varsList[index].value" />
            </t-form-item>
          </t-space>
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="form.remark" :maxlength="255" placeholder="请输入备注" />
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <!-- 消息模板详情 -->
    <t-dialog v-model:visible="viewVisible" header="消息模板详情" width="min(850px, 100%)" :footer="false">
      <div v-if="viewLoading" class="message-template__loading">加载中…</div>
      <t-descriptions v-else :column="2" bordered size="small">
        <t-descriptions-item label="消息模板ID">{{ form.id }}</t-descriptions-item>
        <t-descriptions-item label="模板名称">{{ form.templateName }}</t-descriptions-item>
        <t-descriptions-item label="消息配置ID">{{ form.messageConfigId || '—' }}</t-descriptions-item>
        <t-descriptions-item label="消息key">{{ form.messageKey }}</t-descriptions-item>
        <t-descriptions-item label="消息类型">
          <s-tag-dict dict-type="sys_message_type" :value="form.messageType" />
        </t-descriptions-item>
        <t-descriptions-item label="模板类型">
          <s-tag-dict dict-type="sys_message_template_mode" :value="form.templateMode" />
        </t-descriptions-item>
        <t-descriptions-item label="标题">{{ form.title }}</t-descriptions-item>
        <t-descriptions-item label="签名">{{ form.signature }}</t-descriptions-item>
        <t-descriptions-item label="模板ID">{{ form.templateId }}</t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag :theme="form.status === 1 ? 'success' : 'default'" variant="light">
            {{ form.status === 1 ? '正常' : '停用' }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="内容" :span="2">{{ form.content }}</t-descriptions-item>
        <t-descriptions-item label="输入变量" :span="2">{{ form.varsJson }}</t-descriptions-item>
        <t-descriptions-item label="备注" :span="2">{{ form.remark }}</t-descriptions-item>
        <t-descriptions-item label="更新时间">{{ form.updatedAt }}</t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ form.createdAt }}</t-descriptions-item>
      </t-descriptions>
    </t-dialog>

    <!-- 消息模板测试（账号 + vars 键值对编辑，结果展示） -->
    <t-dialog
      v-model:visible="testVisible"
      header="消息模板测试"
      width="560px"
      destroy-on-close
      :confirm-btn="{ content: '发送测试', theme: 'primary', loading: testLoading }"
      @confirm="onConfirmTest"
    >
      <t-form :data="testForm" label-width="96px">
        <t-form-item label="发送账号" required>
          <t-input v-model="testForm.account" clearable placeholder="请输入手机号/邮箱" />
        </t-form-item>
        <t-form-item label="变量属性">
          <div class="message-template__vars">
            <div v-for="(item, index) in testForm.varsList" :key="index" class="message-template__var-row">
              <t-input v-model="item.key" placeholder="变量名" style="width: 160px" />
              <t-input v-model="item.value" placeholder="变量值" style="flex: 1" />
              <t-link theme="danger" @click="testForm.varsList.splice(index, 1)">删除</t-link>
            </div>
            <t-link theme="primary" @click="testForm.varsList.push({ key: '', value: '' })">添加变量</t-link>
          </div>
        </t-form-item>
      </t-form>
      <div v-if="testResult" class="message-template__test-result">
        <t-space align="center">
          <t-tag :theme="testResult.isSuccess ? 'success' : 'danger'" variant="light">
            {{ testResult.isSuccess ? '发送成功' : '发送失败' }}
          </t-tag>
          <span>耗时 {{ testResult.costTime }} 毫秒</span>
          <span>发送记录ID {{ testResult.logId }}</span>
        </t-space>
        <pre v-if="testResult.responseBody" class="message-template__test-body">{{ testResult.responseBody }}</pre>
      </div>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除消息模板"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'SystemMessageTemplate' });
import type { FormInstanceFunctions, FormRule, SelectValue, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { getMessageFieldConfigs, listMessageConfigs } from '@/api/system/messageConfig';
import type { MessageConfigDTO, MessageFieldConfig } from '@/api/system/messageConfigModel';
import { listMessageKeys } from '@/api/system/messageKey';
import type { MessageKeyDTO } from '@/api/system/messageKeyModel';
import {
  createMessageTemplate,
  getMessageTemplate,
  listMessageTemplates,
  removeMessageTemplate,
  sendMessageTest,
  updateMessageTemplate,
} from '@/api/system/messageTemplate';
import type {
  MessageSendTestResult,
  MessageTemplateDTO,
  MessageTemplateVar,
} from '@/api/system/messageTemplateModel';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import { useConfigStore, useDict } from '@/store';

// 状态语义与 0023 迁移一致：1 正常 0 停用（本仓无对应字典，硬编码选项）
const statusOptions = [
  { label: '正常', value: 1 },
  { label: '停用', value: 0 },
];

interface TemplateForm {
  id?: string;
  templateName: string;
  messageConfigId?: string;
  messageKeyId?: string;
  messageKey: string;
  messageType: string;
  templateMode: string;
  title: string;
  signature: string;
  templateId: string;
  content: string;
  varsJson: string;
  varsList: MessageTemplateVar[];
  status: number;
  remark: string;
  createdAt?: string;
  updatedAt?: string;
}

const messageTypeDict = useDict('sys_message_type');
const templateModeDict = useDict('sys_message_template_mode');

const loading = ref(false);
const saving = ref(false);
const list = ref<MessageTemplateDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');
const messageType = ref<string | undefined>();
const messageKeyId = ref<string | undefined>();
const status = ref<number | undefined>();

const drawerVisible = ref(false);
const formRef = ref<FormInstanceFunctions>();
const form = ref<TemplateForm>({
  templateName: '',
  messageKey: '',
  messageType: '',
  templateMode: '',
  title: '',
  signature: '',
  templateId: '',
  content: '',
  varsJson: '',
  varsList: [],
  status: 1,
  remark: '',
});

const viewVisible = ref(false);
const viewLoading = ref(false);

const confirmVisible = ref(false);
const removeTargets = ref<MessageTemplateDTO[]>([]);

const messageKeys = ref<MessageKeyDTO[]>([]);
const messageConfigs = ref<MessageConfigDTO[]>([]);
const messageFieldConfigs = ref<Record<string, MessageFieldConfig>>({});

const testVisible = ref(false);
const testLoading = ref(false);
const testResult = ref<MessageSendTestResult | null>(null);
const testForm = ref<{ messageTemplateId?: string; account: string; varsList: MessageTemplateVar[] }>({
  account: '',
  varsList: [],
});

const rules: Record<string, FormRule[]> = {
  templateName: [{ required: true, message: '模板名称不能为空' }],
  messageConfigId: [{ required: true, message: '消息配置不能为空' }],
  messageKeyId: [{ required: true, message: '消息key不能为空' }],
  messageType: [{ required: true, message: '消息类型不能为空' }],
  title: [{ required: true, message: '标题不能为空' }],
  templateMode: [{ required: true, message: '模板类型不能为空' }],
  status: [{ required: true, message: '状态不能为空' }],
  templateId: [{ required: true, message: '模板ID不能为空' }],
  content: [{ required: true, message: '内容不能为空' }],
};

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'messageType', label: '消息类型', type: 'select', options: messageTypeDict.options.value },
  {
    name: 'messageKeyId',
    label: '消息key',
    type: 'select',
    options: messageKeys.value.map((item) => ({ label: item.name, value: item.id })),
  },
  { name: 'status', label: '状态', type: 'select', options: statusOptions },
]);

const columns = [
  { colKey: 'templateName', title: '模板名称', minWidth: 150, ellipsis: true },
  { colKey: 'messageKey', title: '消息key', minWidth: 140, ellipsis: true },
  { colKey: 'messageType', title: '消息类型', width: 100, cell: 'messageType' },
  { colKey: 'templateMode', title: '模板类型', width: 120, cell: 'templateMode' },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'content', title: '内容', minWidth: 200, ellipsis: true },
  { colKey: 'remark', title: '备注', minWidth: 120, ellipsis: true },
  { colKey: 'updatedAt', title: '更新时间', width: 170 },
  { colKey: 'createdAt', title: '创建时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 200, fixed: 'right', cell: 'ops' },
];

/** 模板类型选项：按所选消息配置对应 supplier 的 templateMode 能力过滤（字段配置未就绪时不过滤） */
const templateModeOptions = computed(() => {
  const options = templateModeDict.options.value;
  const config = messageConfigs.value.find((item) => String(item.id) === String(form.value.messageConfigId));
  const field = config ? messageFieldConfigs.value[config.supplierType] : undefined;
  if (!field) return options;
  return options.filter(
    (item) =>
      (item.value === 'TEMPLATE_ID' && field.templateMode.supportTemplateId) ||
      (item.value === 'TEMPLATE_CONTENT' && field.templateMode.supportTemplateContent),
  );
});

const maxVarsLabelWidth = computed(() => {
  if (!form.value.varsList.length) return 6;
  return Math.max(...form.value.varsList.map((item) => (item.key ?? '').length), 6);
});

const confirmBody = computed(() =>
  removeTargets.value.length === 1
    ? `确认删除消息模板「${removeTargets.value[0]?.templateName ?? ''}」？删除后不可恢复。`
    : `确认删除选中的 ${removeTargets.value.length} 条消息模板？删除后不可恢复。`,
);

function toNum(v: unknown): number | undefined {
  if (v === undefined || v === null || v === '') return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

/** 提取 ${...} 变量名（去重） */
function getVars(...contents: Array<string | undefined>): string[] {
  const re = /\$\{([^}]+)\}/g;
  const out: string[] = [];
  contents.forEach((content) => {
    if (!content) return;
    let m: RegExpExecArray | null = re.exec(content);
    while (m) {
      out.push(m[1]);
      m = re.exec(content);
    }
  });
  return out.filter((value, index) => out.indexOf(value) === index);
}

/** 内容/标题变更后重建变量列表（保留已填值，默认值为 ${key}） */
function handleVarsChange() {
  const source = form.value.messageType === 'SMS' ? [form.value.content] : [form.value.title, form.value.content];
  const keys = getVars(...source);
  form.value.varsList = keys.map<MessageTemplateVar>((key) => {
    const prev = form.value.varsList.find((item) => item.key === key);
    return { key, value: prev?.value ?? `\${${key}}` };
  });
}

function parseVarsList(json?: string): MessageTemplateVar[] {
  if (!json) return [];
  try {
    const parsed = JSON.parse(json);
    return Array.isArray(parsed) ? (parsed as MessageTemplateVar[]) : [];
  } catch {
    return [];
  }
}

async function load() {
  loading.value = true;
  try {
    const res = await listMessageTemplates({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      messageType: messageType.value,
      messageKeyId: messageKeyId.value,
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
  messageType.value = params.messageType === undefined || params.messageType === '' ? undefined : String(params.messageType);
  messageKeyId.value = params.messageKeyId === undefined || params.messageKeyId === '' ? undefined : String(params.messageKeyId);
  status.value = toNum(params.status);
  page.value = 1;
  load();
}

function onReset() {
  keyword.value = '';
  messageType.value = undefined;
  messageKeyId.value = undefined;
  status.value = undefined;
  page.value = 1;
  load();
}

async function loadKeyOptions() {
  try {
    const res = await listMessageKeys({ page: 1, pageSize: 500 });
    messageKeys.value = res.list ?? [];
  } catch {
    messageKeys.value = [];
  }
}

async function loadConfigOptions(type?: string) {
  if (!type) {
    messageConfigs.value = [];
    return;
  }
  try {
    const res = await listMessageConfigs({ page: 1, pageSize: 500, messageType: type });
    messageConfigs.value = res.list ?? [];
  } catch {
    messageConfigs.value = [];
  }
}

async function onMessageTypeChange(v: SelectValue) {
  form.value.messageConfigId = undefined;
  form.value.templateMode = '';
  await loadConfigOptions(v ? String(v) : undefined);
}

function onMessageConfigChange() {
  form.value.templateMode = '';
}

/** 新增 / 编辑（先取详情再开抽屉） */
async function open(row?: MessageTemplateDTO) {
  saving.value = true;
  try {
    await loadKeyOptions();
    if (row) {
      const data = await getMessageTemplate(row.id);
      form.value = {
        id: data.id,
        templateName: data.templateName,
        messageConfigId: data.messageConfigId || undefined,
        messageKeyId: data.messageKeyId || undefined,
        messageKey: data.messageKey,
        messageType: data.messageType,
        templateMode: data.templateMode,
        title: data.title,
        signature: data.signature,
        templateId: data.templateId,
        content: data.content,
        varsJson: data.varsJson,
        varsList: parseVarsList(data.varsJson),
        status: data.status,
        remark: data.remark,
        createdAt: data.createdAt,
        updatedAt: data.updatedAt,
      };
      await loadConfigOptions(data.messageType);
    } else {
      form.value = {
        templateName: '',
        messageKey: '',
        messageType: '',
        templateMode: '',
        title: '',
        signature: '',
        templateId: '',
        content: '',
        varsJson: '',
        varsList: [],
        status: 1,
        remark: '',
      };
      messageConfigs.value = [];
    }
    drawerVisible.value = true;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取详情失败');
  } finally {
    saving.value = false;
  }
}

async function view(row: MessageTemplateDTO) {
  viewLoading.value = true;
  viewVisible.value = true;
  try {
    const data = await getMessageTemplate(row.id);
    form.value = {
      ...form.value,
      id: data.id,
      templateName: data.templateName,
      messageConfigId: data.messageConfigId || undefined,
      messageKeyId: data.messageKeyId || undefined,
      messageKey: data.messageKey,
      messageType: data.messageType,
      templateMode: data.templateMode,
      title: data.title,
      signature: data.signature,
      templateId: data.templateId,
      content: data.content,
      varsJson: data.varsJson,
      varsList: parseVarsList(data.varsJson),
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
      templateName: (form.value.templateName ?? '').trim(),
      messageConfigId: form.value.messageConfigId,
      messageKeyId: form.value.messageKeyId,
      messageKey: form.value.messageKey ?? '',
      messageType: form.value.messageType ?? '',
      templateMode: form.value.templateMode ?? '',
      title: form.value.title ?? '',
      signature: form.value.signature ?? '',
      templateId: form.value.templateId ?? '',
      content: form.value.content ?? '',
      varsJson: JSON.stringify(form.value.varsList ?? []),
      status: Number(form.value.status),
      remark: form.value.remark,
    };
    if (payload.id) await updateMessageTemplate(payload);
    else await createMessageTemplate(payload);
    MessagePlugin.success('保存成功');
    drawerVisible.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  } finally {
    saving.value = false;
  }
}

/** 测试：解析模板变量为键值对（含 ${key:默认值} 默认值提取） */
async function handleTest(row: MessageTemplateDTO) {
  testResult.value = null;
  testForm.value = { messageTemplateId: undefined, account: '', varsList: [] };
  testVisible.value = true;
  try {
    const data = await getMessageTemplate(row.id);
    const vars: MessageTemplateVar[] = [];
    const merge = (value?: string) => {
      if (!value) return;
      getVars(value).forEach((raw) => {
        if (raw.includes(':')) {
          const [k, d] = raw.split(':');
          vars.push({ key: k, value: d ?? '' });
        } else {
          vars.push({ key: raw, value: '' });
        }
      });
    };
    parseVarsList(data.varsJson).forEach((item) => merge(item.value));
    merge(data.title);
    // 去重（后出现的默认值覆盖先出现的）
    const dedup = new Map<string, MessageTemplateVar>();
    vars.forEach((item) => dedup.set(item.key, item));
    testForm.value = {
      messageTemplateId: data.id,
      account: '',
      varsList: [...dedup.values()].filter((item) => item.key),
    };
  } catch (e) {
    MessagePlugin.error((e as Error).message || '获取模板失败');
    testVisible.value = false;
  }
}

async function onConfirmTest() {
  const account = (testForm.value.account ?? '').trim();
  if (!account) {
    MessagePlugin.warning('发送账号不能为空');
    return;
  }
  for (const item of testForm.value.varsList) {
    if (!item.key?.trim()) {
      MessagePlugin.warning('变量名不能为空');
      return;
    }
    if (!item.value?.trim()) {
      MessagePlugin.warning(`变量 ${item.key} 的值不能为空`);
      return;
    }
  }
  if (!testForm.value.messageTemplateId) {
    MessagePlugin.warning('模板不存在');
    return;
  }
  testLoading.value = true;
  try {
    const vars: Record<string, string> = {};
    testForm.value.varsList.forEach((item) => {
      vars[item.key.trim()] = item.value ?? '';
    });
    testResult.value = await sendMessageTest({
      messageTemplateId: testForm.value.messageTemplateId,
      account,
      vars,
    });
  } catch (e) {
    MessagePlugin.error((e as Error).message || '发送测试失败');
  } finally {
    testLoading.value = false;
  }
}

async function onConfirmDelete() {
  if (!removeTargets.value.length) return;
  saving.value = true;
  try {
    await Promise.all(removeTargets.value.map((row) => removeMessageTemplate(row.id)));
    MessagePlugin.success(`已删除 ${removeTargets.value.length} 条消息模板`);
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
  removeTargets.value = selected as MessageTemplateDTO[];
  confirmVisible.value = true;
}

async function initAux() {
  try {
    messageFieldConfigs.value = await getMessageFieldConfigs();
  } catch {
    messageFieldConfigs.value = {};
  }
  await loadKeyOptions();
}

onMounted(() => {
  load();
  initAux();
});
</script>
<style scoped>
.message-template__help-warn {
  color: var(--td-error-color);
}

.message-template__loading {
  color: var(--td-text-color-secondary);
  padding: var(--td-comp-padding-condensed) 0;
}

.message-template__vars {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-xs);
  width: 100%;
}

.message-template__var-row {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
}

.message-template__test-result {
  margin-top: var(--td-comp-margin-l);
  padding-top: var(--td-comp-margin-s);
  border-top: 1px solid var(--td-component-border);
}

.message-template__test-body {
  margin-top: var(--td-comp-margin-s);
  padding: var(--td-comp-padding-condensed);
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-medium);
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}
</style>
