<template>
  <div class="content-card operlog-page">
    <t-tabs v-model="tab">
      <!-- ===== Tab 1：操作日志 ===== -->
      <t-tab-panel value="oper" label="操作日志">
        <template #panel>
          <s-table
            row-key="id"
            :columns="operColumns as any"
            :data="operList as any"
            :total="operTotal"
            :page="operPage"
            :page-size="operPageSize"
            :loading="operLoading"
            keyword-placeholder="请输入 traceId"
            :search-fields="operSearchFields"
            :default-params="defaultLogParams"
            @search="onOperSearch"
            @reset="
              () => {
                operQuery = { ...defaultLogParams };
                operPage = 1;
                loadOper();
              }
            "
            @refresh="loadOper"
            @page-change="
              (p, ps) => {
                operPage = p;
                operPageSize = ps;
                loadOper();
              }
            "
          >
            <template #toolbar>
              <t-popconfirm content="确认清空全部操作日志？该操作不可恢复" theme="danger" @confirm="cleanAll('oper')">
                <t-button v-permission="'system:operlog:clean'" variant="outline" theme="danger">清空日志</t-button>
              </t-popconfirm>
            </template>
            <template #batch="{ selected, clear }">
              <t-button
                size="small"
                variant="outline"
                theme="danger"
                :loading="batchLoading"
                :disabled="batchLoading"
                @click="batchClean('oper', selected, clear)"
              >
                批量删除
              </t-button>
            </template>
            <template #operType="{ row }">
              <s-tag-dict dict-type="sys_oper_type" :value="(row as OperLogDTO).operType" />
            </template>
            <template #operStatus="{ row }">
              <s-tag-dict dict-type="sys_oper_status" :value="(row as OperLogDTO).operStatus" />
            </template>
            <template #costMs="{ row }">
              <t-tag :theme="costTheme((row as OperLogDTO).costMs)" variant="light">
                {{ (row as OperLogDTO).costMs }}ms
              </t-tag>
            </template>
            <template #ops="{ row }">
              <t-link theme="primary" @click="openDetail('oper', (row as OperLogDTO).id)">查看</t-link>
            </template>
          </s-table>
        </template>
      </t-tab-panel>

      <!-- ===== Tab 2：错误日志 ===== -->
      <t-tab-panel value="error" label="错误日志">
        <template #panel>
          <s-table
            row-key="id"
            :columns="errorColumns as any"
            :data="errorList as any"
            :total="errorTotal"
            :page="errorPage"
            :page-size="errorPageSize"
            :loading="errorLoading"
            keyword-placeholder="请输入 traceId"
            :search-fields="errorSearchFields"
            :default-params="defaultLogParams"
            @search="onErrorSearch"
            @reset="
              () => {
                errorQuery = { ...defaultLogParams };
                errorPage = 1;
                loadError();
              }
            "
            @refresh="loadError"
            @page-change="
              (p, ps) => {
                errorPage = p;
                errorPageSize = ps;
                loadError();
              }
            "
          >
            <template #toolbar>
              <t-popconfirm content="确认清空全部错误日志？该操作不可恢复" theme="danger" @confirm="cleanAll('error')">
                <t-button v-permission="'system:operlog:clean'" variant="outline" theme="danger">清空日志</t-button>
              </t-popconfirm>
            </template>
            <template #batch="{ selected, clear }">
              <t-button
                size="small"
                variant="outline"
                theme="danger"
                :loading="batchLoading"
                :disabled="batchLoading"
                @click="batchClean('error', selected, clear)"
              >
                批量删除
              </t-button>
            </template>
            <template #errLevel="{ row }">
              <s-tag-dict dict-type="sys_error_level" :value="(row as ErrorLogDTO).errLevel" />
            </template>
            <template #handleStatus="{ row }">
              <s-tag-dict dict-type="sys_handle_status" :value="(row as ErrorLogDTO).handleStatus" />
            </template>
            <template #errType="{ row }">
              <t-tag :theme="(row as ErrorLogDTO).errType === 2 ? 'danger' : 'warning'" variant="light">
                {{ (row as ErrorLogDTO).errType === 2 ? 'panic' : '业务异常' }}
              </t-tag>
            </template>
            <template #ops="{ row }">
              <t-space>
                <t-link theme="primary" @click="openDetail('error', (row as ErrorLogDTO).id)">查看</t-link>
                <t-link
                  v-if="(row as ErrorLogDTO).handleStatus === 0"
                  v-permission="'system:operlog:handle'"
                  theme="primary"
                  @click="markHandled(row as ErrorLogDTO)"
                >
                  处理
                </t-link>
              </t-space>
            </template>
          </s-table>
        </template>
      </t-tab-panel>
    </t-tabs>

    <log-detail-drawer :id="detailId" v-model:visible="detailVisible" :kind="detailKind" />
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { cleanErrorLogs, cleanOperLogs, handleErrorLog, listErrorLogs, listOperLogs } from '@/api/system/log';
import type { ErrorLogDTO, OperLogDTO } from '@/api/system/logModel';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useDict } from '@/store';
import { todayRange } from '@/utils/date';

import LogDetailDrawer from '../components/LogDetailDrawer.vue';

const operTypeDict = useDict('sys_oper_type');
const operStatusDict = useDict('sys_oper_status');
const errLevelDict = useDict('sys_error_level');
const handleStatusDict = useDict('sys_handle_status');

const tab = ref('oper');

// 默认查今天：日志表数据量大，避免无时间条件的全量扫描拖垮系统
const defaultLogParams = { logTime: todayRange() };

// ===== 操作日志 tab =====
const operLoading = ref(false);
const operList = ref<OperLogDTO[]>([]);
const operTotal = ref(0);
const operPage = ref(1);
const operPageSize = ref(useConfigStore().pageDefault);
const operQuery = ref<Record<string, unknown>>({ ...defaultLogParams });

const httpMethodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((v) => ({ label: v, value: v }));

const operSearchFields = computed<STableSearchField[]>(() => [
  { name: 'userName', label: '操作人员', type: 'input', placeholder: '请输入操作人员' },
  { name: 'operType', label: '操作类型', type: 'select', options: operTypeDict.options.value },
  { name: 'execStatus', label: '执行状态', type: 'select', options: operStatusDict.options.value },
  { name: 'method', label: '执行方法', type: 'input', placeholder: '接口路径关键字' },
  { name: 'httpMethod', label: '请求类型', type: 'select', options: httpMethodOptions },
  { name: 'logTime', label: '日志时间', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
]);

const operColumns = [
  { colKey: 'traceId', title: 'TraceId', width: 170, ellipsis: true },
  { colKey: 'userName', title: '操作人', width: 100 },
  { colKey: 'operType', title: '类型', width: 90, cell: 'operType' },
  { colKey: 'requestUri', title: '执行方法', minWidth: 220, ellipsis: true },
  { colKey: 'remark', title: '备注', minWidth: 160, ellipsis: true },
  { colKey: 'operStatus', title: '状态', width: 90, cell: 'operStatus' },
  { colKey: 'costMs', title: '执行时长', width: 100, cell: 'costMs' },
  { colKey: 'createdAt', title: '记录时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 80, fixed: 'right', cell: 'ops' },
];

// ===== 系统异常 tab =====
const errorLoading = ref(false);
const errorList = ref<ErrorLogDTO[]>([]);
const errorTotal = ref(0);
const errorPage = ref(1);
const errorPageSize = ref(useConfigStore().pageDefault);
const errorQuery = ref<Record<string, unknown>>({ ...defaultLogParams });

const errorSearchFields = computed<STableSearchField[]>(() => [
  { name: 'userName', label: '用户名称', type: 'input', placeholder: '请输入用户名称' },
  { name: 'errKeyword', label: '异常信息', type: 'input', placeholder: '请输入信息' },
  { name: 'handleStatus', label: '处理状态', type: 'select', options: handleStatusDict.options.value },
  { name: 'errLevel', label: '异常级别', type: 'select', options: errLevelDict.options.value },
  { name: 'httpMethod', label: '请求类型', type: 'select', options: httpMethodOptions },
  { name: 'logTime', label: '日志时间', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
]);

const errorColumns = [
  { colKey: 'traceId', title: 'TraceId', width: 170, ellipsis: true },
  { colKey: 'userName', title: '用户名称', width: 100 },
  { colKey: 'errType', title: '日志类型', width: 100, cell: 'errType' },
  { colKey: 'requestUri', title: '请求地址', minWidth: 200, ellipsis: true },
  { colKey: 'errMsg', title: '异常信息', minWidth: 200, ellipsis: true },
  { colKey: 'errLevel', title: '异常级别', width: 90, cell: 'errLevel' },
  { colKey: 'handleStatus', title: '处理状态', width: 90, cell: 'handleStatus' },
  { colKey: 'createdAt', title: '记录时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 130, fixed: 'right', cell: 'ops' },
];

// ===== 详情抽屉 =====
const detailVisible = ref(false);
const detailKind = ref<'oper' | 'error'>('oper');
const detailId = ref<string | undefined>();

function openDetail(kind: 'oper' | 'error', id: string) {
  detailKind.value = kind;
  detailId.value = id;
  detailVisible.value = true;
}

function costTheme(ms: number): 'success' | 'warning' | 'danger' {
  if (ms < 100) return 'success';
  if (ms < 300) return 'warning';
  return 'danger';
}

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

async function loadOper() {
  operLoading.value = true;
  try {
    const q = operQuery.value;
    const res = await listOperLogs({
      page: operPage.value,
      pageSize: operPageSize.value,
      traceId: (q.keyword as string) || undefined,
      userName: (q.userName as string) || undefined,
      operType: toNum(q.operType),
      execStatus: toNum(q.execStatus),
      method: (q.method as string) || undefined,
      httpMethod: (q.httpMethod as string) || undefined,
      ...timeRange(q),
    });
    operList.value = res.list || [];
    operTotal.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    operLoading.value = false;
  }
}

function onOperSearch(params: Record<string, unknown>) {
  operQuery.value = params;
  operPage.value = 1;
  loadOper();
}

async function loadError() {
  errorLoading.value = true;
  try {
    const q = errorQuery.value;
    const res = await listErrorLogs({
      page: errorPage.value,
      pageSize: errorPageSize.value,
      traceId: (q.keyword as string) || undefined,
      userName: (q.userName as string) || undefined,
      keyword: (q.errKeyword as string) || undefined,
      handleStatus: toNum(q.handleStatus),
      errLevel: toNum(q.errLevel),
      httpMethod: (q.httpMethod as string) || undefined,
      ...timeRange(q),
    });
    errorList.value = res.list || [];
    errorTotal.value = res.total || 0;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    errorLoading.value = false;
  }
}

function onErrorSearch(params: Record<string, unknown>) {
  errorQuery.value = params;
  errorPage.value = 1;
  loadError();
}

// ===== 清空 / 批量删除 / 处理 =====
const batchLoading = ref(false);

async function cleanAll(kind: 'oper' | 'error') {
  try {
    const n = kind === 'oper' ? await cleanOperLogs({ all: true }) : await cleanErrorLogs({ all: true });
    MessagePlugin.success(`已清空 ${n} 条日志`);
    if (kind === 'oper') {
      operPage.value = 1;
      await loadOper();
    } else {
      errorPage.value = 1;
      await loadError();
    }
  } catch (e) {
    MessagePlugin.error((e as Error).message || '清空失败');
  }
}

async function batchClean(kind: 'oper' | 'error', selected: unknown[], clear: () => void) {
  if (!selected.length) return;
  const ids = (selected as Array<{ id: string }>).map((r) => r.id);
  batchLoading.value = true;
  try {
    const n = kind === 'oper' ? await cleanOperLogs({ ids }) : await cleanErrorLogs({ ids });
    MessagePlugin.success(`已删除 ${n} 条日志`);
    clear();
    if (kind === 'oper') await loadOper();
    else await loadError();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '批量删除失败');
  } finally {
    batchLoading.value = false;
  }
}

async function markHandled(row: ErrorLogDTO) {
  try {
    await handleErrorLog({ id: row.id, handleRemark: '' });
    MessagePlugin.success('已标记为已处理');
    await loadError();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '标记失败');
  }
}

onMounted(() => {
  loadOper();
  // error tab 首次切换时加载；预取一次让切换无感
  loadError();
});
</script>
<style scoped lang="less">
/*
 * tab 内嵌 s-table 的高度链修复：
 * .s-table 依赖 height:100%（flex:1 1 0 撑开内层表格），而 .t-tab-panel 默认
 * auto 高度会把链路压塌（t-table 折成 0、分页条错位盖住数据行）。
 * 将 content-card → t-tabs → content → panel 全链改为确定高度的 flex 列。
 */
.operlog-page {
  display: flex;
  flex-direction: column;

  :deep(.t-tabs) {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    min-height: 0;
  }

  :deep(.t-tabs__content) {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    min-height: 0;
    height: auto !important; // 兜底：切换动画可能写入内联定高
    overflow: visible;
    // 整体留白：tab 栏与表格四周的呼吸间距
    padding: 12px;
  }

  :deep(.t-tab-panel) {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    min-height: 0;
    height: auto;

    .s-table {
      flex: 1 1 auto;
      min-height: 0;
    }
  }
}
</style>
