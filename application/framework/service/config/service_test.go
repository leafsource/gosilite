package config

import (
	"context"
	"errors"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

type fakeConfigRepo struct {
	keys map[string]int64
}

func (f *fakeConfigRepo) Page(ctx context.Context, page, pageSize int, keyword string) ([]*model.SystemConfig, int64, error) {
	return nil, 0, nil
}
func (f *fakeConfigRepo) GetByID(ctx context.Context, id int64) (*model.SystemConfig, error) {
	return &model.SystemConfig{ID: id, ConfigKey: "k"}, nil
}
func (f *fakeConfigRepo) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	if id, ok := f.keys[key]; ok {
		return &model.SystemConfig{ID: id, ConfigKey: key}, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeConfigRepo) ListByType(ctx context.Context, configType string) ([]*model.SystemConfig, error) {
	if configType == "global" {
		return []*model.SystemConfig{{ID: 1, ConfigKey: "sys.app.name", ConfigValue: "gosilite", ConfigType: "global"}}, nil
	}
	return nil, nil
}
func (f *fakeConfigRepo) ExistsKey(ctx context.Context, key string, excludeID int64) (bool, error) {
	id, ok := f.keys[key]
	return ok && id != excludeID, nil
}
func (f *fakeConfigRepo) Create(ctx context.Context, m *model.SystemConfig) error {
	f.keys[m.ConfigKey] = m.ID
	return nil
}
func (f *fakeConfigRepo) Save(ctx context.Context, m *model.SystemConfig) error   { return nil }
func (f *fakeConfigRepo) SoftDelete(ctx context.Context, id int64) error          { return nil }

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func TestCreateConfigRejectsDuplicateKey(t *testing.T) {
	repo := &fakeConfigRepo{keys: map[string]int64{"sys.page.size": 1}}
	svc := NewService(repo)
	err := svc.Create(context.Background(), &model.SystemConfig{
		ConfigName: "x", ConfigKey: "sys.page.size", ConfigValue: "1",
	})
	if errCode(err) != errno.ErrSystemConfigKeyExists {
		t.Fatalf("want key exists, got %v", err)
	}
}

func TestCreateConfigRequiresKey(t *testing.T) {
	svc := NewService(&fakeConfigRepo{keys: map[string]int64{}})
	err := svc.Create(context.Background(), &model.SystemConfig{ConfigName: "x", ConfigKey: ""})
	if errCode(err) != errno.ErrSystemInvalidParam {
		t.Fatalf("want invalid, got %v", err)
	}
}

func TestCreateConfigOK(t *testing.T) {
	repo := &fakeConfigRepo{keys: map[string]int64{}}
	svc := NewService(repo)
	err := svc.Create(context.Background(), &model.SystemConfig{
		ConfigName: "n", ConfigKey: "sys.new", ConfigValue: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListGlobalReturnsGlobalType(t *testing.T) {
	svc := NewService(&fakeConfigRepo{keys: map[string]int64{}})
	list, err := svc.ListGlobal(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(list) != 1 || list[0].ConfigKey != "sys.app.name" || list[0].ConfigType != "global" {
		t.Fatalf("want global list with sys.app.name, got %+v", list)
	}
}
