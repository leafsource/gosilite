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
      keyword-placeholder="请输入用户名"
      :search-fields="searchFields"
      :default-params="defaultLogParams"
      @search="onSearch"
      @reset="
        () => {
          query = { ...defaultLogParams };
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
      <template #status="{ row }">
        <s-tag-dict dict-type="sys_login_status" :value="(row as LoginLogDTO).status" />
      </template>
      <template #operInfo="{ row }">
        <span v-if="(row as LoginLogDTO).status === 0" class="loginlog__ok">登录成功</span>
        <span v-else-if="(row as LoginLogDTO).status === 2" class="loginlog__logout">退出登录</span>
        <span v-else class="loginlog__fail">{{ (row as LoginLogDTO).failReason || '登录失败' }}</span>
      </template>
    </s-table>
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemLogLoginlog' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { listLoginLogs } from '@/api/system/log';
import type { LoginLogDTO } from '@/api/system/logModel';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useDict } from '@/store';
import { todayRange } from '@/utils/date';

const statusDict = useDict('sys_login_status');

// 默认查今天：日志表数据量大，避免无时间条件的全量扫描拖垮系统
const defaultLogParams = { logTime: todayRange() };

const loading = ref(false);
const list = ref<LoginLogDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const query = ref<Record<string, unknown>>({ ...defaultLogParams });

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'status', label: '登录状态', type: 'select', options: statusDict.options.value },
  { name: 'clientIp', label: '客户端 IP', type: 'input', placeholder: '请输入 IP' },
  { name: 'logTime', label: '登录时间', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
]);

const columns = [
  { colKey: 'userName', title: '登录用户', width: 130 },
  { colKey: 'status', title: '登录状态', width: 100, cell: 'status' },
  { colKey: 'operInfo', title: '操作信息', minWidth: 200, ellipsis: true, cell: 'operInfo' },
  { colKey: 'browser', title: '浏览器', width: 100 },
  { colKey: 'os', title: '操作系统', width: 110 },
  { colKey: 'clientIp', title: '客户端 IP', width: 130 },
  { colKey: 'traceId', title: 'TraceId', width: 170, ellipsis: true },
  { colKey: 'createdAt', title: '登录时间', width: 170 },
];

function toNum(v: unknown): number | undefined {
  if (v === undefined || v === null || v === '') return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

function timeRange(params: Record<string, unknown>): { startTime?: string; endTime?: string } {
  const arr = Array.isArray(params.logTime) ? (params.logTime as string[]) : [];
  if (!arr.length) return {};
  return { startTime: arr[0], endTime: arr[1] };
}

async function load() {
  loading.value = true;
  try {
    const q = query.value;
    const res = await listLoginLogs({
      page: page.value,
      pageSize: pageSize.value,
      userName: (q.keyword as string) || undefined,
      status: toNum(q.status),
      clientIp: (q.clientIp as string) || undefined,
      ...timeRange(q),
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
  query.value = params;
  page.value = 1;
  load();
}

onMounted(load);
</script>
<style scoped lang="less">
.loginlog__fail {
  color: var(--td-error-color);
}

.loginlog__ok {
  color: var(--td-success-color);
}

.loginlog__logout {
  color: var(--td-text-color-secondary);
}
</style>
