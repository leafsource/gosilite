package ossconfig

import (
	"context"
	"errors"
	"testing"

	"gosilite/application/framework/dal/model"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"
)

type fakeRepo struct {
	rows    map[int64]*model.SystemOssConfig
	keys    map[string]int64
	created *model.SystemOssConfig

	saveErr    error
	changedID  int64
	deletedIDs []int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{rows: map[int64]*model.SystemOssConfig{}, keys: map[string]int64{}}
}

func (f *fakeRepo) Page(context.Context, int, int, string) ([]*model.SystemOssConfig, int64, error) {
	return nil, 0, nil
}
func (f *fakeRepo) GetByID(_ context.Context, id int64) (*model.SystemOssConfig, error) {
	if m, ok := f.rows[id]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeRepo) ExistsKey(_ context.Context, key string, excludeID int64) (bool, error) {
	if id, ok := f.keys[key]; ok && id != excludeID {
		return true, nil
	}
	return false, nil
}
func (f *fakeRepo) Create(_ context.Context, m *model.SystemOssConfig) error {
	f.created = m
	m.ID = int64(len(f.rows) + 1)
	f.rows[m.ID] = m
	f.keys[m.ConfigKey] = m.ID
	return nil
}
func (f *fakeRepo) SaveEditable(_ context.Context, m *model.SystemOssConfig) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	row := f.rows[m.ID]
	if row == nil {
		return errors.New("not found")
	}
	row.AccessKey = m.AccessKey
	row.BucketName = m.BucketName
	row.Endpoint = m.Endpoint
	return nil
}
func (f *fakeRepo) SoftDelete(_ context.Context, ids []int64) error {
	f.deletedIDs = ids
	for _, id := range ids {
		delete(f.rows, id)
	}
	return nil
}
func (f *fakeRepo) ClearDefault(context.Context) error { return nil }
func (f *fakeRepo) SetDefault(context.Context, int64) error {
	return nil
}
func (f *fakeRepo) ChangeDefaultTx(_ context.Context, id int64) error {
	if _, ok := f.rows[id]; !ok {
		return errors.New("not found")
	}
	for _, m := range f.rows {
		m.Status = "0"
	}
	f.rows[id].Status = "1"
	f.changedID = id
	return nil
}

func errCode(err error) int32 {
	var se errorx.StatusError
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func TestCreateRejectsDuplicateKey(t *testing.T) {
	repo := newFakeRepo()
	repo.keys["local"] = 1
	svc := NewService(repo, nil)

	err := svc.Create(context.Background(), &model.SystemOssConfig{ConfigKey: "local", BucketName: "b", Endpoint: "e"})
	if errCode(err) != int32(errno.ErrOssConfigKeyExists) {
		t.Fatalf("expect key exists, got %v", err)
	}
}

func TestCreateForcesNonDefaultStatus(t *testing.T) {
	repo := newFakeRepo()
	invalidated := false
	svc := NewService(repo, func() { invalidated = true })

	m := &model.SystemOssConfig{ConfigKey: "k2", BucketName: "b", Endpoint: "e", Status: "1"}
	if err := svc.Create(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if repo.created.Status != "0" {
		t.Fatalf("new config must not be default, got %q", repo.created.Status)
	}
	if !invalidated {
		t.Fatal("invalidate not fired")
	}
}

func TestCreateValidatesRequired(t *testing.T) {
	svc := NewService(newFakeRepo(), nil)
	for _, m := range []*model.SystemOssConfig{
		{ConfigKey: "", BucketName: "b", Endpoint: "e"},
		{ConfigKey: "k", BucketName: "", Endpoint: "e"},
		{ConfigKey: "k", BucketName: "b", Endpoint: ""},
	} {
		if err := svc.Create(context.Background(), m); errCode(err) != int32(errno.ErrOssConfigInvalid) {
			t.Fatalf("expect invalid for %+v, got %v", m, err)
		}
	}
}

func TestUpdateKeepsConfigKey(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[1] = &model.SystemOssConfig{ID: 1, ConfigKey: "original", Status: "1", BucketName: "b", Endpoint: "e"}
	repo.keys["original"] = 1
	svc := NewService(repo, nil)

	m := &model.SystemOssConfig{ID: 1, ConfigKey: "hacked", BucketName: "b2", Endpoint: "e2", AccessKey: "ak"}
	if err := svc.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if repo.rows[1].ConfigKey != "original" {
		t.Fatalf("config_key must not change, got %q", repo.rows[1].ConfigKey)
	}
	if repo.rows[1].BucketName != "b2" {
		t.Fatalf("editable fields should update")
	}
}

func TestDeleteRejectsDefault(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[1] = &model.SystemOssConfig{ID: 1, ConfigKey: "d", Status: "1"}
	svc := NewService(repo, nil)

	err := svc.Delete(context.Background(), []int64{1})
	if errCode(err) != int32(errno.ErrOssConfigIsDefault) {
		t.Fatalf("expect default reject, got %v", err)
	}
	if len(repo.deletedIDs) != 0 {
		t.Fatal("default must not be deleted")
	}
}

func TestDeleteNonDefault(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[2] = &model.SystemOssConfig{ID: 2, ConfigKey: "x", Status: "0"}
	invalidated := false
	svc := NewService(repo, func() { invalidated = true })

	if err := svc.Delete(context.Background(), []int64{2}); err != nil {
		t.Fatal(err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 2 {
		t.Fatalf("delete not applied: %v", repo.deletedIDs)
	}
	if !invalidated {
		t.Fatal("invalidate not fired")
	}
}

func TestChangeStatusSwitchesDefault(t *testing.T) {
	repo := newFakeRepo()
	repo.rows[1] = &model.SystemOssConfig{ID: 1, ConfigKey: "a", Status: "1"}
	repo.rows[2] = &model.SystemOssConfig{ID: 2, ConfigKey: "b", Status: "0"}
	invalidated := false
	svc := NewService(repo, func() { invalidated = true })

	m, err := svc.ChangeStatus(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != "1" || repo.rows[1].Status != "0" {
		t.Fatalf("default not switched: a=%q b=%q", repo.rows[1].Status, repo.rows[2].Status)
	}
	if repo.changedID != 2 || !invalidated {
		t.Fatalf("changedID=%d invalidated=%v", repo.changedID, invalidated)
	}
}

func TestChangeStatusNotFound(t *testing.T) {
	svc := NewService(newFakeRepo(), nil)
	if _, err := svc.ChangeStatus(context.Background(), 99); errCode(err) != int32(errno.ErrOssConfigNotFound) {
		t.Fatalf("expect not found, got %v", err)
	}
}
