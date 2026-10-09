package log

import (
	log "gosilite/api/model/framework/log"
	"gosilite/application/framework/dal/model"
)

const timeLayout = "2006-01-02 15:04:05"

func toOperDTO(m *model.SystemOperLog) *log.OperLogDTO {
	dto := &log.OperLogDTO{
		ID: m.ID, TraceId: m.TraceID, UserName: m.UserName,
		OperType: int32(m.OperType), RequestUri: m.RequestURI,
		HttpMethod: m.HTTPMethod, Module: m.Module, OperStatus: int32(m.OperStatus),
		Remark: m.Remark, ErrorMsg: m.ErrorMsg, CostMs: m.CostMs, ClientIp: m.ClientIP,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	return dto
}

func toOperDTOs(rows []*model.SystemOperLog) []*log.OperLogDTO {
	out := make([]*log.OperLogDTO, 0, len(rows))
	for _, m := range rows {
		out = append(out, toOperDTO(m))
	}
	return out
}

func toOperDetailDTO(m *model.SystemOperLog) *log.OperLogDetailDTO {
	return &log.OperLogDetailDTO{
		Base:           toOperDTO(m),
		RequestHeaders: m.RequestHeaders,
		RequestBody:    m.RequestBody,
		ResponseBody:   m.ResponseBody,
	}
}

func toErrorDTO(m *model.SystemErrorLog) *log.ErrorLogDTO {
	dto := &log.ErrorLogDTO{
		ID: m.ID, TraceId: m.TraceID, UserName: m.UserName,
		ErrType: int32(m.ErrType), ErrLevel: int32(m.ErrLevel),
		HandleStatus: int32(m.HandleStatus), ErrCode: m.ErrCode, ErrMsg: m.ErrMsg,
		RequestUri: m.RequestURI, HttpMethod: m.HTTPMethod, Module: m.Module,
		CostMs: m.CostMs, ClientIp: m.ClientIP,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	// 历史行可能存了零值时间（0001-01-01），未处理时统一回空
	if !m.HandledAt.IsZero() && m.HandledAt.Year() > 1 {
		dto.HandledAt = m.HandledAt.Format(timeLayout)
	}
	return dto
}

func toErrorDTOs(rows []*model.SystemErrorLog) []*log.ErrorLogDTO {
	out := make([]*log.ErrorLogDTO, 0, len(rows))
	for _, m := range rows {
		out = append(out, toErrorDTO(m))
	}
	return out
}

func toErrorDetailDTO(m *model.SystemErrorLog) *log.ErrorLogDetailDTO {
	return &log.ErrorLogDetailDTO{
		Base:           toErrorDTO(m),
		Stack:          m.Stack,
		RequestHeaders: m.RequestHeaders,
		RequestBody:    m.RequestBody,
		HandleRemark:   m.HandleRemark,
	}
}

func toLoginDTO(m *model.SystemLoginLog) *log.LoginLogDTO {
	dto := &log.LoginLogDTO{
		ID: m.ID, TraceId: m.TraceID, UserName: m.UserName, UserId: m.UserID,
		Status: int32(m.Status), FailReason: m.FailReason, ClientIp: m.ClientIP,
		UserAgent: m.UserAgent, Browser: m.Browser, Os: m.Os,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	return dto
}

func toLoginDTOs(rows []*model.SystemLoginLog) []*log.LoginLogDTO {
	out := make([]*log.LoginLogDTO, 0, len(rows))
	for _, m := range rows {
		out = append(out, toLoginDTO(m))
	}
	return out
}
