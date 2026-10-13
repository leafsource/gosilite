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
      keyword-placeholder="标题/正文"
      :search-fields="searchFields"
      @search="onSearch"
      @reset="
        () => {
          keyword = '';
          noticeType = undefined;
          status = undefined;
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
        <t-button v-permission="'system:notice:add'" theme="primary" @click="open()">新增通知</t-button>
      </template>
      <template #batch="{ selected, clear }">
        <t-button
          size="small"
          variant="outline"
          theme="danger"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchRemove(selected, clear)"
        >
          批量删除
        </t-button>
        <t-button
          size="small"
          variant="outline"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchSetStatus(selected, 0, clear)"
        >
          批量发布
        </t-button>
        <t-button
          size="small"
          variant="outline"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="batchSetStatus(selected, 1, clear)"
        >
          批量撤回
        </t-button>
      </template>
      <template #noticeType="{ row }">
        <s-tag-dict dict-type="sys_notice_type" :value="row.noticeType" />
      </template>
      <template #status="{ row }">
        <s-tag-dict dict-type="sys_notice_status" :value="row.status" />
      </template>
      <template #ops="{ row }">
        <t-space>
          <t-link theme="primary" @click="view(row as NoticeDTO)">查看</t-link>
          <t-link v-permission="'system:notice:edit'" theme="primary" @click="open(row as NoticeDTO)">编辑</t-link>
          <t-link v-permission="'system:notice:remove'" theme="danger" @click="del(row as NoticeDTO)">删除</t-link>
        </t-space>
      </template>
    </s-table>

    <!-- 字段 4 个（标题/类型/状态/正文）≤5 → t-dialog（s-form-dialog） -->
    <s-form-dialog v-model:visible="dialog" :title="form.id ? '编辑通知' : '新增通知'" width="560px" @confirm="save">
      <t-form :data="form" label-width="80px">
        <t-form-item label="标题">
          <t-input v-model="form.title" :maxlength="128" clearable placeholder="请输入标题" />
        </t-form-item>
        <t-form-item label="类型">
          <s-select-dict v-model="form.noticeType" dict-type="sys_notice_type" value-type="number" />
        </t-form-item>
        <t-form-item label="状态">
          <s-radio-dict v-model="form.status" dict-type="sys_notice_status" value-type="number" />
        </t-form-item>
        <t-form-item label="正文">
          <t-textarea
            v-model="form.content"
            :maxlength="2000"
            :autosize="{ minRows: 5, maxRows: 12 }"
            placeholder="请输入通知正文"
          />
        </t-form-item>
      </t-form>
    </s-form-dialog>

    <!-- 只读查看：SNoticeDetail（attach=body + 统一样式） -->
    <s-notice-detail v-model:visible="viewVisible" :detail="viewRow ?? null" />
  </div>
</template>
<script setup lang="ts">

defineOptions({ name: 'SystemNotice' });
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import { createNotice, listNotices, removeNotice, updateNotice } from '@/api/system/notice';
import type { NoticeDTO, NoticeSaveParams } from '@/api/system/noticeModel';
import SFormDialog from '@/components/s-form-dialog/index.vue';
import SNoticeDetail from '@/components/s-notice-detail/index.vue';
import SRadioDict from '@/components/s-radio-dict/index.vue';
import SSelectDict from '@/components/s-select-dict/index.vue';
import STable from '@/components/s-table/index.vue';
import type { STableSearchField } from '@/components/s-table/sTable';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useDict } from '@/store';

const noticeTypeDict = useDict('sys_notice_type');
const statusDict = useDict('sys_notice_status');

const loading = ref(false);
const list = ref<NoticeDTO[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');
const noticeType = ref<number | undefined>();
const status = ref<number | undefined>();

const dialog = ref(false);
const form = ref<Partial<NoticeDTO>>({});

const viewVisible = ref(false);
const viewRow = ref<NoticeDTO | undefined>();

const searchFields = computed<STableSearchField[]>(() => [
  { name: 'noticeType', label: '类型', type: 'select', options: noticeTypeDict.options.value },
  { name: 'status', label: '状态', type: 'select', options: statusDict.options.value },
]);

const columns = [
  { colKey: 'title', title: '标题', minWidth: 200, ellipsis: true },
  { colKey: 'noticeType', title: '类型', width: 90, cell: 'noticeType' },
  { colKey: 'status', title: '状态', width: 90, cell: 'status' },
  { colKey: 'publisher', title: '发布人', width: 110 },
  { colKey: 'createdAt', title: '创建时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 170, fixed: 'right', cell: 'ops' },
];

function toNum(v: unknown): number | undefined {
  if (v === undefined || v === null || v === '') return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

async function load() {
  loading.value = true;
  try {
    const res = await listNotices({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      noticeType: noticeType.value,
      status: status.value,
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
  noticeType.value = toNum(params.noticeType);
  status.value = toNum(params.status);
  page.value = 1;
  load();
}

function open(row?: NoticeDTO) {
  form.value = row ? { ...row } : { title: '', noticeType: 1, status: 0, content: '' };
  dialog.value = true;
}

function view(row: NoticeDTO) {
  viewRow.value = { ...row };
  viewVisible.value = true;
}

async function save() {
  try {
    const payload: NoticeSaveParams = {
      id: form.value.id,
      title: (form.value.title ?? '').trim(),
      noticeType: Number(form.value.noticeType ?? 1),
      content: form.value.content ?? '',
      status: Number(form.value.status ?? 0),
    };
    if (payload.id) await updateNotice(payload);
    else await createNotice(payload);
    MessagePlugin.success('保存成功');
    dialog.value = false;
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '保存失败');
  }
}

async function del(row: NoticeDTO) {
  try {
    await removeNotice(row.id);
    MessagePlugin.success('已删除');
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '删除失败');
  }
}

const batchLoading = ref(false);

// 批量槽位给的是 Record<string, unknown>[]；收 unknown[] 后一次断言（interface 无隐式索引签名）
async function batchRemove(selected: unknown[], clear: () => void) {
  if (!selected.length) return;
  const rows = selected as NoticeDTO[];
  batchLoading.value = true;
  try {
    await Promise.all(rows.map((row) => removeNotice(row.id)));
    MessagePlugin.success(`已删除 ${rows.length} 条通知`);
    clear();
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '批量删除失败');
  } finally {
    batchLoading.value = false;
  }
}

async function batchSetStatus(selected: unknown[], next: number, clear: () => void) {
  if (!selected.length) return;
  const rows = selected as NoticeDTO[];
  batchLoading.value = true;
  try {
    await Promise.all(
      rows.map((row) =>
        updateNotice({ id: row.id, title: row.title, noticeType: row.noticeType, content: row.content, status: next }),
      ),
    );
    MessagePlugin.success(`${statusDict.labelOf(next)}成功 ${rows.length} 条通知`);
    clear();
    await load();
  } catch (e) {
    MessagePlugin.error((e as Error).message || '批量操作失败');
  } finally {
    batchLoading.value = false;
  }
}

onMounted(load);
</script>
