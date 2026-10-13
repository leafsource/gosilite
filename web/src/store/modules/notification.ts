import { defineStore } from 'pinia';

import { getNoticeUnreadCount, listNoticeInbox, markNoticeRead } from '@/api/notice';
import type { NoticeInboxItem } from '@/api/noticeModel';

/**
 * 通知中心 store（真实数据，已去 mock）：
 * - unread：Header 角标未读数（已发布 − 本人已读）
 * - recent：铃铛下拉最近 10 条（不含 content，正文走 getNoticeDetail）
 * 失败不写入，保留现值，下次触点可重试。
 */
export const useNotificationStore = defineStore('notification', {
  state: () => ({
    unread: 0,
    recent: [] as NoticeInboxItem[],
    loading: false,
  }),
  actions: {
    async loadUnread() {
      try {
        const res = await getNoticeUnreadCount();
        this.unread = res.count ?? 0;
      } catch {
        // 保留现值
      }
    },
    async loadRecent() {
      if (this.loading) return;
      this.loading = true;
      try {
        const res = await listNoticeInbox({ page: 1, pageSize: 10 });
        this.recent = res.list || [];
        this.unread = res.unread ?? this.unread;
      } catch {
        // 保留旧数据
      } finally {
        this.loading = false;
      }
    },
    /** 标记单条已读（下拉内「设为已读」） */
    async markOne(id: string) {
      await markNoticeRead(id);
      this.applyRead(id);
    },
    /** 全部已读 */
    async markAll() {
      await markNoticeRead();
      this.recent.forEach((item) => {
        item.read = true;
      });
      this.unread = 0;
    },
    /** 详情打开后服务端已自动已读，同步本地角标与列表状态 */
    applyRead(id: string) {
      const item = this.recent.find((i) => i.id === id);
      if (item && !item.read) {
        item.read = true;
        this.unread = Math.max(0, this.unread - 1);
      }
    },
  },
  persist: false,
});
