namespace go framework.menu

// Meta 元数据
struct Meta {
    1: string title            // 标题
    2: string icon            // 图标
    3: bool hidden = false    // 是否隐藏
    4: bool single = false    // 是否单级
    5: i32 orderNo            // 排序号
    6: bool keepAlive = false // 是否缓存
    7: bool hiddenBreadcrumb = false // 隐藏面包屑
    8: bool expanded = false  // 是否展开
    9: string frameSrc        // iframe地址
    10: bool frameBlank = false // iframe是否新窗口打开
}

// Route 路由
struct Route {
    1: string path            // 路径
    2: string component       // 组件
    3: string redirect        // 重定向
    4: Meta meta              // 元数据
    5: list<Route> children   // 子路由（注意：Thrift中list元素本身不能是optional，但整个list可以是optional）
    6: string name            // 名称
}

// 响应结构
struct RouteResponse {
    1: list<Route> routes     // 路由列表
}


// 服务定义
service MenuController {
    RouteResponse GetRoutes()(api.get = "/menu/routes")
}