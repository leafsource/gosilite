// 通知中心（用户消费侧）类型
export interface NoticeInboxItem {
  id: string;
  title: string;
  noticeType: number;
  publisher: string;
  createdAt: string;
  read: boolean;
}

export interface NoticeInboxResult {
  list: NoticeInboxItem[];
  total: number;
  unread: number;
}

export interface NoticeDetail {
  id: string;
  title: string;
  noticeType: number;
  content: string;
  publisher: string;
  createdAt: string;
  /** 打开详情即已读，服务端返回恒为 true */
  read: boolean;
}

export interface UnreadCountResult {
  count: number;
}
