# Controller 与 Service 层解耦重构文档

## 概述

本次重构旨在解决 Controller 层与 Service 层耦合度过高的问题，通过引入 DTO 层和统一的服务层，实现更清晰的架构分层。

## 重构目标

1. **降低耦合度**：将业务逻辑从 Controller 层移至 Service 层
2. **引入 DTO 层**：统一请求和响应的数据结构
3. **提高可测试性**：业务逻辑与 HTTP 处理分离，便于单元测试
4. **统一错误处理**：使用预定义的业务错误类型

## 架构变化

### 重构前

```
HTTP Request → Router → Middleware → Controller (业务逻辑) → Database
```

### 重构后

```
HTTP Request → Router → Middleware → Controller (HTTP处理) → Service (业务逻辑) → Database
                                                    ↓
                                                DTO (数据传输)
```

## 新增文件

### pkg/dto/ 包

创建了统一的数据传输对象（DTO）包，包含以下文件：

- **common.go**: 基础响应结构、分页请求/响应、时间戳字段等
- **user.go**: 用户相关的 DTO（登录、注册、用户信息等）
- **passage.go**: 文章相关的 DTO（文章信息、访问控制等）
- **music.go**: 音乐相关的 DTO
- **attachment.go**: 附件相关的 DTO
- **comment.go**: 评论相关的 DTO
- **errors.go**: 统一的错误定义（ValidationError、BusinessError）

### service/ 包新增服务

- **auth_service.go**: 认证服务
  - 用户登录
  - 密码解密（ECC 加密）
  - Token 验证
  - 权限检查
  - ECC 会话管理

- **user_service.go**: 用户服务
  - 用户注册
  - 用户信息验证
  - 用户更新/删除
  - 用户列表查询

- **passage_service.go**: 文章服务
  - 文章访问权限检查
  - 文章创建/更新/删除
  - 文章列表查询

## 重构的 Controller 文件

### 1. controller/login.go

**重构前**：
- 包含完整的登录验证逻辑
- 直接操作数据库
- 包含密码解密逻辑
- 包含用户状态检查

**重构后**：
- 只负责 HTTP 请求/响应处理
- 使用 `authService.Login()` 处理业务逻辑
- 代码从 200+ 行减少到约 80 行

### 2. controller/register.go

**重构前**：
- 包含用户名、邮箱、密码验证逻辑
- 直接操作数据库
- 包含密码哈希逻辑
- 包含用户重复检查

**重构后**：
- 只负责 HTTP 请求/响应处理
- 使用 `userService.Register()` 处理业务逻辑
- 代码从 250+ 行减少到约 50 行

### 3. controller/passage.go

**重构前**：
- `PassageDetailHandler` 包含文章权限检查逻辑
- 直接检查文章状态和可见性
- 包含复杂的条件判断

**重构后**：
- 使用 `passageService.CheckAccess()` 处理权限检查
- Controller 只负责响应格式化
- 权限逻辑集中在 Service 层

## DTO 示例

### 请求 DTO

```go
type LoginRequest struct {
    Username          string `json:"username" binding:"required"`
    Password          string `json:"password"`
    EncryptedPassword string `json:"encrypted_password"`
    SessionID         string `json:"session_id"`
    ClientPublicKey   string `json:"client_public_key"`
    Algorithm         string `json:"algorithm"`
}
```

### 响应 DTO

```go
type LoginResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    User      *UserDTO  `json:"user"`
}
```

### 业务错误

```go
var (
    ErrUsernameRequired = &BusinessError{Code: "USERNAME_REQUIRED", Message: "用户名不能为空"}
    ErrUsernameInvalid  = &BusinessError{Code: "USERNAME_INVALID", Message: "用户名格式不正确"}
    ErrUsernameExists   = &BusinessError{Code: "USERNAME_EXISTS", Message: "用户名已存在"}
    // ...
)
```

## Service 层设计原则

### 1. 单一职责

每个 Service 只负责一个领域的业务逻辑：
- `AuthService`: 认证和授权
- `UserService`: 用户管理
- `PassageService`: 文章管理

### 2. 错误处理

Service 层返回业务错误（BusinessError），Controller 层转换为 HTTP 状态码：

```go
// Service 层
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
    if req.Username == "" {
        return nil, dto.ErrUsernameRequired
    }
    // ...
}

// Controller 层
func handleLoginError(w http.ResponseWriter, err error) {
    if bizErr, ok := err.(*dto.BusinessError); ok {
        statusCode := getStatusCodeForError(bizErr.Code)
        sendErrorResponse(w, statusCode, bizErr.Message, bizErr.Code)
        return
    }
    sendErrorResponse(w, http.StatusInternalServerError, "服务器内部错误", "")
}
```

### 3. 数据转换

Service 层负责将数据库模型转换为 DTO：

```go
func (s *UserService) toDTO(user *db.User) *dto.UserDTO {
    return &dto.UserDTO{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        Role:      user.Role,
        Status:    user.Status,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
}
```

## Controller 层职责

重构后的 Controller 层只负责：

1. **HTTP 请求解析**：解析请求体、查询参数、路径参数
2. **参数验证**：基础格式验证（如 JSON 格式）
3. **调用 Service**：将业务逻辑委托给 Service 层
4. **HTTP 响应**：将 Service 返回的 DTO 转换为 HTTP 响应
5. **错误处理**：将业务错误转换为 HTTP 错误响应

### 示例

```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求
    var req dto.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendErrorResponse(w, http.StatusBadRequest, "请求格式错误", "")
        return
    }

    // 2. 调用 Service
    resp, err := authService.Login(&req)
    if err != nil {
        handleLoginError(w, err)
        return
    }

    // 3. 返回响应
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "登录成功",
        "token":   resp.Token,
        "user":    resp.User,
    })
}
```

## 后续重构建议

### 高优先级

1. **重构其他 Controller**：
   - `controller/music.go`: 音乐上传和管理逻辑
   - `controller/upload.go`: 文件上传验证逻辑
   - `controller/attachment.go`: 附件权限检查逻辑
   - `controller/admin/passages.go`: 文章管理逻辑
   - `controller/admin/users.go`: 用户管理逻辑

2. **创建更多 Service**：
   - `MusicService`: 音乐管理服务
   - `UploadService`: 文件上传服务
   - `CommentService`: 评论服务

### 中优先级

3. **统一验证逻辑**：
   - 创建 `ValidationService` 集中处理所有验证逻辑
   - 支持自定义验证规则

4. **添加单元测试**：
   - 为 Service 层编写单元测试
   - 为 Controller 层编写集成测试

### 低优先级

5. **优化性能**：
   - 添加缓存层
   - 优化数据库查询

6. **文档完善**：
   - 添加 API 文档
   - 添加架构设计文档

## 注意事项

1. **向后兼容**：重构过程中确保 API 接口保持不变
2. **渐进式重构**：不要一次性重构所有 Controller，逐步进行
3. **测试覆盖**：每次重构后都要进行充分测试
4. **代码审查**：重构代码需要经过代码审查

## 总结

通过本次重构，我们实现了：

- ✅ 创建了统一的 DTO 层
- ✅ 将业务逻辑从 Controller 移至 Service 层
- ✅ 统一了错误处理机制
- ✅ 重构了 3 个核心 Controller（login、register、passage）
- ✅ 提高了代码的可测试性和可维护性

下一步建议继续重构其他 Controller，完善 Service 层的实现。