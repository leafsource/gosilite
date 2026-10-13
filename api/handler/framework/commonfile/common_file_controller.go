package commonfile

import (
	"context"
	"io"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	commonfile "gosilite/api/model/framework/commonfile"
	ossfile "gosilite/application/framework/service/ossfile"
	R "gosilite/pkg/utils/response"
)

// 统一文件服务（spec system-oss S2.11）：业务侧不再自建 multipart handler，
// 上传/下载全部收口至 /common/file/*，台账由 ossfile.Service.Upload 自动记录。
// 路由与 DTO 由 hz 从 idl/framework/common_file.thrift 生成。

// Upload 统一上传（multipart 字段 file，≤20MB，登录即可）
// @Summary 统一文件上传
// @Description 统一文件上传接口（写桶+自动记录文件台账，登录即可）
// @Tags 系统管理
// @Produce json
// @Param file formData file true "上传文件"
// @Success 200 {object} map[string]interface{}
// @Router /common/file/upload [POST]
func Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		hlog.CtxErrorf(ctx, "common file: get upload file failed, err=%v", err)
		R.InvalidParamRequestResponse(c, "missing file field")
		return
	}
	src, err := file.Open()
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	defer src.Close()

	// 多读 1 字节：超限交给 Service 校验（≤20MB），内存有界
	content, err := io.ReadAll(io.LimitReader(src, int64(ossfile.MaxUploadBytes)+1))
	if err != nil {
		R.Error(ctx, c, err)
		return
	}

	vo, err := ossfile.SystemOssSVC.Upload(ctx, file.Filename, file.Header.Get("Content-Type"), content)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	R.SuccessWithData(ctx, c, &commonfile.CommonFileDTO{
		ID:           vo.ID,
		URL:          vo.URL,
		Key:          vo.FileName,
		OriginalName: vo.OriginalName,
		Size:         vo.Size,
		ContentType:  vo.ContentType,
	})
}

// Download 统一下载（按台账 id 回流附件，非 JSON 信封）
// @Summary 统一文件下载
// @Description 统一文件下载接口（按台账 id 流式下载）
// @Tags 系统管理
// @Produce json
// @Param data query commonfile.CommonFileDownloadRequest false "查询参数"
// @Success 200 {file} file
// @Router /common/file/download [GET]
func Download(ctx context.Context, c *app.RequestContext) {
	var req commonfile.CommonFileDownloadRequest
	if err := c.BindAndValidate(&req); err != nil {
		R.InvalidParamRequestResponse(c, err.Error())
		return
	}
	originalName, contentType, content, err := ossfile.SystemOssSVC.Download(ctx, req.ID)
	if err != nil {
		R.Error(ctx, c, err)
		return
	}
	disposition := `attachment; filename="download"`
	if isASCII(originalName) {
		disposition = `attachment; filename="` + escapeQuotes(originalName) + `"`
	} else {
		disposition = `attachment; filename*=UTF-8''` + urlEncode(originalName)
	}
	c.Response.Header.Set("Content-Disposition", disposition)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.SetContentType(contentType + "; charset=UTF-8")
	c.SetStatusCode(200)
	c.Response.SetBody(content)
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
