<template>
  <div class="content-card dict-page">
    <!-- 左：字典分组 -->
    <div class="dict-page__groups">
      <div class="dict-page__groups-toolbar">
        <span class="dict-page__groups-title">字典分组</span>
        <t-tooltip content="新增分组">
          <t-button v-permission="'system:dict:add'" variant="outline" shape="square" size="small" @click="openGroup()">
            <add-icon />
          </t-button>
        </t-tooltip>
      </div>
      <div
        class="dict-page__group-item dict-page__group-item--static"
        :class="{ 'is-active': activeGroupId === null }"
        @click="selectGroup(null)"
      >
        <span class="dict-page__group-name">全部分组</span>
      </div>
      <div
        v-for="g in groups"
        :key="g.id"
        class="dict-page__group-item"
        :class="{
          'is-active': activeGroupId === Number(g.id),
          'is-dragging': dragGroupId === g.id,
          'is-drop-target': dragOverGroupId === g.id && dragGroupId !== g.id,
        }"
        draggable="true"
        @click="selectGroup(Number(g.id))"
        @dragstart="onGroupDragStart(g)"
        @dragover.prevent="onGroupDragOver(g)"
        @drop.prevent="onGroupDrop(g)"
        @dragend="onGroupDragEnd"
      >
        <span class="dict-page__group-grip"><move-icon size="16" /></span>
        <span class="dict-page__group-body">
          <span class="dict-page__group-name">{{ g.groupName }}</span>
          <t-typography-text v-if="g.remark" theme="secondary" class="dict-page__group-remark">
            {{ g.remark }}
          </t-typography-text>
        </span>
        <span v-if="groupMenuVisible()" class="dict-page__group-ops" @click.stop>
          <t-dropdown trigger="click" :options="groupMenuOptions()" @click="(opt) => onGroupMenu(opt, g)">
            <t-button variant="text" shape="square" size="small">
              <more-icon />
            </t-button>
          </t-dropdown>
        </span>
      </div>
      <div
        class="dict-page__group-item dict-page__group-item--static"
        :class="{ 'is-active': activeGroupId === 0 }"
        @click="selectGroup(0)"
      >
        <span class="dict-page__group-name">未分组</span>
      </div>
    </div>

    <!-- 右：字典类型 -->
    <div class="dict-page__main">
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
        <template #groupName="{ row }">
          {{ groupOf(row.groupId) }}
        </template>
        <template #dictType="{ row }">
          <t-link theme="primary" @click="onSelectType(row as DictTypeDTO)">{{ row.dictType }}</t-link>
        </template>
        <template #status="{ row }">
          <s-tag-dict dict-type="sys_status" :value="row.status" />
        </template>
        <template #ops="{ row }">
          <t-space>
            <t-link theme="primary" @click="onSelectType(row as DictTypeDTO)">列表</t-link>
            <t-link v-permission="'system:dict:edit'" theme="primary" @click="openType(row as DictTypeDTO)"
              >编辑</t-link
            >
            <t-link v-permission="'system:dict:remove'" theme="danger" @click="delType(row as DictTypeDTO)"
              >删除</t-link
            >
          </t-space>
        </template>
      </s-table>
    </div>

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
        <template #dictLabel="{ row }">
          <span v-if="(row as DictDataDTO).theme === 'info'">{{ row.dictLabel }}</span>
          <t-tag
            v-else
            :theme="rowTheme(row as DictDataDTO)"
            :color="(row as DictDataDTO).color || undefined"
            variant="light"
          >
            {{ row.dictLabel }}
          </t-tag>
        </template>
        <template #status="{ row }">
          <s-tag-dict dict-type="sys_status" :value="row.status" />
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

    <!-- 字典类型表单 6 字段 >5 → 抽屉 -->
    <s-form-drawer
      v-model:visible="typeDialog"
      :title="typeForm.id ? '编辑字典类型' : '新增字典类型'"
      @confirm="saveType"
    >
      <t-form :data="typeForm" label-width="90px">
        <t-form-item label="字典名称"><t-input v-model="typeForm.dictName" /></t-form-item>
        <t-form-item label="字典类型"><t-input v-model="typeForm.dictType" :disabled="!!typeForm.id" /></t-form-item>
        <t-form-item label="分组">
          <t-select v-model="typeForm.groupId" :options="groupSelectOptions" placeholder="选择分组（可未分组）" />
        </t-form-item>
        <t-form-item label="排序"><t-input-number v-model="typeForm.orderNum" theme="column" /></t-form-item>
        <t-form-item label="状态">
          <s-radio-dict v-model="typeForm.status" dict-type="sys_status" />
        </t-form-item>
        <t-form-item label="备注"><t-textarea v-model="typeForm.remark" /></t-form-item>
      </t-form>
    </s-form-drawer>

    <!-- 数据项表单 6 字段（主题+颜色合一行）>5 → 抽屉 -->
    <s-form-drawer
      v-model:visible="dataDialog"
      :title="dataForm.id ? '编辑字典数据' : '新增字典数据'"
      @confirm="saveData"
    >
      <t-form :data="dataForm" label-width="90px">
        <t-form-item label="标签"><t-input v-model="dataForm.dictLabel" /></t-form-item>
        <t-form-item label="值"><t-input v-model="dataForm.dictValue" /></t-form-item>
        <t-form-item label="排序"><t-input-number v-model="dataForm.listOrder" theme="column" /></t-form-item>
        <t-form-item label="状态">
          <s-radio-dict v-model="dataForm.status" dict-type="sys_status" />
        </t-form-item>
        <t-form-item label="默认">
          <t-radio-group v-model="dataForm.isDefault">
            <t-radio :value="1">是</t-radio>
            <t-radio :value="0">否</t-radio>
          </t-radio-group>
        </t-form-item>
        <t-form-item label="主题">
          <div class="dict-page__theme-row">
            <t-radio-group v-model="dataForm.theme" @change="onThemePick">
              <t-radio v-for="t in themeOptions" :key="t.value" :value="t.value">
                <span v-if="t.swatch" class="dict-page__theme-swatch" :style="{ backgroundColor: t.swatch }" />
                {{ t.label }}
              </t-radio>
            </t-radio-group>
            <!-- 与主题互斥：填色则取消勾选，勾主题则清空颜色（watch） -->
            <t-color-picker v-model="dataForm.color" format="HEX" clearable />
          </div>
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <s-form-dialog v-model:visible="groupDialog" :title="groupForm.id ? '编辑分组' : '新增分组'" @confirm="saveGroup">
      <t-form ref="groupFormRef" :data="groupForm" :rules="groupRules" label-width="90px">
        <t-form-item label="分组名称" required>
          <t-input v-model="groupForm.groupName" placeholder="请输入分组名称" />
        </t-form-item>
        <t-form-item label="备注"><t-textarea v-model="groupForm.remark" /></t-form-item>
      </t-form>
    </s-form-dialog>
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemDict' });
import { AddIcon, MoreIcon, MoveIcon } from 'tdesign-icons-vue-next';
import type { DropdownOption, FormInstanceFunctions, FormRule } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref, watch } from 'vue';

import {
  createDictData,
  createDictGroup,
  createDictType,
  listDictData,
  listDictGroups,
  listDictTypes,
  removeDictData,
  removeDictGroup,
  removeDictType,
  updateDictData,
  updateDictGroup,
  updateDictType,
} from '@/api/system/dict';
import type { DictDataDTO, DictGroupDTO, DictTypeDTO } from '@/api/system/dictModel';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import { hasPerm } from '@/directives/permission';
import { useConfigStore, useDictStore } from '@/store';

const dictStore = useDictStore();

/** 主题单选（与颜色互斥）：info=纯文本、default=中性标签；色块预览与渲染一致 */
const themeOptions = [
  { value: 'info', label: '纯文本', swatch: '' },
  { value: 'default', label: '默认', swatch: 'var(--td-text-color-placeholder)' },
  { value: 'primary', label: '主色', swatch: 'var(--td-brand-color)' },
  { value: 'success', label: '成功', swatch: 'var(--td-success-color)' },
  { value: 'danger', label: '危险', swatch: 'var(--td-error-color)' },
  { value: 'warning', label: '警告', swatch: 'var(--td-warning-color)' },
];

/** 勾选主题 → 清空自定义色（互斥） */
function onThemePick(v: string | number | boolean) {
  if (v) dataForm.value.color = '';
}

/** 数据项行标签主题：合法令牌透传，default/空/未知回落 default（info 由模板分支处理） */
const TAG_THEMES = ['primary', 'success', 'danger', 'warning'];
function rowTheme(row: DictDataDTO): 'default' | 'primary' | 'success' | 'danger' | 'warning' {
  const t = row.theme || '';
  return (TAG_THEMES.includes(t) ? t : 'default') as 'default' | 'primary' | 'success' | 'danger' | 'warning';
}

const loading = ref(false);
const types = ref<DictTypeDTO[]>([]);
const typeTotal = ref(0);
const typePage = ref(1);
const typePageSize = ref(useConfigStore().pageDefault);
const typeKeyword = ref('');

const groups = ref<DictGroupDTO[]>([]);
/** null=全部分组；0=未分组；>0=指定分组 */
const activeGroupId = ref<number | null>(null);

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

// 互斥：填入自定义色 → 取消全部主题勾选；清空颜色保持未勾选，渲染回落中性 default
watch(
  () => dataForm.value.color,
  (c) => {
    if (c) dataForm.value.theme = '';
  },
);
const groupDialog = ref(false);
const groupForm = ref<Partial<DictGroupDTO>>({});
const groupFormRef = ref<FormInstanceFunctions>();
const groupRules: Record<string, FormRule[]> = {
  groupName: [{ required: true, message: '请输入分组名称', type: 'error' }],
};
const dragGroupId = ref<string | null>(null);
const dragOverGroupId = ref<string | null>(null);
const groupSorting = ref(false);

const typeColumns = [
  { colKey: 'dictName', title: '字典名称', cell: 'dictName' },
  { colKey: 'dictType', title: '字典类型', cell: 'dictType' },
  { colKey: 'groupName', title: '分组', width: 110, cell: 'groupName' },
  { colKey: 'orderNum', title: '排序', width: 80 },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'ops', title: '操作', width: 180, fixed: 'right', cell: 'ops' },
];

const dataColumns = [
  { colKey: 'dictLabel', title: '标签', cell: 'dictLabel' },
  { colKey: 'dictValue', title: '值' },
  { colKey: 'listOrder', title: '排序', width: 80 },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'ops', title: '操作', width: 180, fixed: 'right', cell: 'ops' },
];

const groupSelectOptions = computed(() => [
  { label: '未分组', value: 0 },
  ...groups.value.map((g) => ({ label: g.groupName, value: Number(g.id) })),
]);

function groupOf(groupId: number | undefined): string {
  const id = Number(groupId ?? 0);
  if (!id) return '未分组';
  const hit = groups.value.find((g) => Number(g.id) === id);
  return hit ? hit.groupName : String(id);
}

async function loadGroups() {
  try {
    const res = await listDictGroups();
    groups.value = res.list || [];
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载分组失败');
  }
}

// ===== 分组拖拽排序（move-icon 手柄，drop 后持久化 orderNum） =====
function onGroupDragStart(g: DictGroupDTO) {
  dragGroupId.value = g.id;
}

function onGroupDragOver(g: DictGroupDTO) {
  dragOverGroupId.value = g.id;
}

function onGroupDragEnd() {
  dragGroupId.value = null;
  dragOverGroupId.value = null;
}

async function onGroupDrop(target: DictGroupDTO) {
  const from = dragGroupId.value;
  dragGroupId.value = null;
  dragOverGroupId.value = null;
  if (!from || from === target.id || groupSorting.value) return;
  const list = [...groups.value];
  const fi = list.findIndex((g) => g.id === from);
  const ti = list.findIndex((g) => g.id === target.id);
  if (fi < 0 || ti < 0) return;
  const [moved] = list.splice(fi, 1);
  list.splice(ti, 0, moved);
  groupSorting.value = true;
  try {
    for (let i = 0; i < list.length; i += 1) {
      const g = list[i];
      if (g.orderNum === i + 1) continue;
      await updateDictGroup({ ...g, orderNum: i + 1 });
    }
    await loadGroups();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存分组顺序失败');
    await loadGroups();
  } finally {
    groupSorting.value = false;
  }
}

function selectGroup(id: number | null) {
  if (activeGroupId.value === id) return;
  activeGroupId.value = id;
  typePage.value = 1;
  loadTypes();
}

async function loadTypes() {
  loading.value = true;
  try {
    const res = await listDictTypes({
      page: typePage.value,
      pageSize: typePageSize.value,
      keyword: typeKeyword.value || undefined,
      // null=全部不传；0 保留（未分组）
      groupId: activeGroupId.value === null ? undefined : activeGroupId.value,
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
  const defaultGroup = activeGroupId.value !== null && activeGroupId.value > 0 ? activeGroupId.value : 0;
  typeForm.value = row
    ? { ...row }
    : { status: 0, orderNum: 0, dictName: '', dictType: '', remark: '', groupId: defaultGroup };
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

function openGroup(row?: DictGroupDTO) {
  groupForm.value = row ? { ...row } : { groupName: '', orderNum: groups.value.length + 1, remark: '' };
  groupDialog.value = true;
}

async function saveGroup() {
  const validateResult = await groupFormRef.value?.validate();
  if (validateResult !== true) return;
  try {
    if (groupForm.value.id) await updateDictGroup(groupForm.value as DictGroupDTO);
    else await createDictGroup(groupForm.value as DictGroupDTO);
    MessagePlugin.success('保存成功');
    groupDialog.value = false;
    await loadGroups();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存分组失败');
  }
}

async function delGroup(row: DictGroupDTO) {
  try {
    await removeDictGroup(row.id);
    MessagePlugin.success('已删除');
    if (activeGroupId.value === Number(row.id)) {
      activeGroupId.value = null;
      typePage.value = 1;
      await loadTypes();
    }
    await loadGroups();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除分组失败');
  }
}

// ===== 分组行 more 菜单（编辑 / 删除） =====
function groupMenuVisible(): boolean {
  return hasPerm('system:dict:edit') || hasPerm('system:dict:remove');
}

function groupMenuOptions(): DropdownOption[] {
  const options: DropdownOption[] = [];
  if (hasPerm('system:dict:edit')) {
    options.push({ content: '编辑', value: 'edit' });
  }
  if (hasPerm('system:dict:remove')) {
    options.push({ content: '删除', value: 'delete', theme: 'error' });
  }
  return options;
}

function onGroupMenu(option: DropdownOption, row: DictGroupDTO) {
  switch (String(option.value)) {
    case 'edit':
      openGroup(row);
      break;
    case 'delete':
      delGroup(row);
      break;
    default:
      break;
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
        theme: 'default',
        color: '',
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
  loadGroups();
  loadTypes();
});
</script>
<style lang="less" scoped>
.dict-page {
  display: flex;
  gap: var(--td-comp-margin-xl);
  min-height: 480px;
}

.dict-page__groups {
  width: 200px;
  flex-shrink: 0;
  border-right: 1px solid var(--td-border-level-1-color);
  padding-right: var(--td-comp-margin-m);
}

.dict-page__groups-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--td-comp-margin-m);
}

.dict-page__groups-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.dict-page__group-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--td-comp-margin-s);
  padding: 8px 12px;
  border-radius: var(--td-radius-medium);
  cursor: pointer;
  color: var(--td-text-color-primary);
  font-size: 13px;

  &:hover {
    background-color: var(--td-bg-color-container-hover);
  }

  &.is-active {
    background-color: var(--td-brand-color-light);
    color: var(--td-brand-color);
    font-weight: 600;
  }

  &.is-dragging {
    opacity: 0.55;
  }

  &.is-drop-target {
    outline: 1px dashed var(--td-brand-color);
    outline-offset: -1px;
  }
}

// 静态项（全部分组/未分组）让开 grip 宽度，与分组名称左对齐
.dict-page__group-item--static {
  padding-left: 36px;
}

.dict-page__group-grip {
  display: inline-flex;
  flex-shrink: 0;
  color: var(--td-text-color-placeholder);
  cursor: grab;

  &:active {
    cursor: grabbing;
  }
}

.dict-page__group-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.dict-page__group-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dict-page__group-remark {
  font-size: 12px;
  font-weight: 400;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin: 0;
}

.dict-page__group-ops {
  display: none;
  flex-shrink: 0;
}

.dict-page__group-item:hover .dict-page__group-ops,
.dict-page__group-item.is-active .dict-page__group-ops {
  display: inline-flex;
}

.dict-page__main {
  flex: 1;
  min-width: 0;
}

// 主题单选与取色器同一行（互斥单选）
.dict-page__theme-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--td-comp-margin-m);
}

// 主题单选的色块预览
.dict-page__theme-swatch {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  margin-right: 6px;
  vertical-align: -1px;
}
</style>
