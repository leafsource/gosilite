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
      keyword-placeholder="配置key/endpoint/桶名"
      @search="onSearch"
      @reset="
        () => {
          keyword = '';
          page = 1;
          load();
        }
      "
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
        <t-button v-permission="'system:ossConfig:add'" theme="primary" @click="open()">新增配置</t-button>
        <t-button v-permission="'system:ossConfig:test'" variant="outline" @click="openTest()">连通测试</t-button>
      </template>
      <template #configKey="{ row }">
        <t-space>
          <span>{{ (row as OssConfigDTO).configKey }}</span>
          <t-tag v-if="(row as OssConfigDTO).status === '1'" theme="success" size="small" variant="light">
            默认
          </t-tag>
        </t-space>
      </template>
      <template #accessPolicy="{ row }">
        <t-tag :theme="policyTheme((row as OssConfigDTO).accessPolicy)" variant="light">
          {{ policyLabel((row as OssConfigDTO).accessPolicy) }}
        </t-tag>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link
            v-if="(row as OssConfigDTO).status !== '1'"
            v-permission="'system:ossConfig:edit'"
            theme="primary"
            @click="setDefault(row as OssConfigDTO)"
          >
            设为默认
          </t-link>
          <t-link v-permission="'system:ossConfig:edit'" theme="primary" @click="open(row as OssConfigDTO)"
            >编辑</t-link
          >
          <t-link v-permission="'system:ossConfig:remove'" theme="danger" @click="del(row as OssConfigDTO)"
            >删除</t-link
          >
        </t-space>
      </template>
    </s-table>

    <!-- 字段 11 个 >5 → s-form-drawer（AGENTS 前端规则） -->
    <s-form-drawer
      v-model:visible="drawer"
      :title="form.id ? '编辑配置' : '新增配置'"
      size="560px"
      :loading="saving"
      @confirm="save"
    >
      <t-form :data="form" :rules="rules" label-align="right" label-width="96px">
        <t-form-item label="配置key" name="configKey" :help="form.id ? '建库后不可修改' : '唯一标识，如 local / minio'">
          <t-input
            v-model="form.configKey"
            :maxlength="64"
            :disabled="!!form.id"
            clearable
            placeholder="如 local"
          />
        </t-form-item>
        <t-form-item label="访问站点" name="endpoint" help="S3 端点，如 http://127.0.0.1:19000">
          <t-input v-model="form.endpoint" :maxlength="255" clearable placeholder="http://127.0.0.1:19000" />
        </t-form-item>
        <t-form-item label="自定义域名" name="domain" help="拼访问 URL 用；留空则用访问站点">
          <t-input v-model="form.domain" :maxlength="255" clearable placeholder="如 cdn.example.com" />
        </t-form-item>
        <t-form-item label="accessKey" name="accessKey">
          <t-input v-model="form.accessKey" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="secretKey" name="secretKey">
          <t-input v-model="form.secretKey" type="password" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="桶名称" name="bucketName">
          <t-input v-model="form.bucketName" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="前缀" name="prefix" help="对象 key 前缀，如 app1/">
          <t-input v-model="form.prefix" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="区域" name="region" help="留空通常即可">
          <t-input v-model="form.region" :maxlength="64" clearable />
        </t-form-item>
        <t-form-item label="是否HTTPS" name="isHttps">
          <t-radio-group v-model="form.isHttps" :options="httpsOptions" />
        </t-form-item>
        <t-form-item label="桶权限" name="accessPolicy" help="private 桶读取时自动转 120s 预签名 URL">
          <t-radio-group v-model="form.accessPolicy" :options="policyOptions" />
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-input v-model="form.remark" :maxlength="500" clearable />
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <!-- 连通测试（不写库） -->
    <t-dialog
      v-model:visible="testVisible"
      header="连通测试"
      width="520px"
      :confirm-btn="{ content: '测试', loading: testing }"
      @confirm="runTest"
    >
      <t-form :data="testForm" label-align="right" label-width="80px">
        <t-form-item label="accessKey">
          <t-input v-model="testForm.accessKey" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="secretKey">
          <t-input v-model="testForm.secretKey" type="password" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="桶名称">
          <t-input v-model="testForm.bucketName" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="访问站点">
          <t-input v-model="testForm.endpoint" :maxlength="255" clearable placeholder="http://127.0.0.1:19000" />
        </t-form-item>
        <t-form-item label="区域">
          <t-input v-model="testForm.region" :maxlength="64" clearable />
        </t-form-item>
      </t-form>
      <div v-if="testResult" class="oss-config__test-result">
        <t-tag :theme="testResult.connected ? 'success' : 'danger'" variant="light">
          {{ testResult.connected ? '连接正常' : '连接失败' }}
        </t-tag>
        <span class="oss-config__test-message">{{ testResult.message }}</span>
      </div>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除配置"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemStorageOssConfig' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  changeOssConfigDefault,
  createOssConfig,
  getOssConfig,
  listOssConfigs,
  removeOssConfig,
  testOssConfig,
  updateOssConfig,
} from '@/api/system/ossConfig';
import type {
  OssConfigDTO,
  OssConfigSaveParams,
  OssConfigTestParams,
  OssConfigTestResult,
} from '@/api/system/ossConfigModel';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import STable from '@/components/s-table/index.vue';
import { t } from '@/locales';

const loading = ref(false);
const saving = ref(false);
const list = ref<OssConfigDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');

const drawer = ref(false);
const form = ref<Partial<OssConfigSaveParams>>({});
const rules = {
  configKey: [{ required: true, message: '请输入配置key', type: 'error' as const }],
  endpoint: [{ required: true, message: '请输入访问站点', type: 'error' as const }],
  bucketName: [{ required: true, message: '请输入桶名称', type: 'error' as const }],
};

const confirmVisible = ref(false);
const deleteTarget = ref<OssConfigDTO | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确认删除配置「${deleteTarget.value.configKey}」？` : '',
);

const testVisible = ref(false);
const testing = ref(false);
const testForm = ref<OssConfigTestParams>({ accessKey: '', secretKey: '', bucketName: '', endpoint: '' });
const testResult = ref<OssConfigTestResult | null>(null);

const httpsOptions = [
  { label: 'HTTP', value: 'N' },
  { label: 'HTTPS', value: 'Y' },
];
const policyOptions = [
  { label: 'private', value: '0' },
  { label: 'public', value: '1' },
  { label: 'custom', value: '2' },
];

const columns = [
  { colKey: 'configKey', title: '配置key', width: 140, cell: 'configKey' },
  { colKey: 'endpoint', title: '访问站点', minWidth: 220, ellipsis: true },
  { colKey: 'bucketName', title: '桶名称', width: 130, ellipsis: true },
  { colKey: 'accessPolicy', title: '桶权限', width: 100, cell: 'accessPolicy' },
  { colKey: 'domain', title: '自定义域名', minWidth: 160, ellipsis: true },
  { colKey: 'remark', title: '备注', minWidth: 140, ellipsis: true },
  { colKey: 'updatedAt', title: '更新时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 200, fixed: 'right', cell: 'ops' },
];

function policyLabel(v: string) {
  return policyOptions.find((o) => o.value === v)?.label ?? v;
}
function policyTheme(v: string): 'success' | 'warning' | 'default' {
  if (v === '1') return 'success';
  if (v === '0') return 'warning';
  return 'default';
}

async function load() {
  loading.value = true;
  try {
    const res = await listOssConfigs({ page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined });
    list.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || t('common.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function onSearch(params: Record<string, unknown>) {
  keyword.value = String(params.keyword ?? '');
  page.value = 1;
  load();
}

async function open(row?: OssConfigDTO) {
  if (row) {
    try {
      const detail = await getOssConfig(row.id);
      form.value = {
        id: detail.id,
        configKey: detail.configKey,
        accessKey: detail.accessKey,
        secretKey: detail.secretKey,
        bucketName: detail.bucketName,
        prefix: detail.prefix,
        endpoint: detail.endpoint,
        domain: detail.domain,
        isHttps: detail.isHttps,
        region: detail.region,
        accessPolicy: detail.accessPolicy,
        remark: detail.remark,
      };
    } catch (e) {
      MessagePlugin.error((e as Error).message || '加载失败');
      return;
    }
  } else {
    form.value = {
      configKey: '',
      accessKey: '',
      secretKey: '',
      bucketName: '',
      prefix: '',
      endpoint: 'http://127.0.0.1:19000',
      domain: '',
      isHttps: 'N',
      region: '',
      accessPolicy: '1',
      remark: '',
    };
  }
  drawer.value = true;
}

async function save() {
  saving.value = true;
  try {
    const payload: OssConfigSaveParams = {
      id: form.value.id,
      configKey: (form.value.configKey ?? '').trim(),
      accessKey: form.value.accessKey ?? '',
      secretKey: form.value.secretKey ?? '',
      bucketName: (form.value.bucketName ?? '').trim(),
      prefix: form.value.prefix ?? '',
      endpoint: (form.value.endpoint ?? '').trim(),
      domain: form.value.domain ?? '',
      isHttps: form.value.isHttps ?? 'N',
      region: form.value.region ?? '',
      accessPolicy: form.value.accessPolicy ?? '1',
      remark: form.value.remark ?? '',
    };
    if (payload.id) await updateOssConfig(payload);
    else await createOssConfig(payload);
    MessagePlugin.success('保存成功');
    drawer.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function setDefault(row: OssConfigDTO) {
  try {
    await changeOssConfigDefault(row.id);
    MessagePlugin.success(`已将「${row.configKey}」设为默认，即时生效`);
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '设置失败');
  }
}

function del(row: OssConfigDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeOssConfig([deleteTarget.value.id]);
    MessagePlugin.success('已删除');
    confirmVisible.value = false;
    deleteTarget.value = null;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  } finally {
    saving.value = false;
  }
}

function openTest() {
  testForm.value = { accessKey: '', secretKey: '', bucketName: '', endpoint: 'http://127.0.0.1:19000', region: '' };
  testResult.value = null;
  testVisible.value = true;
}

async function runTest() {
  testing.value = true;
  try {
    testResult.value = await testOssConfig(testForm.value);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '测试失败');
  } finally {
    testing.value = false;
  }
}

onMounted(load);
</script>

<style scoped>
.oss-config__test-result {
  display: flex;
  gap: var(--td-comp-margin-s);
  align-items: center;
  margin-top: var(--td-comp-margin-l);
}

.oss-config__test-message {
  color: var(--td-text-color-secondary);
  font-size: 13px;
}
</style>
