<<<<<<< HEAD
# Todo API
get /api/health
get /api/docs
get /api/todos
post /api/todos
get /api/todos/:id
put /api/todos/:id
delete /api/todos/:id
patch /api/todos/:id/toggle
get /api/todos/statistics


main.go
├── 数据模型 (Models)
│   ├── Todo              # 任务结构体
│   ├── TodoCreateRequest # 创建请求结构体
│   ├── TodoUpdateRequest # 更新请求结构体
│   ├── APIResponse      # API响应结构体
│   └── PaginatedResponse # 分页响应结构体
├── 服务层 (Services)
│   ├── TodoService       # 任务服务接口
│   └── TodoServiceImpl  # 任务服务实现
├── 路由层 (Handlers)
│   ├── handleHome        # 首页处理
│   ├── handleHealth      # 健康检查
│   ├── handleListTodos   # 获取任务列表
│   ├── handleCreateTodo  # 创建任务
│   ├── handleGetTodo     # 获取任务详情
│   ├── handleUpdateTodo  # 更新任务
│   ├── handleDeleteTodo  # 删除任务
│   ├── handleToggleTodo  # 切换任务状态
│   └── handleTodoStatistics # 统计信息
└── 中间件 (Middleware)
    ├── corsMiddleware    # CORS 跨域中间件
    ├── loggingMiddleware # 日志中间件
    └── errorHandlerMiddleware # 错误处理中间件

接口api.html 可以导入foxapi等测试接口软件测试

