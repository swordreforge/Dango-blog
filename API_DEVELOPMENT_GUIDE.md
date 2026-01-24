# MyBlog API 开发文档

## 📋 目录

- [概述](#概述)
- [认证机制](#认证机制)
- [通用响应格式](#通用响应格式)
- [API 接口列表](#api-接口列表)
  - [认证相关](#认证相关)
  - [文章管理](#文章管理)
  - [附件管理](#附件管理)
  - [用户管理](#用户管理)
  - [评论管理](#评论管理)
  - [系统设置](#系统设置)
- [错误码](#错误码)
- [开发规范](#开发规范)

---

## 概述

MyBlog 是一个基于 Go 语言开发的博客系统，提供完整的 RESTful API 接口用于前端交互。

### 基础信息

- **Base URL**: `http://localhost:8080`
- **API 前缀**: `/api`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

---

## 认证机制

### JWT Token 认证

大部分 API 需要使用 JWT Token 进行认证。Token 通过登录接口获取，并在后续请求中通过 HTTP Header 传递。

#### 请求头格式

```
Authorization: Bearer {token}
```

#### Token 获取

通过登录接口获取 Token，详见 [认证相关](#认证相关)。

#### Token 有效期

- 默认有效期：24 小时
- 过期后需要重新登录获取新 Token

---

## 通用响应格式

### 成功响应

```json
{
  "success": true,
  "message": "操作成功",
  "code": "SUCCESS",
  "data": { ... }
}
```

### 分页响应

```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": [ ... ],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### 错误响应

```json
{
  "success": false,
  "message": "错误描述",
  "code": "ERROR_CODE",
  "error": "详细错误信息"
}
```

---

## API 接口列表

### 认证相关

#### 1. 用户注册

**接口**: `POST /api/register`

**请求参数**:
```json
{
  "username": "用户名",
  "password": "密码",
  "email": "邮箱（可选）"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "注册成功",
  "code": "REGISTER_SUCCESS",
  "data": {
    "id": 1,
    "username": "用户名",
    "email": "邮箱"
  }
}
```

#### 2. 用户登录

**接口**: `POST /api/login`

**请求参数**:
```json
{
  "username": "用户名",
  "password": "密码"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "登录成功",
  "code": "LOGIN_SUCCESS",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "用户名",
      "role": "admin"
    }
  }
}
```

#### 3. 用户登出

**接口**: `POST /api/logout`

**认证**: 需要

**响应示例**:
```json
{
  "success": true,
  "message": "登出成功",
  "code": "LOGOUT_SUCCESS"
}
```

---

### 文章管理

#### 1. 获取文章列表

**接口**: `GET /api/passages`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | int | 否 | 每页数量（默认20） |
| offset | int | 否 | 偏移量（默认0） |
| status | string | 否 | 状态筛选（published/unpublished） |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": [
    {
      "id": 1,
      "title": "文章标题",
      "content": "文章内容",
      "status": "published",
      "created_at": "2026-01-24T10:00:00Z",
      "updated_at": "2026-01-24T10:00:00Z"
    }
  ],
  "total": 100
}
```

#### 2. 获取单篇文章

**接口**: `GET /api/passages/{id}`

**路径参数**:
- `id`: 文章ID

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": {
    "id": 1,
    "title": "文章标题",
    "content": "文章内容",
    "status": "published",
    "created_at": "2026-01-24T10:00:00Z",
    "updated_at": "2026-01-24T10:00:00Z"
  }
}
```

#### 3. 创建文章

**接口**: `POST /api/passages`

**认证**: 需要（管理员）

**请求参数**:
```json
{
  "title": "文章标题",
  "content": "文章内容",
  "status": "published"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "创建成功",
  "code": "CREATE_SUCCESS",
  "data": {
    "id": 1,
    "title": "文章标题",
    "content": "文章内容",
    "status": "published"
  }
}
```

#### 4. 更新文章

**接口**: `PUT /api/passages/{id}`

**认证**: 需要（管理员）

**路径参数**:
- `id`: 文章ID

**请求参数**:
```json
{
  "title": "更新后的标题",
  "content": "更新后的内容",
  "status": "published"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "更新成功",
  "code": "UPDATE_SUCCESS",
  "data": {
    "id": 1,
    "title": "更新后的标题",
    "content": "更新后的内容",
    "status": "published"
  }
}
```

#### 5. 删除文章

**接口**: `DELETE /api/passages/{id}`

**认证**: 需要（管理员）

**路径参数**:
- `id`: 文章ID

**响应示例**:
```json
{
  "success": true,
  "message": "删除成功",
  "code": "DELETE_SUCCESS"
}
```

---

### 附件管理

#### 1. 上传附件

**接口**: `POST /api/attachments`

**认证**: 需要（管理员）

**请求格式**: `multipart/form-data`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 上传的文件 |
| passage_id | int | 否 | 关联文章ID |

**限制**:
- 最大文件大小：500MB
- 支持的文件类型：jpg, jpeg, png, gif, bmp, svg, webp, pdf, doc, docx, xls, xlsx, ppt, pptx, mp4, webm, mp3, flac, zip, rar, 7z, tar, gz

**响应示例**:
```json
{
  "success": true,
  "message": "上传成功",
  "code": "UPLOAD_SUCCESS",
  "data": {
    "id": 1,
    "fileName": "原始文件名.jpg",
    "storedName": "原始文件名-20260124-100000.jpg",
    "path": "attachments/2026/01/24/原始文件名-20260124-100000.jpg",
    "url": "/attachments/2026/01/24/原始文件名-20260124-100000.jpg",
    "size": 1024000,
    "fileType": "image",
    "contentType": "image/jpeg",
    "passageId": 1
  }
}
```

#### 2. 获取附件列表

**接口**: `GET /api/attachments`

**认证**: 需要

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| passage_id | int | 否 | 按文章ID筛选 |
| limit | int | 否 | 每页数量（默认20） |
| offset | int | 否 | 偏移量（默认0） |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": [
    {
      "id": 1,
      "file_name": "原始文件名.jpg",
      "stored_name": "原始文件名-20260124-100000.jpg",
      "file_path": "attachments/2026/01/24/原始文件名-20260124-100000.jpg",
      "file_type": "image",
      "content_type": "image/jpeg",
      "file_size": 1024000,
      "passage_id": 1,
      "visibility": "public",
      "show_in_passage": true,
      "uploaded_at": "2026-01-24T10:00:00Z"
    }
  ],
  "total": 10
}
```

#### 3. 下载附件

**接口**: `GET /api/attachments/download?id={id}`

**路径参数**:
- `id`: 附件ID

**认证**:
- `public` 附件：无需认证
- `protected` 附件：需要登录
- `private` 附件：需要管理员权限

**响应**: 文件流

#### 4. 按日期获取附件

**接口**: `GET /api/attachments/by-date?year={年}&month={月}&day={日}`

**认证**: 无需认证

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| year | string | 是 | 年份 |
| month | string | 是 | 月份 |
| day | string | 是 | 日期 |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": [
    {
      "id": 1,
      "fileName": "文件名.jpg",
      "url": "/attachments/2026/01/24/文件名.jpg",
      "fileType": "image",
      "fileSize": 1024000
    }
  ],
  "total": 5
}
```

#### 5. 删除附件

**接口**: `DELETE /api/attachments?id={id}`

**认证**: 需要（管理员）

**请求参数**:
- `id`: 附件ID

**响应示例**:
```json
{
  "success": true,
  "message": "删除成功",
  "code": "DELETE_SUCCESS"
}
```

#### 6. 更新附件权限（管理员）

**接口**: `PATCH /api/admin/attachments?id={id}`

**认证**: 需要（管理员）

**请求参数**:
```json
{
  "visibility": "public",
  "show_in_passage": true
}
```

**字段说明**:
- `visibility`: 可见性（public/protected/private）
- `show_in_passage`: 是否在文章中显示

**响应示例**:
```json
{
  "success": true,
  "message": "更新成功",
  "code": "UPDATE_SUCCESS"
}
```

---

### 用户管理

#### 1. 获取用户列表

**接口**: `GET /api/users`

**认证**: 需要（管理员）

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | int | 否 | 每页数量（默认20） |
| offset | int | 否 | 偏移量（默认0） |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": [
    {
      "id": 1,
      "username": "用户名",
      "email": "邮箱",
      "role": "admin",
      "created_at": "2026-01-24T10:00:00Z"
    }
  ],
  "total": 10
}
```

#### 2. 更新用户信息

**接口**: `PUT /api/users/{id}`

**认证**: 需要（管理员或用户本人）

**请求参数**:
```json
{
  "username": "新用户名",
  "email": "新邮箱"
}
```

#### 3. 删除用户

**接口**: `DELETE /api/users/{id}`

**认证**: 需要（管理员）

---

### 评论管理

#### 1. 获取评论列表

**接口**: `GET /api/comments`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| passage_id | int | 否 | 按文章ID筛选 |
| limit | int | 否 | 每页数量（默认20） |
| offset | int | 否 | 偏移量（默认0） |

#### 2. 创建评论

**接口**: `POST /api/comments`

**认证**: 需要

**请求参数**:
```json
{
  "passage_id": 1,
  "content": "评论内容"
}
```

#### 3. 删除评论

**接口**: `DELETE /api/comments/{id}`

**认证**: 需要（管理员或评论作者）

---

### 系统设置

#### 1. 获取系统设置

**接口**: `GET /api/settings`

**认证**: 需要（管理员）

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "code": "SUCCESS",
  "data": {
    "site_name": "我的博客",
    "site_description": "博客描述",
    "background_attachment": "fixed",
    "attachment_default_visibility": "public",
    "attachment_max_size": 524288000,
    "attachment_allowed_types": "jpg,jpeg,png,gif,mp4,mp3,pdf,doc,docx,xls,xlsx,ppt,pptx,zip,rar,7z,tar,gz"
  }
}
```

#### 2. 更新系统设置

**接口**: `PUT /api/settings`

**认证**: 需要（管理员）

**请求参数**:
```json
{
  "site_name": "我的博客",
  "site_description": "博客描述",
  "background_attachment": "fixed"
}
```

---

## 错误码

### 通用错误码

| 错误码 | HTTP状态码 | 说明 |
|--------|-----------|------|
| SUCCESS | 200 | 操作成功 |
| METHOD_NOT_ALLOWED | 405 | 请求方法不允许 |
| UNAUTHORIZED | 401 | 未认证 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_FOUND | 404 | 资源不存在 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |

### 认证错误码

| 错误码 | 说明 |
|--------|------|
| INVALID_TOKEN | Token 无效 |
| TOKEN_EXPIRED | Token 已过期 |
| INVALID_CREDENTIALS | 用户名或密码错误 |
| USER_EXISTS | 用户已存在 |
| USER_NOT_FOUND | 用户不存在 |

### 文章错误码

| 错误码 | 说明 |
|--------|------|
| PASSAGE_NOT_FOUND | 文章不存在 |
| INVALID_PASSAGE_ID | 无效的文章ID |
| CREATE_FAILED | 创建文章失败 |
| UPDATE_FAILED | 更新文章失败 |
| DELETE_FAILED | 删除文章失败 |

### 附件错误码

| 错误码 | 说明 |
|--------|------|
| NO_FILE_PROVIDED | 未提供文件 |
| INVALID_FILE_TYPE | 不支持的文件类型 |
| FILE_TOO_LARGE | 文件过大 |
| UPLOAD_FAILED | 上传失败 |
| ATTACHMENT_NOT_FOUND | 附件不存在 |
| FILE_NOT_FOUND | 文件不存在 |
| INVALID_VISIBILITY | 无效的可见性值 |

---

## 开发规范

### 请求规范

1. **HTTP 方法**
   - GET：获取资源
   - POST：创建资源
   - PUT：更新资源（完整更新）
   - PATCH：更新资源（部分更新）
   - DELETE：删除资源

2. **参数传递**
   - 路径参数：用于资源标识（如 `/api/passages/{id}`）
   - 查询参数：用于筛选和分页（如 `?limit=20&offset=0`）
   - 请求体：用于提交数据（JSON 格式）

3. **内容类型**
   - 请求：`application/json` 或 `multipart/form-data`
   - 响应：`application/json`

### 认证规范

1. **需要认证的接口**必须在请求头中包含 Token
2. **管理员接口**需要 `role` 为 `admin`
3. Token 过期后需要重新登录获取

### 错误处理

1. 所有错误响应都包含 `success: false`
2. 错误信息包含 `code` 和 `message`
3. 部分错误包含详细的 `error` 信息

### 分页规范

1. 所有列表接口支持分页
2. 默认每页 20 条记录
3. 使用 `limit` 和 `offset` 参数控制分页
4. 响应包含 `total` 字段表示总记录数

### 文件上传规范

1. 使用 `multipart/form-data` 格式
2. 文件大小限制：500MB（可配置）
3. 支持的文件类型：jpg, jpeg, png, gif, bmp, svg, webp, pdf, doc, docx, xls, xlsx, ppt, pptx, mp4, webm, mp3, flac, zip, rar, 7z, tar, gz
4. 上传后自动生成唯一文件名（原名-时间戳.扩展名）

---

## 开发示例

### 使用 cURL

#### 登录获取 Token

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

#### 使用 Token 获取文章列表

```bash
curl -X GET http://localhost:8080/api/passages \
  -H "Authorization: Bearer {token}"
```

#### 上传附件

```bash
curl -X POST http://localhost:8080/api/attachments \
  -H "Authorization: Bearer {token}" \
  -F "file=@/path/to/file.jpg" \
  -F "passage_id=1"
```

### 使用 JavaScript (Fetch)

```javascript
// 登录
async function login(username, password) {
  const response = await fetch('http://localhost:8080/api/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });
  const data = await response.json();
  return data.data.token;
}

// 获取文章列表
async function getPassages(token) {
  const response = await fetch('http://localhost:8080/api/passages', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  return await response.json();
}

// 上传附件
async function uploadAttachment(token, file, passageId) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('passage_id', passageId);

  const response = await fetch('http://localhost:8080/api/attachments', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });
  return await response.json();
}
```

---

## 更新日志

### v1.0.0 (2026-01-24)

- 初始版本
- 完成认证、文章、附件、用户、评论、设置等核心 API
- 支持 JWT 认证
- 支持文件上传和管理
- 支持三级权限系统（public/protected/private）

---

## 联系方式

如有问题或建议，请联系开发团队。