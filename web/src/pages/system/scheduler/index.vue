<template>
  <div class="content-card">
    <t-tabs v-model="activeTab" :destroy-on-hide="false">
      <t-tab-panel value="jobs" label="任务管理">
        <template #panel>
          <s-table row-key="id" :columns="jobColumns" :data="jobTableData" :total="jobsTotal" :page="jobsPage" :page-size="jobsPageSize" :loading="jobsLoading" :show-search="false" :selectable="false" @refresh="loadJobs" @page-change="onJobsPageChange">
            <template #toolbar>
              <t-button v-permission="'system:scheduler:add'" theme="primary" @click="openJobForm()">新建任务</t-button>
            </template>
            <template #status="{ row }">
              <t-switch
                v-permission="'system:scheduler:edit'"
                :value="Boolean((row as SchedulerJob).status)"
                :loading="statusChangingId === (row as SchedulerJob).id"
                @change="(value) => changeJobStatus(row as SchedulerJob, Boolean(value))"
              />
            </template>
            <template #op="{ row }">
              <t-space size="small">
                <t-link v-permission="'system:scheduler:edit'" theme="primary" @click="openJobForm(row as SchedulerJob)">编辑</t-link>
                <t-dropdown trigger="click" :options="jobMoreOptions(row as SchedulerJob)" @click="(option) => onJobMore(option, row as SchedulerJob)">
                  <t-button variant="text" shape="square" size="small"><more-icon /></t-button>
                </t-dropdown>
              </t-space>
            </template>
          </s-table>
        </template>
      </t-tab-panel>

      <t-tab-panel value="runs" label="运行记录">
        <template #panel>
          <s-table row-key="id" :columns="runColumns" :data="runTableData" :total="runsTotal" :page="runsPage" :page-size="runsPageSize" :loading="runsLoading" :show-search="true" keyword-placeholder="请输入运行 ID" :search-fields="runSearchFields" :default-params="runSearchParams" :selectable="false" @search="onRunSearch" @reset="onRunReset" @refresh="loadRuns" @page-change="onRunsPageChange">
            <template #status="{ row }">
              <t-tag :theme="runStatusTheme((row as SchedulerRun).status)" variant="light">{{ runStatusText((row as SchedulerRun).status) }}</t-tag>
            </template>
            <template #triggeredAt="{ row }">{{ formatRunTime((row as SchedulerRun).triggeredAt) }}</template>
            <template #finishedAt="{ row }">{{ formatRunTime((row as SchedulerRun).finishedAt) }}</template>
            <template #op="{ row }">
              <t-space size="small">
                <t-link theme="primary" @click="openRunDetail(row as SchedulerRun)">详情</t-link>
                <t-link theme="primary" @click="openRunLogs(row as SchedulerRun)">日志</t-link>
                <t-link v-if="canRetry(row as SchedulerRun)" theme="primary" @click="retryRun(row as SchedulerRun)">重试</t-link>
                <t-link v-if="canStop(row as SchedulerRun)" theme="danger" @click="stopRun(row as SchedulerRun)">停止</t-link>
              </t-space>
            </template>
          </s-table>
        </template>
      </t-tab-panel>
    </t-tabs>

    <t-drawer v-model:visible="jobFormVisible" :header="jobForm.id ? '编辑任务' : '新建任务'" size="500px" @confirm="submitJob">
      <t-form :data="jobForm" label-width="90px">
        <t-form-item label="任务键"><t-select v-model="jobForm.jobKey" :options="jobKeyOptions" placeholder="请选择已注册任务" /></t-form-item>
        <t-form-item label="名称"><t-input v-model="jobForm.jobName" :maxlength="255" clearable placeholder="请输入任务名称" /></t-form-item>
        <t-form-item label="Cron"><t-input v-model="jobForm.cron" placeholder="0 0 * * * *（秒 分 时 日 月 周）" /></t-form-item>
        <t-form-item label="超时秒数"><t-input v-model="jobForm.timeoutSec" type="number" placeholder="60" /></t-form-item>
        <t-form-item label="重试次数"><t-input v-model="jobForm.retryCount" type="number" placeholder="0" /></t-form-item>
        <t-form-item label="最大并发实例"><t-input v-model="jobForm.maxInstances" type="number" placeholder="1" /></t-form-item>
        <t-form-item label="参数 JSON"><t-textarea v-model="jobForm.paramsJson" :autosize="{ minRows: 3, maxRows: 8 }" placeholder="{}" /></t-form-item>
        <t-form-item label="状态"><t-select v-model="jobForm.status" :options="jobStatusOptions" /></t-form-item>
      </t-form>
    </t-drawer>

    <t-drawer v-model:visible="runDetailVisible" header="运行详情" size="680px" :footer="false">
      <t-descriptions v-if="selectedRun" :column="1" bordered>
        <t-descriptions-item label="任务名称">{{ selectedRun.jobName || '—' }}</t-descriptions-item>
        <t-descriptions-item label="任务键">{{ selectedRun.jobKey }}</t-descriptions-item>
        <t-descriptions-item label="运行 ID">{{ selectedRun.id }}</t-descriptions-item>
        <t-descriptions-item label="状态">{{ runStatusText(selectedRun.status) }}</t-descriptions-item>
        <t-descriptions-item label="触发方式">{{ selectedRun.trigger }}</t-descriptions-item>
        <t-descriptions-item label="开始时间">{{ formatRunTime(selectedRun.triggeredAt) }}</t-descriptions-item>
        <t-descriptions-item label="结束时间">{{ formatRunTime(selectedRun.finishedAt) }}</t-descriptions-item>
        <t-descriptions-item label="超时秒数">{{ selectedRun.timeoutSec }}</t-descriptions-item>
        <t-descriptions-item label="配置版本">{{ selectedRun.configVersion }}</t-descriptions-item>
        <t-descriptions-item label="重试次数">{{ selectedRun.retryCount }}</t-descriptions-item>
      </t-descriptions>
      <section v-if="selectedRun" class="snapshot-section"><h4>运行快照参数</h4><pre class="run-content">{{ selectedRun.paramsJson || '{}' }}</pre></section>
      <section v-if="selectedRun?.output" class="snapshot-section"><h4>执行结果</h4><pre class="run-content">{{ selectedRun.output }}</pre></section>
      <section v-if="selectedRun?.error" class="snapshot-section"><h4>错误信息</h4><pre class="run-content">{{ selectedRun.error }}</pre></section>
    </t-drawer>

    <t-drawer v-model:visible="runLogsVisible" :header="`${selectedRun?.jobName || ''} · 执行日志`" size="860px" :footer="false" @close="closeRunLogs">
      <div class="log-actions"><span>{{ isActiveRun(selectedRun) ? '正在实时订阅日志' : '' }}</span><t-space><t-button variant="outline" size="small" @click="copyLogs">复制</t-button><t-button variant="outline" size="small" @click="downloadLogs">下载</t-button></t-space></div>
      <pre class="run-content log-content">{{ formattedRunLogs || '暂无日志' }}</pre>
    </t-drawer>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'MonitorScheduler' });

import { MoreIcon } from 'tdesign-icons-vue-next';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { DropdownOption } from 'tdesign-vue-next';

import { listSchedulerJobs, listSchedulerRunLogs, listSchedulerRuns, removeSchedulerJob, retrySchedulerRun, runSchedulerJob, saveSchedulerJob, schedulerKeys, stopSchedulerRun, streamSchedulerRunLogs, toggleSchedulerJob } from '@/api/system/scheduler';
import type { SchedulerJob, SchedulerJobSave, SchedulerRun, SchedulerRunLog } from '@/api/system/schedulerModel';
import STable from '@/components/s-table/index.vue';
import type { STableColumn, STableSearchField } from '@/components/s-table/sTable';
import { useConfigStore } from '@/store';
import type { SSEConnection } from '@/utils/request/sse';

const configStore = useConfigStore();
const activeTab = ref('jobs');
const jobsLoading = ref(false);
const jobs = ref<SchedulerJob[]>([]);
const jobsTotal = ref(0);
const jobsPage = ref(1);
const jobsPageSize = ref(configStore.pageDefault);
const statusChangingId = ref('');
const jobFormVisible = ref(false);
const jobForm = ref<SchedulerJobSave>(createJobForm());
const jobKeys = ref<string[]>([]);
const runSearchParams = ref<Record<string, unknown>>({});
const runsLoading = ref(false);
const runs = ref<SchedulerRun[]>([]);
const runsTotal = ref(0);
const runsPage = ref(1);
const runsPageSize = ref(configStore.pageDefault);
const runDetailVisible = ref(false);
const runLogsVisible = ref(false);
const selectedRun = ref<SchedulerRun>();
const runLogs = ref<SchedulerRunLog[]>([]);
let logStream: SSEConnection | undefined;

const jobTableData = computed(() => jobs.value as unknown as Array<Record<string, unknown>>);
const runTableData = computed(() => runs.value as unknown as Array<Record<string, unknown>>);
const jobKeyOptions = computed(() => jobKeys.value.map((key) => ({ label: key, value: key })));
const runJobOptions = computed(() => jobs.value.map((job) => ({ label: `${job.jobName}（${job.jobKey}）`, value: job.id })));
const runSearchFields = computed<STableSearchField[]>(() => [
  { name: 'jobId', label: '任务', type: 'select', options: runJobOptions.value, placeholder: '请选择任务' },
]);
const formattedRunLogs = computed(() => runLogs.value.map((log) => `${formatRunTime(log.createdAt)} [${log.level.toUpperCase()}] [${log.stage}] ${log.message}`).join('\n'));
const jobStatusOptions = [{ label: '启用', value: 1 }, { label: '停用', value: 0 }];
const jobColumns: STableColumn[] = [
  { colKey: 'jobKey', title: '任务键', minWidth: 180, ellipsis: true }, { colKey: 'jobName', title: '任务名称', minWidth: 150, ellipsis: true }, { colKey: 'cron', title: 'Cron（服务器本地时间）', minWidth: 200, ellipsis: true }, { colKey: 'timeoutSec', title: '超时（秒）', width: 110 }, { colKey: 'maxInstances', title: '最大并发', width: 100 }, { colKey: 'retryCount', title: '重试次数', width: 100 }, { colKey: 'status', title: '状态', width: 90, cell: 'status' },  { colKey: 'op', title: '操作', width: 290, fixed: 'right', cell: 'op' },
];
const runColumns: STableColumn[] = [
  { colKey: 'jobName', title: '任务名称', minWidth: 150, ellipsis: true }, { colKey: 'jobKey', title: '任务键', minWidth: 160, ellipsis: true }, { colKey: 'id', title: '运行 ID', minWidth: 160, ellipsis: true }, { colKey: 'trigger', title: '触发方式', width: 100 }, { colKey: 'retryCount', title: '重试次数', width: 100 }, { colKey: 'status', title: '状态', width: 110, cell: 'status' }, { colKey: 'triggeredAt', title: '触发时间', minWidth: 180, cell: 'triggeredAt' }, { colKey: 'finishedAt', title: '结束时间', minWidth: 180, cell: 'finishedAt' }, { colKey: 'op', title: '操作', width: 180, fixed: 'right', cell: 'op' },
];

function createJobForm(): SchedulerJobSave { return { jobKey: '', jobName: '', cron: '0 0 * * * *', timeoutSec: 60, paramsJson: '{}', retryCount: 0, maxInstances: 1, status: 1 }; }
async function loadJobs() { jobsLoading.value = true; try { const result = await listSchedulerJobs({ page: jobsPage.value, pageSize: jobsPageSize.value }); jobs.value = result.list ?? []; jobsTotal.value = result.total ?? 0; } catch (error) { MessagePlugin.error((error as Error).message || '加载任务失败'); } finally { jobsLoading.value = false; } }
async function loadJobKeys() { try { const result = await schedulerKeys(); jobKeys.value = result.list ?? []; } catch (error) { MessagePlugin.error((error as Error).message || '加载任务键失败'); } }
function onJobsPageChange(nextPage: number, nextPageSize: number) { jobsPage.value = nextPage; jobsPageSize.value = nextPageSize; void loadJobs(); }
function showJobRuns(job: SchedulerJob) { runSearchParams.value = { jobId: job.id }; runsPage.value = 1; activeTab.value = 'runs'; void loadRuns(); }
function onRunSearch(params: Record<string, unknown>) { runSearchParams.value = params; runsPage.value = 1; void loadRuns(); }
function onRunReset() { runSearchParams.value = {}; runsPage.value = 1; void loadRuns(); }
async function loadRuns() { const jobId = String(runSearchParams.value.jobId ?? '') || undefined; runsLoading.value = true; try { const result = await listSchedulerRuns({ jobId, page: runsPage.value, pageSize: runsPageSize.value }); runs.value = result.list ?? []; runsTotal.value = result.total ?? 0; } catch (error) { MessagePlugin.error((error as Error).message || '加载运行记录失败'); } finally { runsLoading.value = false; } }
function onRunsPageChange(nextPage: number, nextPageSize: number) { runsPage.value = nextPage; runsPageSize.value = nextPageSize; void loadRuns(); }
function openJobForm(job?: SchedulerJob) { jobForm.value = job ? { id: job.id, jobKey: job.jobKey, jobName: job.jobName, cron: job.cron, timeoutSec: job.timeoutSec, paramsJson: job.paramsJson, retryCount: job.retryCount, maxInstances: job.maxInstances, status: job.status } : createJobForm(); jobFormVisible.value = true; }
async function submitJob() { try { await saveSchedulerJob({ ...jobForm.value, jobKey: jobForm.value.jobKey.trim(), jobName: jobForm.value.jobName.trim(), cron: jobForm.value.cron.trim(), paramsJson: jobForm.value.paramsJson.trim(), timeoutSec: Number(jobForm.value.timeoutSec), retryCount: Number(jobForm.value.retryCount), maxInstances: Number(jobForm.value.maxInstances), status: Number(jobForm.value.status) }); MessagePlugin.success('保存成功'); jobFormVisible.value = false; await loadJobs(); } catch (error) { MessagePlugin.error((error as Error).message || '保存任务失败'); } }
async function changeJobStatus(job: SchedulerJob, enabled: boolean) { if (Boolean(job.status) === enabled) return; statusChangingId.value = job.id; try { await toggleSchedulerJob(job.id); MessagePlugin.success(enabled ? '任务已启用' : '任务已停用'); await loadJobs(); } catch (error) { MessagePlugin.error((error as Error).message || '更新任务状态失败'); } finally { statusChangingId.value = ''; } }
function jobMoreOptions(job: SchedulerJob): DropdownOption[] { return [
  { content: '运行记录', value: 'runs' },
  { content: '立即运行', value: 'run' },
  { content: '删除', value: 'remove', theme: 'error' },
]; }
function onJobMore(option: DropdownOption, job: SchedulerJob) { const action = String(option.value ?? ''); if (action === 'runs') showJobRuns(job); else if (action === 'run') void runJob(job); else if (action === 'remove') void removeJob(job); }
async function runJob(job: SchedulerJob) { try { await runSchedulerJob(job.id); MessagePlugin.success('任务已执行'); } catch (error) { MessagePlugin.error((error as Error).message || '执行任务失败'); } }
async function removeJob(job: SchedulerJob) { try { await removeSchedulerJob(job.id); if (jobs.value.length === 1 && jobsPage.value > 1) jobsPage.value -= 1; await loadJobs(); MessagePlugin.success('删除成功'); } catch (error) { MessagePlugin.error((error as Error).message || '删除任务失败'); } }
function openRunDetail(run: SchedulerRun) { selectedRun.value = run; runDetailVisible.value = true; }
function isActiveRun(run?: SchedulerRun) { return run?.status === 'running' || run?.status === 'stopping'; }
function canStop(run: SchedulerRun) { return isActiveRun(run); }
function canRetry(run: SchedulerRun) { return !isActiveRun(run); }
async function retryRun(run: SchedulerRun) { try { await retrySchedulerRun(run.id); MessagePlugin.success('已创建重试运行'); await loadRuns(); } catch (error) { MessagePlugin.error((error as Error).message || '重试失败'); } }
async function stopRun(run: SchedulerRun) { if (!canStop(run)) return; try { await stopSchedulerRun(run.id); MessagePlugin.success('已请求停止'); await loadRuns(); } catch (error) { MessagePlugin.error((error as Error).message || '停止失败'); } }
async function openRunLogs(run: SchedulerRun) { closeRunLogs(); selectedRun.value = run; runLogsVisible.value = true; try { const result = await listSchedulerRunLogs({ runId: run.id, page: 1, pageSize: 200 }); runLogs.value = result.list ?? []; if (isActiveRun(run)) { logStream = streamSchedulerRunLogs(run.id, appendLog, { afterId: runLogs.value.at(-1)?.id, onError: (error) => MessagePlugin.error(error.message || '订阅执行日志失败') }); } } catch (error) { MessagePlugin.error((error as Error).message || '加载执行日志失败'); } }
function appendLog(log: SchedulerRunLog) { if (!runLogs.value.some((item) => item.id === log.id)) runLogs.value.push(log); }
function closeRunLogs() { logStream?.close(); logStream = undefined; }
async function copyLogs() { try { await navigator.clipboard.writeText(formattedRunLogs.value); MessagePlugin.success('日志已复制'); } catch { MessagePlugin.error('复制日志失败'); } }
function downloadLogs() { const blob = new Blob([formattedRunLogs.value], { type: 'text/plain;charset=utf-8' }); const url = URL.createObjectURL(blob); const link = document.createElement('a'); link.href = url; link.download = `scheduler-run-${selectedRun.value?.id || 'logs'}.log`; document.body.appendChild(link); link.click(); link.remove(); URL.revokeObjectURL(url); }
function runStatusTheme(status: string): 'success' | 'danger' | 'warning' | 'default' { if (status === 'succeeded') return 'success'; if (status === 'failed') return 'danger'; if (['timed_out', 'running', 'stopping'].includes(status)) return 'warning'; return 'default'; }
function runStatusText(status: string) { return { running: '运行中', stopping: '停止中', stopped: '已停止', succeeded: '成功', failed: '失败', timed_out: '超时', skipped_locked: '锁占用' }[status] ?? status; }
function formatRunTime(value: string): string { if (!value) return '—'; const date = new Date(value); if (Number.isNaN(date.getTime())) return value; const pad = (part: number) => String(part).padStart(2, '0'); return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`; }
onMounted(async () => { await Promise.all([loadJobs(), loadJobKeys(), loadRuns()]); });
onUnmounted(closeRunLogs);
</script>

<style scoped>
.content-card {
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
    height: auto !important;
    overflow: visible;
    padding: 12px 0 0;
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

.run-filter { display: flex; align-items: center; margin: 0 0 16px; }
.snapshot-section { margin-top: 20px; }
.snapshot-section h4 { margin: 0 0 8px; }
.run-content { margin: 0; padding: 12px; border-radius: 4px; background: var(--td-bg-color-secondarycontainer); white-space: pre-wrap; word-break: break-word; }
.log-actions { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; color: var(--td-text-color-secondary); }
.log-content { min-height: 360px; max-height: calc(100vh - 190px); overflow: auto; }
</style>
