<template>
  <div class="content-card">
    <s-table
      row-key="id"
      :columns="typeColumns as any"
      :data="types as any"
      :total="typeTotal"
      :page="typePage"
      :page-size="typePageSize"
      :loading="loading"
      keyword-placeholder="字典名称/类型"
      @search="onTypeSearch"
      @reset="
        () => {
          typeKeyword = '';
          typePage = 1;
          loadTypes();
        }
      "
      @refresh="loadTypes"
      @page-change="
        (p, ps) => {
          typePage = p;
          typePageSize = ps;
          loadTypes();
        }
      "
    >
      <template #toolbar>
        <t-button v-permission="'system:dict:add'" theme="primary" @click="openType()">新增字典类型</t-button>
      </template>
      <template #status="{ row }">
        <t-tag :theme="row.status === 0 ? 'success' : 'danger'" variant="light">
          {{ statusDict.labelOf(row.status) }}
        </t-tag>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link theme="primary" @click="onSelectType(row as DictTypeDTO)">数据项</t-link>
          <t-link v-permission="'system:dict:edit'" theme="primary" @click="openType(row as DictTypeDTO)">编辑</t-link>
          <t-link v-permission="'system:dict:remove'" theme="danger" @click="delType(row as DictTypeDTO)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <t-drawer
      v-model:visible="dataDrawerVisible"
      :header="selectedType ? `字典数据：${selectedType.dictName}（${selectedType.dictType}）` : '字典数据'"
      size="large"
      :footer="false"
    >
      <s-table
        row-key="id"
        :columns="dataColumns as any"
        :data="dataList as any"
        :total="dataTotal"
        :page="dataPage"
        :page-size="dataPageSize"
        :loading="loading"
        keyword-placeholder=""
        @search="
          () => {
            dataPage = 1;
            loadDataList();
          }
        "
        @reset="
          () => {
            dataPage = 1;
            loadDataList();
          }
        "
        @refresh="loadDataList"
        @page-change="
          (p, ps) => {
            dataPage = p;
            dataPageSize = ps;
            loadDataList();
          }
        "
      >
        <template #toolbar>
          <t-button v-permission="'system:dict:add'" theme="primary" :disabled="!selectedType" @click="openData()"
            >新增数据项</t-button
          >
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 0 ? 'success' : 'danger'" variant="light">
            {{ statusDict.labelOf(row.status) }}
          </t-tag>
        </template>
        <template #ops="{ row }">
          <t-space>
            <t-link v-permission="'system:dict:edit'" theme="primary" @click="openData(row as DictDataDTO)"
              >编辑</t-link
            >
            <t-link v-permission="'system:dict:remove'" theme="danger" @click="delData(row as DictDataDTO)"
              >删除</t-link
            >
          </t-space>
        </template>
      </s-table>
    </t-drawer>

    <t-dialog v-model:visible="typeDialog" :header="typeForm.id ? '编辑字典类型' : '新增字典类型'" @confirm="saveType">
      <t-form :data="typeForm" label-width="90px">
        <t-form-item label="字典名称"><t-input v-model="typeForm.dictName" /></t-form-item>
        <t-form-item label="字典类型"><t-input v-model="typeForm.dictType" :disabled="!!typeForm.id" /></t-form-item>
        <t-form-item label="排序"><t-input-number v-model="typeForm.orderNum" theme="column" /></t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="typeForm.status" :options="statusOptions" />
        </t-form-item>
        <t-form-item label="备注"><t-textarea v-model="typeForm.remark" /></t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog v-model:visible="dataDialog" :header="dataForm.id ? '编辑字典数据' : '新增字典数据'" @confirm="saveData">
      <t-form :data="dataForm" label-width="90px">
        <t-form-item label="标签"><t-input v-model="dataForm.dictLabel" /></t-form-item>
        <t-form-item label="值"><t-input v-model="dataForm.dictValue" /></t-form-item>
        <t-form-item label="排序"><t-input-number v-model="dataForm.listOrder" theme="column" /></t-form-item>
        <t-form-item label="状态">
          <t-radio-group v-model="dataForm.status" :options="statusOptions" />
        </t-form-item>
        <t-form-item label="默认">
          <t-radio-group v-model="dataForm.isDefault">
            <t-radio :value="1">是</t-radio>
            <t-radio :value="0">否</t-radio>
          </t-radio-group>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import {
  createDictData,
  createDictType,
  listDictData,
  listDictTypes,
  removeDictData,
  removeDictType,
  updateDictData,
  updateDictType,
} from '@/api/system/dict';
import type { DictDataDTO, DictTypeDTO } from '@/api/system/dictModel';
import STable from '@/components/s-table/index.vue';
import { useConfigStore, useDict, useDictStore } from '@/store';

const statusDict = useDict('sys_status');
const statusOptions = computed(() => statusDict.options.value.map((o) => ({ label: o.label, value: Number(o.value) })));
const dictStore = useDictStore();

const loading = ref(false);
const types = ref<DictTypeDTO[]>([]);
const typeTotal = ref(0);
const typePage = ref(1);
const typePageSize = ref(useConfigStore().pageDefault);
const typeKeyword = ref('');

const selectedType = ref<DictTypeDTO | null>(null);
const dataList = ref<DictDataDTO[]>([]);
const dataTotal = ref(0);
const dataPage = ref(1);
const dataPageSize = ref(useConfigStore().pageDefault);
const dataDrawerVisible = ref(false);

const typeDialog = ref(false);
const typeForm = ref<Partial<DictTypeDTO>>({});
const dataDialog = ref(false);
const dataForm = ref<Partial<DictDataDTO>>({});

const typeColumns = [
  { colKey: 'dictName', title: '字典名称', cell: 'dictName' },
  { colKey: 'dictType', title: '字典类型' },
  { colKey: 'orderNum', title: '排序', width: 80 },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'ops', title: '操作', width: 180, fixed: 'right', cell: 'ops' },
];

const dataColumns = [
  { colKey: 'dictLabel', title: '标签' },
  { colKey: 'dictValue', title: '值' },
  { colKey: 'listOrder', title: '排序', width: 80 },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'ops', title: '操作', width: 180, fixed: 'right', cell: 'ops' },
];

async function loadTypes() {
  loading.value = true;
  try {
    const res = await listDictTypes({
      page: typePage.value,
      pageSize: typePageSize.value,
      keyword: typeKeyword.value || undefined,
    });
    types.value = res.list || [];
    typeTotal.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载字典类型失败');
  } finally {
    loading.value = false;
  }
}

async function loadDataList() {
  if (!selectedType.value) {
    dataList.value = [];
    dataTotal.value = 0;
    return;
  }
  loading.value = true;
  try {
    const res = await listDictData({
      page: dataPage.value,
      pageSize: dataPageSize.value,
      dictType: selectedType.value.dictType,
    });
    dataList.value = res.list || [];
    dataTotal.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载字典数据失败');
  } finally {
    loading.value = false;
  }
}

function onTypeSearch(params: Record<string, unknown>) {
  typeKeyword.value = String(params.keyword ?? '');
  typePage.value = 1;
  loadTypes();
}

function onSelectType(row: DictTypeDTO) {
  selectedType.value = row;
  dataPage.value = 1;
  dataDrawerVisible.value = true;
  loadDataList();
}

function openType(row?: DictTypeDTO) {
  typeForm.value = row ? { ...row } : { status: 0, orderNum: 0, dictName: '', dictType: '', remark: '' };
  typeDialog.value = true;
}

async function saveType() {
  try {
    if (typeForm.value.id) await updateDictType(typeForm.value as DictTypeDTO);
    else await createDictType(typeForm.value as DictTypeDTO);
    MessagePlugin.success('保存成功');
    typeDialog.value = false;
    await loadTypes();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  }
}

async function delType(row: DictTypeDTO) {
  try {
    await removeDictType(row.id);
    MessagePlugin.success('已删除');
    dictStore.invalidate();
    if (selectedType.value?.id === row.id) {
      selectedType.value = null;
      dataList.value = [];
      dataDrawerVisible.value = false;
    }
    await loadTypes();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  }
}

function openData(row?: DictDataDTO) {
  dataForm.value = row
    ? { ...row }
    : {
        dictType: selectedType.value?.dictType,
        dictLabel: '',
        dictValue: '',
        listOrder: 0,
        status: 0,
        isDefault: 0,
        cssClass: '',
      };
  dataDialog.value = true;
}

async function saveData() {
  try {
    if (dataForm.value.id) await updateDictData(dataForm.value as DictDataDTO);
    else await createDictData(dataForm.value as DictDataDTO);
    MessagePlugin.success('保存成功');
    dictStore.invalidate(dataForm.value.dictType);
    dataDialog.value = false;
    await loadDataList();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  }
}

async function delData(row: DictDataDTO) {
  try {
    await removeDictData(row.id);
    MessagePlugin.success('已删除');
    dictStore.invalidate(row.dictType);
    await loadDataList();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  }
}

onMounted(() => {
  loadTypes();
});
</script>
