// 通知公告类型
export interface NoticeDTO {
  id: string;
  title: string;
  noticeType: number;
  content: string;
  status: number;
  publisher: string;
  createdAt: string;
  updatedAt: string;
}

export interface NoticeSaveParams {
  id?: string;
  title: string;
  noticeType: number;
  content: string;
  status: number;
}

export interface NoticeListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  noticeType?: number;
  status?: number;
}
