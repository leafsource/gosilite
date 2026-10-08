<template>
  <div class="menu-page">
    <!-- 树形表格：s-table tree 模式（内部 EnhancedTable），全量树数据不走分页 -->
    <s-table
      row-key="id"
      :columns="columns"
      :data="filteredRows as any"
      :loading="loading"
      :tree="treeConfig"
      :paginated="false"
      :selectable="false"
      keyword-placeholder="菜单名称/权限标识"
      @search="(p) => (filterText = String(p.keyword ?? ''))"
      @reset="() => (filterText = '')"
      @refresh="loadData"
    >
      <template #toolbar>
        <t-button v-permission="'system:menu:add'" theme="primary" @click="openCreate()">新增菜单</t-button>
      </template>
      <template #menuType="{ row }">
        <s-tag-dict dict-type="sys_menu_type" :value="row.menuType" />
      </template>
      <template #actionType="{ row }">
        <s-tag-dict v-if="row.menuType === 'F'" dict-type="sys_action_type" :value="row.actionType" />
      </template>
      <template #visible="{ row }">
        <t-tag :theme="row.visible ? 'success' : 'default'" variant="light">
          {{ row.visible ? '显示' : '隐藏' }}
        </t-tag>
      </template>
      <template #status="{ row }">
        <s-tag-dict dict-type="sys_status" :value="row.status" />
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:menu:add'" theme="primary" @click="openCreate(row.id)">新增子菜单</t-link>
          <t-link v-permission="'system:menu:edit'" theme="primary" @click="openEdit(row)">编辑</t-link>
          <t-link v-permission="'system:menu:remove'" theme="danger" @click="openDelete(row)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 11 字段 >5 → 抽屉 -->
    <s-form-drawer
      v-model:visible="dialogVisible"
      :title="isEdit ? '编辑菜单' : '新增菜单'"
      :loading="saving"
      @confirm="onSave"
    >
      <t-form ref="formRef" :data="form" :rules="rules" label-width="90px">
        <t-form-item label="上级菜单" name="parentId">
          <t-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            clearable
            placeholder="选择上级菜单"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="菜单类型" name="menuType">
          <s-radio-dict v-model="form.menuType" dict-type="sys_menu_type" />
        </t-form-item>
        <t-form-item label="菜单名称" name="menuName">
          <t-input v-model="form.menuName" placeholder="请输入菜单名称" />
        </t-form-item>
        <!-- 字段显隐按类型：M/C 需路由+组件；F 不进路由，必填权限标识+操作类型 -->
        <t-form-item v-if="form.menuType !== 'F'" label="路由路径" name="path">
          <t-input v-model="form.path" :placeholder="pathPlaceholder" />
        </t-form-item>
        <t-form-item v-if="form.menuType !== 'F'" label="组件路径" name="component">
          <t-input v-model="form.component" :placeholder="componentPlaceholder" />
        </t-form-item>
        <t-form-item v-if="form.menuType !== 'M'" label="权限标识" name="perms">
          <t-input v-model="form.perms" :placeholder="permsPlaceholder" />
        </t-form-item>
        <t-form-item v-if="form.menuType === 'F'" label="操作类型" name="actionType">
          <s-select-dict v-model="form.actionType" dict-type="sys_action_type" placeholder="请选择操作类型" />
        </t-form-item>
        <t-form-item v-if="form.menuType !== 'F'" label="图标" name="icon">
          <s-select-icon v-model="form.icon" />
        </t-form-item>
        <t-form-item label="排序" name="orderNum">
          <t-input-number v-model="form.orderNum" :min="0" theme="column" style="width: 100%" />
        </t-form-item>
        <t-form-item v-if="form.menuType !== 'F'" label="是否可见" name="visible">
          <t-switch v-model="form.visible">
            <template #label>{{ form.visible ? '显示' : '隐藏' }}</template>
          </t-switch>
        </t-form-item>
        <t-form-item label="状态" name="status">
          <s-radio-dict v-model="form.status" dict-type="sys_status" />
        </t-form-item>
      </t-form>
    </s-form-drawer>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除菜单"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions, FormRule, PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { MenuNode, MenuSaveParams, MenuTreeItem } from '@/api/system/menu';
import { createMenu, fetchMenuTree, removeMenu, updateMenu } from '@/api/system/menu';
import SFormDrawer from '@/components/s-form-drawer/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import SSelectDict from '@/components/s-select-dict/index.vue';
import SSelectIcon from '@/components/s-select-icon/index.vue';
import STable from '@/components/s-table/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';

const columns: PrimaryTableCol[] = [
  { title: '菜单名称', colKey: 'label', minWidth: 220 },
  { title: '类型', colKey: 'menuType', width: 90 },
  { title: '操作类型', colKey: 'actionType', width: 100 },
  { title: '权限标识', colKey: 'perms', width: 180 },
  { title: '路径', colKey: 'path', width: 160 },
  { title: '组件', colKey: 'component', width: 180 },
  { title: '排序', colKey: 'orderNum', width: 80 },
  { title: '可见', colKey: 'visible', width: 80 },
  { title: '状态', colKey: 'status', width: 80 },
  { title: '操作', colKey: 'op', width: 220, fixed: 'right' },
];

const treeConfig = {
  childrenKey: 'children',
  treeNodeColumnIndex: 0,
  defaultExpandAll: true,
};

const loading = ref(false);
const saving = ref(false);
const rows = ref<MenuNode[]>([]);
const filterText = ref('');

function normalize(nodes: MenuNode[]): MenuNode[] {
  return nodes.map((node) => ({
    ...node,
    children: node.children && node.children.length ? normalize(node.children) : undefined,
  }));
}

const filteredRows = computed<MenuNode[]>(() => {
  const keyword = filterText.value.trim().toLowerCase();
  if (!keyword) return rows.value;
  const walk = (nodes: MenuNode[]): MenuNode[] =>
    nodes
      .map((node) => ({ ...node, children: node.children ? walk(node.children) : node.children }))
      .filter(
        (node) =>
          node.label.toLowerCase().includes(keyword) ||
          (node.perms || '').toLowerCase().includes(keyword) ||
          (node.children ? node.children.length > 0 : false),
      );
  return walk(rows.value);
});

async function loadData() {
  loading.value = true;
  try {
    const res = await fetchMenuTree();
    rows.value = normalize(res.list || []);
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

// ===== 新增 / 编辑 =====
const INITIAL_FORM: MenuSaveParams = {
  parentId: '0',
  menuName: '',
  menuType: 'C',
  path: '',
  component: '',
  perms: '',
  icon: '',
  orderNum: 0,
  visible: true,
  status: 0,
  actionType: '',
};
const dialogVisible = ref(false);
const form = ref<MenuSaveParams>({ ...INITIAL_FORM });
const formRef = ref<FormInstanceFunctions>();
const isEdit = computed(() => form.value.id != null);

// 必填规则随类型变：M/C 要路由+组件；F 要权限标识+操作类型（与后端 validateMenu 同规则）
const rules = computed<Record<string, FormRule[]>>(() => {
  const isButton = form.value.menuType === 'F';
  return {
    menuName: [{ required: true, whitespace: true, message: '请输入菜单名称' }],
    path: [{ required: true, whitespace: true, message: '请输入路由路径' }],
    component: [{ required: true, whitespace: true, message: '请输入组件路径' }],
    perms: isButton ? [{ required: true, whitespace: true, message: '请输入权限标识' }] : [],
    actionType: isButton ? [{ required: true, message: '请选择操作类型' }] : [],
  };
});

// 根目录显式成项（value='0'），避免 t-tree-select 回显原始值 0
const parentOptions = computed<MenuTreeItem[]>(() => {
  const root: MenuTreeItem = { value: '0', label: '根目录' };
  const source = [root, ...toTreeItems(rows.value)];
  if (form.value.id == null) return source;
  return pruneNode(source, form.value.id);
});

const pathPlaceholder = computed(() => (form.value.menuType === 'M' ? '如 /system（目录）' : '如 /system/user'));
const componentPlaceholder = computed(() =>
  form.value.menuType === 'M' ? '目录固定填 layout' : '如 system/user/index',
);
const permsPlaceholder = computed(() =>
  form.value.menuType === 'F' ? '如 system:user:add（按钮必填）' : '如 system:user:list（选填）',
);

function toTreeItems(nodes: MenuNode[]): MenuTreeItem[] {
  return nodes.map((node) => ({
    value: node.id,
    label: node.label,
    menuType: node.menuType,
    children: node.children && node.children.length ? toTreeItems(node.children) : undefined,
  }));
}

function pruneNode(items: MenuTreeItem[], excludeId: string): MenuTreeItem[] {
  return items
    .filter((item) => item.value !== excludeId)
    .map((item) => ({ ...item, children: item.children ? pruneNode(item.children, excludeId) : undefined }));
}

function openCreate(parentId?: string) {
  form.value = { ...INITIAL_FORM, parentId: parentId ?? '0' };
  dialogVisible.value = true;
}

function openEdit(row: MenuNode) {
  form.value = {
    id: row.id,
    parentId: row.parentId,
    menuName: row.label,
    menuType: row.menuType,
    path: row.path,
    component: row.component,
    perms: row.perms,
    icon: row.icon,
    orderNum: row.orderNum,
    visible: row.visible,
    status: row.status,
    actionType: row.actionType ?? '',
  };
  dialogVisible.value = true;
}

async function onSave() {
  // t-form rules 校验（含必填星号），失败时错误文案内联展示
  const validateResult = await formRef.value?.validate();
  if (validateResult !== true) return;
  saving.value = true;
  try {
    const payload: MenuSaveParams = { ...form.value, parentId: form.value.parentId || '0' };
    // 归一不适用字段，与后端校验保持一致
    if (payload.menuType === 'M') {
      payload.perms = '';
      payload.actionType = '';
    }
    if (payload.menuType === 'C') {
      payload.actionType = '';
    }
    if (payload.menuType === 'F') {
      payload.path = '';
      payload.component = '';
      payload.icon = '';
    }
    if (payload.id != null) {
      await updateMenu(payload);
    } else {
      await createMenu(payload);
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

// ===== 删除 =====
const confirmVisible = ref(false);
const deleteTarget = ref<MenuNode | null>(null);
const confirmBody = computed(() =>
  deleteTarget.value ? `确定删除菜单「${deleteTarget.value.label}」吗？存在子菜单时将被拒绝。` : '',
);

function openDelete(row: MenuNode) {
  deleteTarget.value = row;
  confirmVisible.value = true;
}

async function onConfirmDelete() {
  if (!deleteTarget.value) return;
  saving.value = true;
  try {
    await removeMenu(deleteTarget.value.id);
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
.menu-page {
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
}
</style>
