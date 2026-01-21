// Package middleware 提供HTTP中间件功能
// 本包导出所有中间件函数，方便在其他包中使用
package middleware

// 所有中间件函数都已在各自的独立文件中定义
// 本文件作为包的入口点，确保所有中间件可以被外部导入使用
//
// 可用的中间件:
//   - Recovery: 恢复中间件，捕获panic (recovery.go)
//   - Logging: 日志中间件 (logging.go)
//   - CORS: 跨域中间件 (cors.go)
//   - AuthMiddleware: 认证中间件 (auth.go)
//   - RateLimitMiddleware: 限流中间件 (ratelimit.go)
//   - VisitorTracking: 访客追踪中间件 (visitor.go)
//   - RequireAdmin: 检查是否为管理员 (context.go)
//   - RequireAuth: 检查是否已认证 (context.go)
//   - CheckPassageAccess: 文章访问权限检查 (passage.go)
//
// Context 工具函数 (context.go):
//   - GetUserID: 从context中获取用户ID
//   - GetUsername: 从context中获取用户名
//   - GetRole: 从context中获取用户角色