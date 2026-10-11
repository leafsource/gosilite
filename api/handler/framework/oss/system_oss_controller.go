package oss

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	oss "gosilite/api/model/framework/oss"
	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	ossfile "gosilite/application/framework/service/ossfile"
	R "gosilite/pkg/utils/response"
)

// 路由与请求/响应 DTO 均由 hz 从 idl/framework/system_oss.thrift 生成
// （api/model/framework/oss + api/router/framework/oss），
// 本文件只保留 Handler 实现（绑定 → 调 Service → 回包）。
// spec: docs/compose/spec/system-oss.md S2.5

const timeLayout = "2006-01-02 15:04:05"

// List 台账分页列表
// @Summary 文件管理列表
// @Description OSS文件台账分页列表接口
// @Tags 系统管理
// @Produce json
// @Param data query oss.OssListRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	var req oss.OssListRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	f := repository.OssFileFilter{
		Page:         int(req.Page),
		PageSize:     int(req.PageSize),
		FileName:     req.FileName,
		OriginalName: req.OriginalName,
		CreateBy:     req.CreateBy,
		Service:      req.Service,
		BeginTime:    req.BeginTime,
		EndTime:      req.EndTime,
	}
	if req.FileSuffix != "" {
		// 小写归一（源码 queryPageList 语义）
		for _, s := range splitTrim(req.FileSuffix) {
			f.FileSuffix = append(f.FileSuffix, toLower(s))
		}
	}
	list, total, err := ossfile.SystemOssSVC.Page(ctx, f)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &oss.OssListResponse{List: toDTOs(list), Total: total})
}

// Get 台账详情
// @Summary 文件详情
// @Description OSS文件台账详情接口
// @Tags 系统管理
// @Produce json
// @Param data query oss.OssIdRequest false "查询参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	var req oss.OssIdRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	vo, err := ossfile.SystemOssSVC.Get(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(vo))
}

// Update 编辑（originalName / isLock）
// @Summary 文件修改
// @Description OSS文件台账修改接口（原名/锁定）
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body oss.OssSaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	var req oss.OssSaveRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	m := &model.SystemOss{ID: req.ID, OriginalName: req.OriginalName, IsLock: req.IsLock}
	if err := ossfile.SystemOssSVC.Update(ctx, m); err != nil {
		R.Error(ctx, c, err)
		return
	}
	vo, err := ossfile.SystemOssSVC.Get(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, toDTO(vo))
}

// Remove 批量删除台账
// @Summary 文件删除
// @Description OSS文件台账删除接口（锁定文件拒绝）
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param data body oss.OssIdsRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Router /system/oss [DELETE]
func Remove(ctx context.Context, c *app.RequestContext) {
	var req oss.OssIdsRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	if err := ossfile.SystemOssSVC.Delete(ctx, req.Ids); err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, int32(len(req.Ids)))
}

// --- 转换/辅助 ---

func toDTO(vo *ossfile.FileVO) *oss.OssDTO {
	m := vo.SystemOss
	dto := &oss.OssDTO{
		ID:           m.ID,
		FileName:     m.FileName,
		OriginalName: m.OriginalName,
		FileSuffix:   m.FileSuffix,
		URL:          vo.URL,
		Size:         m.Size,
		ContentType:  m.ContentType,
		IsLock:       m.IsLock,
		Service:      m.Service,
		CreateBy:     m.CreateBy,
		RuleUrls:     vo.RuleUrls,
	}
	if !m.CreatedAt.IsZero() {
		dto.CreatedAt = m.CreatedAt.Format(timeLayout)
	}
	if !m.UpdatedAt.IsZero() {
		dto.UpdatedAt = m.UpdatedAt.Format(timeLayout)
	}
	return dto
}

func toDTOs(list []*ossfile.FileVO) []*oss.OssDTO {
	out := make([]*oss.OssDTO, 0, len(list))
	for _, vo := range list {
		out = append(out, toDTO(vo))
	}
	return out
}

func splitTrim(s string) []string {
	var out []string
	for _, p := range splitComma(s) {
		if t := trimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// 以下小工具避免为几个字符串操作引包（保持 handler 轻依赖）
func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func toLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func escapeQuotes(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			out = append(out, '\\')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func urlEncode(s string) string {
	// 简化实现：空格→%20，非安全字符按 UTF-8 字节百分号编码
	const hex = "0123456789ABCDEF"
	var out []byte
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.' || ch == '~' {
			out = append(out, ch)
			continue
		}
		out = append(out, '%', hex[ch>>4], hex[ch&0xf])
	}
	return string(out)
}

var _ = time.Now
