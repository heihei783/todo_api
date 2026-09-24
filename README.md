# Todo API


## API 接口

### 基础接口

| 方法     | 接口                      | 说明           |
| ------ | ----------------------- | ------------ |
| GET    | `/api/health`           | 健康检查         |
| GET    | `/api/docs`             | API 文档       |
| GET    | `/api/todos`            | 获取 Todo 列表   |
| POST   | `/api/todos`            | 创建 Todo      |
| GET    | `/api/todos/:id`        | 获取 Todo 详情   |
| PUT    | `/api/todos/:id`        | 更新 Todo      |
| DELETE | `/api/todos/:id`        | 删除 Todo      |
| PATCH  | `/api/todos/:id/toggle` | 切换 Todo 状态   |
| GET    | `/api/todos/statistics` | 获取 Todo 统计信息 |



## 项目结构


```text
main.go
│
├── 📦 数据模型（Models）
│   ├── Todo
│   │   └── Todo 任务结构体
│   │
│   ├── TodoCreateRequest
│   │   └── 创建 Todo 请求结构体
│   │
│   ├── TodoUpdateRequest
│   │   └── 更新 Todo 请求结构体
│   │
│   ├── APIResponse
│   │   └── 统一 API 响应结构体
│   │
│   └── PaginatedResponse
│       └── 分页响应结构体
│
├── 🔧 服务层（Services）
│   ├── TodoService
│   │   └── Todo 服务接口
│   │
│   └── TodoServiceImpl
│       └── Todo 服务具体实现
│
├── 🌐 路由层（Handlers）
│   ├── handleHome
│   │   └── 首页处理
│   │
│   ├── handleHealth
│   │   └── 健康检查
│   │
│   ├── handleListTodos
│   │   └── 获取 Todo 列表
│   │
│   ├── handleCreateTodo
│   │   └── 创建 Todo
│   │
│   ├── handleGetTodo
│   │   └── 获取 Todo 详情
│   │
│   ├── handleUpdateTodo
│   │   └── 更新 Todo
│   │
│   ├── handleDeleteTodo
│   │   └── 删除 Todo
│   │
│   ├── handleToggleTodo
│   │   └── 切换 Todo 状态
│   │
│   └── handleTodoStatistics
│       └── 获取 Todo 统计信息
│
└── 🛡️ 中间件（Middleware）
    ├── corsMiddleware
    │   └── CORS 跨域中间件
    │
    ├── loggingMiddleware
    │   └── 日志中间件
    │
    └── errorHandlerMiddleware
        └── 错误处理中间件
```





