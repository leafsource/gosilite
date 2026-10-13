<template>
  <t-drawer
    :visible="visible"
    :header="drawerHeader"
    size="760px"
    :footer="false"
    attach="body"
    @update:visible="(v: boolean) => emit('update:visible', v)"
  >
    <div v-if="loading" class="log-detail__loading">加载中…</div>
    <template v-else-if="kind === 'oper' && oper">
      <t-descriptions :column="2" bordered size="small">
        <t-descriptions-item label="接口" :span="2">
          <code>{{ oper.base.httpMethod }} {{ oper.base.requestUri }}</code>
        </t-descriptions-item>
        <t-descriptions-item label="操作人">{{ oper.base.userName || '—' }}</t-descriptions-item>
        <t-descriptions-item label="记录时间">{{ oper.base.createdAt }}</t-descriptions-item>
        <t-descriptions-item label="客户端 IP">{{ oper.base.clientIp || '—' }}</t-descriptions-item>
        <t-descriptions-item label="执行时长">{{ oper.base.costMs }} ms</t-descriptions-item>
        <t-descriptions-item label="模块">{{ oper.base.module || '—' }}</t-descriptions-item>
        <t-descriptions-item label="类型">
          <s-tag-dict dict-type="sys_oper_type" :value="oper.base.operType" />
        </t-descriptions-item>
        <t-descriptions-item label="状态" :span="2">
          <s-tag-dict dict-type="sys_oper_status" :value="oper.base.operStatus" />
          <span v-if="oper.base.errorMsg" class="log-detail__err"> {{ oper.base.errorMsg }}</span>
        </t-descriptions-item>
        <t-descriptions-item label="TraceId" :span="2">
          <span class="log-detail__mono">{{ oper.base.traceId || '—' }}</span>
        </t-descriptions-item>
        <t-descriptions-item v-if="oper.base.remark" label="备注" :span="2">
          {{ oper.base.remark }}
        </t-descriptions-item>
      </t-descriptions>

      <h4 class="log-detail__section">请求参数</h4>
      <div class="log-detail__block">
        <div class="log-detail__block-title">Header 参数</div>
        <table v-if="headerRows.length" class="log-detail__kv">
          <thead>
            <tr>
              <th>参数名</th>
              <th>请求值</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in headerRows" :key="h.key">
              <td>{{ h.key }}</td>
              <td class="log-detail__mono">{{ h.value }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="log-detail__empty">无</div>
      </div>

      <div class="log-detail__block">
        <div class="log-detail__block-title">Body 参数</div>
        <pre v-if="prettyBody" class="log-detail__pre">{{ prettyBody }}</pre>
        <div v-else class="log-detail__empty">{{ bodyEmptyText }}</div>
      </div>

      <h4 class="log-detail__section">返回响应</h4>
      <div class="log-detail__block">
        <pre v-if="prettyResp" class="log-detail__pre">{{ prettyResp }}</pre>
        <div v-else class="log-detail__empty">无</div>
      </div>
    </template>

    <template v-else-if="kind === 'error' && err">
      <t-descriptions :column="2" bordered size="small">
        <t-descriptions-item label="接口" :span="2">
          <code>{{ err.base.httpMethod }} {{ err.base.requestUri }}</code>
        </t-descriptions-item>
        <t-descriptions-item label="异常级别">
          <s-tag-dict dict-type="sys_error_level" :value="err.base.errLevel" />
        </t-descriptions-item>
        <t-descriptions-item label="处理状态">
          <s-tag-dict dict-type="sys_handle_status" :value="err.base.handleStatus" />
        </t-descriptions-item>
        <t-descriptions-item label="异常来源">
          {{ err.base.errType === 2 ? '系统 panic' : '业务异常' }}
        </t-descriptions-item>
        <t-descriptions-item label="错误码">{{ err.base.errCode || '—' }}</t-descriptions-item>
        <t-descriptions-item label="操作人">{{ err.base.userName || '—' }}</t-descriptions-item>
        <t-descriptions-item label="记录时间">{{ err.base.createdAt }}</t-descriptions-item>
        <t-descriptions-item label="客户端 IP">{{ err.base.clientIp || '—' }}</t-descriptions-item>
        <t-descriptions-item label="执行时长">{{ err.base.costMs }} ms</t-descriptions-item>
        <t-descriptions-item label="TraceId" :span="2">
          <span class="log-detail__mono">{{ err.base.traceId || '—' }}</span>
        </t-descriptions-item>
        <t-descriptions-item label="异常信息" :span="2">
          <span class="log-detail__err">{{ err.base.errMsg || '—' }}</span>
        </t-descriptions-item>
        <t-descriptions-item v-if="err.base.handledAt" label="处理时间">
          {{ err.base.handledAt }}
        </t-descriptions-item>
        <t-descriptions-item v-if="err.handleRemark" label="处理备注" :span="2">
          {{ err.handleRemark }}
        </t-descriptions-item>
      </t-descriptions>

      <h4 class="log-detail__section">堆栈信息（Stack Trace）</h4>
      <pre v-if="err.stack" class="log-detail__pre log-detail__pre--stack">{{ err.stack }}</pre>
      <div v-else class="log-detail__empty log-detail__empty--hint">
        未采集到堆栈信息。请结合错误码、异常信息与下方请求参数定位；也可用 TraceId
        在操作日志中检索同一次请求的记录。
      </div>

      <h4 class="log-detail__section">请求参数</h4>
      <div class="log-detail__block">
        <div class="log-detail__block-title">Header 参数</div>
        <table v-if="headerRows.length" class="log-detail__kv">
          <thead>
            <tr>
              <th>参数名</th>
              <th>请求值</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in headerRows" :key="h.key">
              <td>{{ h.key }}</td>
              <td class="log-detail__mono">{{ h.value }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="log-detail__empty">无</div>
      </div>

      <div class="log-detail__block">
        <div class="log-detail__block-title">Body 参数</div>
        <pre v-if="prettyBody" class="log-detail__pre">{{ prettyBody }}</pre>
        <div v-else class="log-detail__empty">{{ bodyEmptyText }}</div>
      </div>
    </template>

    <div v-else class="log-detail__loading">暂无数据</div>
  </t-drawer>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { getErrorLog, getOperLog } from '@/api/system/log';
import type { ErrorLogDetailDTO, OperLogDetailDTO } from '@/api/system/logModel';
import STagDict from '@/components/s-tag-dict/index.vue';

defineOptions({ name: 'LogDetailDrawer' });

const props = defineProps<{
  visible: boolean;
  kind: 'oper' | 'error';
  id?: string;
}>();

const emit = defineEmits<(e: 'update:visible', v: boolean) => void>();

const loading = ref(false);
const oper = ref<OperLogDetailDTO | null>(null);
const err = ref<ErrorLogDetailDTO | null>(null);

const drawerHeader = computed(
  () => `${props.kind === 'oper' ? '操作详情' : '错误详情'}${props.id ? ` #${props.id}` : ''}`,
);

function pretty(raw: string): string {
  if (!raw) return '';
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

const headersRaw = computed(
  () => (props.kind === 'oper' ? oper.value?.requestHeaders : err.value?.requestHeaders) ?? '',
);
const bodyRaw = computed(() => (props.kind === 'oper' ? oper.value?.requestBody : err.value?.requestBody) ?? '');

const prettyBody = computed(() => pretty(bodyRaw.value));

const prettyResp = computed(() => pretty(oper.value?.responseBody ?? ''));

const bodyEmptyText = '无';

const headerRows = computed(() =>
  headersRaw.value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const i = line.indexOf(':');
      return i > 0 ? { key: line.slice(0, i), value: line.slice(i + 1).trim() } : { key: line, value: '' };
    }),
);

watch(
  () => [props.visible, props.id] as const,
  async ([open, id]) => {
    if (!open || !id) return;
    loading.value = true;
    oper.value = null;
    err.value = null;
    try {
      if (props.kind === 'oper') oper.value = await getOperLog(id);
      else err.value = await getErrorLog(id);
    } catch {
      // 详情加载失败由空态兜底
    } finally {
      loading.value = false;
    }
  },
  { immediate: true },
);
</script>
<style scoped lang="less">
.log-detail {
  &__loading {
    padding: 32px 0;
    text-align: center;
    color: var(--td-text-color-placeholder);
  }

  &__section {
    margin: 20px 0 8px;
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-color-primary);
  }

  &__block {
    margin-bottom: 12px;
    padding: 12px;
    border: 1px solid var(--td-component-border);
    border-radius: var(--td-radius-default);
    background: var(--td-bg-color-container);
  }

  &__block-title {
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 600;
    color: var(--td-text-color-secondary);
  }

  &__kv {
    width: 100%;
    font-size: 13px;
    border-collapse: collapse;

    th,
    td {
      padding: 6px 10px;
      border: 1px solid var(--td-component-border);
      text-align: left;
      word-break: break-all;
    }

    th {
      background: var(--td-bg-color-secondarycontainer);
      color: var(--td-text-color-secondary);
      font-weight: 500;
    }
  }

  &__pre {
    margin: 0;
    padding: 10px 12px;
    max-height: 320px;
    overflow: auto;
    font-family: var(--td-font-family-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    border-radius: var(--td-radius-small);
    background: var(--td-bg-color-secondarycontainer);

    &--stack {
      max-height: 320px;
      // 堆栈保持原始缩进，不折行；超宽横向滚动，便于对照排错
      white-space: pre;
      word-break: normal;
      color: var(--td-error-color);
    }
  }

  &__empty {
    padding: 8px 0;
    font-size: 13px;
    color: var(--td-text-color-placeholder);

    &--hint {
      padding: 12px;
      border: 1px dashed var(--td-component-border);
      border-radius: var(--td-radius-default);
      line-height: 1.7;
    }
  }

  &__mono {
    font-family: var(--td-font-family-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
    font-size: 12px;
    word-break: break-all;
  }

  &__err {
    margin-left: 8px;
    color: var(--td-error-color);
  }
}
</style>
