<template>
  <t-drawer
    v-model:visible="visibleProxy"
    size="large"
    :header="renderHeader"
    :footer="false"
    destroy-on-close
    @close="onClose"
  >
    <t-loading :loading="loading" text="加载中…" style="height: 100%">
      <div class="auth-drawer">
        <div class="auth-drawer__tabs" role="tablist">
          <button
            type="button"
            class="auth-drawer__tab"
            :class="{ 'is-active': activeTab === 'menu' }"
            role="tab"
            :aria-selected="activeTab === 'menu'"
            @click="activeTab = 'menu'"
          >
            菜单权限
            <span class="auth-drawer__tab-count">{{ menuStats.menus }}</span>
          </button>
          <button
            type="button"
            class="auth-drawer__tab"
            :class="{ 'is-active': activeTab === 'dept' }"
            role="tab"
            :aria-selected="activeTab === 'dept'"
            @click="activeTab = 'dept'"
          >
            数据权限
            <span v-if="isCustomScope" class="auth-drawer__tab-count">{{ deptCheckedIds.length }}</span>
          </button>
        </div>

        <div class="auth-drawer__body">
          <menu-perm-table v-show="activeTab === 'menu'" ref="menuTableRef" :menus="menuNodes" @stats="onMenuStats" />

          <div v-show="activeTab === 'dept'" class="auth-drawer__scope">
            <s-radio-dict v-model="dataScope" dict-type="sys_data_scope" />
            <template v-if="isCustomScope">
              <div class="auth-drawer__dept-toolbar">
                <t-input v-model="deptFilter" placeholder="搜索部门" clearable />
                <t-button variant="outline" size="small" @click="deptCheckedIds = collectTreeIds(deptTreeData)">
                  全选
                </t-button>
                <t-button variant="outline" size="small" @click="deptCheckedIds = []">清空</t-button>
              </div>
              <t-tree
                :data="filteredDeptTree"
                checkable
                value-mode="all"
                hover
                :value="deptCheckedIds"
                :expanded="deptExpanded"
                @expand="onDeptExpand"
                @change="onDeptCheck"
              />
            </template>
            <div v-else class="auth-drawer__scope-hint">选择「自定义部门」后，可在此勾选该角色可见的部门。</div>
          </div>
        </div>

        <div class="auth-drawer__footer">
          <span class="auth-drawer__summary">
            菜单 <b>{{ menuStats.menus }}</b> · 权限 <b>{{ menuStats.perms }}</b> · 数据范围
            <b>{{ dataScopeLabel }}</b>
            <template v-if="isCustomScope"
              >（部门 <b>{{ deptCheckedIds.length }}</b
              >）</template
            >
          </span>
          <div class="auth-drawer__actions">
            <t-button variant="outline" @click="visibleProxy = false">取消</t-button>
            <t-button theme="primary" :loading="saving" @click="onSave">保存</t-button>
          </div>
        </div>
      </div>
    </t-loading>
  </t-drawer>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, h, ref, watch } from 'vue';

import type { DeptNode, DeptTreeItem } from '@/api/system/dept';
import { fetchDeptTree } from '@/api/system/dept';
import type { RoleAuthMenuNode, RoleDTO } from '@/api/system/role';
import { fetchRoleAuthorize, saveRoleAuthorize } from '@/api/system/role';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import { useDict } from '@/store';

import MenuPermTable from './MenuPermTable.vue';

defineOptions({ name: 'AuthorizeDrawer' });

const props = defineProps<{
  visible: boolean;
  role: RoleDTO | null;
}>();
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'saved'): void;
}>();

const visibleProxy = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
});

const loading = ref(false);
const saving = ref(false);
const activeTab = ref<'menu' | 'dept'>('menu');
const menuTableRef = ref<InstanceType<typeof MenuPermTable> | null>(null);
const menuNodes = ref<RoleAuthMenuNode[]>();
const menuStats = ref({ menus: 0, perms: 0 });

// 数据权限：sys_data_scope 字典（1-5），5=自定义部门才展开部门树
const CUSTOM_DATA_SCOPE = 5;
const dataScopeDict = useDict('sys_data_scope');
const dataScope = ref(1);
const isCustomScope = computed(() => dataScope.value === CUSTOM_DATA_SCOPE);
const dataScopeLabel = computed(() => dataScopeDict.labelOf(dataScope.value));

// 部门树全量不变，会话内缓存
let deptTreeCache: DeptTreeItem[] | null = null;

// ===== 数据部门 =====
const deptTreeData = ref<DeptTreeItem[]>([]);
const deptCheckedIds = ref<string[]>([]);
const deptExpanded = ref<string[]>([]);
const deptFilter = ref('');

const filteredDeptTree = computed(() => filterTree(deptTreeData.value, deptFilter.value.trim()));

interface TreeNode {
  value: string;
  label: string;
  children?: TreeNode[];
}

// id 统一归一化为 string 参与树勾选/展开比对，保存时再转回 number（后端 []int64）
function collectTreeIds(nodes: TreeNode[], acc: string[] = []): string[] {
  for (const node of nodes) {
    acc.push(node.value);
    if (node.children) collectTreeIds(node.children, acc);
  }
  return acc;
}

function filterTree(nodes: TreeNode[], keyword: string): TreeNode[] {
  if (!keyword) return nodes;
  const lower = keyword.toLowerCase();
  const walk = (items: TreeNode[]): TreeNode[] =>
    items
      .map((item) => ({ ...item, children: item.children ? walk(item.children) : item.children }))
      .filter((item) => item.label.toLowerCase().includes(lower) || (item.children ? item.children.length > 0 : false));
  return walk(nodes);
}

function deptNodesToTree(nodes: DeptNode[]): DeptTreeItem[] {
  return nodes.map((node) => ({
    value: String(node.id),
    label: node.label,
    children: node.children && node.children.length ? deptNodesToTree(node.children) : undefined,
  }));
}

function onDeptExpand(value: Array<string | number>) {
  deptExpanded.value = value.map((item) => String(item));
}

function onDeptCheck(value: Array<string | number>) {
  deptCheckedIds.value = value.map((item) => String(item));
}

// ===== 加载 =====
async function load() {
  if (!props.role) return;
  loading.value = true;
  try {
    if (!deptTreeCache) {
      const deptRes = await fetchDeptTree();
      deptTreeCache = deptNodesToTree(deptRes.list || []);
    }
    const snap = await fetchRoleAuthorize(props.role.id);
    menuNodes.value = snap.menus || [];
    dataScope.value = snap.dataScope || 1;
    deptTreeData.value = deptTreeCache;
    deptCheckedIds.value = (snap.deptIds || []).map((id) => String(id));
    deptExpanded.value = collectTreeIds(deptTreeData.value);
    deptFilter.value = '';
    activeTab.value = 'menu';
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) load();
  },
);

function onMenuStats(stats: { menus: number; perms: number }) {
  menuStats.value = stats;
}

function onClose() {
  emit('update:visible', false);
}

// ===== 保存（单次 PUT：菜单 + 数据范围 + 自定义部门） =====
async function onSave() {
  if (!props.role) return;
  saving.value = true;
  try {
    const menuIds = menuTableRef.value?.collect() ?? [];
    // 部门 id 归一化为 string 参与树比对，提交前转回 i64；非自定义范围提交空，服务端也会清空
    const deptIds = isCustomScope.value ? deptCheckedIds.value.map((id) => Number(id)) : [];
    await saveRoleAuthorize({
      roleId: props.role.id,
      menuIds,
      deptIds,
      dataScope: dataScope.value,
    });
    MessagePlugin.success('授权已保存');
    visibleProxy.value = false;
    emit('saved');
  } catch (e) {
    MessagePlugin.error(errMsg(e));
  } finally {
    saving.value = false;
  }
}

// header 渲染在抽屉头部（scoped 样式无法命中），故用内联样式
function renderHeader() {
  const role = props.role;
  return h('div', { style: 'display:flex;align-items:baseline;gap:8px' }, [
    h('span', { style: 'font-size:16px;font-weight:600' }, '授权'),
    role
      ? h('span', { style: 'font-size:13px;color:var(--td-text-color-secondary)' }, `${role.name}（${role.code}）`)
      : null,
  ]);
}

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}
</script>
<style lang="less" scoped>
.auth-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.auth-drawer__tabs {
  display: flex;
  flex: none;
  gap: 4px;
  padding: 4px;
  margin-bottom: var(--td-comp-margin-m);
  background: var(--td-bg-color-component);
  border-radius: 999px;
  width: fit-content;
}

.auth-drawer__tab {
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 6px 16px;
  color: var(--td-text-color-secondary);
  font-size: 13px;
  cursor: pointer;
  background: transparent;
  border: none;
  border-radius: 999px;
  transition: all 0.15s ease;

  &:hover {
    color: var(--td-text-color);
  }

  &.is-active {
    color: var(--td-brand-color);
    font-weight: 600;
    background: var(--td-bg-color-container);
    box-shadow: var(--td-shadow-1);
  }
}

.auth-drawer__tab-count {
  min-width: 18px;
  padding: 0 5px;
  color: inherit;
  font-size: 11px;
  line-height: 18px;
  text-align: center;
  background: var(--td-bg-color-component);
  border-radius: 9px;
}

.auth-drawer__tab.is-active .auth-drawer__tab-count {
  background: var(--td-brand-color-light);
}

.auth-drawer__body {
  flex: 1;
  min-height: 0;
}

.auth-drawer__scope {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-m);
  height: 100%;
  overflow-y: auto;
}

.auth-drawer__scope-hint {
  padding: 24px 0;
  color: var(--td-text-color-placeholder);
  font-size: 13px;
  text-align: center;
}

.auth-drawer__dept-toolbar {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-m);

  .t-input {
    width: 240px;
  }
}

.auth-drawer__footer {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: space-between;
  padding-top: var(--td-comp-margin-m);
  margin-top: var(--td-comp-margin-xl);
  border-top: 1px solid var(--td-border-level-1-color);
}

.auth-drawer__summary {
  color: var(--td-text-color-secondary);
  font-size: 13px;

  b {
    color: var(--td-brand-color);
    font-weight: 600;
  }
}

.auth-drawer__actions {
  display: flex;
  gap: var(--td-comp-margin-s);
}
</style>
