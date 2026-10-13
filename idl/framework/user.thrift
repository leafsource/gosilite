namespace go framework.user

// 登录请求
struct LoginRequest {
    1: required string username
    2: required string password
}

// 用户信息
struct UserProfile {
    1: required i64 id (go.tag = "json:\"id,string\"")
    2: required string account
    3: required string username
    4: required string avatarURL
}

// 个人资料（GET /user/profile）
struct ProfileResponse {
    1: required i64 id (go.tag = "json:\"id,string\"")
    2: required string account
    3: required string username
    4: required string avatarURL
    5: required string email
    6: required string mobilePhone
    7: required string state
    // yyyy-MM-dd HH:mm:ss
    8: required string createdAt
}

// 登录用户响应
struct LoginUserResponse {
    1: required UserProfile user
    2: required list<string> roles
    3: required list<string> permissions
}

// 更新头像（统一文件服务两步链路第二步）：按台账 fileId 回查 key 更新 avatar_url
struct UpdateProfileAvatarRequest {
    1: required i64 fileId (go.tag = "json:\"fileId,string\"")
}

struct UpdateProfileAvatarResponse {
    1: required string avatarURL
}


// 刷新会话请求
struct RefreshRequest {
    1: required string refreshToken
}

// 会话令牌响应
struct SessionTokenResponse {
    1: required string accessToken
    2: required string tokenType
    3: required i64 expiresIn
    4: required string refreshToken
}

// 退出登录请求
struct LogoutRequest {
    1: optional string refreshToken
}

// 修改密码请求
struct ChangePasswordRequest {
    1: required string currentPassword
    2: required string newPassword
}

// 服务定义
service UserController {
    // 登录
    SessionTokenResponse Login(1: LoginRequest req)(api.post="/user/login")

    // 获取用户信息（含 roles / permissions）
    LoginUserResponse GetInfo()(api.get="/user/info")

    // 获取用户资料
    ProfileResponse GetProfile()(api.get="/user/profile")

    // 刷新会话
    SessionTokenResponse Refresh(1: RefreshRequest req)(api.post="/user/refresh")

    // 退出登录
    LogoutRequest Logout(1: LogoutRequest req)(api.post="/user/logout")

    // 修改密码
    ChangePasswordRequest ChangePassword(1: ChangePasswordRequest req)(api.post="/user/password")

    // 更新头像（fileId 来自 /common/file/upload，服务端回查台账防客户端塞任意 url）
    UpdateProfileAvatarResponse UpdateProfileAvatar(1: UpdateProfileAvatarRequest req)(api.put="/user/profile/avatar")
}