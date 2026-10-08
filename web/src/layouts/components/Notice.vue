<template>
  <t-popup v-model:visible="popupVisible" expand-animation placement="bottom-right" trigger="click">
    <template #content>
      <div class="header-msg">
        <div class="header-msg-top">
          <p>{{ t('layout.notice.title') }}</p>
          <t-button
            v-if="unread > 0"
            class="clear-btn"
            variant="text"
            theme="primary"
            :loading="marking"
            @click="onMarkAll"
          >
            {{ t('layout.notice.clear') }}
          </t-button>
        </div>
        <t-list v-if="recent.length > 0" class="narrow-scrollbar" :split="false">
          <t-list-item v-for="item in recent" :key="item.id" @click="openDetail(item.id)">
            <div>
              <p class="msg-content">
                <span v-if="!item.read" class="unread-dot" />
                {{ item.title }}
              </p>
              <p class="msg-type">{{ typeLabel(item.noticeType) }} · {{ item.publisher }}</p>
            </div>
            <p class="msg-time">{{ item.createdAt }}</p>
            <template #action>
              <t-button v-if="!item.read" size="small" variant="outline" @click.stop="onMarkOne(item.id)">
                {{ t('layout.notice.setRead') }}
              </t-button>
            </template>
          </t-list-item>
        </t-list>

        <div v-else class="empty-list">
          <img src="https://tdesign.gtimg.com/pro-template/personal/nothing.png" alt="空" />
          <p>{{ t('layout.notice.emptyNotice') }}</p>
        </div>
        <div v-if="recent.length > 0" class="header-msg-bottom">
          <t-button class="header-msg-bottom-link" variant="text" theme="default" block @click="goAll">
            {{ t('layout.notice.viewAll') }}
          </t-button>
        </div>
      </div>
    </template>
    <t-badge :count="unread" :offset="[4, 4]">
      <t-button theme="default" shape="square" variant="text">
        <t-icon name="mail" />
      </t-button>
    </t-badge>
  </t-popup>

  <!-- 详情（打开即自动已读）：SNoticeDetail 内部 attach=body，脱离 header 裁剪 -->
  <s-notice-detail v-model:visible="detailVisible" :detail="detail ?? null" />
</template>
<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { getNoticeDetail } from '@/api/notice';
import type { NoticeDetail } from '@/api/noticeModel';
import SNoticeDetail from '@/components/s-notice-detail/index.vue';
import { t } from '@/locales';
import { useDict, useNotificationStore } from '@/store';

const router = useRouter();
const store = useNotificationStore();
const { unread, recent } = storeToRefs(store);
const typeDict = useDict('sys_notice_type');

const popupVisible = ref(false);
const marking = ref(false);
const detailVisible = ref(false);
const detail = ref<NoticeDetail | undefined>();

const typeLabel = (noticeType: number) => typeDict.labelOf(noticeType);

onMounted(() => {
  store.loadUnread();
  store.loadRecent();
});

const openDetail = async (id: string) => {
  // 先关铃铛浮层：popup z-index(5500) 高于 dialog(2500)，不关会压在详情与页面之上
  popupVisible.value = false;
  try {
    detail.value = await getNoticeDetail(id);
    detailVisible.value = true;
    store.applyRead(id);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '加载失败');
  }
};

const onMarkOne = async (id: string) => {
  try {
    await store.markOne(id);
  } catch (e) {
    MessagePlugin.error((e as Error).message || '操作失败');
  }
};

const onMarkAll = async () => {
  marking.value = true;
  try {
    await store.markAll();
    MessagePlugin.success('已全部标为已读');
  } catch (e) {
    MessagePlugin.error((e as Error).message || '操作失败');
  } finally {
    marking.value = false;
  }
};

const goAll = () => {
  router.push('/notice/list');
};
</script>
<style lang="less" scoped>
.header-msg {
  width: 400px;
  margin: calc(0px - var(--td-comp-paddingTB-xs)) calc(0px - var(--td-comp-paddingLR-s));

  .empty-list {
    text-align: center;
    padding: var(--td-comp-paddingTB-xxl) 0;
    font: var(--td-font-body-medium);
    color: var(--td-text-color-secondary);

    img {
      width: var(--td-comp-size-xxxxl);
    }

    p {
      margin-top: var(--td-comp-margin-xs);
    }
  }

  &-top {
    position: relative;
    font: var(--td-font-title-medium);
    color: var(--td-text-color-primary);
    text-align: left;
    padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-xl) 0;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .clear-btn {
      right: calc(var(--td-comp-paddingTB-l) - var(--td-comp-paddingLR-xl));
    }
  }

  &-bottom {
    align-items: center;
    display: flex;
    justify-content: center;
    padding: var(--td-comp-paddingTB-s) var(--td-comp-paddingLR-s);
    border-top: 1px solid var(--td-component-stroke);

    &-link {
      text-decoration: none;
      cursor: pointer;
      color: var(--td-text-color-placeholder);
    }
  }

  .t-list {
    height: calc(100% - 104px);
    padding: var(--td-comp-margin-s) var(--td-comp-margin-s);
  }

  .t-list-item {
    overflow: hidden;
    width: 100%;
    padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
    border-radius: var(--td-radius-default);
    font: var(--td-font-body-medium);
    color: var(--td-text-color-primary);
    cursor: pointer;
    transition: background-color 0.2s linear;

    &:hover {
      background-color: var(--td-bg-color-container-hover);

      .msg-content {
        color: var(--td-brand-color);
      }

      .t-list-item__action {
        button {
          bottom: var(--td-comp-margin-l);
          opacity: 1;
        }
      }

      .msg-time {
        bottom: -6px;
        opacity: 0;
      }
    }

    .msg-content {
      margin-bottom: var(--td-comp-margin-s);
      display: flex;
      align-items: center;
      gap: var(--td-comp-margin-xs);
    }

    .unread-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: var(--td-brand-color);
      flex: none;
    }

    .msg-type {
      color: var(--td-text-color-secondary);
    }

    .t-list-item__action {
      button {
        opacity: 0;
        position: absolute;
        right: var(--td-comp-margin-xxl);
        bottom: -6px;
      }
    }

    .msg-time {
      transition: all 0.2s ease;
      opacity: 1;
      position: absolute;
      right: var(--td-comp-margin-xxl);
      bottom: var(--td-comp-margin-l);
      color: var(--td-text-color-secondary);
    }
  }
}
</style>
