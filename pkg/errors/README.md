# 错误处理包使用指南

## 概述

`pkg/errors` 包提供了统一的错误处理机制，包括：

- **AppError 接口**：定义了应用错误的标准接口
- **预定义错误常量**：常见的业务错误和 HTTP 错误
- **HTTP 响应函数**：用于发送标准化的错误响应

## 基本用法

### 1. 使用预定义错误

```go
import "your-project/pkg/errors"

// 在业务逻辑中使用
if user == nil {
    return nil, errors.ErrUserNotFound
}

if password != user.Password {
    return errors.ErrPasswordIncorrect
}
```

### 2. 创建自定义错误

```go
// 创建简单错误
err := errors.New("CUSTOM_ERROR", "自定义错误消息")

// 创建带 HTTP 状态码的错误
err := errors.NewWithStatus("CUSTOM_ERROR", "自定义错误消息", http.StatusBadRequest)

// 创建带详细信息的错误
err := errors.NewWithDetails("CUSTOM_ERROR", "自定义错误消息", "详细信息")

// 包装已知错误
err := errors.Wrap(originalErr, "CODE", "包装错误消息")

// 包装错误并指定 HTTP 状态码
err := errors.WrapWithStatus(originalErr, "CODE", "包装错误消息", http.StatusBadRequest)

// 包装错误并添加详细信息
err := errors.WrapWithDetails(originalErr, "CODE", "包装错误消息", "详细信息")
```

### 3. 在 HTTP 处理器中使用

```go
import "your-project/pkg/errors"

func Handler(w http.ResponseWriter, r *http.Request) {
    result, err := service.DoSomething()
    if err != nil {
        // 自动处理错误并发送响应
        errors.SendError(w, err)
        return
    }

    // 返回成功响应
    json.NewEncoder(w).Encode(result)
}
```

### 4. 使用快捷方法

```go
// 发送 400 错误
errors.SendBadRequest(w, "INVALID_PARAM", "参数无效")

// 发送 401 错误
errors.SendUnauthorized(w, "NOT_LOGGED_IN", "请先登录")

// 发送 403 错误
errors.SendForbidden(w, "NO_PERMISSION", "权限不足")

// 发送 404 错误
errors.SendNotFound(w, "RESOURCE_NOT_FOUND", "资源不存在")

// 发送 409 错误
errors.SendConflict(w, "DUPLICATE_ENTRY", "资源已存在")

// 发送 500 错误
errors.SendInternalError(w, "INTERNAL_ERROR", "服务器内部错误")

// 发送带原始错误的 500 错误
errors.SendInternalErrorWithError(w, "INTERNAL_ERROR", "服务器内部错误", originalErr)
```

### 5. 验证错误

```go
import "your-project/pkg/errors"

// 创建验证错误
func ValidateUser(user *User) error {
    if user.Username == "" {
        return errors.NewValidationError("username", "用户名不能为空")
    }
    if len(user.Username) < 3 {
        return errors.NewValidationError("username", "用户名至少需要3个字符")
    }
    if user.Email == "" {
        return errors.NewValidationError("email", "邮箱不能为空")
    }
    return nil
}

// 判断错误类型
if errors.IsValidationError(err) {
    // 处理验证错误
}
```

### 6. 错误判断和转换

```go
import "your-project/pkg/errors"

// 判断是否为应用错误
if errors.IsAppError(err) {
    // 获取错误代码
    code := errors.GetCode(err)
    // 获取 HTTP 状态码
    status := errors.GetHTTPStatus(err)
}

// 转换为应用错误
if appErr, ok := errors.AsAppError(err); ok {
    fmt.Printf("Code: %s, Message: %s\n", appErr.Code(), appErr.Message())
    fmt.Printf("HTTP Status: %d\n", appErr.HTTPStatus())
    fmt.Printf("Details: %s\n", appErr.Details())
}
```

## 预定义错误列表

### 通用错误

| 错误常量 | 代码 | HTTP 状态码 | 说明 |
|---------|------|-----------|------|
| `ErrInternal` | INTERNAL_ERROR | 500 | 服务器内部错误 |
| `ErrBadRequest` | BAD_REQUEST | 400 | 请求参数错误 |
| `ErrUnauthorized` | UNAUTHORIZED | 401 | 未授权 |
| `ErrForbidden` | FORBIDDEN | 403 | 禁止访问 |
| `ErrNotFound` | NOT_FOUND | 404 | 资源不存在 |
| `ErrConflict` | CONFLICT | 409 | 资源冲突 |
| `ErrMethodNotAllowed` | METHOD_NOT_ALLOWED | 405 | 方法不允许 |
| `ErrServiceUnavailable` | SERVICE_UNAVAILABLE | 503 | 服务不可用 |
| `ErrRequestTimeout` | REQUEST_TIMEOUT | 408 | 请求超时 |
| `ErrTooManyRequests` | TOO_MANY_REQUESTS | 429 | 请求过多 |

### 业务错误

#### 用户相关

| 错误常量 | 代码 | HTTP 状态码 | 说明 |
|---------|------|-----------|------|
| `ErrUsernameRequired` | USERNAME_REQUIRED | 400 | 用户名不能为空 |
| `ErrUsernameInvalid` | USERNAME_INVALID | 400 | 用户名格式不正确 |
| `ErrUsernameExists` | USERNAME_EXISTS | 409 | 用户名已存在 |
| `ErrEmailRequired` | EMAIL_REQUIRED | 400 | 邮箱不能为空 |
| `ErrEmailInvalid` | EMAIL_INVALID | 400 | 邮箱格式不正确 |
| `ErrEmailExists` | EMAIL_EXISTS | 409 | 邮箱已被使用 |
| `ErrPasswordRequired` | PASSWORD_REQUIRED | 400 | 密码不能为空 |
| `ErrPasswordTooShort` | PASSWORD_TOO_SHORT | 400 | 密码长度不足 |
| `ErrPasswordIncorrect` | PASSWORD_INCORRECT | 401 | 密码错误 |
| `ErrUserNotFound` | USER_NOT_FOUND | 404 | 用户不存在 |
| `ErrUserInactive` | USER_INACTIVE | 403 | 用户账户未激活 |
| `ErrCannotDeleteAdmin` | CANNOT_DELETE_ADMIN | 403 | 无法删除管理员账户 |

#### 文章相关

| 错误常量 | 代码 | HTTP 状态码 | 说明 |
|---------|------|-----------|------|
| `ErrPassageNotFound` | PASSAGE_NOT_FOUND | 404 | 文章不存在 |
| `ErrPassageNotPublished` | PASSAGE_NOT_PUBLISHED | 404 | 文章尚未发布 |
| `ErrPassagePrivate` | PASSAGE_PRIVATE | 403 | 文章为私密文章 |

#### 文件相关

| 错误常量 | 代码 | HTTP 状态码 | 说明 |
|---------|------|-----------|------|
| `ErrFileTooLarge` | FILE_TOO_LARGE | 400 | 文件过大 |
| `ErrUnsupportedFileType` | UNSUPPORTED_FILE_TYPE | 400 | 不支持的文件类型 |

#### 会话相关

| 错误常量 | 代码 | HTTP 状态码 | 说明 |
|---------|------|-----------|------|
| `ErrSessionExpired` | SESSION_EXPIRED | 401 | 会话已过期 |
| `ErrSessionNotFound` | SESSION_NOT_FOUND | 401 | 会话不存在 |

## 迁移指南

### 从 `pkg/dto/errors.go` 迁移

**旧代码：**
```go
import "your-project/pkg/dto"

if username == "" {
    return nil, dto.ErrUsernameRequired
}
```

**新代码：**
```go
import "your-project/pkg/errors"

if username == "" {
    return nil, errors.ErrUsernameRequired
}
```

### 从 `pkg/response/errors.go` 迁移

**旧代码：**
```go
import "your-project/pkg/response"

response.Error(w, http.StatusBadRequest, "INVALID_PARAM", "参数无效")
```

**新代码：**
```go
import "your-project/pkg/errors"

errors.SendBadRequest(w, "INVALID_PARAM", "参数无效")
```

## 最佳实践

1. **在业务层使用预定义错误**：优先使用预定义的错误常量，保持错误代码一致
2. **在服务层包装错误**：使用 `Wrap` 或 `WrapWithDetails` 包装底层错误，保留错误链
3. **在控制器层发送响应**：使用 `SendError` 或快捷方法发送 HTTP 响应
4. **提供有意义的错误消息**：错误消息应该对用户友好，避免暴露内部实现细节
5. **使用验证错误**：对于输入验证，使用 `ValidationError` 提供字段级别的错误信息
6. **记录日志**：在发送错误响应前，记录错误日志以便调试

## 示例项目

```go
// service/user.go
package service

import "your-project/pkg/errors"

func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        return nil, errors.Wrap(err, "DB_ERROR", "获取用户失败")
    }
    if user == nil {
        return nil, errors.ErrUserNotFound
    }
    return user, nil
}

// controller/user.go
package controller

import "your-project/pkg/errors"

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
    id := extractID(r)
    
    user, err := c.service.GetUser(id)
    if err != nil {
        errors.SendError(w, err)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}
```