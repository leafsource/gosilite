package notice

import (
	notice "gosilite/api/model/framework/notice"
	"gosilite/application/framework/dal/model"
	noticeSvc "gosilite/application/framework/service/notice"
)

const timeLayout = "2006-01-02 15:04:05"

func toNoticeDTO(m *model.SystemNotice) *notice.NoticeDTO {
	dto := &notice.NoticeDTO{
		ID:         m.ID,
		Title:      m.Title,
		NoticeType: int32(m.NoticeType),
		Content:    m.Content,
		Status:     int32(m.Status),
		Publisher:  m.Publisher,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toNoticeDTOs(list []*model.SystemNotice) []*notice.NoticeDTO {
	out := make([]*notice.NoticeDTO, 0, len(list))
	for _, m := range list {
		out = append(out, toNoticeDTO(m))
	}
	return out
}

func fromNoticeSaveRequest(req *notice.NoticeSaveRequest) *model.SystemNotice {
	return &model.SystemNotice{
		Title:      req.Title,
		NoticeType: int16(req.NoticeType),
		Content:    req.Content,
		Status:     int16(req.Status),
	}
}

// ========== 消费侧（通知中心） ==========

func toInboxItemDTOs(items []noticeSvc.InboxItem) []*notice.NoticeInboxItemDTO {
	out := make([]*notice.NoticeInboxItemDTO, 0, len(items))
	for _, it := range items {
		m := it.Notice
		dto := &notice.NoticeInboxItemDTO{
			ID: m.ID, Title: m.Title, NoticeType: int32(m.NoticeType),
			Publisher: m.Publisher, Read_: it.Read,
		}
		if !m.CreatedAt.IsZero() {
			dto.CreatedAt = m.CreatedAt.Format(timeLayout)
		}
		out = append(out, dto)
	}
	return out
}

func toNoticeDetailDTO(m *model.SystemNotice) *notice.NoticeDetailDTO {
	dto := &notice.NoticeDetailDTO{
		ID: m.ID, Title: m.Title, NoticeType: int32(m.NoticeType),
		Content: m.Content, Publisher: m.Publisher,
		// Detail 在服务端已 upsert 已读，返回恒为 true
		Read_: true,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	return dto
}
