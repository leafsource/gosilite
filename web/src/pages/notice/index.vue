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
      :selectable="false"
      keyword-placeholder="标题"
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
      <template #noticeType="{ row }">
        <s-tag-dict dict-type="sys_notice_type" :value="row.noticeType" />
      </template>
      <template #read="{ row }">
        <t-tag :theme="row.read ? 'default' : 'primary'" variant="light">
          {{ row.read ? '已读' : '未读' }}
        </t-tag>
      </template>
      <template #ops="{ row }">
        <t-link theme="primary" @click="openDetail(row as NoticeInboxItem)">查看</t-link>
      </template>
    </s-table>

    <s-notice-detail v-model:visible="detailVisible" :detail="detail ?? null" />
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, ref } from 'vue';

import { getNoticeDetail, listNoticeInbox } from '@/api/notice';
import type { NoticeDetail, NoticeInboxItem } from '@/api/noticeModel';
import SNoticeDetail from '@/components/s-notice-detail/index.vue';
import STable from '@/components/s-table/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import { useConfigStore, useNotificationStore } from '@/store';

const notificationStore = useNotificationStore();

const loading = ref(false);
const list = ref<NoticeInboxItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(useConfigStore().pageDefault);
const keyword = ref('');

const detailVisible = ref(false);
const detail = ref<NoticeDetail | undefined>();

const columns = [
  { colKey: 'title', title: '标题', minWidth: 220, ellipsis: true },
  { colKey: 'noticeType', title: '类型', width: 90, cell: 'noticeType' },
  { colKey: 'read', title: '状态', width: 90, cell: 'read' },
  { colKey: 'publisher', title: '发布人', width: 110 },
  { colKey: 'createdAt', title: '发布时间', width: 170 },
  { colKey: 'ops', title: '操作', width: 100, fixed: 'right', cell: 'ops' },
];

async function load() {
  loading.value = true;
  try {
    const res = await listNoticeInbox({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
    });
    list.value = res.list || [];
    total.value = res.total || 0;
    notificationStore.unread = res.unread ?? notificationStore.unread;
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

async function openDetail(row: NoticeInboxItem) {
  try {
    detail.value = await getNoticeDetail(row.id);
    detailVisible.value = true;
    // 服务端已自动已读：同步铃铛角标与当前行状态
    notificationStore.applyRead(row.id);
    const hit = list.value.find((i) => i.id === row.id);
    if (hit) hit.read = true;
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  }
}

onMounted(load);
</script>
