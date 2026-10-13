<template>
  <div class="content-card">
    <s-table
      row-key="id"
      :columns="ledgerColumns as any"
      :data="ledgerList as any"
      :total="ledgerTotal"
      :page="ledgerPage"
      :page-size="ledgerPageSize"
      :loading="ledgerLoading"
      keyword-placeholder="原名模糊搜索"
      :search-fields="ledgerSearchFields"
      @search="onLedgerSearch"
      @reset="
        () => {
          ledgerKeyword = '';
          ledgerFileName = '';
          ledgerSuffix = undefined;
          ledgerCreateBy = '';
          ledgerService = '';
          ledgerTime = [];
          ledgerPage = 1;
          loadLedger();
        }
      "
      @refresh="loadLedger"
      @page-change="
        (p, ps) => {
          ledgerPage = p;
          ledgerPageSize = ps;
          loadLedger();
        }
      "
    >
      <template #toolbar>
        <t-button theme="primary" @click="imageUploadVisible = true">上传图片</t-button>
        <t-button theme="primary" @click="fileUploadVisible = true">上传附件</t-button>
        <t-button variant="outline" @click="previewEnabled = !previewEnabled">
          预览开关 : {{ previewEnabled ? '启用' : '禁用' }}
        </t-button>
      </template>
      <template #batch="{ selected, clear }">
        <t-button
          size="small"
          variant="outline"
          theme="danger"
          :loading="batchDeleting"
          @click="batchRemove(selected, clear)"
        >
          批量删除
        </t-button>
      </template>
      <template #size="{ row }">{{ formatSize((row as OssFileDTO).size) }}</template>
      <template #url="{ row }">
        <s-image-preview
          v-if="previewEnabled && checkImageSuffix((row as OssFileDTO).fileSuffix)"
          :src="(row as OssFileDTO).url"
          width="60px"
          height="60px"
        />
        <t-tooltip v-else :content="(row as OssFileDTO).url" placement="top">
          <div class="t-text-ellipsis" v-text="(row as OssFileDTO).url" />
        </t-tooltip>
      </template>
      <template #isLock="{ row }">
        <t-tag v-if="(row as OssFileDTO).isLock" theme="warning" variant="light">已锁定</t-tag>
        <t-tag v-else theme="default" variant="light">正常</t-tag>
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link v-permission="'system:oss:download'" theme="primary" @click="download(row as OssFileDTO)"
            >下载</t-link
          >
          <t-link v-permission="'system:oss:edit'" theme="primary" @click="openEdit(row as OssFileDTO)">编辑</t-link>
          <t-link
            v-permission="'system:oss:remove'"
            theme="danger"
            @click="
              removeTarget = row as OssFileDTO;
              confirmVisible = true;
            "
            >删除</t-link
          >
        </t-space>
      </template>
    </s-table>

    <!-- 编辑（原名 + 锁定，2 字段 ≤5 → t-dialog） -->
    <t-dialog
      v-model:visible="editVisible"
      header="编辑文件"
      width="460px"
      :confirm-btn="{ content: '保存', loading: saving }"
      @confirm="saveEdit"
    >
      <t-form :data="editForm" label-width="80px">
        <t-form-item label="原名">
          <t-input v-model="editForm.originalName" :maxlength="255" clearable />
        </t-form-item>
        <t-form-item label="锁定">
          <t-switch v-model="editForm.isLock" />
          <span class="oss-lock-tip">锁定后不可删除</span>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 上传（s-upload-image / s-upload-file：复刻 ruoyi，内部走 common/file 并自动入台账） -->
    <!-- 上传图片 -->
    <t-dialog
      v-model:visible="imageUploadVisible"
      header="上传图片"
      width="360px"
      :footer="false"
      @close="imageUrl = ''"
    >
      <div class="oss-upload-pane">
        <s-upload-image
          v-model="imageUrl"
          :limit="1"
          :is-show-tip="false"
          :show-image-file-name="false"
          @success="onImageUploaded"
        />
        <div v-if="imageUrl" class="oss-upload__result">
          <t-tag theme="success" variant="light">上传成功</t-tag>
          <t-link theme="primary" @click="copyUrl(imageUrl)">复制URL</t-link>
        </div>
      </div>
    </t-dialog>

    <!-- 上传附件 -->
    <t-dialog
      v-model:visible="fileUploadVisible"
      header="上传附件"
      width="460px"
      :footer="false"
      @close="fileValue = ''"
    >
      <div class="oss-upload-pane">
        <s-upload-file v-model="fileValue" :limit="1" :is-show-tip="false" @success="onFileUploaded" />
      </div>
    </t-dialog>

    <t-dialog
      v-model:visible="confirmVisible"
      header="删除文件"
      :body="confirmBody"
      :confirm-btn="{ content: '删除', theme: 'danger', loading: saving }"
      @confirm="onConfirmDelete"
    />
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: 'SystemStorageOss' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { downloadCommonFile } from '@/api/common/file';
import { listOssFiles, removeOssFile, updateOssFile } from '@/api/system/oss';
import { listOssConfigs } from '@/api/system/ossConfig';
import type { OssFileDTO, OssFileSaveParams } from '@/api/system/ossModel';
import SImagePreview from '@/components/s-image-preview/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import SUploadFile from '@/components/s-upload-file/index.vue';
import SUploadImage from '@/components/s-upload-image/index.vue';

// ===== 台账 =====
const ledgerLoading = ref(false);
const ledgerList = ref<OssFileDTO[]>([]);
const ledgerTotal = ref(0);
const ledgerPage = ref(1);
const ledgerPageSize = ref(10);
const ledgerKeyword = ref('');
const ledgerFileName = ref('');
const ledgerService = ref('');
const ledgerSuffix = ref<string | undefined>();
const ledgerCreateBy = ref('');
const ledgerTime = ref<string[]>([]);

// 上传（图片/附件分组件，spec 对齐 ruoyi 文件管理页）
const imageUploadVisible = ref(false);
const fileUploadVisible = ref(false);
const imageUrl = ref('');
const fileValue = ref('');
// 预览开关：控制表格图片预览列（对齐源码「预览开关」）
const previewEnabled = ref(true);

const saving = ref(false);
const batchDeleting = ref(false);

const editVisible = ref(false);
const editForm = ref<OssFileSaveParams>({ id: '', originalName: '', isLock: false });

const confirmVisible = ref(false);
const removeTarget = ref<OssFileDTO | null>(null);
const confirmBody = computed(() =>
  removeTarget.value ? `确认删除文件「${removeTarget.value.originalName}」？删除后不可恢复。` : '',
);

// 统一文件服务（spec S2.11）：t-upload 原生 XHR 不经 axios 拦截器，
// action 必须拼 VITE_API_URL_PREFIX、token 必须手动带（漏带 Authorization 会 401）

// 服务商下拉：从 OSS配置 列表取 config_key（对齐源码服务商筛选为下拉）
const serviceOptions = ref<Array<{ label: string; value: string }>>([]);
async function loadServiceOptions() {
  try {
    const res = await listOssConfigs({ page: 1, pageSize: 200 });
    serviceOptions.value = (res.list || []).map((c) => ({ label: c.configKey, value: c.configKey }));
  } catch {
    // 配置列表不可用时退化为空选项，不影响主列表
  }
}

const suffixOptions = [
  { label: '图片 (png/jpg/gif/webp)', value: 'png,jpg,jpeg,gif,webp' },
  { label: '文档 (pdf/doc/xlsx)', value: 'pdf,doc,docx,xls,xlsx' },
  { label: '压缩包 (zip/rar/7z)', value: 'zip,rar,7z' },
];

const ledgerSearchFields = computed<STableSearchField[]>(() => [
  { name: 'fileName', label: '文件名', type: 'input', placeholder: '对象 key 模糊' },
  { name: 'fileSuffix', label: '后缀', type: 'select', options: suffixOptions },
  { name: 'createBy', label: '上传人', type: 'input' },
  { name: 'service', label: '服务商', type: 'select', options: serviceOptions.value },
  { name: 'ledgerTime', label: '上传时间', type: 'daterange', placeholder: ['开始日期', '结束日期'] },
]);

const ledgerColumns = computed(() => [
  { colKey: 'fileName', title: '文件名', minWidth: 220, ellipsis: true },
  { colKey: 'originalName', title: '原名', minWidth: 200, ellipsis: true },
  { colKey: 'fileSuffix', title: '后缀', width: 80 },
  { colKey: 'size', title: '大小', width: 100, cell: 'size' },
  { colKey: 'contentType', title: '类型', width: 140, ellipsis: true },
  ...(previewEnabled.value ? [{ colKey: 'url', title: '预览', width: 150, cell: 'url' }] : []),
  { colKey: 'service', title: '服务商', width: 100 },
  { colKey: 'isLock', title: '状态', width: 90, cell: 'isLock' },
  { colKey: 'createBy', title: '上传人', width: 100 },
  { colKey: 'createdAt', title: '上传时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 160, fixed: 'right', cell: 'ops' },
]);

function formatSize(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / 1024 ** i).toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}

// 图片后缀白名单（对齐 ruoyi checkFileSuffix）
function checkImageSuffix(fileSuffix: string): boolean {
  const arr = ['png', 'jpg', 'jpeg', 'ico', 'bmp', 'webp', 'gif', 'svg'];
  return arr.some((type) => fileSuffix?.includes(type));
}

async function copyUrl(url: string) {
  try {
    await navigator.clipboard.writeText(url);
    MessagePlugin.success('已复制');
  } catch {
    MessagePlugin.warning('复制失败，请手动复制');
  }
}

async function loadLedger() {
  ledgerLoading.value = true;
  try {
    const [beginTime, endTime] = ledgerTime.value ?? [];
    const res = await listOssFiles({
      page: ledgerPage.value,
      pageSize: ledgerPageSize.value,
      fileName: ledgerFileName.value || undefined,
      originalName: ledgerKeyword.value || undefined,
      fileSuffix: ledgerSuffix.value || undefined,
      createBy: ledgerCreateBy.value || undefined,
      service: ledgerService.value || undefined,
      beginTime: beginTime ? `${beginTime} 00:00:00` : undefined,
      endTime: endTime ? `${endTime} 23:59:59` : undefined,
    });
    ledgerList.value = res.list || [];
    ledgerTotal.value = res.total || 0;
    // 删除后可能停在越界页：回退到末页重拉
    if (ledgerList.value.length === 0 && ledgerPage.value > 1 && ledgerTotal.value > 0) {
      ledgerPage.value = Math.max(1, Math.ceil(ledgerTotal.value / ledgerPageSize.value));
      await loadLedger();
    }
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  } finally {
    ledgerLoading.value = false;
  }
}

function onLedgerSearch(params: Record<string, unknown>) {
  ledgerKeyword.value = String(params.keyword ?? '');
  ledgerFileName.value = String(params.fileName ?? '');
  ledgerSuffix.value = params.fileSuffix === undefined ? undefined : String(params.fileSuffix);
  ledgerCreateBy.value = String(params.createBy ?? '');
  ledgerService.value = String(params.service ?? '');
  ledgerTime.value = Array.isArray(params.ledgerTime) ? (params.ledgerTime as string[]) : [];
  ledgerPage.value = 1;
  loadLedger();
}

function openEdit(row: OssFileDTO) {
  editForm.value = { id: row.id, originalName: row.originalName, isLock: row.isLock };
  editVisible.value = true;
}

async function saveEdit() {
  if (!editForm.value.originalName?.trim()) {
    MessagePlugin.warning('原名不能为空');
    return;
  }
  saving.value = true;
  try {
    await updateOssFile({
      id: editForm.value.id,
      originalName: editForm.value.originalName.trim(),
      isLock: !!editForm.value.isLock,
    });
    MessagePlugin.success('保存成功');
    editVisible.value = false;
    await loadLedger();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function download(row: OssFileDTO) {
  try {
    // blob 下载带 Authorization（window.open 直开会 401，spec S2.11 注记）
    await downloadCommonFile(row.id, row.originalName);
    MessagePlugin.success('下载已开始');
  } catch (e) {
    MessagePlugin.error((e as Error).message || '下载失败');
  }
}

async function onConfirmDelete() {
  if (!removeTarget.value) return;
  saving.value = true;
  try {
    await removeOssFile([removeTarget.value.id]);
    MessagePlugin.success('已删除');
    confirmVisible.value = false;
    removeTarget.value = null;
    await loadLedger();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  } finally {
    saving.value = false;
  }
}

async function batchRemove(selected: unknown[], clear: () => void) {
  if (!selected.length) return;
  const rows = selected as OssFileDTO[];
  batchDeleting.value = true;
  try {
    await removeOssFile(rows.map((r) => r.id));
    MessagePlugin.success(`已删除 ${rows.length} 个文件`);
    clear();
    await loadLedger();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  } finally {
    batchDeleting.value = false;
  }
}

// 上传成功（组件默认不弹通用提示，由页面给出业务提示 + 关窗 + 刷新台账）
function onImageUploaded() {
  MessagePlugin.success('图片上传成功');
  imageUploadVisible.value = false;
  imageUrl.value = '';
  loadLedger();
}

function onFileUploaded(files: { name: string }[]) {
  MessagePlugin.success(`附件「${files[0]?.name ?? ''}」上传成功`);
  fileUploadVisible.value = false;
  fileValue.value = '';
  loadLedger();
}

onMounted(() => {
  loadLedger();
  loadServiceOptions();
});
</script>
<style scoped>
.oss-lock-tip {
  margin-left: var(--td-comp-margin-s);
  color: var(--td-text-color-secondary);
  font-size: 12px;
}

.oss-upload-pane {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-m);
  align-items: flex-start;
  padding: var(--td-comp-margin-xs) 0;
}

.oss-upload__result {
  display: flex;
  gap: var(--td-comp-margin-s);
  align-items: center;
  margin-top: var(--td-comp-margin-m);
}
</style>
