namespace go framework.syscfg

struct PageQuery {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
}

struct IdRequest {
    1: i64 id
}

struct DictTypeDTO {
    1: i64 id
    2: string dictName
    3: string dictType
    4: i32 status
    5: i32 orderNum
    6: string remark
    7: i64 groupId
}

struct DictTypeListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
    // 缺省=全部；0=未分组；>0=指定分组
    4: optional i64 groupId
}

struct DictTypeListResponse {
    1: list<DictTypeDTO> list
    2: i64 total
}

struct DictTypeSaveRequest {
    1: optional i64 id
    2: string dictName
    3: string dictType
    4: i32 status
    5: i32 orderNum
    6: optional string remark
    7: i64 groupId
}

struct DictGroupDTO {
    1: i64 id
    2: string groupName
    3: i32 orderNum
    4: string remark
}

struct DictGroupListResponse {
    1: list<DictGroupDTO> list
}

struct DictGroupSaveRequest {
    1: optional i64 id
    2: string groupName
    3: i32 orderNum
    4: optional string remark
}

struct DictDataDTO {
    1: i64 id
    2: string dictType
    3: string dictLabel
    4: string dictValue
    5: string cssClass
    6: i32 listOrder
    7: i32 status
    8: i32 isDefault
}

struct DictDataListRequest {
    1: i32 page = 1
    2: i32 pageSize = 50
    3: optional string dictType
}

struct DictDataListResponse {
    1: list<DictDataDTO> list
    2: i64 total
}

struct DictDataSaveRequest {
    1: optional i64 id
    2: string dictType
    3: string dictLabel
    4: string dictValue
    5: optional string cssClass
    6: i32 listOrder
    7: i32 status
    8: i32 isDefault
}

struct DictOptionsRequest {
    1: string dictType
}

struct DictOptionDTO {
    1: string label
    2: string value
}

struct DictOptionsResponse {
    1: list<DictOptionDTO> list
}

struct ConfigDTO {
    1: i64 id
    2: string configName
    3: string configKey
    4: string configValue
    5: string configType
    6: string remark
    7: i32 status
}

struct ConfigListResponse {
    1: list<ConfigDTO> list
    2: i64 total
}

struct ConfigSaveRequest {
    1: optional i64 id
    2: string configName
    3: string configKey
    4: string configValue
    5: optional string configType
    6: optional string remark
    7: i32 status
}

struct ConfigKeyRequest {
    1: string key
}

service SystemDictController {
    DictTypeListResponse TypeList(1: DictTypeListRequest req)(api.get = "/system/dict/type/list")
    DictTypeDTO TypeCreate(1: DictTypeSaveRequest req)(api.post = "/system/dict/type")
    DictTypeDTO TypeUpdate(1: DictTypeSaveRequest req)(api.put = "/system/dict/type")
    i32 TypeRemove(1: IdRequest req)(api.delete = "/system/dict/type")

    DictGroupListResponse GroupList(1: PageQuery req)(api.get = "/system/dict/group/list")
    DictGroupDTO GroupCreate(1: DictGroupSaveRequest req)(api.post = "/system/dict/group")
    DictGroupDTO GroupUpdate(1: DictGroupSaveRequest req)(api.put = "/system/dict/group")
    i32 GroupRemove(1: IdRequest req)(api.delete = "/system/dict/group")

    DictDataListResponse DataList(1: DictDataListRequest req)(api.get = "/system/dict/data/list")
    DictDataDTO DataCreate(1: DictDataSaveRequest req)(api.post = "/system/dict/data")
    DictDataDTO DataUpdate(1: DictDataSaveRequest req)(api.put = "/system/dict/data")
    i32 DataRemove(1: IdRequest req)(api.delete = "/system/dict/data")
    DictOptionsResponse DataOptions(1: DictOptionsRequest req)(api.get = "/system/dict/data/options")
}

service SystemConfigController {
    ConfigListResponse List(1: PageQuery req)(api.get = "/system/config/list")
    ConfigDTO Create(1: ConfigSaveRequest req)(api.post = "/system/config")
    ConfigDTO Update(1: ConfigSaveRequest req)(api.put = "/system/config")
    i32 Remove(1: IdRequest req)(api.delete = "/system/config")
    ConfigDTO GetByKey(1: ConfigKeyRequest req)(api.get = "/system/config/key")
}
