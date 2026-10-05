// pkg/gorm/plugin/hooks_simple.go
package hook

import (
	"context"
	"fmt"
	"gosilite/application/base/ctxutil"
	"gosilite/pkg/genid"
	"reflect"
	"time"

	"gorm.io/gorm"
)

// 全局Hook插件
type BaseModelPlugin struct{}

func (p *BaseModelPlugin) Name() string {
	return "base_model_plugin"
}

func (p *BaseModelPlugin) Initialize(db *gorm.DB) error {
	_ = db.Callback().Create().Before("gorm:create").Register("before_create", beforeCreate)
	_ = db.Callback().Update().Before("gorm:update").Register("before_update", beforeUpdate)
	_ = db.Callback().Delete().Before("gorm:delete").Register("before_delete", beforeDelete)
	return nil
}

// 创建前的Hook
func beforeCreate(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	processValues(db.Statement.ReflectValue, func(value reflect.Value) {
		setCreateFields(value, db.Statement.Context)
	})
}

// 更新前的Hook
func beforeUpdate(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	processValues(db.Statement.ReflectValue, func(value reflect.Value) {
		setUpdateFields(value, db.Statement.Context)
	})
}

// 删除前的Hook
func beforeDelete(db *gorm.DB) {
	if db.Statement.Schema == nil || db.Statement.Unscoped {
		return
	}

	processValues(db.Statement.ReflectValue, func(value reflect.Value) {
		setDeleteFields(value, db.Statement.Context)
	})
}

// 处理值的通用函数
func processValues(reflectValue reflect.Value, process func(reflect.Value)) {
	val := reflect.Indirect(reflectValue)
	if !val.IsValid() {
		return
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		// 批量操作
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			processElem(elem, process)
		}
	case reflect.Struct:
		// 单个操作
		process(val)
	case reflect.Ptr:
		// 指针类型
		elem := val.Elem()
		if elem.IsValid() && elem.Kind() == reflect.Struct {
			process(elem)
		}
	}
}

// 处理元素
func processElem(elem reflect.Value, process func(reflect.Value)) {
	if elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			return
		}
		elem = elem.Elem()
	}

	if elem.IsValid() && elem.Kind() == reflect.Struct {
		process(elem)
	}
}

// 设置创建字段
func setCreateFields(value reflect.Value, ctx context.Context) {
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}

	now := time.Now()
	var userID int64
	if ctx != nil {
		if user := ctxutil.GetLoginUser(ctx); user != nil {
			userID = user.ID
		}
	}

	// 使用更安全的方式设置字段
	trySetField(value, "ID", func(field reflect.Value) bool {
		if field.Int() == 0 {
			id := genid.Snowflake.GenerateId()
			fmt.Println("雪花ID： " + fmt.Sprintf("%d", id))
			field.SetInt(id)
			return true
		}
		return false
	})

	trySetField(value, "CreatedAt", func(field reflect.Value) bool {
		if field.Type().String() == "time.Time" && field.Interface().(time.Time).IsZero() {
			field.Set(reflect.ValueOf(now))
			return true
		}
		return false
	})

	trySetField(value, "UpdatedAt", func(field reflect.Value) bool {
		if field.Type().String() == "time.Time" {
			field.Set(reflect.ValueOf(now))
			return true
		}
		return false
	})

	if userID > 0 {
		trySetField(value, "CreatedBy", func(field reflect.Value) bool {
			if field.Int() == 0 {
				field.SetInt(userID)
				return true
			}
			return false
		})

		trySetField(value, "UpdatedBy", func(field reflect.Value) bool {
			field.SetInt(userID)
			return true
		})
	}
}

// 设置更新字段
func setUpdateFields(value reflect.Value, ctx context.Context) {
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}

	now := time.Now()
	var userID int64
	if ctx != nil {
		if user := ctxutil.GetLoginUser(ctx); user != nil {
			userID = user.ID
		}
	}

	trySetField(value, "UpdatedAt", func(field reflect.Value) bool {
		if field.Type().String() == "time.Time" {
			field.Set(reflect.ValueOf(now))
			return true
		}
		return false
	})

	if userID > 0 {
		trySetField(value, "UpdatedBy", func(field reflect.Value) bool {
			field.SetInt(userID)
			return true
		})
	}
}

// 设置删除字段
func setDeleteFields(value reflect.Value, ctx context.Context) {
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}

	now := time.Now()
	var userID int64
	if ctx != nil {
		if user := ctxutil.GetLoginUser(ctx); user != nil {
			userID = user.ID
		}
	}

	trySetField(value, "DeletedAt", func(field reflect.Value) bool {
		if field.Type().String() == "gorm.DeletedAt" {
			field.Set(reflect.ValueOf(gorm.DeletedAt{
				Time:  now,
				Valid: true,
			}))
			return true
		}
		return false
	})

	if userID > 0 {
		trySetField(value, "DeletedBy", func(field reflect.Value) bool {
			field.SetInt(userID)
			return true
		})
	}
}

// 尝试设置字段的安全函数
func trySetField(value reflect.Value, fieldName string, setter func(field reflect.Value) bool) bool {
	field := value.FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return false
	}
	return setter(field)
}
