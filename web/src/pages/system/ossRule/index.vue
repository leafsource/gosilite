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
      keyword-placeholder="规则名"
      :search-fields="searchFields"
      @search="onSearch"
      @reset="
        () => {
          ruleName = '';
          domain = '';
          status = undefined;
          isDefault = undefined;
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
        <t-button v-permission="'system:ossRule:add'" theme="primary" @click="open()">新增规则</t-button>
        <t-button
          v-permission="'system:ossRule:refresh'"
          variant="outline"
          :loading="refreshing"
          @click="refreshCache"
        >
          刷新缓存
        </t-button>
      </template>
      <template #mimeType="{ row }">
        <span class="oss-rule__mime" :title="(row as OssRuleDTO).mimeType">
          {{ (row as OssRuleDTO).mimeType }}
        </span>
      </template>
      <template #rule="{ row }">
        <span class="oss-rule__rule" :title="(row as OssRuleDTO).rule">{{ (row as OssRuleDTO).rule }}</span>
      </template>
      <template #isOverwrite="{ row }">
        <t-switch
          :value="(row as OssRuleDTO).isOverwrite === 'Y'"
          :disabled="!hasEdit"
          size="small"
          @change="(v) => toggleOverwrite(row as OssRuleDTO, v as boolean)"
        />
      </template>
      <template #isDefault="{ row }">
        <t-tag v-if="(row as OssRuleDTO).isDefault === 'Y'" theme="success" variant="light">默认</t-tag>
        <t-tag v-else theme="default" variant="light">普通</t-tag>
      </template>
      <template #status="{ row }">
        <t-switch
          :value="(row as OssRuleDTO).status === '1'"
          :disabled="!hasEdit"
          size="small"
          @change="(v) => toggleStatus(row as OssRuleDTO, v as boolean)"
        />
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:ossRule:edit'" theme="primary" @click="open(row as OssRuleDTO)">编辑</t-link>
          <t-link v-permission="'system:ossRule:remove'" theme="danger" @click="del(row as OssRuleDTO)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 字段 8 个 >5 → s-form-drawer（AGENTS 前端规则） -->
    <s-form-drawer
      v-model:visible="drawer"
      :title="form.id ? '编辑规则' : '新增规则'"
      size="520px"
      :loading="saving"
      @confirm="save"
    >
      <t-form :data="form" :rules="rules" label-align="right" label-width="96px" @submit="save">
        <t-form-item label="规则名称" name="ruleName" help="例如：80x80，则字段名称将输出字段名_80x80">
          <t-input v-model="form.ruleName" :maxlength="255" clearable placeholder="如 80x80" />
        </t-form-item>
        <t-form-item label="匹配域名" name="domain" help="只有 URL 中包含匹配的域名，该规则才生效">
          <t-input v-model="form.domain" :maxlength="255" clearable placeholder="如 cdn.example.com" />
        </t-form-item>
        <t-form-item label="媒体类型" name="mimeType" help="逗号分隔；匹配 contentType 或文件后缀，如 image,.png">
          <t-input v-model="form.mimeType" :maxlength="255" clearable placeholder="如 image,.png" />
        </t-form-item>
        <t-form-item label="规则" name="rule" help="内置 domain、path、filename、url 变量，使用 #{#url} 的方式引用">
          <t-textarea
            v-model="form.rule"
            :maxlength="255"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="如 #{#url}?x-oss-process=image/resize,w_80"
          />
        </t-form-item>
        <t-form-item label="是否默认" name="isDefault">
          <t-radio-group v-model="form.isDefault" :options="ynOptions" />
        </t-form-item>
        <t-form-item label="启用状态" name="status">
          <t-radio-group v-model="form.status" :options="enableOptions" />
        </t-form-item>
        <t-form-item label="规则顺序" name="ruleSort">
          <t-input-number v-model="form.ruleSort" :min="0" :max="9999" style="width: 100%" />
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-input v-model="form.remark" :maxlength="500" clearable />
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除规则"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'SystemStorageOssRule' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  changeOssRuleOverwrite,
  createOssRule,
  listOssRules,
  refreshOssRuleCache,
  removeOssRule,
  updateOssRule,
} from '@/api/system/ossRule';
import type { OssRuleDTO, OssRuleListParams, OssRuleSaveParams } from '@/api/system/ossRuleModel';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import { hasPerm } from '@/directives/permission';

const loading = ref(false);
const refreshing = ref(false);
const list = ref<OssRuleDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const ruleName = ref('');
const domain = ref('');
const status = ref<string | undefined>();
const isDefault = ref<string | undefined>();

const drawer = ref(false);
const form = ref<Partial<OssRuleSaveParams>>({});
const rules = {
  ruleName: [{ required: true, message: '请输入规则名称', type: 'error' as const }],
  domain: [{ required: true, message: '请输入匹配域名', type: 'error' as const }],
  mimeType: [{ required: true, message: '请输入媒体类型', type: 'error' as const }],
  rule: [{ required: true, message: '请输入规则', type: 'error' as const }],
};

const ynOptions = [
  { label: '是', value: 'Y' },
  { label: '否', value: 'N' },
];
const enableOptions = [
  { label: '启用', value: '1' },
  { label: '停用', value: '0' },
];

const hasEdit = computed(() => hasPerm('system:ossRule:edit'));

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'domain', label: '匹配域名', type: 'input' },
  { name: 'status', label: '状态', type: 'select', options: enableOptions },
  { name: 'isDefault', label: '默认', type: 'select', options: ynOptions },
]);

const columns = [
  { colKey: 'ruleName', title: '规则名称', width: 120 },
  { colKey: 'domain', title: '匹配域名', minWidth: 160, ellipsis: true },
  { colKey: 'mimeType', title: '媒体类型', width: 140, cell: 'mimeType' },
  { colKey: 'rule', title: '规则', minWidth: 240, cell: 'rule' },
  { colKey: 'isOverwrite', title: '覆盖字段值', width: 110, cell: 'isOverwrite' },
  { colKey: 'isDefault', title: '默认', width: 80, cell: 'isDefault' },
  { colKey: 'status', title: '启用', width: 80, cell: 'status' },
  { colKey: 'ruleSort', title: '顺序', width: 70 },
  { colKey: 'ops', title: '操作', width: 120, fixed: 'right', cell: 'ops' },
];

async function load() {
  loading.value = true;
  try {
    const params: OssRuleListParams = {
      page: page.value,
      pageSize: pageSize.value,
      ruleName: ruleName.value || undefined,
      domain: domain.value || undefined,
      status: status.value,
      isDefault: isDefault.value,
    };
    const res = await listOssRules(params);
    list.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function onSearch(params: Record<string, unknown>) {
  ruleName.value = String(params.keyword ?? '');
  domain.value = String(params.domain ?? '');
  status.value = params.status === undefined ? undefined : String(params.status);
  isDefault.value = params.isDefault === undefined ? undefined : String(params.isDefault);
  page.value = 1;
  load();
}

function open(row?: OssRuleDTO) {
  form.value = row
    ? {
        id: row.id,
        ruleName: row.ruleName,
        domain: row.domain,
        mimeType: row.mimeType,
        rule: row.rule,
        isDefault: row.isDefault,
        status: row.status,
        ruleSort: row.ruleSort,
        remark: row.remark,
      }
    : { ruleName: '', domain: '', mimeType: 'image', rule: '', isDefault: 'N', status: '1', ruleSort: 0, remark: '' };
  drawer.value = true;
}

async function save() {
  saving.value = true;
  try {
    const payload: OssRuleSaveParams = {
      id: form.value.id,
      ruleName: (form.value.ruleName ?? '').trim(),
      domain: (form.value.domain ?? '').trim(),
      mimeType: (form.value.mimeType ?? '').trim(),
      rule: (form.value.rule ?? '').trim(),
      isDefault: form.value.isDefault ?? 'N',
      status: form.value.status ?? '1',
      ruleSort: Number(form.value.ruleSort ?? 0),
      remark: form.value.remark ?? '',
    };
    if (payload.id) await updateOssRule(payload);
    else await createOssRule(payload);
    MessagePlugin.success('保存成功');
    drawer.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  } finally {
    saving.value = false;
  }
}

const confirmVisible = ref(false);
const saving = ref(false);
const deleteTarget = ref<OssRuleDTO | null>(null);
const confirmBody = computed(() => (deleteTarget.value ? `确认删除规则「${deleteTarget.value.ruleName}」？` : ''));

function del(row: OssRuleDTO) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeOssRule([deleteTarget.value.id]);
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

async function toggleOverwrite(row: OssRuleDTO, on: boolean) {
  try {
    await changeOssRuleOverwrite(row.id, on ? 'Y' : 'N');
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '修改失败');
  }
}

async function toggleStatus(row: OssRuleDTO, on: boolean) {
  try {
    await updateOssRule({
      id: row.id,
      ruleName: row.ruleName,
      domain: row.domain,
      mimeType: row.mimeType,
      rule: row.rule,
      isDefault: row.isDefault,
      status: on ? '1' : '0',
      ruleSort: row.ruleSort,
      remark: row.remark,
    });
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '修改失败');
  }
}

async function refreshCache() {
  refreshing.value = true;
  try {
    await refreshOssRuleCache();
    MessagePlugin.success('缓存已刷新');
  } catch (e) {
    MessagePlugin.error((e as Error).message || '刷新失败');
  } finally {
    refreshing.value = false;
  }
}

onMounted(load);
</script>

<style scoped>
.oss-rule__mime,
.oss-rule__rule {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
</style>
