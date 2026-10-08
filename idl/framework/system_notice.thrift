namespace go framework.notice

struct NoticeListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
    // 缺省=全部；1=通知 2=公告
    4: optional i32 noticeType
    // 缺省=全部；0=已发布 1=已撤回
    5: optional i32 status
}

struct NoticeIdRequest {
    1: i64 id
}

struct NoticeDTO {
    1: i64 id
    2: string title
    // sys_notice_type: 1 通知 2 公告
    3: i32 noticeType
    4: string content
    // sys_notice_status: 0 已发布 1 已撤回
    5: i32 status
    6: string publisher
    7: string createdAt
    8: string updatedAt
}

struct NoticeListResponse {
    1: list<NoticeDTO> list
    2: i64 total
}

struct NoticeSaveRequest {
    1: optional i64 id
    2: string title
    3: i32 noticeType
    4: string content
    5: i32 status
}

// ===== 消费侧（通知中心，login-only） =====

struct NoticeInboxRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
}

struct NoticeInboxItemDTO {
    1: i64 id
    2: string title
    3: i32 noticeType
    4: string publisher
    5: string createdAt
    6: bool read
}

struct NoticeInboxListResponse {
    1: list<NoticeInboxItemDTO> list
    2: i64 total
    // 未读数（已发布 − 本人已读），供列表页角标
    3: i64 unread
}

struct NoticeDetailRequest {
    1: i64 id
}

struct NoticeDetailDTO {
    1: i64 id
    2: string title
    3: i32 noticeType
    4: string content
    5: string publisher
    6: string createdAt
    // 打开详情即已读，返回恒为 true
    7: bool read
}

struct UnreadCountResponse {
    1: i64 count
}

// id 缺省 = 全部已读
struct NoticeReadRequest {
    1: optional i64 id
}

service SystemNoticeController {
    NoticeListResponse List(1: NoticeListRequest req)(api.get = "/system/notice/list")
    NoticeDTO Get(1: NoticeIdRequest req)(api.get = "/system/notice")
    NoticeDTO Create(1: NoticeSaveRequest req)(api.post = "/system/notice")
    NoticeDTO Update(1: NoticeSaveRequest req)(api.put = "/system/notice")
    i32 Remove(1: NoticeIdRequest req)(api.delete = "/system/notice")

    // 通知中心消费侧（login-only：mw 保持 nil，登录由 rootMw 提供）
    NoticeInboxListResponse Inbox(1: NoticeInboxRequest req)(api.get = "/notice/inbox")
    NoticeDetailDTO Detail(1: NoticeDetailRequest req)(api.get = "/notice/detail")
    UnreadCountResponse UnreadCount()(api.get = "/notice/unread-count")
    i32 Read(1: NoticeReadRequest req)(api.post = "/notice/read")
}
