<template>
  <div class="content-card">
    <!-- 概览：连通性 / 配置 / 用量 -->
    <t-card class="storage-overview" :bordered="false" :title="t('pages.storageOverview.title')">
      <t-row :gutter="[16, 16]">
        <t-col :span="3">
          <div class="storage-stat">
            <div class="storage-stat__label">Endpoint</div>
            <div class="storage-stat__value" :title="info?.endpoint">
              {{ info?.endpoint || '-' }}
            </div>
          </div>
        </t-col>
        <t-col :span="2">
          <div class="storage-stat">
            <div class="storage-stat__label">{{ t('pages.storageOverview.bucket') }}</div>
            <div class="storage-stat__value">{{ info?.bucket || '-' }}</div>
          </div>
        </t-col>
        <t-col :span="3">
          <div class="storage-stat">
            <div class="storage-stat__label">{{ t('pages.storageOverview.domain') }}</div>
            <div class="storage-stat__value" :title="info?.domain">
              {{ info?.domain || '-' }}
            </div>
          </div>
        </t-col>
        <t-col :span="2">
          <div class="storage-stat">
            <div class="storage-stat__label">{{ t('pages.storageOverview.status') }}</div>
            <t-tag v-if="info?.connected" theme="success" variant="light">
              {{ t('pages.storageOverview.connected') }}
            </t-tag>
            <t-tag v-else-if="info" theme="danger" variant="light">
              {{ t('pages.storageOverview.disconnected') }}
            </t-tag>
            <span v-else>-</span>
          </div>
        </t-col>
        <t-col :span="2">
          <div class="storage-stat">
            <div class="storage-stat__label">{{ t('pages.storageOverview.objectCount') }}</div>
            <div class="storage-stat__value">
              {{ info ? info.objectCount : '-' }}
            </div>
          </div>
        </t-col>
        <t-col :span="2">
          <div class="storage-stat">
            <div class="storage-stat__label">{{ t('pages.storageOverview.totalSize') }}</div>
            <div class="storage-stat__value">
              {{ info ? formatSize(info.totalSize) : '-' }}
            </div>
          </div>
        </t-col>
      </t-row>
      <div v-if="info && !info.connected && info.message" class="storage-overview__error">
        {{ info.message }}
      </div>
      <template #actions>
        <t-button size="small" variant="outline" :loading="infoLoading" @click="loadInfo">
          {{ t('pages.storageOverview.refresh') }}
        </t-button>
      </template>
    </t-card>

    <s-table
      row-key="key"
      :columns="columns as any"
      :data="list as any"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :selectable="false"
      :keyword-placeholder="t('pages.storageTable.keywordPlaceholder')"
      :search-fields="searchFields"
      @search="onSearch"
      @reset="
        () => {
          prefix = '';
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
      <template #toolbar>
        <t-button v-permission="'system:storage:upload'" theme="primary" @click="uploadVisible = true">
          {{ t('pages.storageTable.uploadTest') }}
        </t-button>
      </template>
      <template #size="{ row }">{{ formatSize(row.size) }}</template>
      <template #ops="{ row }">
        <t-link v-permission="'system:storage:delete'" theme="danger" @click="askRemove(row as StorageObject)">
          {{ t('pages.storageTable.remove') }}
        </t-link>
      </template>
    </s-table>

    <!-- 删除确认：对象删除不可恢复，强确认 -->
    <t-dialog
      v-model:visible="removeVisible"
      :header="t('pages.storageTable.removeTitle')"
      :confirm-btn="{
        content: t('pages.storageTable.remove'),
        theme: 'danger',
        loading: removing,
        disabled: removing,
      }"
      :close-on-confirm="false"
      @confirm="doRemove"
    >
      <p class="storage-remove-tip">
        {{ t('pages.storageTable.removeConfirm', { key: removeRow?.key }) }}
      </p>
    </t-dialog>

    <!-- 上传测试 -->
    <t-dialog
      v-model:visible="uploadVisible"
      :header="t('pages.storageTable.uploadTitle')"
      :footer="false"
      width="380px"
      attach="body"
    >
      <div class="storage-upload">
        <s-upload-image
          action="/system/storage/upload"
          field="file"
          :url="uploadedUrl"
          size="96px"
          :max-size-mb="10"
          :upload-text="t('pages.storageTable.uploadAction')"
          @success="onUploaded"
        />
        <div v-if="uploadedUrl" class="storage-upload__url" :title="uploadedUrl">
          {{ uploadedUrl }}
        </div>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';

import { getStorageInfo, listStorageObjects, removeStorageObject } from '@/api/system/storage';
import type { StorageInfo, StorageObject } from '@/api/system/storageModel';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import SUploadImage from '@/components/s-upload-image/index.vue';
import { t } from '@/locales';
import { useConfigStore } from '@/store';

defineOptions({ name: 'SystemStorage' });

const info = ref<StorageInfo | null>(null);
const infoLoading = ref(false);

const loading = ref(false);
const list = ref<StorageObject[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const prefix = ref('');
const keyword = ref('');

const removeVisible = ref(false);
const removing = ref(false);
const removeRow = ref<StorageObject | null>(null);

const uploadVisible = ref(false);
const uploadedUrl = ref('');

const searchFields = computed<STableSearchField[]>(() => [
  {
    name: 'prefix',
    label: t('pages.storageTable.prefix'),
    type: 'input',
    placeholder: t('pages.storageTable.prefixPlaceholder'),
  },
]);

const columns = [
  { colKey: 'key', title: 'Key', minWidth: 280, ellipsis: true },
  { colKey: 'size', title: t('pages.storageTable.size'), width: 110, cell: 'size' },
  { colKey: 'lastModified', title: t('pages.storageTable.lastModified'), width: 180 },
  { colKey: 'type', title: t('pages.storageTable.type'), width: 110 },
  { colKey: 'ops', title: t('pages.storageTable.ops'), width: 90, fixed: 'right', cell: 'ops' },
];

function formatSize(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / 1024 ** i).toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}

async function loadInfo() {
  infoLoading.value = true;
  try {
    info.value = await getStorageInfo();
  } catch (e) {
    MessagePlugin.error((e as Error).message || t('pages.storageOverview.loadFailed'));
  } finally {
    infoLoading.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const res = await listStorageObjects({
      page: page.value,
      pageSize: pageSize.value,
      prefix: prefix.value || undefined,
      keyword: keyword.value || undefined,
    });
    list.value = res.list || [];
    total.value = res.total || 0;
    // 删除后可能停在越界页：回退到末页重拉
    if (list.value.length === 0 && page.value > 1 && total.value > 0) {
      page.value = Math.max(1, Math.ceil(total.value / pageSize.value));
      await load();
      return;
    }
  } catch (e) {
    MessagePlugin.error((e as Error).message || t('pages.storageTable.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function onSearch(params: Record<string, unknown>) {
  prefix.value = String(params.prefix ?? '');
  keyword.value = String(params.keyword ?? '');
  page.value = 1;
  load();
}

function askRemove(row: StorageObject) {
  removeRow.value = row;
  removeVisible.value = true;
}

async function doRemove() {
  if (!removeRow.value) return;
  removing.value = true;
  try {
    await removeStorageObject(removeRow.value.key);
    MessagePlugin.success(t('pages.storageTable.removeSuccess'));
    removeVisible.value = false;
    removeRow.value = null;
    await Promise.all([load(), loadInfo()]);
  } catch (e) {
    MessagePlugin.error((e as Error).message || t('pages.storageTable.removeFailed'));
  } finally {
    removing.value = false;
  }
}

function onUploaded(payload: { url: string }) {
  uploadedUrl.value = payload.url;
  MessagePlugin.success(t('pages.storageTable.uploadSuccess'));
  // 不自动关闭：让用户能看到 webURI，确认后手动关闭
  Promise.all([load(), loadInfo()]);
}

onMounted(() => {
  loadInfo();
  load();
});
</script>

<style lang="less" scoped>
.storage-overview {
  margin-bottom: var(--td-comp-margin-m, 12px);

  :deep(.t-card__body) {
    padding-bottom: var(--td-comp-paddingTB-s, 8px);
  }

  &__error {
    margin-top: var(--td-comp-margin-s, 8px);
    font-size: var(--td-font-size-body-small, 12px);
    color: var(--td-error-color, #d54941);
  }
}

.storage-stat {
  min-width: 0;

  &__label {
    font-size: var(--td-font-size-body-small, 12px);
    color: var(--td-text-color-secondary);
    margin-bottom: 4px;
  }

  &__value {
    font-size: var(--td-font-size-body-medium, 14px);
    color: var(--td-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.storage-remove-tip {
  word-break: break-all;
  line-height: 1.6;
}

.storage-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--td-comp-margin-m, 12px);
  padding: var(--td-comp-paddingTB-m, 12px) 0;

  &__url {
    max-width: 320px;
    font-size: var(--td-font-size-body-small, 12px);
    color: var(--td-text-color-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
