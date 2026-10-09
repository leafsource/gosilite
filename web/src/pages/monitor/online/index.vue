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
      keyword-placeholder="请输入登录名称"
      :search-fields="searchFields"
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
      <template #deptName="{ row }">
        <span v-if="(row as OnlineDTO).deptName">{{ (row as OnlineDTO).deptName }}</span>
        <span v-else class="online__empty">—</span>
      </template>
      <template #ops="{ row }">
        <t-popconfirm content="确认强制退出该会话？该设备将立即掉线" theme="danger" @confirm="kick(row as OnlineDTO)">
          <t-link v-permission="'system:online:force'" theme="danger">强退</t-link>
        </t-popconfirm>
      </template>
    </s-table>
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { forceOffline, listOnline } from '@/api/system/online';
import type { OnlineDTO } from '@/api/system/onlineModel';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import { useConfigStore } from '@/store';

const loading = ref(false);
const list = ref<OnlineDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');

const searchFields = computed<STableSearchField[]>(() => []);

const columns = [
  { colKey: 'userName', title: '登录名称', width: 140 },
  { colKey: 'deptName', title: '部门名称', width: 160, cell: 'deptName' },
  { colKey: 'clientIp', title: '客户端 IP', width: 140 },
  { colKey: 'browser', title: '浏览器', width: 110 },
  { colKey: 'os', title: '操作系统', width: 120 },
  { colKey: 'loginAt', title: '登录时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 100, fixed: 'right', cell: 'ops' },
];

async function load() {
  loading.value = true;
  try {
    const res = await listOnline({
      page: page.value,
      pageSize: pageSize.value,
      userName: keyword.value || undefined,
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
  page.value = 1;
  load();
}

async function kick(row: OnlineDTO) {
  try {
    await forceOffline(row.id);
    MessagePlugin.success(`已强退 ${row.userName}`);
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '强退失败');
  }
}

onMounted(load);
</script>
<style scoped lang="less">
.online__empty {
  color: var(--td-text-color-placeholder);
}
</style>
