package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"text/template"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var generateConfig = map[string]map[string]map[string]any{
	"application/framework/dal/query": {
		"system_user": {},
		"system_user_role": {
			"repository": false,
			"service":    false,
		},
		"system_role": {},
		"system_role_permission": {
			"repository": false,
			"service":    false,
		},
		"system_menu": {
			"repository": false,
			"service":    false,
		},
		"system_dept": {
			"repository": false,
			"service":    false,
		},
		"system_post": {
			"repository": false,
			"service":    false,
		},
		"system_user_post": {
			"repository": false,
			"service":    false,
		},
		"system_role_menu": {
			"repository": false,
			"service":    false,
		},
		"system_role_dept": {
			"repository": false,
			"service":    false,
		},
		"system_dict_type": {
			"repository": false,
			"service":    false,
		},
		"system_dict_data": {
			"repository": false,
			"service":    false,
		},
		"system_dict_group": {
			"repository": false,
			"service":    false,
		},
		"system_config": {
			"repository": false,
			"service":    false,
		},
		"system_notice": {
			"repository": false,
			"service":    false,
		},
		"system_notice_read": {
			"repository": false,
			"service":    false,
		},
		"system_oss_config": {
			"repository": false,
			"service":    false,
		},
		"system_oss": {
			"repository": false,
			"service":    false,
		},
		"system_oss_rule": {
			"repository": false,
			"service":    false,
		},
		"system_oper_log": {
			"repository": false,
			"service":    false,
		},
		"system_error_log": {
			"repository": false,
			"service":    false,
		},
		"system_login_log": {
			"repository": false,
			"service":    false,
		},
	},
}

func main() {
	GetCodegen()
}

// TableInfo 表信息结构体
type TableInfo struct {
	TableName      string
	FileStructName string
	StructName     string
	RepositoryName string
	ServiceName    string
	PackageName    string
	ServicePackage string
	FullPackage    string
	ModelPackage   string
	QueryPackage   string
}

func GetCodegen() {

	// 初始化数据库连接（默认使用本地 docker-compose 开发库，可通过 GOSILITE_CODEGEN_DSN 覆盖）
	dsn := os.Getenv("GOSILITE_CODEGEN_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 user=gosilite password=gosilite dbname=gosilite port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	rootPath, err := findProjectRoot()
	if err != nil {
		panic(err)
	}

	for path, mapping := range generateConfig {
		fmt.Printf("Processing path: %s\n", filepath.Join(rootPath, path))

		// 检查是否已存在生成的文件
		parts := strings.Split(path, "/")
		fullPackage := strings.Join(parts[:len(parts)-1], "/")
		modelPath := filepath.Join(rootPath, fullPackage, "model")
		//queryPath := filepath.Join(rootPath, path)

		// 检查model目录和query目录是否已存在文件
		//modelFiles, _ := os.ReadDir(modelPath)
		//queryFiles, _ := os.ReadDir(queryPath)

		// if len(modelFiles) > 0 && len(queryFiles) > 0 {
		// 	fmt.Printf("GORM生成文件已存在，跳过生成: %s\n", path)
		// } else {
		// 生成GORM模型和查询代码
		g := gen.NewGenerator(gen.Config{
			OutPath: filepath.Join(rootPath, path),
			Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		g.UseDB(db)
		g.WithOpts(gen.FieldType("deleted_at", "gorm.DeletedAt"))

		var resolveType func(typ reflect.Type, required bool) string
		resolveType = func(typ reflect.Type, required bool) string {
			switch typ.Kind() {
			case reflect.Ptr:
				return resolveType(typ.Elem(), false)
			case reflect.Slice:
				return "[]" + resolveType(typ.Elem(), required)
			default:
				prefix := "*"
				if required {
					prefix = ""
				}

				if strings.HasSuffix(typ.PkgPath(), modelPath) {
					return prefix + typ.Name()
				}

				return prefix + typ.String()
			}
		}

		genModify := func(col string, model any) func(f gen.Field) gen.Field {
			return func(f gen.Field) gen.Field {
				if f.ColumnName != col {
					return f
				}

				st := reflect.TypeOf(model)
				// f.Name = st.Name()
				f.Type = resolveType(st, true)
				f.GORMTag.Set("serializer", "json")
				return f
			}
		}

		timeModify := func(f gen.Field) gen.Field {
			if f.ColumnName == "updated_at" {
				// https://gorm.io/zh_CN/docs/models.html#%E5%88%9B%E5%BB%BA-x2F-%E6%9B%B4%E6%96%B0%E6%97%B6%E9%97%B4%E8%BF%BD%E8%B8%AA%EF%BC%88%E7%BA%B3%E7%A7%92%E3%80%81%E6%AF%AB%E7%A7%92%E3%80%81%E7%A7%92%E3%80%81Time%EF%BC%89
				f.GORMTag.Set("autoUpdateTime", "milli")
			}
			if f.ColumnName == "created_at" {
				f.GORMTag.Set("autoCreateTime", "milli")
			}
			return f
		}
		var models []any
		// 显式列类型覆盖：gorm gen 对个别列名推断有误（如 system_oss.service varchar 被判成 bool）
		fieldTypes := map[string]map[string]string{
			"system_oss": {"service": "string"},
		}
		for table, col2Model := range mapping {
			opts := make([]gen.ModelOpt, 0, len(col2Model)+1)
			for column, m := range col2Model {
				// "repository"/"service" 是本文件的生成开关（bool），不是列映射，跳过
				if column == "repository" || column == "service" {
					if _, isBool := m.(bool); isBool {
						continue
					}
				}
				cp := m
				opts = append(opts, gen.FieldModify(genModify(column, cp)))
			}
			for column, typ := range fieldTypes[table] {
				opts = append(opts, gen.FieldType(column, typ))
			}
			opts = append(opts, gen.FieldModify(timeModify))
			models = append(models, g.GenerateModel(table, opts...))
		}

		g.ApplyBasic(models...)
		g.Execute()
		//}

		// 生成Repository和Service代码
		for table, config := range mapping {
			fileStructName := strings.TrimPrefix(table, "system_")
			structName := tableNameToStructName(table)

			// 提取包名信息
			parts := strings.Split(path, "/")
			// 构建完整包名
			fullPackage := strings.Join(parts[:len(parts)-1], "/")
			modelPackage := fullPackage + "/model"
			queryPackage := fullPackage + "/query"
			ServicePackage := strings.Join(parts[:len(parts)-2], "/")

			// 构建表信息
			tableInfo := TableInfo{
				TableName:      table,
				FileStructName: fileStructName,
				StructName:     structName,
				RepositoryName: structName + "Repository",
				ServiceName:    structName + "Service",
				PackageName:    parts[len(parts)-2], // 获取倒数第二个部分作为包名
				ServicePackage: ServicePackage,
				FullPackage:    fullPackage,
				ModelPackage:   modelPackage,
				QueryPackage:   queryPackage,
			}

			isRepo, rExists := config["repository"]
			isService, sExists := config["service"]
			if !rExists || isRepo.(bool) {
				// 生成Repository代码
				generateRepository(rootPath, fileStructName, fullPackage, tableInfo)
			}

			if (!sExists) || isService.(bool) {
				// 生成Service代码
				generateService(rootPath, fileStructName, ServicePackage, tableInfo)
			}
		}
	}
}

func findProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// 从当前文件路径开始向上查找go.mod文件
	currentDir := filepath.Dir(filename)
	for {
		// 检查是否存在go.mod文件
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		// 检查是否到达文件系统根目录
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("could not find go.mod file in any parent directory")
		}

		currentDir = parentDir
	}
}

// 根据表名生成对应的结构体名称
func tableNameToStructName(tableName string) string {
	parts := strings.Split(tableName, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// 生成Repository代码
func generateRepository(rootPath, fileStructName, fullPackage string, tableInfo TableInfo) {
	// 读取模板文件
	tmplPath := filepath.Join(rootPath, "script/codegen/templates/repository.go.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		panic(fmt.Errorf("读取repository模板失败: %w", err))
	}

	// 解析模板
	tmpl, err := template.New("repository").Parse(string(tmplContent))
	if err != nil {
		panic(fmt.Errorf("解析repository模板失败: %w", err))
	}

	// 确定输出路径
	repoPath := fullPackage + "/repository"
	outputFile := fileStructName + "_repository.go"
	outputPath := filepath.Join(rootPath, repoPath, outputFile)

	// 检查文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("文件已存在，跳过生成: %s\n", outputPath)
		return
	}

	// 创建输出目录
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Errorf("创建repository目录失败: %w", err))
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		panic(fmt.Errorf("创建repository文件失败: %w", err))
	}
	defer file.Close()

	// 执行模板生成
	if err := tmpl.Execute(file, tableInfo); err != nil {
		panic(fmt.Errorf("生成repository代码失败: %w", err))
	}

	fmt.Printf("生成Repository文件成功: %s\n", outputPath)
}

// 生成Service代码
func generateService(rootPath, fileStructName, servicePackage string, tableInfo TableInfo) {
	// 单文件 Service（无 service_impl）
	generateServiceBase(rootPath, fileStructName, servicePackage, tableInfo)
}

// 生成service.go文件（具体 Service + NewService）
func generateServiceBase(rootPath, fileStructName, servicePackage string, tableInfo TableInfo) {
	// 读取模板文件
	tmplPath := filepath.Join(rootPath, "script/codegen/templates/service.go.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		panic(fmt.Errorf("读取service模板失败: %w", err))
	}

	// 解析模板
	tmpl, err := template.New("service").Parse(string(tmplContent))
	if err != nil {
		panic(fmt.Errorf("解析service模板失败: %w", err))
	}

	// 确定输出路径
	servicePath := servicePackage + "/service/" + fileStructName
	outputFile := "service.go"
	outputPath := filepath.Join(rootPath, servicePath, outputFile)

	// 检查文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("文件已存在，跳过生成: %s\n", outputPath)
		return
	}

	// 创建输出目录
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Errorf("创建service目录失败: %w", err))
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		panic(fmt.Errorf("创建service文件失败: %w", err))
	}
	defer file.Close()

	// 执行模板生成
	if err := tmpl.Execute(file, tableInfo); err != nil {
		panic(fmt.Errorf("生成service代码失败: %w", err))
	}

	fmt.Printf("生成Service文件成功: %s\n", outputPath)
}

