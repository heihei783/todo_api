package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Todo 任务结构体
type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title" binding:"required,min=1,max=200"`
	Description string     `json:"description" binding:"max=1000"`
	Status      string     `json:"status" binding:"oneof=pending completed cancelled"`
	Priority    string     `json:"priority" binding:"oneof=low medium high"`
	DueDate     *string    `json:"due_date,omitempty"`
	CreateAt    time.Time  `json:"created_at"`
	UpdateAt    time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TodoCreateRequest struct
type TodoCreateRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=1000"`
	Priority    string `json:"priority" binding:"oneof=low medium high"`
	DueDate     string `json:"due_date,omitempty"`
}

// TodoUpdateRequest 更新任务请求
type TodoUpdateRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=pending completed cancelled"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low medium high"`
	DueDate     *string `json:"due_date,omitempty"`
}

// APIResponse 通用API相应
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// 创建数据库
func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title VARCHAR(200) NOT NULL,
	description TEXT,
	status VARCHAR(20) DEFAULT 'pending' CHECK(status IN('pending','completed','cancelled')),
	priority VARCHAR(10) DEFAULT 'medium' CHECK(priority IN('low','medium','high')),
	due_date DATE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	completed_at TIMESTAMP
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建数据库失败：%w", err)
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)",
		"CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority)",
		"CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at)",
	}
	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			return fmt.Errorf("创建索引失败：%w", err)
		}
	}
	return nil
}

// 初始化数据库
func initDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal("连接数据库失败；", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败", err)
	}

	if err := createTables(db); err != nil {
		log.Fatal("创建数据库表失败：", err)
	}

	if err := seedData(db); err != nil {
		log.Fatal("初始化示例数据示例失败:", err)
	}

	fmt.Println("数据库初始化完成")
	return db

}

// 初始数据
func seedData(db *sql.DB) error {
	var count int
	db.QueryRow("SELECT COUNT (*) FROM todos").Scan(&count)
	if count > 0 {
		return nil
	}

	sampleTodos := []struct {
		title       string
		description string
		status      string
		priority    string
		dueDate     string
	}{
		{"学习Go语言", "完成Go语言基础教程的学习", "completed", "high", "2024-01-15"},
		{"写代码", "实现TODO API项目", "pending", "high", "2024-01-20"},
		{"阅读文档", "阅读Gin框架官方文档", "pending", "medium", "2024-01-25"},
		{"代码审查", "审查团队提交的代码", "completed", "medium", "2024-01-10"},
		{"部署应用", "将应用部署到生产环境", "cancelled", "low", "2024-02-01"},
	}
	query := `
	INSERT INTO todos(title,description,status,priority,due_date) VALUES (?,?,?,?,?)`
	for _, todo := range sampleTodos {
		_, err := db.Exec(query, todo.title, todo.description, todo.status, todo.priority, todo.dueDate)
		if err != nil {
			return fmt.Errorf("插入实例数据失败:%w", err)
		}
	}
	fmt.Println("插入了%d条示例数据", len(sampleTodos))
	return nil
}

// TodoService 任务服务接口
type TodoService interface {
	Create(todo *Todo) error
	GetByID(id int) (*Todo, error)
	Update(todo *Todo) error
	Delete(id int) error
	List(filter TodoFilter) ([]Todo, int, error)
	ToggleStatus(id int, status string) error
}

// TodoFilter 任务查询过滤器
type TodoFilter struct {
	Status   string
	Priority string
	Search   string
	Page     int
	PageSize int
}

// TodoServiceImpl 任务服务实现
type TodoServiceImpl struct {
	db *sql.DB
}

func NewTodoService(db *sql.DB) TodoService {
	return &TodoServiceImpl{db: db}
}

// Create创建任务
func (s *TodoServiceImpl) Create(todo *Todo) error {
	query := `
	INSERT INTO todos (title,description,status,priority,due_date,created_at,updated_at)
	VALUES(?,?,?,?,?,?,?)`
	now := time.Now()
	todo.CreateAt = now
	todo.UpdateAt = now
	if todo.Status == "" {
		todo.Status = "pending"
	}

	result, err := s.db.Exec(query, todo.Title, todo.Description, todo.Status, todo.Priority, todo.DueDate, todo.CreateAt, todo.UpdateAt)
	if err != nil {
		return fmt.Errorf("创建任务失败：%w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取任务ID失败：%w", err)
	}
	todo.ID = int(id)
	return nil
}

// GetByID 获取任务
func (s *TodoServiceImpl) GetByID(id int) (*Todo, error) {
	query := "SELECT * FROM todos WHERE id=?"
	todo := &Todo{}
	var completedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Status, &todo.Priority, &todo.DueDate,
		&todo.CompletedAt, &todo.UpdateAt, &completedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("任务不存在")
		}
		return nil, fmt.Errorf("获取任务失败：%w", err)
	}

	if completedAt.Valid {
		todo.CompletedAt = &completedAt.Time

	}

	return todo, nil

}

// Update 更新任务
func (s *TodoServiceImpl) Update(todo *Todo) error {
	query := `
	UPDATE todos SET title=?,description=?,status=?,priority=?,due_date=?,updated_at=?,completed_at=?
	WHERE id=?`
	now := time.Now()
	todo.UpdateAt = now

	var completedAt interface{}
	if todo.CompletedAt != nil {
		completedAt = todo.CompletedAt
	}

	_, err := s.db.Exec(query, todo.Title, todo.Description, todo.Status, todo.Priority, todo.DueDate, todo.UpdateAt, completedAt, todo.ID)
	if err != nil {
		return fmt.Errorf("更新任务失败：%w", err)
	}
	return nil
}

// Delete 删除任务
func (s *TodoServiceImpl) Delete(id int) error {
	query := `DELETE FROM todos WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除任务失败：%w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("任务不存在")
	}
	return nil
}

// List 获取任务列表
func (s *TodoServiceImpl) List(filter TodoFilter) ([]Todo, int, error) {
	//构建WHERE条件
	whereClauses := " WHERE 1=1"
	args := []interface{}{}

	if filter.Status != "" {
		whereClauses += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Priority != "" {
		whereClauses += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	if filter.Search != "" {
		whereClauses += " AND (title LIKE ? OR description LIKE ?)"
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	//获取总数
	countQuery := "SELECT COUNT(*) FROM todos" + whereClauses
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("获取任务总数失败：%w", err)
	}

	//分页查询
	query := "SELECT * FROM todos " + whereClauses + " ORDER BY id DESC LIMIT ? OFFSET ?"
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := s.db.Query(query, args...)

	if err != nil {
		return nil, 0, fmt.Errorf("分页查询任务失败：%w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		todo := &Todo{}
		var completedAt sql.NullTime

		err := rows.Scan(
			&todo.ID, &todo.Title, &todo.Description, &todo.Status, &todo.Priority,
			&todo.DueDate, &todo.CreateAt, &todo.UpdateAt, &completedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("扫描任务数据失败：%w", err)
		}

		if completedAt.Valid {
			todo.CompletedAt = &completedAt.Time
		}
		todos = append(todos, *todo)

	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("分页查询任务失败：%w", err)
	}
	return todos, total, nil
}

// ToggleStatus 切换任务状态
func (s *TodoServiceImpl) ToggleStatus(id int, status string) error {
	query := " UPDATE todos SET status = ? ,updated_at = ?,completed_at = ? WHERE id = ?"
	now := time.Now()
	var completedAt interface{}
	if status == "completed" {
		completedAt = now
	}

	_, err := s.db.Exec(query, status, now, completedAt, id)
	if err != nil {
		return fmt.Errorf("切换任务失败:%w", err)
	}
	return nil

}

// setupServer 设置HTTP服务
func setupServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(corsMiddleware())
	r.Use(errorHandlerMiddleware())

	setupRoutes(r)

	return r
}

// 路由处理
func setupRoutes(r *gin.Engine) {
	//API路由组
	api := r.Group("/api")
	{
		api.GET("/health", handleHealth)
		api.GET("/docs", handleDocs)
		todos := api.Group("/todos")
		{
			todos.GET("", handleListTodos)
			todos.POST("", handleCreateTodo)
			todos.GET("/:id", HandleGetTodo)
			todos.PUT("/:id", handleUpdateTodo)
			todos.DELETE("/:id", handleDeleteTodo)
			todos.PATCH("/:id/toggle", handleToggleTodo)
			todos.GET("/statistics", handleTodoStatistics)
		}
	}
}

// 实现处理函数

// handlerHome 处理首页
func handlerHome(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "TODO API应用",
		"version": "1.0.0",
		"endpoints": []string{
			"GET/ api/todos -获取任务列表",
			"POST/api/todos - 创建任务",
			"GET/api/todos/:id-获取任务详情",
			"PUT/api/todos/:id - 更新任务",
			"DELETE/api/todos/:id - 删除任务",
			"PATCH/api/todos/:id/toggle - 切换任务状态",
		},
	})
}

// 健康检查
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "服务器运行正常",
		Timestamp: time.Now(),
	})
}

// handleDocs API文档
func handleDocs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":       "TODO API 文档",
		"description": "一个功能完整的任务管理API",
		"version":     "1.0.0",
		"baseUrl":     "http://localhost:8080",
		"endpoints": gin.H{
			"tasks": gin.H{
				"list": gin.H{
					"method":      "GET",
					"path":        "/api/todos",
					"description": "获取任务列表",
					"parameters": gin.H{
						"page": gin.H{
							"type":        "query",
							"description": "页码，默认1",
							"required":    false,
						},
						"page_size": gin.H{
							"type":        "query",
							"description": "每页大小，默认10",
							"required":    false,
						},
						"status": gin.H{
							"type":        "query",
							"description": "任务状态: pending, completed, cancelled",
							"required":    false,
						},
						"priority": gin.H{
							"type":        "query",
							"description": "优先级: low, medium, high",
							"required":    false,
						},
						"search": gin.H{
							"type":        "query",
							"description": "搜索关键词",
							"required":    false,
						},
					},
				},
				"create": gin.H{
					"method":      "POST",
					"path":        "/api/todos",
					"description": "创建新任务",
					"body": gin.H{
						"title": gin.H{
							"type":        "string",
							"description": "任务标题",
							"required":    true,
							"min_length":  1,
							"max_length":  200,
						},
						"description": gin.H{
							"type":        "string",
							"description": "任务描述",
							"required":    false,
							"max_length":  1000,
						},
						"priority": gin.H{
							"type":        "string",
							"description": "优先级: low, medium, high",
							"required":    false,
							"default":     "medium",
						},
						"due_date": gin.H{
							"type":        "string",
							"description": "截止日期 (YYYY-MM-DD)",
							"required":    false,
						},
					},
				},
			},
		},
	})
}

// handleListTodos 获取任务列表
func handleListTodos(c *gin.Context) {
	filter := TodoFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Search:   c.Query("search"),
		Page:     1,
		PageSize: 10,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		filter.Page = page
	}

	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && pageSize > 0 && pageSize <= 100 {

		filter.PageSize = pageSize
	}

	todos, total, err := todoService.List(filter)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "获取任务列表失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "获取任务列表成功",
		Data: PaginatedResponse{
			Items:      todos,
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now(),
	})

}

// handleCreatTodo 创建任务
func handleCreateTodo(c *gin.Context) {
	var req TodoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "请求参数无效",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	//创建任务
	todo := &Todo{
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
	}

	if req.DueDate != "" {
		todo.DueDate = &req.DueDate
	}

	if err := todoService.Create(todo); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "创建任务失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success:   true,
		Message:   "任务创建成功",
		Data:      todo,
		Timestamp: time.Now(),
	})
}

// handleGetTodo 获取任务详情
func HandleGetTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "无效的ID",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	todo, err := todoService.GetByID(id)

	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Message:   "任务不存在",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "获取任务成功",
		Data:      todo,
		Timestamp: time.Now(),
	})
}

// handleUpdateTodo 更新任务
func handleUpdateTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "无效的任务id",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	todo, err := todoService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Message:   "任务不存在",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	var req TodoUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "请求参数无效",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if req.Title != nil {
		todo.Title = *req.Title
	}

	if req.Description != nil {
		todo.Description = *req.Description
	}

	if req.Status != nil {
		todo.Status = *req.Status
		if *req.Status == "completed" {
			now := time.Now()
			todo.CompletedAt = &now
		} else {
			todo.CompletedAt = nil

		}
	}

	if req.Priority != nil {
		todo.Priority = *req.Priority
	}

	if req.DueDate != nil {
		if *req.DueDate == "" {
			todo.DueDate = nil
		} else {
			todo.DueDate = req.DueDate
		}
	}

	if err := todoService.Update(todo); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "更新状态失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "更新状态成功",
		Data:      todo,
		Timestamp: time.Now(),
	})
}

// DeleteTodo 删除任务
func handleDeleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "参数错误",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := todoService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "删除任务失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "删除任务成功",
		Timestamp: time.Now(),
	})
}

// handleToggleTodo 切换任务状态
func handleToggleTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success:   false,
			Message:   "无效的任务ID",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	todo, err := todoService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success:   false,
			Message:   "任务不存在",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	var nowStatus string
	switch todo.Status {
	case "pending":
		nowStatus = "completed"
	case "completed":
		nowStatus = "pending"
	case "cancelled":
		nowStatus = "pending"
	default:
		nowStatus = "pending"
	}

	err = todoService.ToggleStatus(id, nowStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "任务切换状态失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "任务切换状态成功",
		Timestamp: time.Now(),
	})
}

// handleTodoStatistics 获取统计信息
func handleTodoStatistics(c *gin.Context) {
	query := `
	SELECT status,COUNT(*) as count FROM todos GROUP BY status`

	rows, err := todoService.(*TodoServiceImpl).db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "获取统计信息失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	defer rows.Close()

	stats := make(map[string]int)
	total := 0

	for rows.Next() {
		var status string
		var count int

		err := rows.Scan(&status, &count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success:   false,
				Message:   "获取统计信息失败",
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		total += count
		stats[status] = count
	}

	today := time.Now().Format("2006-01-02")
	var todayCount int
	todoService.(*TodoServiceImpl).db.QueryRow(
		"SELECT COUNT(*) FROM todos WHERE DATE(created_at) = ?", today,
	).Scan(&todayCount)

	overdueQuery := "SELECT COUNT(*) FROM todos WHERE status = 'pending' AND due_date < DATA('now)"

	var overdueCount int
	todoService.(*TodoServiceImpl).db.QueryRow(overdueQuery).Scan(&overdueCount)

	statistics := gin.H{
		"total":         total,
		"by_status":     stats,
		"today_tasks":   todayCount,
		"overdue_tasks": overdueCount,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "获取统计信息成功",
		Data:      statistics,
		Timestamp: time.Now(),
	})
}

// cors中间件
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

}

// 日志中间件
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})

}

// 错误处理中间件
func errorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recoverd interface{}) {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success:   false,
			Message:   "服务器内部错误",
			Error:     fmt.Sprintf("%v", recoverd),
			Timestamp: time.Now(),
		})
		c.Abort()
	})
}

var todoService TodoService

func main() {

	db := initDatabase()
	defer db.Close()

	todoService = NewTodoService(db)
	server := setupServer()

	log.Fatal(server.Run(":8080"))
}
