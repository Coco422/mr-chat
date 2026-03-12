# MrChat 架构设计 - 融合 Open WebUI 优秀设计

> **基于**: new-api + Open WebUI 最佳实践
> **技术栈**: Go + Vue + MySQL/PostgreSQL + Redis
> **更新时间**: 2026-01-27

---

## 📋 目录

1. [核心设计理念](#核心设计理念)
2. [多租户与权限系统](#多租户与权限系统)
3. [模型管理与集成](#模型管理与集成)
4. [数据库架构优化](#数据库架构优化)
5. [实时协作能力](#实时协作能力)
6. [API 设计规范](#api-设计规范)
7. [缓存与性能优化](#缓存与性能优化)
8. [安全与审计](#安全与审计)

---

## 🎯 核心设计理念

### 从 Open WebUI 学到的关键原则

#### 1. **用户与认证分离**
```
User 表 (用户信息)
  ↓
Auth 表 (认证凭据)
  ↓
ApiKey 表 (API 密钥)
```

**优势**:
- 支持一个用户多种登录方式（密码、OAuth、LDAP、SSO）
- 认证方式变更不影响用户数据
- 便于实现企业级 SSO 集成

#### 2. **JSON 字段的灵活性**
```go
// 用户设置存储为 JSON
type User struct {
    ID       int
    Username string
    Settings json.RawMessage  // 灵活的用户配置
    Info     json.RawMessage  // 扩展信息
}
```

**优势**:
- 快速迭代新功能，无需频繁迁移数据库
- 用户个性化配置灵活存储
- 减少表字段膨胀

#### 3. **基于组的权限控制 (RBAC + Groups)**
```
User → Groups (多对多)
  ↓
Group → Permissions (JSON)
  ↓
Resource Access Control
```

**优势**:
- 比单纯的角色系统更灵活
- 支持团队协作场景
- 便于实现资源共享

---

## 🔐 多租户与权限系统

### 1. 数据库设计（融合 OUI 设计）

#### 1.1 用户表（增强版）

```sql
CREATE TABLE users (
    -- 基础信息
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),  -- 使用 UUID 而非自增 ID
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100),

    -- 个人资料（借鉴 OUI）
    profile_image_url VARCHAR(500),
    bio TEXT,
    timezone VARCHAR(50) DEFAULT 'UTC',
    locale VARCHAR(10) DEFAULT 'zh-CN',

    -- 角色与状态
    role INT NOT NULL DEFAULT 100 COMMENT '1=Root, 10=Admin, 100=User',
    status INT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Disabled, 3=Pending',

    -- 额度管理
    quota BIGINT NOT NULL DEFAULT 0,
    used_quota BIGINT NOT NULL DEFAULT 0,
    request_count INT NOT NULL DEFAULT 0,

    -- 分组
    primary_group VARCHAR(64) NOT NULL DEFAULT 'default',

    -- 邀请系统
    aff_code VARCHAR(32) UNIQUE,
    inviter_id VARCHAR(36),

    -- 灵活配置（JSON）
    settings JSON COMMENT '用户偏好设置',
    metadata JSON COMMENT '扩展元数据',

    -- 在线状态（借鉴 OUI）
    last_active_at BIGINT,

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_primary_group (primary_group),
    INDEX idx_status (status),
    INDEX idx_last_active (last_active_at),
    FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 1.2 认证表（分离设计）

```sql
CREATE TABLE auths (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36) NOT NULL,

    -- 密码认证
    password_hash VARCHAR(255),

    -- OAuth 提供商
    oauth_provider VARCHAR(50) COMMENT 'github, google, microsoft, etc.',
    oauth_id VARCHAR(255),
    oauth_data JSON,

    -- 状态
    active BOOLEAN DEFAULT TRUE,
    verified BOOLEAN DEFAULT FALSE,

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    UNIQUE KEY uk_user_provider (user_id, oauth_provider),
    UNIQUE KEY uk_oauth (oauth_provider, oauth_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 1.3 组管理表（新增）

```sql
CREATE TABLE groups (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- 所有者
    owner_id VARCHAR(36) NOT NULL,

    -- 权限配置（JSON）
    permissions JSON COMMENT '组权限配置',

    -- 计费倍率
    quota_multiplier FLOAT DEFAULT 1.0 COMMENT '额度倍率（折扣）',

    -- 状态
    status INT DEFAULT 1,

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    INDEX idx_owner (owner_id),
    INDEX idx_name (name),
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE group_members (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,

    -- 成员角色
    role VARCHAR(20) DEFAULT 'member' COMMENT 'owner, admin, member',

    -- 时间戳
    joined_at BIGINT NOT NULL,

    UNIQUE KEY uk_group_user (group_id, user_id),
    INDEX idx_group_id (group_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 1.4 API 密钥表（增强版）

```sql
CREATE TABLE api_keys (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36) NOT NULL,

    -- 密钥信息
    `key` VARCHAR(64) UNIQUE NOT NULL COMMENT 'sk-xxx',
    name VARCHAR(100),

    -- 状态
    status INT DEFAULT 1 COMMENT '1=Active, 2=Disabled, 3=Expired',

    -- 额度限制
    quota_limit BIGINT DEFAULT 0 COMMENT '0=无限制',
    quota_used BIGINT DEFAULT 0,

    -- 速率限制
    rate_limit INT DEFAULT 60 COMMENT '每分钟请求数',

    -- 模型访问控制
    allowed_models JSON COMMENT '允许的模型列表',

    -- IP 白名单
    ip_whitelist JSON COMMENT 'IP 白名单',

    -- 过期时间
    expires_at BIGINT,

    -- 时间戳
    created_at BIGINT NOT NULL,
    last_used_at BIGINT,

    INDEX idx_user_id (user_id),
    INDEX idx_key (`key`),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. Go 模型定义

#### 2.1 用户模型

```go
// model/user.go
package model

import (
    "encoding/json"
    "time"
)

type User struct {
    ID              string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
    Username        string          `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email           string          `gorm:"uniqueIndex;size:100;not null" json:"email"`
    DisplayName     string          `gorm:"size:100" json:"display_name"`

    // 个人资料
    ProfileImageURL string          `gorm:"size:500" json:"profile_image_url"`
    Bio             string          `gorm:"type:text" json:"bio"`
    Timezone        string          `gorm:"size:50;default:UTC" json:"timezone"`
    Locale          string          `gorm:"size:10;default:zh-CN" json:"locale"`

    // 角色与状态
    Role            int             `gorm:"not null;default:100" json:"role"`
    Status          int             `gorm:"not null;default:1;index" json:"status"`

    // 额度
    Quota           int64           `gorm:"not null;default:0" json:"quota"`
    UsedQuota       int64           `gorm:"not null;default:0" json:"used_quota"`
    RequestCount    int             `gorm:"not null;default:0" json:"request_count"`

    // 分组
    PrimaryGroup    string          `gorm:"size:64;not null;default:default;index" json:"primary_group"`

    // 邀请
    AffCode         string          `gorm:"size:32;uniqueIndex" json:"aff_code,omitempty"`
    InviterID       *string         `gorm:"type:varchar(36)" json:"inviter_id,omitempty"`

    // JSON 字段
    Settings        json.RawMessage `gorm:"type:json" json:"settings,omitempty"`
    Metadata        json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`

    // 在线状态
    LastActiveAt    *int64          `gorm:"index" json:"last_active_at,omitempty"`

    // 时间戳
    CreatedAt       int64           `gorm:"not null" json:"created_at"`
    UpdatedAt       int64           `json:"updated_at,omitempty"`

    // 关联
    Groups          []Group         `gorm:"many2many:group_members;" json:"groups,omitempty"`
    Auths           []Auth          `gorm:"foreignKey:UserID" json:"-"`
    ApiKeys         []ApiKey        `gorm:"foreignKey:UserID" json:"-"`
}

// 用户设置结构
type UserSettings struct {
    Theme           string `json:"theme"`              // light, dark, auto
    Language        string `json:"language"`           // zh-CN, en-US
    DefaultModel    string `json:"default_model"`      // 默认模型
    StreamEnabled   bool   `json:"stream_enabled"`     // 是否启用流式输出
    NotifyEnabled   bool   `json:"notify_enabled"`     // 是否启用通知
}

// BeforeCreate 钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    if u.CreatedAt == 0 {
        u.CreatedAt = time.Now().Unix()
    }
    return nil
}

// 获取用户设置
func (u *User) GetSettings() (*UserSettings, error) {
    if u.Settings == nil {
        return &UserSettings{}, nil
    }

    var settings UserSettings
    err := json.Unmarshal(u.Settings, &settings)
    return &settings, err
}

// 更新用户设置
func (u *User) UpdateSettings(settings *UserSettings) error {
    data, err := json.Marshal(settings)
    if err != nil {
        return err
    }
    u.Settings = data
    return nil
}

// 检查是否在线（3分钟内活跃）
func (u *User) IsOnline() bool {
    if u.LastActiveAt == nil {
        return false
    }
    return time.Now().Unix()-*u.LastActiveAt < 180
}
```

#### 2.2 认证模型

```go
// model/auth.go
package model

type Auth struct {
    ID            string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
    UserID        string          `gorm:"type:varchar(36);not null;index" json:"user_id"`

    // 密码认证
    PasswordHash  string          `gorm:"size:255" json:"-"`

    // OAuth
    OAuthProvider string          `gorm:"size:50" json:"oauth_provider,omitempty"`
    OAuthID       string          `gorm:"size:255" json:"oauth_id,omitempty"`
    OAuthData     json.RawMessage `gorm:"type:json" json:"oauth_data,omitempty"`

    // 状态
    Active        bool            `gorm:"default:true" json:"active"`
    Verified      bool            `gorm:"default:false" json:"verified"`

    // 时间戳
    CreatedAt     int64           `gorm:"not null" json:"created_at"`
    UpdatedAt     int64           `json:"updated_at,omitempty"`

    // 关联
    User          User            `gorm:"foreignKey:UserID" json:"-"`
}

// OAuth 数据结构
type OAuthData struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
    ExpiresAt    int64  `json:"expires_at,omitempty"`
    Email        string `json:"email,omitempty"`
    Name         string `json:"name,omitempty"`
    Avatar       string `json:"avatar,omitempty"`
}
```

#### 2.3 组模型

```go
// model/group.go
package model

type Group struct {
    ID               string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
    Name             string          `gorm:"size:100;not null;index" json:"name"`
    Description      string          `gorm:"type:text" json:"description,omitempty"`

    // 所有者
    OwnerID          string          `gorm:"type:varchar(36);not null;index" json:"owner_id"`

    // 权限配置
    Permissions      json.RawMessage `gorm:"type:json" json:"permissions,omitempty"`

    // 计费倍率
    QuotaMultiplier  float64         `gorm:"default:1.0" json:"quota_multiplier"`

    // 状态
    Status           int             `gorm:"default:1" json:"status"`

    // 时间戳
    CreatedAt        int64           `gorm:"not null" json:"created_at"`
    UpdatedAt        int64           `json:"updated_at,omitempty"`

    // 关联
    Owner            User            `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
    Members          []User          `gorm:"many2many:group_members;" json:"members,omitempty"`
}

type GroupMember struct {
    ID       string `gorm:"primaryKey;type:varchar(36)" json:"id"`
    GroupID  string `gorm:"type:varchar(36);not null;index" json:"group_id"`
    UserID   string `gorm:"type:varchar(36);not null;index" json:"user_id"`
    Role     string `gorm:"size:20;default:member" json:"role"` // owner, admin, member
    JoinedAt int64  `gorm:"not null" json:"joined_at"`
}

// 组权限结构
type GroupPermissions struct {
    CanCreateChat    bool     `json:"can_create_chat"`
    CanShareChat     bool     `json:"can_share_chat"`
    CanUseModels     []string `json:"can_use_models"`
    MaxTokensPerDay  int      `json:"max_tokens_per_day"`
    MaxChatsPerDay   int      `json:"max_chats_per_day"`
}
```

### 3. 权限控制中间件

```go
// middleware/permission.go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "mrchat/model"
)

// 角色常量
const (
    RoleRoot  = 1
    RoleAdmin = 10
    RoleUser  = 100
)

// 要求特定角色
func RequireRole(minRole int) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
            c.Abort()
            return
        }

        u := user.(*model.User)
        if u.Role > minRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// 检查组成员资格
func RequireGroupMember(groupID string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*model.User)

        var member model.GroupMember
        err := model.DB.Where("group_id = ? AND user_id = ?", groupID, user.ID).
            First(&member).Error

        if err != nil {
            c.JSON(http.StatusForbidden, gin.H{"error": "不是组成员"})
            c.Abort()
            return
        }

        c.Set("group_member", &member)
        c.Next()
    }
}

// 检查模型访问权限
func CheckModelAccess(modelName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*model.User)

        // Root 和 Admin 无限制
        if user.Role <= RoleAdmin {
            c.Next()
            return
        }

        // 检查用户组权限
        hasAccess := checkUserModelAccess(user.ID, modelName)
        if !hasAccess {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "无权访问此模型",
                "model": modelName,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

func checkUserModelAccess(userID, modelName string) bool {
    // 查询用户所属组的权限
    var groups []model.Group
    model.DB.Joins("JOIN group_members ON groups.id = group_members.group_id").
        Where("group_members.user_id = ?", userID).
        Find(&groups)

    for _, group := range groups {
        var perms model.GroupPermissions
        if err := json.Unmarshal(group.Permissions, &perms); err == nil {
            for _, allowedModel := range perms.CanUseModels {
                if allowedModel == modelName || allowedModel == "*" {
                    return true
                }
            }
        }
    }

    return false
}
```

---

## 🤖 模型管理与集成

### 1. 模型配置表设计（借鉴 OUI）

```sql
CREATE TABLE models (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),

    -- 模型标识
    model_id VARCHAR(100) UNIQUE NOT NULL COMMENT '模型ID，如 gpt-4',
    model_name VARCHAR(200) NOT NULL COMMENT '显示名称',
    provider VARCHAR(50) NOT NULL COMMENT '提供商：openai, anthropic, google, etc.',

    -- 模型配置
    base_url VARCHAR(500) COMMENT 'API 基础URL',
    api_version VARCHAR(20) COMMENT 'API 版本',

    -- 能力标识
    capabilities JSON COMMENT '模型能力：vision, function_calling, streaming, etc.',

    -- 计费配置
    pricing JSON COMMENT '定价信息',
    quota_type INT DEFAULT 1 COMMENT '1=Token, 2=Request, 3=Time',

    -- 访问控制
    is_public BOOLEAN DEFAULT TRUE COMMENT '是否公开可用',
    allowed_groups JSON COMMENT '允许的用户组',

    -- 参数限制
    max_tokens INT DEFAULT 4096,
    context_length INT DEFAULT 4096,

    -- 系统提示词
    system_prompt TEXT COMMENT '默认系统提示词',

    -- 状态
    status INT DEFAULT 1 COMMENT '1=Active, 2=Disabled, 3=Maintenance',

    -- 元数据
    metadata JSON COMMENT '扩展元数据',

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    INDEX idx_model_id (model_id),
    INDEX idx_provider (provider),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. Go 模型定义

```go
// model/ai_model.go
package model

type AIModel struct {
    ID            string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
    ModelID       string          `gorm:"uniqueIndex;size:100;not null" json:"model_id"`
    ModelName     string          `gorm:"size:200;not null" json:"model_name"`
    Provider      string          `gorm:"size:50;not null;index" json:"provider"`

    // 配置
    BaseURL       string          `gorm:"size:500" json:"base_url,omitempty"`
    APIVersion    string          `gorm:"size:20" json:"api_version,omitempty"`

    // 能力
    Capabilities  json.RawMessage `gorm:"type:json" json:"capabilities,omitempty"`

    // 计费
    Pricing       json.RawMessage `gorm:"type:json" json:"pricing,omitempty"`
    QuotaType     int             `gorm:"default:1" json:"quota_type"`

    // 访问控制
    IsPublic      bool            `gorm:"default:true" json:"is_public"`
    AllowedGroups json.RawMessage `gorm:"type:json" json:"allowed_groups,omitempty"`

    // 参数
    MaxTokens     int             `gorm:"default:4096" json:"max_tokens"`
    ContextLength int             `gorm:"default:4096" json:"context_length"`

    // 系统提示词
    SystemPrompt  string          `gorm:"type:text" json:"system_prompt,omitempty"`

    // 状态
    Status        int             `gorm:"default:1;index" json:"status"`

    // 元数据
    Metadata      json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`

    // 时间戳
    CreatedAt     int64           `gorm:"not null" json:"created_at"`
    UpdatedAt     int64           `json:"updated_at,omitempty"`
}

// 模型能力
type ModelCapabilities struct {
    Vision          bool `json:"vision"`
    FunctionCalling bool `json:"function_calling"`
    Streaming       bool `json:"streaming"`
    JSON            bool `json:"json_mode"`
}

// 定价信息
type ModelPricing struct {
    InputPrice  float64 `json:"input_price"`   // 每百万 token 价格
    OutputPrice float64 `json:"output_price"`  // 每百万 token 价格
    Currency    string  `json:"currency"`      // USD, CNY
}
```

### 3. 统一模型接口（借鉴 OUI 的 OpenAI 兼容设计）

```go
// service/model_service.go
package service

import (
    "context"
    "encoding/json"
    "mrchat/model"
)

// 统一的聊天请求接口
type ChatRequest struct {
    Model       string         `json:"model"`
    Messages    []ChatMessage  `json:"messages"`
    Stream      bool           `json:"stream,omitempty"`
    Temperature float64        `json:"temperature,omitempty"`
    MaxTokens   int            `json:"max_tokens,omitempty"`
    TopP        float64        `json:"top_p,omitempty"`

    // 扩展参数
    SystemPrompt string        `json:"system_prompt,omitempty"`
    Tools        []Tool        `json:"tools,omitempty"`
}

type ChatMessage struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"`
}

type ChatResponse struct {
    ID      string        `json:"id"`
    Model   string        `json:"model"`
    Choices []Choice      `json:"choices"`
    Usage   UsageInfo     `json:"usage"`
}

type Choice struct {
    Index        int         `json:"index"`
    Message      ChatMessage `json:"message"`
    FinishReason string      `json:"finish_reason"`
}

type UsageInfo struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// 模型服务接口
type ModelService interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest, callback func(chunk string) error) error
    GetModels() ([]*model.AIModel, error)
    GetModel(modelID string) (*model.AIModel, error)
}
```

### 4. 多提供商适配器（Adapter Pattern）

```go
// service/adapter/base.go
package adapter

import (
    "context"
    "mrchat/service"
)

// 基础适配器接口
type ModelAdapter interface {
    Chat(ctx context.Context, req *service.ChatRequest) (*service.ChatResponse, error)
    ChatStream(ctx context.Context, req *service.ChatRequest, callback func(string) error) error
    ValidateConfig() error
}

// 适配器工厂
type AdapterFactory struct {
    adapters map[string]ModelAdapter
}

func NewAdapterFactory() *AdapterFactory {
    return &AdapterFactory{
        adapters: make(map[string]ModelAdapter),
    }
}

func (f *AdapterFactory) Register(provider string, adapter ModelAdapter) {
    f.adapters[provider] = adapter
}

func (f *AdapterFactory) GetAdapter(provider string) (ModelAdapter, error) {
    adapter, ok := f.adapters[provider]
    if !ok {
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }
    return adapter, nil
}
```

### 5. OpenAI 适配器实现

```go
// service/adapter/openai.go
package adapter

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "mrchat/service"
)

type OpenAIAdapter struct {
    APIKey  string
    BaseURL string
    client  *http.Client
}

func NewOpenAIAdapter(apiKey, baseURL string) *OpenAIAdapter {
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }

    return &OpenAIAdapter{
        APIKey:  apiKey,
        BaseURL: baseURL,
        client:  &http.Client{Timeout: 60 * time.Second},
    }
}

func (a *OpenAIAdapter) Chat(ctx context.Context, req *service.ChatRequest) (*service.ChatResponse, error) {
    // 构建 OpenAI 请求
    openaiReq := map[string]interface{}{
        "model":    req.Model,
        "messages": req.Messages,
    }

    if req.Temperature > 0 {
        openaiReq["temperature"] = req.Temperature
    }
    if req.MaxTokens > 0 {
        openaiReq["max_tokens"] = req.MaxTokens
    }

    // 发送请求
    body, _ := json.Marshal(openaiReq)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        a.BaseURL+"/chat/completions", bytes.NewBuffer(body))

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

    resp, err := a.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("OpenAI API error: %s", string(bodyBytes))
    }

    // 解析响应
    var chatResp service.ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
        return nil, err
    }

    return &chatResp, nil
}

func (a *OpenAIAdapter) ChatStream(ctx context.Context, req *service.ChatRequest,
    callback func(string) error) error {
    // 流式实现
    openaiReq := map[string]interface{}{
        "model":    req.Model,
        "messages": req.Messages,
        "stream":   true,
    }

    body, _ := json.Marshal(openaiReq)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        a.BaseURL+"/chat/completions", bytes.NewBuffer(body))

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

    resp, err := a.client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 读取 SSE 流
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }

        line = strings.TrimSpace(line)
        if !strings.HasPrefix(line, "data: ") {
            continue
        }

        data := strings.TrimPrefix(line, "data: ")
        if data == "[DONE]" {
            break
        }

        // 回调处理每个 chunk
        if err := callback(data); err != nil {
            return err
        }
    }

    return nil
}

func (a *OpenAIAdapter) ValidateConfig() error {
    if a.APIKey == "" {
        return fmt.Errorf("API key is required")
    }
    return nil
}
```

### 6. 模型管理服务实现

```go
// service/model_manager.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "mrchat/model"
    "mrchat/service/adapter"
    "sync"
    "time"
)

type ModelManager struct {
    factory       *adapter.AdapterFactory
    modelCache    map[string]*model.AIModel
    cacheMutex    sync.RWMutex
    cacheExpiry   time.Duration
}

func NewModelManager() *ModelManager {
    factory := adapter.NewAdapterFactory()

    // 注册适配器
    factory.Register("openai", adapter.NewOpenAIAdapter(
        os.Getenv("OPENAI_API_KEY"),
        os.Getenv("OPENAI_BASE_URL"),
    ))
    factory.Register("anthropic", adapter.NewAnthropicAdapter(
        os.Getenv("ANTHROPIC_API_KEY"),
    ))

    return &ModelManager{
        factory:     factory,
        modelCache:  make(map[string]*model.AIModel),
        cacheExpiry: 5 * time.Minute,
    }
}

// 获取模型配置（带缓存）
func (m *ModelManager) GetModel(modelID string) (*model.AIModel, error) {
    // 尝试从缓存获取
    m.cacheMutex.RLock()
    if cached, ok := m.modelCache[modelID]; ok {
        m.cacheMutex.RUnlock()
        return cached, nil
    }
    m.cacheMutex.RUnlock()

    // 从数据库查询
    var aiModel model.AIModel
    err := model.DB.Where("model_id = ? AND status = 1", modelID).First(&aiModel).Error
    if err != nil {
        return nil, fmt.Errorf("model not found: %s", modelID)
    }

    // 写入缓存
    m.cacheMutex.Lock()
    m.modelCache[modelID] = &aiModel
    m.cacheMutex.Unlock()

    return &aiModel, nil
}

// 执行聊天请求
func (m *ModelManager) Chat(ctx context.Context, req *ChatRequest, userID string) (*ChatResponse, error) {
    // 1. 获取模型配置
    aiModel, err := m.GetModel(req.Model)
    if err != nil {
        return nil, err
    }

    // 2. 检查用户权限
    if err := m.checkUserAccess(userID, aiModel); err != nil {
        return nil, err
    }

    // 3. 获取适配器
    modelAdapter, err := m.factory.GetAdapter(aiModel.Provider)
    if err != nil {
        return nil, err
    }

    // 4. 注入系统提示词（如果配置了）
    if aiModel.SystemPrompt != "" && req.SystemPrompt == "" {
        req.SystemPrompt = aiModel.SystemPrompt
    }

    // 5. 调用适配器
    resp, err := modelAdapter.Chat(ctx, req)
    if err != nil {
        return nil, err
    }

    return resp, nil
}

// 检查用户访问权限
func (m *ModelManager) checkUserAccess(userID string, aiModel *model.AIModel) error {
    // 公开模型直接允许
    if aiModel.IsPublic {
        return nil
    }

    // 检查用户组权限
    var user model.User
    if err := model.DB.Preload("Groups").First(&user, "id = ?", userID).Error; err != nil {
        return fmt.Errorf("user not found")
    }

    // Root 和 Admin 无限制
    if user.Role <= 10 {
        return nil
    }

    // 解析允许的组
    var allowedGroups []string
    if aiModel.AllowedGroups != nil {
        json.Unmarshal(aiModel.AllowedGroups, &allowedGroups)
    }

    // 检查用户是否在允许的组中
    for _, group := range user.Groups {
        for _, allowed := range allowedGroups {
            if group.ID == allowed {
                return nil
            }
        }
    }

    return fmt.Errorf("access denied to model: %s", aiModel.ModelID)
}
```

---

## 💾 数据库架构优化

### 1. 对话表设计（增强版）

```sql
CREATE TABLE conversations (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36) NOT NULL,

    -- 基础信息
    title VARCHAR(255) NOT NULL,

    -- 模型配置
    model_id VARCHAR(100) NOT NULL,

    -- 状态与分类
    status INT DEFAULT 1 COMMENT '1=Active, 2=Archived, 3=Deleted',
    folder_id VARCHAR(36) COMMENT '文件夹ID',

    -- 标签（借鉴 OUI）
    tags JSON COMMENT '标签数组',

    -- 统计信息
    message_count INT DEFAULT 0,
    total_tokens INT DEFAULT 0,

    -- 分享（借鉴 OUI）
    share_id VARCHAR(32) UNIQUE COMMENT '分享ID',
    is_public BOOLEAN DEFAULT FALSE,

    -- 固定与归档
    pinned BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    last_message_at BIGINT,

    INDEX idx_user_id (user_id),
    INDEX idx_user_updated (user_id, updated_at),
    INDEX idx_user_pinned (user_id, pinned),
    INDEX idx_folder (folder_id),
    INDEX idx_share_id (share_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 文件夹表（借鉴 OUI）
CREATE TABLE folders (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    parent_id VARCHAR(36) COMMENT '父文件夹ID',

    -- 排序
    sort_order INT DEFAULT 0,

    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT,

    INDEX idx_user_id (user_id),
    INDEX idx_parent (parent_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. 消息表优化（支持多模态）

```sql
CREATE TABLE messages (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    conversation_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,

    -- 消息内容
    role VARCHAR(20) NOT NULL COMMENT 'user, assistant, system',
    content TEXT NOT NULL,

    -- 多模态内容（借鉴 OUI）
    content_type VARCHAR(20) DEFAULT 'text' COMMENT 'text, image, file, code',
    attachments JSON COMMENT '附件列表',

    -- Token 统计
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,

    -- 模型信息
    model_id VARCHAR(100),

    -- 额度消耗
    quota_used BIGINT DEFAULT 0,

    -- 元数据
    metadata JSON COMMENT '扩展元数据',

    -- 时间戳
    created_at BIGINT NOT NULL,

    INDEX idx_conversation (conversation_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3. 索引优化策略

```sql
-- 复合索引优化查询性能
CREATE INDEX idx_conv_user_status ON conversations(user_id, status, updated_at DESC);
CREATE INDEX idx_msg_conv_created ON messages(conversation_id, created_at ASC);
CREATE INDEX idx_user_active ON users(status, last_active_at DESC);

-- 全文索引（用于搜索）
ALTER TABLE conversations ADD FULLTEXT INDEX ft_title (title);
ALTER TABLE messages ADD FULLTEXT INDEX ft_content (content);
```

---

## 🔄 实时协作能力

### 1. WebSocket 集成（借鉴 OUI 的 Socket.IO 设计）

```go
// service/websocket/hub.go
package websocket

import (
    "encoding/json"
    "sync"
)

// WebSocket Hub 管理所有连接
type Hub struct {
    // 用户连接映射
    clients map[string]*Client  // userID -> Client

    // 房间管理（用于群聊）
    rooms map[string]map[string]*Client  // roomID -> userID -> Client

    // 注册/注销通道
    register   chan *Client
    unregister chan *Client

    // 广播通道
    broadcast chan *Message

    mutex sync.RWMutex
}

type Client struct {
    ID     string
    UserID string
    Hub    *Hub
    Conn   *websocket.Conn
    Send   chan []byte
}

type Message struct {
    Type    string      `json:"type"`
    From    string      `json:"from"`
    To      string      `json:"to,omitempty"`
    Room    string      `json:"room,omitempty"`
    Data    interface{} `json:"data"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]*Client),
        rooms:      make(map[string]map[string]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *Message, 256),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client.UserID] = client
            h.mutex.Unlock()

            // 通知其他用户上线
            h.broadcastUserStatus(client.UserID, "online")

        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client.UserID]; ok {
                delete(h.clients, client.UserID)
                close(client.Send)
            }
            h.mutex.Unlock()

            // 通知其他用户下线
            h.broadcastUserStatus(client.UserID, "offline")

        case message := <-h.broadcast:
            h.handleBroadcast(message)
        }
    }
}

// 广播用户状态
func (h *Hub) broadcastUserStatus(userID, status string) {
    msg := &Message{
        Type: "user_status",
        From: userID,
        Data: map[string]string{
            "user_id": userID,
            "status":  status,
        },
    }

    data, _ := json.Marshal(msg)
    h.mutex.RLock()
    for _, client := range h.clients {
        if client.UserID != userID {
            select {
            case client.Send <- data:
            default:
                close(client.Send)
                delete(h.clients, client.UserID)
            }
        }
    }
    h.mutex.RUnlock()
}

// 发送消息给特定用户
func (h *Hub) SendToUser(userID string, message *Message) error {
    h.mutex.RLock()
    client, ok := h.clients[userID]
    h.mutex.RUnlock()

    if !ok {
        return fmt.Errorf("user not connected: %s", userID)
    }

    data, _ := json.Marshal(message)
    select {
    case client.Send <- data:
        return nil
    default:
        return fmt.Errorf("client send buffer full")
    }
}
```

### 2. 在线状态追踪

```go
// service/presence.go
package service

import (
    "context"
    "time"
    "mrchat/model"
)

type PresenceService struct {
    redis *redis.Client
}

// 更新用户在线状态
func (p *PresenceService) UpdateUserPresence(userID string) error {
    ctx := context.Background()
    key := fmt.Sprintf("presence:%s", userID)

    // 设置 Redis 键，5分钟过期
    err := p.redis.Set(ctx, key, time.Now().Unix(), 5*time.Minute).Err()
    if err != nil {
        return err
    }

    // 异步更新数据库（节流：每分钟最多更新一次）
    go p.throttledDBUpdate(userID)

    return nil
}

// 获取在线用户列表
func (p *PresenceService) GetOnlineUsers() ([]string, error) {
    ctx := context.Background()

    // 扫描所有 presence:* 键
    keys, err := p.redis.Keys(ctx, "presence:*").Result()
    if err != nil {
        return nil, err
    }

    userIDs := make([]string, 0, len(keys))
    for _, key := range keys {
        userID := strings.TrimPrefix(key, "presence:")
        userIDs = append(userIDs, userID)
    }

    return userIDs, nil
}

// 节流更新数据库
func (p *PresenceService) throttledDBUpdate(userID string) {
    ctx := context.Background()
    lockKey := fmt.Sprintf("presence_lock:%s", userID)

    // 尝试获取锁（60秒内只更新一次）
    locked, err := p.redis.SetNX(ctx, lockKey, 1, 60*time.Second).Result()
    if err != nil || !locked {
        return
    }

    // 更新数据库
    now := time.Now().Unix()
    model.DB.Model(&model.User{}).
        Where("id = ?", userID).
        Update("last_active_at", now)
}
```

---

## 🌐 API 设计规范

### 1. RESTful API 路由设计（借鉴 OUI）

```go
// router/api.go
package router

func SetupAPIRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    api.Use(middleware.CORS())

    // 认证相关
    auth := api.Group("/auth")
    {
        auth.POST("/signup", controller.Signup)
        auth.POST("/signin", controller.Signin)
        auth.POST("/signout", controller.Signout)
        auth.POST("/refresh", controller.RefreshToken)
        auth.GET("/oauth/:provider", controller.OAuthRedirect)
        auth.GET("/oauth/:provider/callback", controller.OAuthCallback)
    }

    // 需要认证的路由
    protected := api.Group("")
    protected.Use(middleware.JWTAuth())

    // 用户管理
    users := protected.Group("/users")
    {
        users.GET("/me", controller.GetCurrentUser)
        users.PUT("/me", controller.UpdateProfile)
        users.PUT("/me/settings", controller.UpdateSettings)
        users.GET("/me/quota", controller.GetQuota)
        users.GET("/online", controller.GetOnlineUsers)
    }

    // 对话管理
    conversations := protected.Group("/conversations")
    {
        conversations.GET("", controller.ListConversations)
        conversations.POST("", controller.CreateConversation)
        conversations.GET("/:id", controller.GetConversation)
        conversations.PUT("/:id", controller.UpdateConversation)
        conversations.DELETE("/:id", controller.DeleteConversation)
        conversations.POST("/:id/archive", controller.ArchiveConversation)
        conversations.POST("/:id/share", controller.ShareConversation)
        conversations.GET("/:id/messages", controller.GetMessages)
    }

    // 聊天接口（OpenAI 兼容）
    chat := protected.Group("/chat")
    {
        chat.POST("/completions", middleware.CheckModelAccess(), controller.ChatCompletions)
    }

    // 模型管理
    models := protected.Group("/models")
    {
        models.GET("", controller.ListModels)
        models.GET("/:id", controller.GetModel)

        // 管理员路由
        adminModels := models.Group("")
        adminModels.Use(middleware.RequireRole(middleware.RoleAdmin))
        {
            adminModels.POST("", controller.CreateModel)
            adminModels.PUT("/:id", controller.UpdateModel)
            adminModels.DELETE("/:id", controller.DeleteModel)
        }
    }

    // 组管理
    groups := protected.Group("/groups")
    {
        groups.GET("", controller.ListGroups)
        groups.POST("", controller.CreateGroup)
        groups.GET("/:id", controller.GetGroup)
        groups.PUT("/:id", controller.UpdateGroup)
        groups.DELETE("/:id", controller.DeleteGroup)
        groups.POST("/:id/members", controller.AddGroupMember)
        groups.DELETE("/:id/members/:user_id", controller.RemoveGroupMember)
    }

    // API 密钥管理
    apiKeys := protected.Group("/api-keys")
    {
        apiKeys.GET("", controller.ListAPIKeys)
        apiKeys.POST("", controller.CreateAPIKey)
        apiKeys.DELETE("/:id", controller.DeleteAPIKey)
    }

    // 管理员路由
    admin := protected.Group("/admin")
    admin.Use(middleware.RequireRole(middleware.RoleAdmin))
    {
        admin.GET("/users", controller.AdminListUsers)
        admin.PUT("/users/:id", controller.AdminUpdateUser)
        admin.GET("/stats", controller.GetSystemStats)
        admin.GET("/logs", controller.GetAuditLogs)
    }
}
```

### 2. 统一响应格式

```go
// types/response.go
package types

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    PageSize   int `json:"page_size,omitempty"`
    Total      int `json:"total,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

// 成功响应
func Success(data interface{}) APIResponse {
    return APIResponse{
        Success: true,
        Data:    data,
    }
}

// 分页响应
func SuccessWithMeta(data interface{}, meta *Meta) APIResponse {
    return APIResponse{
        Success: true,
        Data:    data,
        Meta:    meta,
    }
}

// 错误响应
func Error(code, message string) APIResponse {
    return APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
        },
    }
}
```

---

## ⚡ 缓存与性能优化

### 1. 多级缓存架构（借鉴 new-api）

```go
// service/cache/cache.go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// 三级缓存：内存 -> Redis -> 数据库
type CacheManager struct {
    // L1: 内存缓存
    memCache   map[string]*CacheItem
    memMutex   sync.RWMutex
    memTTL     time.Duration

    // L2: Redis 缓存
    redis      *redis.Client
    redisTTL   time.Duration
}

type CacheItem struct {
    Data      interface{}
    ExpiresAt time.Time
}

func NewCacheManager(redis *redis.Client) *CacheManager {
    return &CacheManager{
        memCache:  make(map[string]*CacheItem),
        memTTL:    1 * time.Minute,
        redis:     redis,
        redisTTL:  5 * time.Minute,
    }
}

// 获取缓存（三级查找）
func (c *CacheManager) Get(key string, dest interface{}) error {
    // L1: 内存缓存
    if data, ok := c.getFromMemory(key); ok {
        return json.Unmarshal(data, dest)
    }

    // L2: Redis 缓存
    if data, err := c.getFromRedis(key); err == nil {
        // 回写到内存
        c.setToMemory(key, data)
        return json.Unmarshal(data, dest)
    }

    return fmt.Errorf("cache miss: %s", key)
}

// 设置缓存（写入所有层级）
func (c *CacheManager) Set(key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // 写入内存
    c.setToMemory(key, data)

    // 写入 Redis
    return c.setToRedis(key, data)
}

// 内存缓存操作
func (c *CacheManager) getFromMemory(key string) ([]byte, bool) {
    c.memMutex.RLock()
    defer c.memMutex.RUnlock()

    item, ok := c.memCache[key]
    if !ok || time.Now().After(item.ExpiresAt) {
        return nil, false
    }

    data, _ := json.Marshal(item.Data)
    return data, true
}

func (c *CacheManager) setToMemory(key string, data []byte) {
    c.memMutex.Lock()
    defer c.memMutex.Unlock()

    var value interface{}
    json.Unmarshal(data, &value)

    c.memCache[key] = &CacheItem{
        Data:      value,
        ExpiresAt: time.Now().Add(c.memTTL),
    }
}

// Redis 缓存操作
func (c *CacheManager) getFromRedis(key string) ([]byte, error) {
    ctx := context.Background()
    data, err := c.redis.Get(ctx, key).Bytes()
    return data, err
}

func (c *CacheManager) setToRedis(key string, data []byte) error {
    ctx := context.Background()
    return c.redis.Set(ctx, key, data, c.redisTTL).Err()
}

// 删除缓存
func (c *CacheManager) Delete(key string) error {
    // 删除内存缓存
    c.memMutex.Lock()
    delete(c.memCache, key)
    c.memMutex.Unlock()

    // 删除 Redis 缓存
    ctx := context.Background()
    return c.redis.Del(ctx, key).Err()
}
```

### 2. 用户缓存实现

```go
// model/user_cache.go
package model

import (
    "fmt"
    "mrchat/service/cache"
)

var userCache *cache.CacheManager

func InitUserCache(redis *redis.Client) {
    userCache = cache.NewCacheManager(redis)
}

// 获取用户（带缓存）
func GetUserByID(userID string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", userID)

    // 尝试从缓存获取
    var user User
    if err := userCache.Get(cacheKey, &user); err == nil {
        return &user, nil
    }

    // 从数据库查询
    if err := DB.First(&user, "id = ?", userID).Error; err != nil {
        return nil, err
    }

    // 写入缓存
    userCache.Set(cacheKey, &user)

    return &user, nil
}

// 更新用户（清除缓存）
func (u *User) Update() error {
    if err := DB.Save(u).Error; err != nil {
        return err
    }

    // 清除缓存
    cacheKey := fmt.Sprintf("user:%s", u.ID)
    userCache.Delete(cacheKey)

    return nil
}
```

---

## 🔒 安全与审计

### 1. 审计日志系统（借鉴 OUI）

```sql
CREATE TABLE audit_logs (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36),

    -- 操作信息
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) COMMENT '资源类型',
    resource_id VARCHAR(36) COMMENT '资源ID',

    -- 请求信息
    method VARCHAR(10) COMMENT 'HTTP 方法',
    path VARCHAR(500) COMMENT '请求路径',
    ip_address VARCHAR(50),
    user_agent TEXT,

    -- 操作结果
    status INT COMMENT 'HTTP 状态码',
    error_message TEXT,

    -- 额外数据
    metadata JSON,

    -- 时间戳
    created_at BIGINT NOT NULL,

    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at),
    INDEX idx_resource (resource_type, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```go
// middleware/audit.go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "mrchat/model"
)

func AuditLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 处理请求
        c.Next()

        // 记录审计日志
        go func() {
            var userID string
            if user, exists := c.Get("user"); exists {
                userID = user.(*model.User).ID
            }

            log := &model.AuditLog{
                UserID:       userID,
                Action:       c.Request.Method + " " + c.Request.URL.Path,
                Method:       c.Request.Method,
                Path:         c.Request.URL.Path,
                IPAddress:    c.ClientIP(),
                UserAgent:    c.Request.UserAgent(),
                Status:       c.Writer.Status(),
                CreatedAt:    time.Now().Unix(),
            }

            model.DB.Create(log)
        }()
    }
}
```

### 2. 速率限制（多维度）

```go
// middleware/rate_limit.go
package middleware

import (
    "fmt"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    // 用户级限流器
    userLimiters map[string]*rate.Limiter
    // IP 级限流器
    ipLimiters   map[string]*rate.Limiter
    mutex        sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
    limiter := &RateLimiter{
        userLimiters: make(map[string]*rate.Limiter),
        ipLimiters:   make(map[string]*rate.Limiter),
    }

    // 定期清理过期的限流器
    go limiter.cleanupRoutine()

    return limiter
}

func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    limiter, exists := rl.userLimiters[userID]
    if !exists {
        // 每秒 10 个请求，突发 20 个
        limiter = rate.NewLimiter(10, 20)
        rl.userLimiters[userID] = limiter
    }

    return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // IP 级限流
        ip := c.ClientIP()
        ipLimiter := rl.getIPLimiter(ip)
        if !ipLimiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "IP rate limit exceeded",
            })
            c.Abort()
            return
        }

        // 用户级限流
        if user, exists := c.Get("user"); exists {
            userID := user.(*model.User).ID
            userLimiter := rl.getUserLimiter(userID)
            if !userLimiter.Allow() {
                c.JSON(http.StatusTooManyRequests, gin.H{
                    "error": "User rate limit exceeded",
                })
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

---

## 📊 关键设计对比总结

### Open WebUI vs new-api vs MrChat

| 特性 | Open WebUI | new-api | MrChat（融合设计） |
|------|-----------|---------|------------------|
| **用户系统** | User + Auth 分离 | User 单表 | ✅ User + Auth 分离 |
| **组管理** | ✅ Groups + RBAC | 简单分组 | ✅ Groups + RBAC |
| **模型管理** | ✅ 统一接口 + 配置 | 渠道管理 | ✅ 统一接口 + 适配器 |
| **实时协作** | ✅ Socket.IO | ❌ | ✅ WebSocket Hub |
| **多级缓存** | ❌ | ✅ 三级缓存 | ✅ 三级缓存 |
| **审计日志** | ✅ 完善 | 基础日志 | ✅ 完善审计 |
| **JSON 灵活性** | ✅ 广泛使用 | 部分使用 | ✅ 广泛使用 |
| **API 设计** | OpenAI 兼容 | 多渠道代理 | ✅ OpenAI 兼容 |

---

## 🚀 实施路线图

### Phase 1: 核心基础（2-3周）
- [ ] 数据库设计与迁移（Users, Auths, Groups）
- [ ] JWT 认证系统
- [ ] 基础 CRUD API
- [ ] Redis 缓存集成
- [ ] 基础中间件（认证、日志、CORS）

### Phase 2: 多租户与权限（2周）
- [ ] 组管理功能
- [ ] RBAC 权限系统
- [ ] API Key 管理
- [ ] OAuth 集成（GitHub, Google）
- [ ] 权限中间件

### Phase 3: 模型管理（2周）
- [ ] 模型配置表
- [ ] 适配器工厂
- [ ] OpenAI 适配器
- [ ] Anthropic 适配器
- [ ] 统一聊天接口

### Phase 4: 对话管理（2-3周）
- [ ] 对话 CRUD
- [ ] 消息存储
- [ ] 文件夹管理
- [ ] 分享功能
- [ ] 搜索功能

### Phase 5: 实时功能（1-2周）
- [ ] WebSocket Hub
- [ ] 在线状态追踪
- [ ] 实时消息推送
- [ ] 流式响应

### Phase 6: 安全与监控（1-2周）
- [ ] 审计日志
- [ ] 速率限制
- [ ] Prometheus 指标
- [ ] 健康检查

### Phase 7: 前端开发（3-4周）
- [ ] Vue 3 + Vite 项目初始化
- [ ] 用户认证界面
- [ ] 聊天主界面
- [ ] 管理后台
- [ ] 响应式设计

---

## 💡 核心设计原则总结

### 1. **分离关注点**
- 用户与认证分离
- 业务逻辑与数据访问分离
- 前后端完全分离

### 2. **灵活性优先**
- JSON 字段存储配置
- 适配器模式支持多提供商
- 插件化架构

### 3. **性能优化**
- 多级缓存
- 数据库索引优化
- 异步处理

### 4. **安全第一**
- JWT + API Key 双认证
- RBAC 权限控制
- 审计日志
- 速率限制

### 5. **可扩展性**
- 水平扩展（Redis 支持）
- 模块化设计
- 清晰的接口定义

---

## 📚 参考资源

- **Open WebUI**: https://github.com/open-webui/open-webui
- **new-api**: https://github.com/QuantumNous/new-api
- **Go 最佳实践**: https://github.com/golang-standards/project-layout
- **Vue 3 文档**: https://vuejs.org/
- **Gin 框架**: https://gin-gonic.com/

---

**文档版本**: v2.0
**创建时间**: 2026-01-27
**作者**: Claude Sonnet 4.5
**状态**: ✅ 完成

