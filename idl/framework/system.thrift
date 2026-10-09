namespace go framework.system

// ========== 公共 ==========

struct PageQuery {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: string keyword
}

// ========== 部门 ==========

struct DeptDTO {
    1: i64 id
    2: i64 parentId
    3: string ancestors
    4: string deptName
    5: i32 orderNum
    6: i32 status
}

struct DeptNode {
    1: i64 id
    2: i64 parentId
    3: string label
    4: i32 orderNum
    5: i32 status
    6: list<DeptNode> children
}

struct DeptTreeResponse {
    1: list<DeptNode> list
}

struct DeptListResponse {
    1: list<DeptDTO> list
    2: i64 total
}

struct DeptSaveRequest {
    1: optional i64 id
    2: i64 parentId
    3: string deptName
    4: i32 orderNum
    5: i32 status
}

// ========== 岗位 ==========

struct PostDTO {
    1: i64 id
    2: string postCode
    3: string postName
    4: i32 orderNum
    5: i32 status
}

struct PostListResponse {
    1: list<PostDTO> list
    2: i64 total
}

struct PostSaveRequest {
    1: optional i64 id
    2: string postCode
    3: string postName
    4: i32 orderNum
    5: i32 status
}

// ========== 用户管理 ==========

struct AdminUserDTO {
    1: i64 id
    2: string account
    3: string username
    4: string mobilePhone
    5: string email
    6: string state
    7: i64 deptId
    8: string deptName
    9: list<i64> roleIds
    10: list<i64> postIds
    11: string avatarUrl
}

struct AdminUserListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: optional string keyword
    4: optional i64 deptId
    5: optional string state
}

struct AdminUserListResponse {
    1: list<AdminUserDTO> list
    2: i64 total
}

struct AdminUserSaveRequest {
    1: optional i64 id
    2: string account
    3: string username
    4: optional string mobilePhone
    5: optional string email
    6: optional string password
    7: i64 deptId
    8: string state
    9: list<i64> roleIds
    10: list<i64> postIds
}

struct IdRequest {
    1: i64 id
}

struct IdsRequest {
    1: list<i64> ids
}

struct IdsResponse {
    1: list<i64> ids
}

struct UserStatusRequest {
    1: i64 id
    2: string state
}

// ========== 菜单 ==========

struct MenuDTO {
    1: i64 id
    2: i64 parentId
    3: string menuName
    4: string menuType
    5: string path
    6: string component
    7: string perms
    8: string icon
    9: i32 orderNum
    10: bool visible
    11: i32 status
    12: string actionType
}

struct MenuNode {
    1: i64 id
    2: i64 parentId
    3: string label
    4: string menuType
    5: string perms
    6: i32 orderNum
    7: bool visible
    8: i32 status
    9: string path
    10: string component
    11: string icon
    12: list<MenuNode> children
    13: string actionType
    14: string redirect
}

struct MenuTreeResponse {
    1: list<MenuNode> list
}

struct MenuListResponse {
    1: list<MenuDTO> list
    2: i64 total
}

struct MenuSaveRequest {
    1: optional i64 id
    2: i64 parentId
    3: string menuName
    4: string menuType
    5: optional string path
    6: optional string component
    7: optional string perms
    8: optional string icon
    9: i32 orderNum
    10: bool visible
    11: i32 status
    12: optional string actionType
    13: optional string redirect
}

// ========== 角色 ==========

struct RoleDTO {
    1: i64 id
    2: string name
    3: string code
    4: string description
    5: string state
    6: i32 dataScope
    7: i32 orderNum
}

struct RoleListRequest {
    1: i32 page = 1
    2: i32 pageSize = 10
    3: optional string keyword
}

struct RoleListResponse {
    1: list<RoleDTO> list
    2: i64 total
}

struct RoleSaveRequest {
    1: optional i64 id
    2: string name
    3: string code
    4: optional string description
    5: string state
    6: i32 dataScope
    7: i32 orderNum
}

// 角色授权：权限项（按钮 F 提升后的权限标识）
struct RolePermItem {
    1: i64 id
    2: string menuName
    3: string perms
    4: bool checked
}

// 角色授权：菜单行（F 已提升为 perms，不再单独成行）
struct RoleAuthMenuNode {
    1: i64 id
    2: i64 parentId
    3: string menuName
    4: string menuType
    5: string icon
    6: i32 orderNum
    7: bool checked
    8: list<RolePermItem> perms
    9: list<RoleAuthMenuNode> children
}

// 角色授权快照：一次返回菜单树勾选态 + 数据范围 + 自定义部门
struct RoleAuthorizeResponse {
    1: i32 dataScope
    2: list<RoleAuthMenuNode> menus
    3: list<i64> deptIds
}

// 角色授权保存：菜单 + 数据范围 + 自定义部门一次提交
struct RoleAuthorizeSaveRequest {
    1: i64 roleId
    2: list<i64> menuIds
    3: list<i64> deptIds
    4: i32 dataScope
}

// 部门负责人设置（userId=0 清除）
struct DeptLeaderRequest {
    1: i64 deptId
    2: i64 userId
}

// 服务定义
service SystemController {
    // 部门
    DeptTreeResponse DeptTree()(api.get = "/system/dept/tree")
    DeptListResponse DeptList()(api.get = "/system/dept/list")
    DeptDTO DeptCreate(1: DeptSaveRequest req)(api.post = "/system/dept")
    DeptDTO DeptUpdate(1: DeptSaveRequest req)(api.put = "/system/dept")
    i32 DeptRemove(1: IdRequest req)(api.delete = "/system/dept")
    i32 DeptLeader(1: DeptLeaderRequest req)(api.put = "/system/dept/leader")

    // 岗位
    PostListResponse PostList(1: PageQuery req)(api.get = "/system/post/list")
    PostDTO PostCreate(1: PostSaveRequest req)(api.post = "/system/post")
    PostDTO PostUpdate(1: PostSaveRequest req)(api.put = "/system/post")
    i32 PostRemove(1: IdRequest req)(api.delete = "/system/post")

    // 用户
    AdminUserListResponse UserList(1: AdminUserListRequest req)(api.get = "/system/user/list")
    AdminUserDTO UserCreate(1: AdminUserSaveRequest req)(api.post = "/system/user")
    AdminUserDTO UserUpdate(1: AdminUserSaveRequest req)(api.put = "/system/user")
    i32 UserRemove(1: IdRequest req)(api.delete = "/system/user")
    i32 UserStatus(1: UserStatusRequest req)(api.put = "/system/user/status")
    IdsResponse UserRoles(1: IdRequest req)(api.get = "/system/user/roles")

    // 菜单
    MenuTreeResponse MenuTree()(api.get = "/system/menu/tree")
    MenuListResponse MenuList()(api.get = "/system/menu/list")
    MenuDTO MenuCreate(1: MenuSaveRequest req)(api.post = "/system/menu")
    MenuDTO MenuUpdate(1: MenuSaveRequest req)(api.put = "/system/menu")
    i32 MenuRemove(1: IdRequest req)(api.delete = "/system/menu")

    // 角色
    RoleListResponse RoleList(1: RoleListRequest req)(api.get = "/system/role/list")
    RoleDTO RoleCreate(1: RoleSaveRequest req)(api.post = "/system/role")
    RoleDTO RoleUpdate(1: RoleSaveRequest req)(api.put = "/system/role")
    i32 RoleRemove(1: IdRequest req)(api.delete = "/system/role")
    RoleAuthorizeResponse RoleAuthorize(1: IdRequest req)(api.get = "/system/role/authorize")
    i32 RoleAuthorizeSave(1: RoleAuthorizeSaveRequest req)(api.put = "/system/role/authorize")
}
