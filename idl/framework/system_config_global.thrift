// 全局参数（匿名可读）：独立 IDL 文件 → 独立 Register 与 middleware.go，
// rootMw 不挂登录，实现 config-consumption S2.2 的免登录端点；
// 不得并入 system_dict.thrift（其 Register 整组继承 permission.Default）。
namespace go framework.syscfgglobal

struct ConfigGlobalDTO {
    1: i64 id (go.tag = "json:\"id,string\"")
    2: string configName
    3: string configKey
    4: string configValue
    5: string configType
    6: string remark
    7: i32 status
}

struct ConfigGlobalListResponse {
    1: list<ConfigGlobalDTO> list
    2: i64 total
}

service SystemConfigGlobalController {
    ConfigGlobalListResponse GlobalList()(api.get = "/system/config/global")
}
