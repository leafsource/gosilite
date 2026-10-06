package dict

import (
	"context"
	"errors"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

// fakeDictRepo 仅实现唯一性与校验所需方法
type fakeDictRepo struct {
	types       map[string]*model.SystemDictType // by dict_type
	dataByType  map[string]int64
	saveErr     error
	createCalls int
}

func newFakeDictRepo() *fakeDictRepo {
	return &fakeDictRepo{
		types:      map[string]*model.SystemDictType{},
		dataByType: map[string]int64{},
	}
}

func (f *fakeDictRepo) PageTypes(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemDictType, int64, error) {
	return nil, 0, nil
}

func (f *fakeDictRepo) GetTypeByID(ctx context.Context, id int64) (*model.SystemDictType, error) {
	for _, t := range f.types {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeDictRepo) ExistsType(ctx context.Context, dictType string, excludeID int64) (bool, error) {
	t, ok := f.types[dictType]
	return ok && t.ID != excludeID, nil
}

func (f *fakeDictRepo) CountDataByType(ctx context.Context, dictType string) (int64, error) {
	return f.dataByType[dictType], nil
}

func (f *fakeDictRepo) CreateType(ctx context.Context, m *model.SystemDictType) error {
	f.createCalls++
	f.types[m.DictType] = m
	return nil
}

func (f *fakeDictRepo) SaveType(ctx context.Context, m *model.SystemDictType) error { return nil }
func (f *fakeDictRepo) SoftDeleteType(ctx context.Context, id int64) error         { return nil }

func (f *fakeDictRepo) PageData(ctx context.Context, dictType string, page, pageSize int) ([]*model.SystemDictData, int64, error) {
	return nil, 0, nil
}
func (f *fakeDictRepo) Options(ctx context.Context, dictType string) ([]*model.SystemDictData, error) {
	return nil, nil
}
func (f *fakeDictRepo) GetDataByID(ctx context.Context, id int64) (*model.SystemDictData, error) {
	return nil, errors.New("not found")
}
func (f *fakeDictRepo) CreateData(ctx context.Context, m *model.SystemDictData) error { return nil }
func (f *fakeDictRepo) SaveData(ctx context.Context, m *model.SystemDictData) error   { return nil }
func (f *fakeDictRepo) SoftDeleteData(ctx context.Context, id int64) error            { return nil }

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func TestCreateTypeRejectsDuplicate(t *testing.T) {
	repo := newFakeDictRepo()
	repo.types["sys_dup"] = &model.SystemDictType{ID: 1, DictType: "sys_dup"}
	svc := NewTypeService(repo)

	err := svc.CreateType(context.Background(), &model.SystemDictType{DictName: "x", DictType: "sys_dup"})
	if errCode(err) != errno.ErrSystemDictTypeExists {
		t.Fatalf("want ErrSystemDictTypeExists, got %v", err)
	}
}

func TestCreateTypeRequiresFields(t *testing.T) {
	svc := NewTypeService(newFakeDictRepo())
	err := svc.CreateType(context.Background(), &model.SystemDictType{DictName: "", DictType: ""})
	if errCode(err) != errno.ErrSystemInvalidParam {
		t.Fatalf("want invalid param, got %v", err)
	}
}

func TestCreateTypeOK(t *testing.T) {
	repo := newFakeDictRepo()
	svc := NewTypeService(repo)
	err := svc.CreateType(context.Background(), &model.SystemDictType{DictName: "ok", DictType: "sys_ok"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("createCalls=%d", repo.createCalls)
	}
}

func TestDeleteTypeRejectsWithData(t *testing.T) {
	repo := newFakeDictRepo()
	repo.types["sys_full"] = &model.SystemDictType{ID: 2, DictType: "sys_full"}
	repo.dataByType["sys_full"] = 3
	svc := NewTypeService(repo)

	err := svc.DeleteType(context.Background(), 2)
	if errCode(err) != errno.ErrSystemDictHasData {
		t.Fatalf("want has data, got %v", err)
	}
}
