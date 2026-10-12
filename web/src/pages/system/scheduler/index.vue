<template>
  <div class="content-card">
    <s-table
      row-key="id"
      :columns="columns"
      :data="tableData"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :show-search="false"
      :selectable="false"
      @refresh="load"
      @page-change="onPageChange"
    >
      <template #toolbar>
        <t-button v-permission="'system:scheduler:add'" theme="primary" @click="() => open()">新建任务</t-button>
      </template>
      <template #status="{ row }">
        <t-tag :theme="(row as SchedulerJob).status ? 'success' : 'default'">
          {{ (row as SchedulerJob).status ? '启用' : '停用' }}
        </t-tag>
      </template>
      <template #op="{ row }">
        <t-space>
          <t-link v-permission="'system:scheduler:edit'" theme="primary" @click="open(row as SchedulerJob)">编辑</t-link>
          <t-link v-permission="'system:scheduler:edit'" @click="toggle(row as SchedulerJob)">
            {{ (row as SchedulerJob).status ? '停用' : '启用' }}
          </t-link>
          <t-link theme="primary" @click="openRuns(row as SchedulerJob)">运行记录</t-link>
          <t-link v-permission="'system:scheduler:run'" theme="primary" @click="run(row as SchedulerJob)">运行</t-link>
          <t-link v-permission="'system:scheduler:remove'" theme="danger" @click="remove(row as SchedulerJob)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <t-drawer v-model:visible="visible" :header="form.id ? '编辑任务' : '新建任务'" size="500px" @confirm="submit">
      <t-form :data="form" label-width="90px">
        <t-form-item label="任务键">
          <t-select v-model="form.jobKey" :options="jobKeyOptions" placeholder="请选择已注册任务" />
        </t-form-item>
        <t-form-item label="名称">
          <t-input v-model="form.jobName" :maxlength="255" clearable placeholder="请输入任务名称" />
        </t-form-item>
        <t-form-item label="Cron">
          <t-input v-model="form.cron" placeholder="0 0 * * * *（秒 分 时 日 月 周）" />
        </t-form-item>
        <t-form-item label="超时秒数">
          <t-input v-model="form.timeoutSec" type="number" placeholder="60" />
        </t-form-item>
        <t-form-item label="重试次数">
          <t-input v-model="form.retryCount" type="number" placeholder="0" />
        </t-form-item>
        <t-form-item label="参数 JSON">
          <t-textarea v-model="form.paramsJson" :autosize="{ minRows: 3, maxRows: 8 }" placeholder="{}" />
        </t-form-item>
        <t-form-item label="状态">
          <t-select v-model="form.status" :options="statusOptions" />
        </t-form-item>
      </t-form>
    </t-drawer>

    <t-drawer v-model:visible="runsVisible" :header="`${selectedJob?.jobName ?? ''} · 运行记录`" size="900px" :footer="false" @close="closeRuns">
      <s-table
        row-key="id"
        :columns="runColumns"
        :data="runTableData"
        :total="runsTotal"
        :page="runsPage"
        :page-size="runsPageSize"
        :loading="runsLoading"
        :show-search="false"
        :selectable="false"
        @refresh="loadRuns"
        @page-change="onRunsPageChange"
      >
        <template #status="{ row }">
          <t-tag :theme="runStatusTheme((row as SchedulerRun).status)" variant="light">
            {{ runStatusText((row as SchedulerRun).status) }}
          </t-tag>
        </template>
        <template #triggeredAt="{ row }">{{ formatRunTime((row as SchedulerRun).triggeredAt) }}</template>
        <template #finishedAt="{ row }">{{ formatRunTime((row as SchedulerRun).finishedAt) }}</template>
        <template #detail="{ row }">
          <t-space size="small">
            <t-link theme="primary" @click="openRunLogs(row as SchedulerRun)">实时日志</t-link>
            <t-link v-if="(row as SchedulerRun).output" theme="primary" @click="showRunDetail((row as SchedulerRun).output, '执行结果')">执行结果</t-link>
            <t-link v-if="(row as SchedulerRun).error" theme="danger" @click="showRunDetail((row as SchedulerRun).error, '错误')">错误</t-link>
          </t-space>
        </template>
      </s-table>
    </t-drawer>

    <t-drawer v-model:visible="runLogVisible" header="执行日志" size="760px" :footer="false" @close="closeRunLogs">
      <pre class="run-detail">{{ formattedRunLogs }}</pre>
    </t-drawer>

    <t-drawer v-model:visible="runDetailVisible" :header="runDetailTitle" size="600px" :footer="false">
      <pre class="run-detail">{{ runDetail }}</pre>
    </t-drawer>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorScheduler' });

import { computed, onMounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';

import {
  listSchedulerJobs,
  listSchedulerRunLogs,
  listSchedulerRuns,
  removeSchedulerJob,
  runSchedulerJob,
  saveSchedulerJob,
  schedulerKeys,
  toggleSchedulerJob,
} from '@/api/system/scheduler';
import type { SchedulerJob, SchedulerJobSave, SchedulerRun, SchedulerRunLog } from '@/api/system/schedulerModel';
import STable from '@/components/s-table/index.vue';
import type { STableColumn } from '@/components/s-table/sTable';
import { useConfigStore, useUserStore } from '@/store';

const loading = ref(false);
const jobs = ref<SchedulerJob[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const visible = ref(false);
const runsVisible = ref(false);
const runDetailVisible = ref(false);
const runLogVisible = ref(false);
const runsLoading = ref(false);
const selectedJob = ref<SchedulerJob>();
const runs = ref<SchedulerRun[]>([]);
const runsTotal = ref(0);
const runsPage = ref(1);
const runsPageSize = ref(useConfigStore().pageDefault);
const runDetail = ref('');
const runDetailTitle = ref('');
const runLogs = ref<SchedulerRunLog[]>([]);
const selectedRun = ref<SchedulerRun>();
let runLogAbort: AbortController | undefined;
const keys = ref<string[]>([]);
const form = ref<SchedulerJobSave>(createForm());

const tableData = computed(() => jobs.value as unknown as Array<Record<string, unknown>>);
const runTableData = computed(() => runs.value as unknown as Array<Record<string, unknown>>);
const jobKeyOptions = computed(() => keys.value.map((key) => ({ label: key, value: key })));
const formattedRunLogs = computed(() => runLogs.value.map((log) => `${formatRunTime(log.createdAt)} [${log.level.toUpperCase()}] [${log.stage}] ${log.message}`).join('\n'));
const statusOptions = [
  { label: '启用', value: 1 },
  { label: '停用', value: 0 },
];
const columns: STableColumn[] = [
  { colKey: 'jobKey', title: '任务键', minWidth: 180, ellipsis: true },
  { colKey: 'jobName', title: '任务名称', minWidth: 150, ellipsis: true },
  { colKey: 'cron', title: 'Cron（服务器本地时间）', minWidth: 200, ellipsis: true },
  { colKey: 'timeoutSec', title: '超时（秒）', width: 110 },
  { colKey: 'retryCount', title: '重试次数', width: 100 },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'op', title: '操作', width: 290, fixed: 'right', cell: 'op' },
];
const runColumns: STableColumn[] = [
  { colKey: 'trigger', title: '触发方式', width: 100 },
  { colKey: 'retryCount', title: '重试次数', width: 100 },
  { colKey: 'status', title: '状态', width: 120, cell: 'status' },
  { colKey: 'triggeredAt', title: '触发时间', minWidth: 190, cell: 'triggeredAt' },
  { colKey: 'finishedAt', title: '结束时间', minWidth: 190, cell: 'finishedAt' },
  { colKey: 'detail', title: '详情', width: 120, cell: 'detail' },
];

function createForm(): SchedulerJobSave {
  return {
    jobKey: '',
    jobName: '',
    cron: '0 0 * * * *',
    timeoutSec: 60,
    paramsJson: '{}',
    retryCount: 0,
    status: 1,
  };
}

async function load() {
  loading.value = true;
  try {
    const result = await listSchedulerJobs({ page: page.value, pageSize: pageSize.value });
    jobs.value = result.list ?? [];
    total.value = result.total ?? 0;
  } catch (error) {
    MessagePlugin.error((error as Error).message || '加载任务失败');
  } finally {
    loading.value = false;
  }
}

async function loadKeys() {
  try {
    const result = await schedulerKeys();
    keys.value = result.list ?? [];
  } catch (error) {
    MessagePlugin.error((error as Error).message || '加载任务键失败');
  }
}

function onPageChange(nextPage: number, nextPageSize: number) {
  page.value = nextPage;
  pageSize.value = nextPageSize;
  load();
}

async function openRuns(job: SchedulerJob) {
  selectedJob.value = job;
  runsPage.value = 1;
  runsVisible.value = true;
  await loadRuns();
}

async function loadRuns() {
  if (!selectedJob.value) return;
  runsLoading.value = true;
  try {
    const result = await listSchedulerRuns({ jobId: selectedJob.value.id, page: runsPage.value, pageSize: runsPageSize.value });
    runs.value = result.list ?? [];
    runsTotal.value = result.total ?? 0;
  } catch (error) {
    MessagePlugin.error((error as Error).message || '加载运行记录失败');
  } finally {
    runsLoading.value = false;
  }
}

function onRunsPageChange(nextPage: number, nextPageSize: number) {
  runsPage.value = nextPage;
  runsPageSize.value = nextPageSize;
  loadRuns();
}

function closeRuns() {
  selectedJob.value = undefined;
  runs.value = [];
  runsTotal.value = 0;
}

async function openRunLogs(run: SchedulerRun) {
  selectedRun.value = run;
  runLogs.value = [];
  runLogVisible.value = true;
  try {
    const result = await listSchedulerRunLogs({ runId: run.id, page: 1, pageSize: 200 });
    runLogs.value = result.list ?? [];
    followRunLogs(run.id, runLogs.value.at(-1)?.id);
  } catch (error) {
    MessagePlugin.error((error as Error).message || '加载执行日志失败');
  }
}

function closeRunLogs() {
  runLogAbort?.abort();
  runLogAbort = undefined;
  selectedRun.value = undefined;
}

async function followRunLogs(runId: string, afterId?: string) {
  runLogAbort?.abort();
  runLogAbort = new AbortController();
  const userStore = useUserStore();
  const params = new URLSearchParams({ runId, page: '1', pageSize: '200' });
  if (afterId) params.set('afterId', afterId);
  try {
    const response = await fetch(`/system/scheduler/run/log/stream?${params.toString()}`, {
      headers: { Authorization: `Bearer ${userStore.token}`, Accept: 'text/event-stream', ...(afterId ? { 'Last-Event-ID': afterId } : {}) },
      signal: runLogAbort.signal,
    });
    if (!response.ok || !response.body) throw new Error('订阅执行日志失败');
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    while (runLogVisible.value && !runLogAbort.signal.aborted) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      let separator = buffer.indexOf('\n\n');
      while (separator >= 0) {
        const event = buffer.slice(0, separator);
        buffer = buffer.slice(separator + 2);
        const data = event.split('\n').find((line) => line.startsWith('data: '))?.slice(6);
        if (data) {
          const log = JSON.parse(data) as SchedulerRunLog;
          if (!runLogs.value.some((item) => item.id === log.id)) runLogs.value.push(log);
        }
        separator = buffer.indexOf('\n\n');
      }
    }
  } catch (error) {
    if (!runLogAbort?.signal.aborted) MessagePlugin.error((error as Error).message || '订阅执行日志失败');
  } finally {
    if (runLogVisible.value && !runLogAbort?.signal.aborted && selectedRun.value?.id === runId) {
      window.setTimeout(() => followRunLogs(runId, runLogs.value[runLogs.value.length - 1]?.id), 1000);
    }
  }
}

function runStatusTheme(status: string): 'success' | 'danger' | 'warning' | 'default' {
  if (status === 'succeeded') return 'success';
  if (status === 'timed_out') return 'warning';
  if (status === 'failed') return 'danger';
  return 'default';
}

function runStatusText(status: string) {
  return { succeeded: '成功', failed: '失败', timed_out: '超时', skipped_locked: '锁占用' }[status] ?? status;
}

function formatRunTime(value: string): string {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const pad = (part: number) => String(part).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function showRunDetail(content: string, title: string) {
  runDetail.value = content;
  runDetailTitle.value = title;
  runDetailVisible.value = true;
}

function open(job?: SchedulerJob) {
  form.value = job
    ? {
        id: job.id,
        jobKey: job.jobKey,
        jobName: job.jobName,
        cron: job.cron,
        timeoutSec: job.timeoutSec,
        paramsJson: job.paramsJson,
        retryCount: job.retryCount,
        status: job.status,
      }
    : createForm();
  visible.value = true;
}

async function submit() {
  try {
    await saveSchedulerJob({
      ...form.value,
      jobKey: form.value.jobKey.trim(),
      jobName: form.value.jobName.trim(),
      cron: form.value.cron.trim(),
      paramsJson: form.value.paramsJson.trim(),
      timeoutSec: Number(form.value.timeoutSec),
      retryCount: Number(form.value.retryCount),
      status: Number(form.value.status),
    });
    MessagePlugin.success('保存成功');
    visible.value = false;
    await load();
  } catch (error) {
    MessagePlugin.error((error as Error).message || '保存任务失败');
  }
}

async function toggle(job: SchedulerJob) {
  try {
    await toggleSchedulerJob(job.id);
    MessagePlugin.success(job.status ? '任务已停用' : '任务已启用');
    await load();
  } catch (error) {
    MessagePlugin.error((error as Error).message || '更新任务状态失败');
  }
}

async function run(job: SchedulerJob) {
  try {
    await runSchedulerJob(job.id);
    MessagePlugin.success('任务已执行');
    await load();
  } catch (error) {
    MessagePlugin.error((error as Error).message || '执行任务失败');
  }
}

async function remove(job: SchedulerJob) {
  try {
    await removeSchedulerJob(job.id);
    MessagePlugin.success('任务已删除');
    if (jobs.value.length === 1 && page.value > 1) page.value -= 1;
    await load();
  } catch (error) {
    MessagePlugin.error((error as Error).message || '删除任务失败');
  }
}

onMounted(async () => {
  await Promise.all([load(), loadKeys()]);
});
</script>

<style scoped>
.run-detail {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
