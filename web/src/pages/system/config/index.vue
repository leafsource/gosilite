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
      keyword-placeholder="参数名称/键"
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
        <t-button v-permission="'system:config:add'" theme="primary" @click="open()">新增参数</t-button>
      </template>
      <template #batch="{ selected, clear }">
        <t-button
          size="small"
          variant="outline"
          theme="danger"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchRemove(selected as ConfigDTO[], clear)"
        >
          批量删除
        </t-button>
        <t-button
          size="small"
          variant="outline"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchSetStatus(selected as ConfigDTO[], 0, clear)"
        >
          批量启用
        </t-button>
        <t-button
          size="small"
          variant="outline"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchSetStatus(selected as ConfigDTO[], 1, clear)"
        >
          批量禁用
        </t-button>
      </template>
      <template #configType="{ row }">
        <s-tag-dict dict-type="sys_config_type" :value="row.configType" />
      </template>
      <template #status="{ row }">
        <s-tag-dict dict-type="sys_status" :value="row.status" />
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:config:edit'" theme="primary" @click="open(row as ConfigDTO)">编辑</t-link>
          <t-link v-permission="'system:config:remove'" theme="danger" @click="del(row as ConfigDTO)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <s-form-dialog v-model:visible="dialog" :title="form.id ? '编辑参数' : '新增参数'" @confirm="save">
      <t-form :data="form" label-width="90px">
        <t-form-item label="参数名称"><t-input v-model="form.configName" /></t-form-item>
        <t-form-item label="参数键"><t-input v-model="form.configKey" :disabled="!!form.id" /></t-form-item>
        <t-form-item label="参数值"><t-textarea v-model="form.configValue" /></t-form-item>
        <t-form-item label="类型">
          <s-select-dict v-model="form.configType" dict-type="sys_config_type" />
        </t-form-item>
        <t-form-item label="状态">
          <s-radio-dict v-model="form.status" dict-type="sys_status" />
        </t-form-item>
        <t-form-item label="备注"><t-textarea v-model="form.remark" /></t-form-item>
      </t-form>
    </s-form-dialog>
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemConfig' });
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, ref } from 'vue';

import { createConfig, listConfigs, removeConfig, updateConfig } from '@/api/system/config';
import type { ConfigDTO } from '@/api/system/configModel';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import SSelectDict from '@/components/s-select-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useDict } from '@/store';

const statusDict = useDict('sys_status');

const loading = ref(false);
const list = ref<ConfigDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');

const dialog = ref(false);
const form = ref<Partial<ConfigDTO>>({});

const columns = [
  { colKey: 'configName', title: '参数名称', minWidth: 120 },
  { colKey: 'configKey', title: '参数键', minWidth: 140 },
  { colKey: 'configValue', title: '参数值', minWidth: 140, ellipsis: true },
  { colKey: 'configType', title: '类型', width: 100, cell: 'configType' },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'ops', title: '操作', width: 160, fixed: 'right', cell: 'ops' },
];

async function load() {
  loading.value = true;
  try {
    const res = await listConfigs({ page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined });
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

function open(row?: ConfigDTO) {
  form.value = row
    ? { ...row }
    : { status: 0, configType: 'system', configName: '', configKey: '', configValue: '', remark: '' };
  dialog.value = true;
}

async function save() {
  try {
    if (form.value.id) await updateConfig(form.value as ConfigDTO);
    else await createConfig(form.value as ConfigDTO);
    MessagePlugin.success('保存成功');
    dialog.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  }
}

async function del(row: ConfigDTO) {
  try {
    await removeConfig(row.id);
    MessagePlugin.success('已删除');
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  }
}

const batchLoading = ref(false);

async function batchRemove(selected: ConfigDTO[], clear: () => void) {
  if (!selected.length) return;
  batchLoading.value = true;
  try {
    await Promise.all(selected.map((row) => removeConfig(row.id)));
    MessagePlugin.success(`已删除 ${selected.length} 项参数`);
    clear();
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '批量删除失败');
  } finally {
    batchLoading.value = false;
  }
}

async function batchSetStatus(selected: ConfigDTO[], status: number, clear: () => void) {
  if (!selected.length) return;
  batchLoading.value = true;
  try {
    await Promise.all(selected.map((row) => updateConfig({ ...row, status })));
    MessagePlugin.success(`${statusDict.labelOf(status)}成功 ${selected.length} 项参数`);
    clear();
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '批量操作失败');
  } finally {
    batchLoading.value = false;
  }
}

onMounted(load);
</script>
