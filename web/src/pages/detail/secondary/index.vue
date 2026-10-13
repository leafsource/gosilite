<template>
  <div>
    <div class="secondary-notification">
      <t-tabs v-model="tabValue">
        <t-tab-panel v-for="(tab, tabIndex) in TAB_LIST" :key="tabIndex" :value="tab.value" :label="tab.label">
          <t-list v-if="msgDataList.length > 0" class="secondary-msg-list" :split="true">
            <t-list-item v-for="(item, index) in msgDataList" :key="index">
              <p class="content" :class="[{ unread: item.status }]" @click="setReadStatus(item)">
                <t-tag size="medium" :theme="NOTIFICATION_TYPES.get(item.quality)" variant="light">
                  {{ item.type }}
                </t-tag>
                {{ item.content }}
              </p>
              <template #action>
                <span class="msg-date">{{ item.date }}</span>

                <div class="msg-action">
                  <t-tooltip
                    class="set-read-icon"
                    :overlay-style="{ margin: '6px' }"
                    :content="item.status ? t('pages.detailSecondary.setRead') : t('pages.detailSecondary.setUnread')"
                  >
                    <span class="msg-action-icon" @click="setReadStatus(item)">
                      <t-icon v-if="!!item.status" name="queue" size="16px" />
                      <t-icon v-else name="chat" />
                    </span>
                  </t-tooltip>
                  <t-tooltip :content="t('pages.detailSecondary.delete')" :overlay-style="{ margin: '6px' }">
                    <span @click="handleClickDeleteBtn(item)">
                      <t-icon name="delete" size="16px" />
                    </span>
                  </t-tooltip>
                </div>
              </template>
            </t-list-item>
          </t-list>
          <div v-else class="secondary-msg-list__empty-list">
            <empty-icon></empty-icon>
            <p>{{ t('pages.detailSecondary.empty') }}</p>
          </div>
        </t-tab-panel>
      </t-tabs>
    </div>
    <t-dialog
      v-model:visible="visible"
      :header="t('pages.detailSecondary.deleteConfirm')"
      :body="t('pages.detailSecondary.deleteTip', { content: selectedItem?.content ?? '' })"
      :on-confirm="deleteMsg"
    />
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';

import EmptyIcon from '@/assets/assets-empty.svg?component';
import { NOTIFICATION_TYPES } from '@/constants';
import { t } from '@/locales';
import type { NotificationItem } from '@/types/interface';

defineOptions({
  name: 'DetailSecondary',
});

// starter 演示页（无路由注册）：与通知中心解耦，保留本地 mock，不依赖 notification store
const getMsgData = (): NotificationItem[] => [
  {
    id: '123',
    content: t('pages.notification.msg1'),
    type: t('pages.notification.contract'),
    status: true,
    collected: false,
    date: '2021-01-01 08:00',
    quality: 'high',
  },
  {
    id: '124',
    content: t('pages.notification.msg2'),
    type: t('pages.notification.invoice'),
    status: true,
    collected: false,
    date: '2021-01-01 08:00',
    quality: 'low',
  },
  {
    id: '125',
    content: t('pages.notification.msg3'),
    type: t('pages.notification.meeting'),
    status: false,
    collected: false,
    date: '2021-01-01 08:00',
    quality: 'middle',
  },
  {
    id: '126',
    content: t('pages.notification.msg4'),
    type: t('pages.notification.invoice'),
    status: true,
    collected: false,
    date: '2021-01-01 08:00',
    quality: 'low',
  },
];

const TAB_LIST = [
  {
    label: t('pages.detailSecondary.all'),
    value: 'msgData',
  },
  {
    label: t('pages.detailSecondary.unread'),
    value: 'unreadMsg',
  },
  {
    label: t('pages.detailSecondary.read'),
    value: 'readMsg',
  },
];

const tabValue = ref('msgData');

const visible = ref(false);
const selectedItem = ref<NotificationItem>();

const msgData = ref<NotificationItem[]>(getMsgData());
const unreadMsg = computed(() => msgData.value.filter((item) => item.status));
const readMsg = computed(() => msgData.value.filter((item) => !item.status));

const msgDataList = computed(() => {
  if (tabValue.value === 'msgData') return msgData.value;
  if (tabValue.value === 'unreadMsg') return unreadMsg.value;
  if (tabValue.value === 'readMsg') return readMsg.value;
  return [];
});

const handleClickDeleteBtn = (item: NotificationItem) => {
  visible.value = true;
  selectedItem.value = item;
};

const setReadStatus = (item: NotificationItem) => {
  msgData.value.forEach((e) => {
    if (e.id === item.id) {
      e.status = !e.status;
    }
  });
};

const deleteMsg = () => {
  const item = selectedItem.value;
  msgData.value = msgData.value.filter((e) => e.id !== item?.id);
  visible.value = false;
};
</script>
<style lang="less" scoped>
@import './index.less';
</style>
