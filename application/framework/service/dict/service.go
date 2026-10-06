package dict

import (
	"context"
	"strconv"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// SystemDictTypeSVC / SystemDictDataSVC 门面
var (
	SystemDictTypeSVC *TypeService
	SystemDictDataSVC *DataService
)

// TypeService 字典类型
type TypeService struct {
	repo repository.IDictRepository
}

// DataService 字典数据
type DataService struct {
	repo repository.IDictRepository
}

// NewTypeService 构造
func NewTypeService(repo repository.IDictRepository) *TypeService {
	return &TypeService{repo: repo}
}

// NewDataService 构造
func NewDataService(repo repository.IDictRepository) *DataService {
	return &DataService{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.IDictRepository) {
	SystemDictTypeSVC = NewTypeService(repo)
	SystemDictDataSVC = NewDataService(repo)
}

// PageType 分页
func (s *TypeService) PageType(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemDictType, int64, error) {
	return s.repo.PageTypes(ctx, page, pageSize, keyword)
}

// CreateType 创建（dict_type 唯一）
func (s *TypeService) CreateType(ctx context.Context, d *model.SystemDictType) error {
	if strings.TrimSpace(d.DictType) == "" || strings.TrimSpace(d.DictName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "字典名称/类型不能为空"))
	}
	exists, err := s.repo.ExistsType(ctx, d.DictType, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemDictTypeExists, errorx.KV("type", d.DictType))
	}
	d.DelFlag = 0
	return s.repo.CreateType(ctx, d)
}

// UpdateType 更新
func (s *TypeService) UpdateType(ctx context.Context, d *model.SystemDictType) error {
	if _, err := s.repo.GetTypeByID(ctx, d.ID); err != nil {
		return errorx.New(errno.ErrSystemDictNotFound, errorx.KV("id", formatID(d.ID)))
	}
	exists, err := s.repo.ExistsType(ctx, d.DictType, d.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemDictTypeExists, errorx.KV("type", d.DictType))
	}
	return s.repo.SaveType(ctx, d)
}

// DeleteType 软删，有数据项则拒绝
func (s *TypeService) DeleteType(ctx context.Context, id int64) error {
	old, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		return errorx.New(errno.ErrSystemDictNotFound, errorx.KV("id", formatID(id)))
	}
	n, err := s.repo.CountDataByType(ctx, old.DictType)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(errno.ErrSystemDictHasData, errorx.KV("type", old.DictType))
	}
	return s.repo.SoftDeleteType(ctx, id)
}

// ListData 按字典类型列数据
func (s *DataService) ListData(ctx context.Context, dictType string, page, pageSize int) ([]*model.SystemDictData, int64, error) {
	return s.repo.PageData(ctx, dictType, page, pageSize)
}

// Options 下拉
func (s *DataService) Options(ctx context.Context, dictType string) ([]*model.SystemDictData, error) {
	return s.repo.Options(ctx, dictType)
}

// CreateData 新建
func (s *DataService) CreateData(ctx context.Context, d *model.SystemDictData) error {
	if strings.TrimSpace(d.DictLabel) == "" || strings.TrimSpace(d.DictValue) == "" || strings.TrimSpace(d.DictType) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "标签/值/类型不能为空"))
	}
	d.DelFlag = 0
	return s.repo.CreateData(ctx, d)
}

// UpdateData 更新
func (s *DataService) UpdateData(ctx context.Context, d *model.SystemDictData) error {
	if _, err := s.repo.GetDataByID(ctx, d.ID); err != nil {
		return errorx.New(errno.ErrSystemDictNotFound, errorx.KV("id", formatID(d.ID)))
	}
	return s.repo.SaveData(ctx, d)
}

// DeleteData 软删
func (s *DataService) DeleteData(ctx context.Context, id int64) error {
	return s.repo.SoftDeleteData(ctx, id)
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
