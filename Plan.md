# MrChat - 用户对话管理系统设计文档

## 项目概述

**项目名称**: MrChat
**核心定位**: 基于 Go 的多用户对话管理+余额计费系统
**参考项目**: new-api (AI网关与资产管理系统)
**技术栈**: Go (主导) + Vue 3 + MySQL/PostgreSQL + Redis

## 核心功能模块

### 1. 用户管理模块
- 多种登录方式（密码、OAuth、OIDC）
- 用户角色与权限（Root/Admin/User）
- 用户分组与邀请机制
- 用户配置管理
- 2FA 认证（可选）

### 2. 对话管理模块
- 对话会话 CRUD
- 对话历史存储与检索
- 对话分组/标签
- 对话分享功能
- 对话导出（Markdown/JSON）
- 多模型切换支持
- 上下文管理（Token计数）

### 3. 页面提供模块
- 前端 Web 界面（Vue 3）
- RESTful API 服务
- WebSocket 支持（实时对话）
- 管理后台
- 用户个人中心

### 4. 计费管理模块
- 余额管理（充值、扣费、退款）
- 多种计费方式（Token、次数、时长）
- 计费规则配置
- 消费记录查询
- 账单生成与导出
- 套餐管理

## 从 new-api 借鉴的设计

### 1. 架构设计

#### 1.1 项目结构（借鉴 new-api）
```
mrchat/
├── main.go                    # 程序入口
├── go.mod                     # Go 依赖管理
├── docker-compose.yml         # Docker 部署
│
├── common/                    # 公共工具库（无业务依赖）
│   ├── constants.go          # 全局常量
│   ├── redis.go              # Redis 客户端
│   └── limiter/              # 限流器
│
├── constant/                  # 常量定义（无依赖原则）
│   ├── model.go              # 模型类型常量
│   ├── api_type.go           # API类型枚举
│   └── context_key.go        # 上下文键
│
├── model/                     # 数据模型层（ORM）
│   ├── main.go               # 数据库初始化
│   ├── user.go               # 用户模型
│   ├── conversation.go       # 对话会话模型
│   ├── message.go            # 消息模型
│   ├── token.go              # 令牌模型
│   ├── log.go                # 日志模型
│   ├── option.go             # 配置项模型
│   ├── pricing.go            # 定价模型
│   ├── topup.go              # 充值模型
│   └── *_cache.go            # 缓存层
│
├── controller/                # 控制器层（HTTP 处理）
│   ├── user.go               # 用户管理
│   ├── conversation.go       # 对话管理
│   ├── message.go            # 消息管理
│   ├── chat.go               # 聊天接口
│   ├── billing.go            # 计费管理
│   └── topup.go              # 充值管理
│
├── middleware/                # 中间件层
│   ├── auth.go               # 认证中间件
│   ├── rate-limit.go         # 限流中间件
│   └── logging.go            # 日志中间件
│
├── router/                    # 路由层
│   ├── main.go               # 路由总入口
│   ├── api-router.go         # API 路由
│   └── web-router.go         # 前端路由
│
├── service/                   # 业务服务层
│   ├── conversation.go       # 对话业务逻辑
│   ├── quota.go              # 额度计算
│   ├── ai_proxy.go           # AI API 代理
│   └── token_counter.go      # Token 计数
│
├── setting/                   # 配置管理
│   ├── config/               # 配置体系
│   ├── ratio_setting/        # 计费比率
│   └── system_setting/       # 系统配置
│
├── dto/                       # 数据传输对象
│   ├── user_dto.go
│   ├── conversation_dto.go
│   └── message_dto.go
│
├── types/                     # 类型定义
│   ├── error.go              # 错误类型
│   └── chat_format.go        # 聊天格式
│
├── logger/                    # 日志系统
├── pkg/                       # 第三方包封装
└── web/                       # 前端代码（Vue 3）
    ├── src/
    │   ├── components/       # UI 组件
    │   ├── pages/            # 页面
    │   └── services/         # API 服务
    └── package.json          # 前端依赖
```

#### 1.2 技术栈（参考 new-api）

**后端核心框架**:
- **Gin** (v1.9.1): HTTP Web 框架
- **GORM** (v1.25.2): ORM 框架
  - 支持 MySQL/PostgreSQL/SQLite
- **go-redis** (v8): Redis 客户端

**推荐依赖库**:
```go
// HTTP & 网络
github.com/gin-gonic/gin                    // Web 框架
github.com/gin-contrib/sessions             // Session 管理
github.com/gin-contrib/cors                 // CORS
github.com/gorilla/websocket                // WebSocket

// 数据库 & 缓存
gorm.io/gorm                                // ORM
github.com/go-redis/redis/v8                // Redis

// 认证 & 安全
github.com/golang-jwt/jwt/v5                // JWT
golang.org/x/crypto                         // 加密库

// 支付（可选）
github.com/stripe/stripe-go/v81             // Stripe

// AI & Token
github.com/tiktoken-go/tokenizer            // Token 计数器

// 工具库
github.com/samber/lo                        // 函数式工具
github.com/shopspring/decimal               // 高精度计算
github.com/tidwall/gjson                    // JSON 解析
github.com/google/uuid                      // UUID
github.com/joho/godotenv                    // .env 文件
```

**前端技术栈**（已敲定 Vue 3 + Vite，生态选型后续再补一份 ADR 或在实现时收敛）:
```json
{
  "vue": "^3.x",
  "vue-router": "^4.x",
  "pinia": "^2.x",
  "axios": "^1.x",
  "vue-i18n": "^9.x"
}
```

### 2. 数据库设计（借鉴 new-api）

#### 2.1 核心数据表

##### users（用户表）
```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(20) UNIQUE NOT NULL COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码（哈希）',
    display_name VARCHAR(50) COMMENT '显示名称',
    email VARCHAR(100) UNIQUE COMMENT '邮箱',

    -- 角色与状态
    role INT NOT NULL DEFAULT 10 COMMENT '角色：1=Root, 10=Admin, 100=User',
    status INT NOT NULL DEFAULT 1 COMMENT '状态：1=启用, 2=禁用',

    -- OAuth 登录（可选）
    github_id VARCHAR(50) UNIQUE COMMENT 'GitHub ID',
    discord_id VARCHAR(50) UNIQUE COMMENT 'Discord ID',

    -- 额度管理
    quota BIGINT NOT NULL DEFAULT 0 COMMENT '剩余额度',
    used_quota BIGINT NOT NULL DEFAULT 0 COMMENT '已用额度',
    request_count INT NOT NULL DEFAULT 0 COMMENT '请求次数',

    -- 分组与邀请
    `group` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '用户分组',
    aff_code VARCHAR(32) UNIQUE COMMENT '邀请码',
    aff_count INT NOT NULL DEFAULT 0 COMMENT '邀请人数',
    inviter_id INT COMMENT '邀请人ID',

    -- 其他
    setting TEXT COMMENT '用户设置（JSON）',
    created_at BIGINT NOT NULL COMMENT '创建时间（Unix时间戳）',
    updated_at BIGINT COMMENT '更新时间',

    INDEX idx_username (username),
    INDEX idx_group (`group`),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

##### conversations（对话会话表）
```sql
CREATE TABLE conversations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    title VARCHAR(255) NOT NULL COMMENT '对话标题',

    -- 模型配置
    model_name VARCHAR(50) NOT NULL COMMENT '使用的模型',
    system_prompt TEXT COMMENT '系统提示词',

    -- 状态与分类
    status INT NOT NULL DEFAULT 1 COMMENT '状态：1=正常, 2=已删除',
    tags VARCHAR(255) COMMENT '标签（逗号分隔）',

    -- 统计信息
    message_count INT NOT NULL DEFAULT 0 COMMENT '消息数量',
    total_tokens INT NOT NULL DEFAULT 0 COMMENT '总Token数',

    -- 分享
    share_code VARCHAR(32) UNIQUE COMMENT '分享码',
    is_public TINYINT(1) DEFAULT 0 COMMENT '是否公开',

    -- 时间戳
    created_at BIGINT NOT NULL COMMENT '创建时间',
    updated_at BIGINT COMMENT '更新时间',
    last_message_at BIGINT COMMENT '最后消息时间',

    INDEX idx_user_id (user_id),
    INDEX idx_user_updated (user_id, updated_at),
    INDEX idx_share_code (share_code),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对话会话表';
```

##### messages（消息表）
```sql
CREATE TABLE messages (
    id INT PRIMARY KEY AUTO_INCREMENT,
    conversation_id INT NOT NULL COMMENT '对话ID',
    user_id INT NOT NULL COMMENT '用户ID',

    -- 消息内容
    role VARCHAR(20) NOT NULL COMMENT '角色：user/assistant/system',
    content TEXT NOT NULL COMMENT '消息内容',

    -- Token统计
    prompt_tokens INT NOT NULL DEFAULT 0 COMMENT '输入Token数',
    completion_tokens INT NOT NULL DEFAULT 0 COMMENT '输出Token数',
    total_tokens INT NOT NULL DEFAULT 0 COMMENT '总Token数',

    -- 模型信息
    model_name VARCHAR(50) COMMENT '使用的模型',

    -- 额度消耗
    quota_used INT NOT NULL DEFAULT 0 COMMENT '消耗的额度',

    -- 时间戳
    created_at BIGINT NOT NULL COMMENT '创建时间',

    INDEX idx_conversation_id (conversation_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';
```

##### tokens（令牌表）
```sql
CREATE TABLE tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    `key` CHAR(48) UNIQUE NOT NULL COMMENT 'API Key（sk-xxx）',
    name VARCHAR(50) COMMENT '令牌名称',

    -- 状态
    status INT NOT NULL DEFAULT 1 COMMENT '状态：1=启用, 2=禁用, 3=已过期',

    -- 额度限制
    remain_quota BIGINT NOT NULL DEFAULT 0 COMMENT '剩余额度',
    unlimited_quota TINYINT(1) DEFAULT 0 COMMENT '无限额度',

    -- 模型限制
    model_limits_enabled TINYINT(1) DEFAULT 0 COMMENT '启用模型限制',
    model_limits VARCHAR(1024) COMMENT '允许的模型（逗号分隔）',

    -- 分组
    `group` VARCHAR(64) DEFAULT 'default' COMMENT '分组',

    -- 过期时间
    expired_time BIGINT COMMENT '过期时间（Unix时间戳）',

    -- 时间戳
    created_at BIGINT NOT NULL COMMENT '创建时间',
    accessed_at BIGINT COMMENT '最后访问时间',

    INDEX idx_user_id (user_id),
    INDEX idx_key (`key`),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='令牌表';
```

##### logs（日志表）
```sql
CREATE TABLE logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    token_id INT COMMENT '令牌ID',
    conversation_id INT COMMENT '对话ID',

    -- 类型
    type INT NOT NULL COMMENT '日志类型：1=消费, 2=充值, 3=管理操作',

    -- 模型信息
    model_name VARCHAR(50) COMMENT '模型名称',

    -- 额度消耗
    quota INT NOT NULL DEFAULT 0 COMMENT '额度变动',
    prompt_tokens INT COMMENT '输入Token',
    completion_tokens INT COMMENT '输出Token',

    -- 其他
    content TEXT COMMENT '日志内容/备注',
    ip VARCHAR(50) COMMENT 'IP地址',

    -- 时间戳
    created_at BIGINT NOT NULL COMMENT '创建时间',

    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_type (type),
    INDEX idx_user_model (user_id, model_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志表';
```

##### topups（充值表）
```sql
CREATE TABLE topups (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',

    -- 金额
    amount BIGINT NOT NULL COMMENT '充值额度',
    money DECIMAL(10, 2) NOT NULL COMMENT '金额（元）',

    -- 支付信息
    trade_no VARCHAR(255) UNIQUE COMMENT '交易号',
    payment_method VARCHAR(50) COMMENT '支付方式',

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/completed/failed',

    -- 时间戳
    created_at BIGINT NOT NULL COMMENT '创建时间',
    completed_at BIGINT COMMENT '完成时间',

    INDEX idx_user_id (user_id),
    INDEX idx_trade_no (trade_no),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值表';
```

##### options（配置表）
```sql
CREATE TABLE options (
    `key` VARCHAR(255) PRIMARY KEY COMMENT '配置键',
    value TEXT COMMENT '配置值',
    updated_at BIGINT COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置表';
```

##### models（模型定价表）
```sql
CREATE TABLE models (
    id INT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) UNIQUE NOT NULL COMMENT '模型名称',

    -- 计费类型
    quota_type INT NOT NULL DEFAULT 1 COMMENT '1=Token计费, 2=次数计费, 3=时长计费',

    -- 计费参数
    model_ratio FLOAT NOT NULL DEFAULT 1.0 COMMENT '模型倍率',
    model_price FLOAT COMMENT '单价（$/1M token）',
    completion_ratio FLOAT NOT NULL DEFAULT 1.0 COMMENT '输出倍率',

    -- 启用分组
    enable_groups TEXT COMMENT '启用的分组（JSON数组）',

    -- 其他
    status INT NOT NULL DEFAULT 1 COMMENT '状态：1=启用, 2=禁用',
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    INDEX idx_model_name (model_name),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模型定价表';
```

#### 2.2 数据库设计要点

**索引优化**（借鉴 new-api）:
- 用户查询高频字段：username, group, status
- 对话查询：user_id + updated_at 组合索引
- 消息查询：conversation_id, created_at
- 日志查询：user_id, created_at, type

**分表策略**（未来扩展）:
- messages 表按月分表（message_202601, message_202602...）
- logs 表按月分表

**缓存策略**（借鉴 new-api）:
```
Level 1: Memory Cache (60秒同步)
    ↓
Level 2: Redis Cache (5分钟TTL)
    ↓
Level 3: Database
```

### 3. 计费管理（重点借鉴 new-api）

#### 3.1 计费公式

```
最终额度 = (输入Token × 模型倍率 + 输出Token × 完成倍率) × 分组倍率 × 汇率
```

**示例**:
```
模型: gpt-4
输入: 1000 tokens
输出: 500 tokens
模型倍率: 30
完成倍率: 2
分组倍率: 1.2（VIP用户打8折，则为0.8）
汇率: 7（美元换人民币）

计算:
基础消耗 = 1000 × 30 + 500 × 2 × 30 = 30000 + 30000 = 60000
最终额度 = 60000 × 1.2 × 7 = 504000
```

#### 3.2 计费流程（借鉴 new-api 的预扣模式）

```go
// 1. 预扣额度（按最大Token估算）
func PreConsumeQuota(userId int, estimatedTokens int) error {
    estimatedQuota := CalculateQuota(estimatedTokens, modelName)

    // 检查余额
    if user.Quota < estimatedQuota {
        return errors.New("余额不足")
    }

    // 预扣
    user.Quota -= estimatedQuota
    user.Save()

    // 记录预扣日志
    return nil
}

// 2. 调用 AI API
func CallAIAPI(request) (response, error) {
    // 调用上游API
    return aiClient.Chat(request)
}

// 3. 结算额度（根据实际使用）
func PostConsumeQuota(userId int, actualTokens int, preConsumed int) error {
    actualQuota := CalculateQuota(actualTokens, modelName)

    // 退款差额
    refund := preConsumed - actualQuota
    if refund > 0 {
        user.Quota += refund
        user.UsedQuota += actualQuota
    } else if refund < 0 {
        // 补扣（理论上不应发生）
        user.Quota -= (-refund)
        user.UsedQuota += actualQuota
    }

    user.Save()

    // 记录日志
    return nil
}
```

#### 3.3 计费类型支持

1. **Token 计费**: GPT、Claude 等（按输入/输出Token）
2. **次数计费**: 图像生成、TTS（按调用次数）
3. **时长计费**: 语音通话（按秒计费）

#### 3.4 分组倍率配置

```json
{
  "group_ratios": {
    "default": 1.0,
    "vip": 0.8,       // VIP 8折
    "premium": 0.6,   // Premium 6折
    "free": 1.5       // 免费用户 1.5倍（限制使用）
  }
}
```

### 4. 关键设计模式（借鉴 new-api）

#### 4.1 中间件模式（Middleware Pattern）

```go
// router/api-router.go
apiRouter := r.Group("/api/v1")
apiRouter.Use(middleware.TokenAuth())          // 认证
apiRouter.Use(middleware.RateLimit())          // 限流
apiRouter.Use(middleware.Logging())            // 日志

apiRouter.POST("/chat/completions", controller.Chat)
```

#### 4.2 缓存模式（Cache Pattern）

```go
// model/user_cache.go
var userCache = make(map[int]*User)
var cacheMutex sync.RWMutex

func GetUserById(id int) *User {
    // 1. 尝试从内存缓存获取
    cacheMutex.RLock()
    if user, ok := userCache[id]; ok {
        cacheMutex.RUnlock()
        return user
    }
    cacheMutex.RUnlock()

    // 2. 尝试从Redis获取
    user := GetUserFromRedis(id)
    if user != nil {
        // 写入内存缓存
        cacheMutex.Lock()
        userCache[id] = user
        cacheMutex.Unlock()
        return user
    }

    // 3. 从数据库获取
    DB.First(&user, id)

    // 写入缓存
    SetUserToRedis(id, user)
    cacheMutex.Lock()
    userCache[id] = user
    cacheMutex.Unlock()

    return user
}

// 定时同步缓存
func SyncUserCache(frequency int) {
    ticker := time.NewTicker(time.Duration(frequency) * time.Second)
    for range ticker.C {
        var users []*User
        DB.Find(&users)

        cacheMutex.Lock()
        userCache = make(map[int]*User)
        for _, user := range users {
            userCache[user.Id] = user
        }
        cacheMutex.Unlock()
    }
}
```

#### 4.3 单例模式（Singleton Pattern）

```go
// model/main.go
var DB *gorm.DB  // 全局数据库连接

func InitDB(dsn string) error {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return err
    }

    DB = db

    // 设置连接池
    sqlDB, _ := DB.DB()
    sqlDB.SetMaxIdleConns(100)
    sqlDB.SetMaxOpenConns(1000)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return nil
}
```

### 5. AI API 代理设计

虽然我们不像 new-api 那样做复杂的多渠道中转，但需要代理 AI API 请求以实现计费和日志记录。

#### 5.1 简化版代理架构

```go
// service/ai_proxy.go
type AIProxy struct {
    client *http.Client
}

func (p *AIProxy) ChatCompletion(ctx *gin.Context, req *dto.ChatRequest) (*dto.ChatResponse, error) {
    user := ctx.MustGet("user").(*model.User)

    // 1. Token 计数（输入）
    inputTokens := tokenizer.CountTokens(req.Messages)

    // 2. 预扣额度
    estimatedQuota := service.CalculateQuota(inputTokens, req.Model, user.Group)
    err := service.PreConsumeQuota(user.Id, estimatedQuota)
    if err != nil {
        return nil, err
    }

    // 3. 调用 OpenAI API（或其他AI服务）
    resp, err := p.callOpenAI(req)
    if err != nil {
        // 退款
        service.RefundQuota(user.Id, estimatedQuota)
        return nil, err
    }

    // 4. 结算额度（根据实际使用）
    actualQuota := service.CalculateQuota(resp.Usage.TotalTokens, req.Model, user.Group)
    service.PostConsumeQuota(user.Id, actualQuota, estimatedQuota)

    // 5. 记录日志
    model.CreateLog(&model.Log{
        UserId:           user.Id,
        Type:             1, // 消费
        ModelName:        req.Model,
        Quota:            actualQuota,
        PromptTokens:     resp.Usage.PromptTokens,
        CompletionTokens: resp.Usage.CompletionTokens,
        CreatedAt:        time.Now().Unix(),
    })

    return resp, nil
}

func (p *AIProxy) callOpenAI(req *dto.ChatRequest) (*dto.ChatResponse, error) {
    // 调用 OpenAI API
    url := "https://api.openai.com/v1/chat/completions"

    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

    httpResp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer httpResp.Body.Close()

    var resp dto.ChatResponse
    json.NewDecoder(httpResp.Body).Decode(&resp)

    return &resp, nil
}
```

#### 5.2 流式响应处理

```go
func (p *AIProxy) ChatCompletionStream(ctx *gin.Context, req *dto.ChatRequest) error {
    user := ctx.MustGet("user").(*model.User)

    // 预扣额度
    // ... (同上)

    // 调用 OpenAI Stream API
    httpResp, err := p.callOpenAIStream(req)
    if err != nil {
        return err
    }
    defer httpResp.Body.Close()

    // 设置 SSE 响应头
    ctx.Header("Content-Type", "text/event-stream")
    ctx.Header("Cache-Control", "no-cache")
    ctx.Header("Connection", "keep-alive")

    // 流式转发
    scanner := bufio.NewScanner(httpResp.Body)
    var totalTokens int

    for scanner.Scan() {
        line := scanner.Text()

        // 解析 data: {...}
        if strings.HasPrefix(line, "data: ") {
            ctx.Writer.WriteString(line + "\n\n")
            ctx.Writer.Flush()

            // 统计Token（从最后一条消息获取）
            if strings.Contains(line, "usage") {
                var chunk dto.ChatStreamChunk
                json.Unmarshal([]byte(line[6:]), &chunk)
                if chunk.Usage != nil {
                    totalTokens = chunk.Usage.TotalTokens
                }
            }
        }
    }

    // 结算额度
    service.PostConsumeQuota(user.Id, totalTokens, estimatedQuota)

    return nil
}
```

### 6. 关键功能实现要点

#### 6.1 用户分组与权限

**用户分组**:
- default: 普通用户
- vip: VIP用户（优惠价格）
- premium: 高级用户（更多优惠）
- free: 免费试用（限制额度）

**权限控制**:
```go
// middleware/auth.go
func RequireRole(role int) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*model.User)
        if user.Role > role {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用
adminRouter.Use(middleware.RequireRole(constant.RoleAdmin))
```

#### 6.2 对话历史管理

**设计要点**:
1. 对话与消息分表存储
2. 支持分页查询历史消息
3. 软删除（status=2）而非物理删除
4. 定期清理过期数据（管理员配置）

```go
// service/conversation.go
func GetConversationHistory(conversationId int, limit int, offset int) ([]*model.Message, error) {
    var messages []*model.Message
    err := model.DB.Where("conversation_id = ?", conversationId).
        Order("created_at ASC").
        Limit(limit).
        Offset(offset).
        Find(&messages).Error

    return messages, err
}
```

#### 6.3 Token 计数

使用 tiktoken-go 库计算Token数:

```go
// service/token_counter.go
import "github.com/tiktoken-go/tokenizer"

func CountTokens(messages []dto.Message, model string) int {
    enc, err := tokenizer.ForModel(tokenizer.Model(model))
    if err != nil {
        // 降级：简单估算
        return estimateTokens(messages)
    }

    var total int
    for _, msg := range messages {
        tokens, _ := enc.Encode(msg.Content)
        total += len(tokens)
    }

    return total
}

func estimateTokens(messages []dto.Message) int {
    // 简单估算：1 token ≈ 4 字符
    var total int
    for _, msg := range messages {
        total += len(msg.Content) / 4
    }
    return total
}
```

#### 6.4 限流策略

**多级限流**:
1. **用户级**: 每个用户每分钟最多 X 次请求
2. **IP 级**: 每个 IP 每分钟最多 Y 次请求
3. **模型级**: 某些昂贵模型单独限流

```go
// middleware/rate-limit.go
import "golang.org/x/time/rate"

var userLimiters = make(map[int]*rate.Limiter)
var limiterMutex sync.RWMutex

func RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*model.User)

        limiterMutex.Lock()
        limiter, ok := userLimiters[user.Id]
        if !ok {
            // 每秒1个请求，突发10个
            limiter = rate.NewLimiter(1, 10)
            userLimiters[user.Id] = limiter
        }
        limiterMutex.Unlock()

        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "请求过于频繁"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 7. 与 new-api 的关键差异

| 功能 | new-api | MrChat |
|------|---------|--------|
| **核心定位** | AI网关+资产管理 | 用户对话管理+计费 |
| **渠道管理** | 58种渠道，复杂分发 | 简化：单一或少量AI服务 |
| **对话管理** | 无（纯中转） | **核心功能**：会话、历史、分享 |
| **计费** | 复杂（多渠道、多计费类型） | 借鉴：预扣+结算模式 |
| **用户体系** | 重管理（分组、邀请、多登录） | **完整保留** |
| **前端** | 管理后台为主 | **用户聊天界面为主** + 管理后台 |
| **适配器模式** | 40+适配器 | 不需要（或仅2-3个） |
| **WebSocket** | 无 | **需要**（实时聊天） |

### 8. 开发路线图

#### Phase 1: 基础架构（1-2周）
- [ ] 项目初始化（Go + Gin）
- [ ] 数据库设计与迁移（GORM）
- [ ] 用户管理模块（注册、登录、JWT）
- [ ] 配置管理（环境变量、数据库配置）
- [ ] Redis 缓存集成
- [ ] 基础中间件（认证、日志、CORS）

#### Phase 2: 对话管理（2-3周）
- [ ] 对话会话 CRUD
- [ ] 消息存储与查询
- [ ] 对话历史分页
- [ ] 对话分组/标签
- [ ] 对话分享功能
- [ ] WebSocket 实时通信

#### Phase 3: AI 集成与计费（2周）
- [ ] OpenAI API 代理
- [ ] Token 计数器
- [ ] 计费规则引擎
- [ ] 预扣/结算流程
- [ ] 日志记录
- [ ] 余额管理

#### Phase 4: 前端开发（3-4周）
- [ ] Vue 3 + Vite 项目初始化
- [ ] 用户登录/注册页面
- [ ] 聊天主界面
- [ ] 对话历史侧边栏
- [ ] 设置页面（个人信息、API Key）
- [ ] 管理后台（用户管理、计费配置）

#### Phase 5: 高级功能（2周）
- [ ] 充值管理（支付接入）
- [ ] 邀请机制
- [ ] 用户统计看板
- [ ] 模型切换
- [ ] Markdown 渲染
- [ ] 对话导出

#### Phase 6: 测试与部署（1-2周）
- [ ] 单元测试
- [ ] 压力测试
- [ ] Docker 部署
- [ ] 文档编写

### 9. 技术细节参考

#### 9.1 环境配置（.env 文件）

```env
# 服务器配置
PORT=3000
GIN_MODE=release

# 数据库配置
SQL_DSN=root:password@tcp(localhost:3306)/mrchat?charset=utf8mb4&parseTime=True&loc=Local

# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT 配置
JWT_SECRET=your-secret-key

# AI API 配置
OPENAI_API_KEY=sk-xxxx
OPENAI_BASE_URL=https://api.openai.com/v1

# 计费配置
DEFAULT_QUOTA=100000  # 新用户初始额度
EXCHANGE_RATE=7       # 美元换人民币汇率

# 限流配置
RATE_LIMIT_PER_MINUTE=60

# 缓存配置
CACHE_SYNC_INTERVAL=60  # 秒
```

#### 9.2 项目初始化命令

```bash
# 创建项目
mkdir mrchat && cd mrchat
go mod init mrchat

# 安装依赖
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u github.com/go-redis/redis/v8
go get -u github.com/golang-jwt/jwt/v5
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/tiktoken-go/tokenizer
go get -u github.com/gorilla/websocket
go get -u github.com/joho/godotenv

# 创建目录结构
mkdir -p common constant model controller middleware router service dto types logger pkg web
```

#### 9.3 数据库初始化

```go
// model/main.go
func InitDB() error {
    dsn := os.Getenv("SQL_DSN")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    DB = db

    // 自动迁移
    err = DB.AutoMigrate(
        &User{},
        &Conversation{},
        &Message{},
        &Token{},
        &Log{},
        &Topup{},
        &Option{},
        &Model{},
    )

    return err
}
```

### 10. 待讨论的问题

1. **AI 服务商选择**
   - 仅支持 OpenAI？
   - 还是多支持几个（Claude、Gemini）？

2. **支付接入**
   - Stripe（国际）
   - 易支付（国内）
   - 还是暂不接入，手动充值？

3. **前端框架**
   - Semi Design（字节）
   - Ant Design（阿里）
   - 还是 Material-UI？

4. **部署方式**
   - Docker Compose（单机）
   - Kubernetes（集群）
   - 还是云服务（Vercel/Railway）？

5. **多租户支持**
   - 是否支持多个独立空间（类似 Slack Workspace）？

---

## 总结

本设计文档基于对 new-api 项目的深入分析，借鉴了其优秀的架构设计和实现思路，同时结合 MrChat 的核心需求（对话管理）进行了针对性调整。

**核心借鉴**:
- 数据库设计（用户、额度、日志）
- 计费管理（预扣/结算模式）
- 缓存架构（多级缓存）
- 中间件模式（认证、限流）

**核心差异**:
- 聚焦对话管理而非渠道中转
- 前端以用户聊天界面为主
- 简化渠道管理，降低复杂度

**下一步**:
请确认以上设计方案，我们将进入具体开发阶段。

---

**文档版本**: v1.0
**创建时间**: 2026-01-21
**最后更新**: 2026-01-21
