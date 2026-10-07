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

// SystemDictTypeSVC / SystemDictDataSVC / SystemDictGroupSVC 门面
var (
	SystemDictTypeSVC  *TypeService
	SystemDictDataSVC  *DataService
	SystemDictGroupSVC *GroupService
)

// TypeService 字典类型
type TypeService struct {
	repo repository.IDictRepository
}

// DataService 字典数据
type DataService struct {
	repo repository.IDictRepository
}

// GroupService 字典分组
type GroupService struct {
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

// NewGroupService 构造
func NewGroupService(repo repository.IDictRepository) *GroupService {
	return &GroupService{repo: repo}
}

// InitService 安装门面
func InitService(repo repository.IDictRepository) {
	SystemDictTypeSVC = NewTypeService(repo)
	SystemDictDataSVC = NewDataService(repo)
	SystemDictGroupSVC = NewGroupService(repo)
}

// PageType 分页；groupID：nil=全部，0=未分组，>0=指定分组
func (s *TypeService) PageType(ctx context.Context, page, pageSize int, keyword string, groupID *int64) ([]*model.SystemDictType, int64, error) {
	return s.repo.PageTypes(ctx, page, pageSize, keyword, groupID)
}

// CreateType 创建（dict_type 全局唯一）
func (s *TypeService) CreateType(ctx context.Context, d *model.SystemDictType) error {
	if strings.TrimSpace(d.DictType) == "" || strings.TrimSpace(d.DictName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "字典名称/类型不能为空"))
	}
	if err := s.checkGroup(ctx, d.GroupID); err != nil {
		return err
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
	if err := s.checkGroup(ctx, d.GroupID); err != nil {
		return err
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

// checkGroup groupId>0 时分组必须存在（分组删除时已有字典类型则拒绝，保证不变量）
func (s *TypeService) checkGroup(ctx context.Context, groupID int64) error {
	if groupID <= 0 {
		return nil
	}
	if _, err := s.repo.GetGroupByID(ctx, groupID); err != nil {
		return errorx.New(errno.ErrSystemDictGroupNotFound, errorx.KV("id", formatID(groupID)))
	}
	return nil
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

// List 分组列表（不分页，按 order_num 排序）
func (s *GroupService) List(ctx context.Context) ([]*model.SystemDictGroup, error) {
	return s.repo.ListGroups(ctx)
}

// Create 创建分组（名称必填且唯一）
func (s *GroupService) Create(ctx context.Context, g *model.SystemDictGroup) error {
	if strings.TrimSpace(g.GroupName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "分组名称不能为空"))
	}
	exists, err := s.repo.ExistsGroupName(ctx, g.GroupName, 0)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemDictGroupExists, errorx.KV("name", g.GroupName))
	}
	g.DelFlag = 0
	return s.repo.CreateGroup(ctx, g)
}

// Update 更新分组
func (s *GroupService) Update(ctx context.Context, g *model.SystemDictGroup) error {
	if strings.TrimSpace(g.GroupName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "分组名称不能为空"))
	}
	if _, err := s.repo.GetGroupByID(ctx, g.ID); err != nil {
		return errorx.New(errno.ErrSystemDictGroupNotFound, errorx.KV("id", formatID(g.ID)))
	}
	exists, err := s.repo.ExistsGroupName(ctx, g.GroupName, g.ID)
	if err != nil {
		return err
	}
	if exists {
		return errorx.New(errno.ErrSystemDictGroupExists, errorx.KV("name", g.GroupName))
	}
	return s.repo.SaveGroup(ctx, g)
}

// Delete 软删，组内有字典类型则拒绝
func (s *GroupService) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.GetGroupByID(ctx, id); err != nil {
		return errorx.New(errno.ErrSystemDictGroupNotFound, errorx.KV("id", formatID(id)))
	}
	n, err := s.repo.CountTypesByGroup(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(errno.ErrSystemDictGroupHasTypes, errorx.KV("id", formatID(id)))
	}
	return s.repo.SoftDeleteGroup(ctx, id)
}

func formatID(v int64) string {
	return strconv.FormatInt(v, 10)
}
