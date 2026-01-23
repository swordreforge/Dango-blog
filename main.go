package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myblog-gogogo/config"
	"myblog-gogogo/controller"
	"myblog-gogogo/db"
	"myblog-gogogo/service/kafka"
	"myblog-gogogo/middleware"
	"myblog-gogogo/router"
	"myblog-gogogo/server"
	"myblog-gogogo/service"
	"myblog-gogogo/service/attachment"
	"myblog-gogogo/pkg/beautify"
)

//go:embed template img music
var templateFS embed.FS

func init() {
	// 初始化嵌入的模板文件系统
	controller.SetTemplateFS(templateFS)
}

func main() {
	// 初始化嵌入的模板文件系统
	controller.SetTemplateFS(templateFS)
	// 初始化模板
	controller.InitTemplates()

	// 释放嵌入的资源并创建必要的目录
	beautify.Section("资源初始化")
	beautify.Indent()
	if err := extractEmbeddedResources(); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("资源释放失败: %v", err))
	} else {
		beautify.SuccessLeaf("资源初始化成功")
	}
	beautify.Outdent()

	// 加载配置
	cfg := config.Load()

	// 显示启动标题
	beautify.Header("MyBlog-Gogogo 服务启动")

	// 加载配置
	beautify.Section("配置加载")
	beautify.Indent()
	beautify.Branch("数据库配置")
	beautify.Leaf(fmt.Sprintf("驱动: %s", cfg.DBDriver))
	beautify.Leaf(fmt.Sprintf("连接: %s", cfg.DBConnStr))
	beautify.Branch("服务器配置")
	beautify.Leaf(fmt.Sprintf("端口: %s", cfg.Port))
	if cfg.KafkaBrokers != "" {
		beautify.Leaf(fmt.Sprintf("Kafka: %s", cfg.KafkaBrokers))
	} else {
		beautify.Leaf("Kafka: 未启用")
	}
	beautify.Branch("JWT 配置")
	beautify.Leaf(fmt.Sprintf("Secret: %s...", cfg.JWTSecret[:min(8, len(cfg.JWTSecret))]))
	beautify.Outdent()

	// 初始化 JWT secret
	beautify.Section("JWT 初始化")
	beautify.Indent()
	beautify.Info("正在初始化 JWT secret...")
	controller.InitJWTSecret(cfg.JWTSecret)
	beautify.SuccessLeaf("JWT secret 初始化成功")
	beautify.Outdent()

	// 初始化数据库
	beautify.Section("数据库初始化")
	beautify.Indent()
	beautify.Info("正在初始化数据库...")
	if err := db.InitDB(cfg.DBDriver, cfg.DBConnStr); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("初始化失败: %v", err))
		log.Fatalf("Failed to initialize database: %v", err)
	}
	beautify.SuccessLeaf("数据库初始化成功")
	beautify.Outdent()
	defer db.CloseDB()

	// 数据同步
	beautify.Section("数据同步")
	beautify.Indent()
	// 同步附件
	beautify.Branch("附件同步")
	attachmentService := attachment.NewService()
	if err := attachmentService.SyncToDB(); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("同步失败: %v", err))
	} else {
		beautify.SuccessLeaf("附件同步成功")
	}

	// 同步音乐
	beautify.Branch("音乐同步")
	musicSyncService := service.NewMusicSyncService()
	if err := musicSyncService.SyncMusicFilesToDB(); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("同步失败: %v", err))
	} else {
		beautify.SuccessLeaf("音乐同步成功")
	}

	// 清理音乐标题
	beautify.Branch("清理音乐标题")
	if err := musicSyncService.CleanAllTitles(); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("清理失败: %v", err))
	} else {
		beautify.SuccessLeaf("标题清理完成")
	}
	beautify.Outdent()

	// 服务初始化
	beautify.Section("服务初始化")
	beautify.Indent()
	// GeoIP 服务
	beautify.Branch("GeoIP 服务")
	if err := service.InitGeoIP(); err != nil {
		beautify.ErrorLeaf(fmt.Sprintf("初始化失败: %v", err))
		beautify.Warn("GeoIP 功能已禁用。要启用，请从 MaxMind 下载 GeoLite2-City.mmdb 并放置在 data/ 目录中。")
	} else {
		defer service.CloseGeoIP()
		beautify.SuccessLeaf("GeoIP 服务初始化成功")
	}

	// Kafka 服务
	beautify.Branch("Kafka 服务")
	if cfg.KafkaBrokers != "" {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		for i, broker := range brokers {
			brokers[i] = strings.TrimSpace(broker)
		}

		if err := kafka.InitKafka(brokers, cfg.KafkaGroupID); err != nil {
			beautify.ErrorLeaf(fmt.Sprintf("初始化失败: %v", err))
			beautify.Warn("Kafka 功能已禁用。要启用，请确保 Kafka 正在运行且可访问。")
		} else {
			defer kafka.CloseKafka()
			beautify.SuccessLeaf(fmt.Sprintf("Kafka 服务初始化成功 (Brokers: %v)", brokers))

			// 异步生产者
			beautify.Indent()
			beautify.Branch("异步生产者")
			if err := kafka.InitAsyncProducer(brokers, 1000); err != nil {
				beautify.ErrorLeaf(fmt.Sprintf("初始化失败: %v", err))
				beautify.Warn("异步 Kafka 生产者功能已禁用。")
			} else {
				defer kafka.CloseAsyncProducer()
				beautify.SuccessLeaf("异步 Kafka 生产者初始化成功 (队列大小: 1000)")
			}
			beautify.Outdent()
		}
	} else {
		beautify.Leaf("Kafka 服务未启用（使用 --kafka-brokers 标志启用）")
	}

	// 工作池
	beautify.Branch("工作池")
	workerCount := 2
	queueSize := 1000
	service.InitWorkerPool(workerCount, queueSize)
	defer service.CloseWorkerPool()
	beautify.SuccessLeaf(fmt.Sprintf("工作池初始化成功 (%d workers, 队列: %d)", workerCount, queueSize))
	beautify.Outdent()

	// 初始化关于页面仓库
	database := db.GetDB()
	controller.InitAboutRepositories(database)

	// 启动定期清理过期会话的goroutine
	beautify.Section("后台任务")
	beautify.Indent()
	beautify.Branch("会话清理")
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			controller.CleanupExpiredSessions()
			beautify.Debugf("清理过期 ECC 会话完成。活跃会话: %d", controller.GetSessionCount())
		}
	}()
	beautify.SuccessLeaf("会话清理任务已启动（每 5 分钟）")

	// 启动文件监控
	beautify.Branch("文件监控")
	repo := db.GetPassageRepository()
	syncService := service.NewSyncService(repo)
	go func() {
		if err := syncService.WatchAndSync(); err != nil {
			beautify.Errorf("文件监控错误: %v", err)
		}
	}()
	beautify.SuccessLeaf("文件监控已启动")
	beautify.Outdent()

	// 设置路由
	var handler http.Handler
	var srv *server.Server

	beautify.Section("路由配置")
	beautify.Indent()
	mux := router.SetupRoutes()
	beautify.SuccessLeaf("路由设置完成")

	// 应用中间件链
	beautify.Branch("中间件链")
	beautify.Indent()
	handler = middleware.Recovery(mux)
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)
	handler = middleware.AuthMiddleware(handler)
	handler = middleware.CheckPassageAccess(handler)
	handler = middleware.RateLimitMiddleware(handler)
	handler = middleware.VisitorTracking(handler)

	beautify.Leaf("✓ Recovery")
	beautify.Leaf("✓ Logging")
	beautify.Leaf("✓ CORS")
	beautify.Leaf("✓ Auth")
	beautify.Leaf("✓ Passage Access")
	beautify.Leaf("✓ Rate Limit")
	beautify.Leaf("✓ Visitor Tracking")
	beautify.Outdent()

	// 创建服务器
	beautify.Branch("服务器创建")
	srv = server.New(handler, cfg)
	beautify.SuccessLeaf(fmt.Sprintf("服务器创建完成 (端口: %s)", cfg.Port))

	// 启动服务器
	beautify.Branch("启动服务器")
	go func() {
		if err := srv.Start(); err != nil {
			beautify.ErrorLeaf(fmt.Sprintf("服务器错误: %v", err))
			log.Fatalf("Server error: %v", err)
		}
	}()
	beautify.Outdent()

	// 等待优雅关闭
	beautify.Section("服务运行中")
	beautify.Info("服务器正在运行，按 Ctrl+C 停止...")

	if err := server.GracefulShutdown(srv); err != nil {
		beautify.Errorf("关闭错误: %v", err)
		log.Printf("Shutdown error: %v", err)
	}

	beautify.Success("服务器已优雅关闭")
}

// extractEmbeddedResources 释放嵌入的资源并创建必要的目录
func extractEmbeddedResources() error {
	// 需要创建的目录列表
	dirs := []string{
		"attachments",
		"data",
		"markdown",
	}

	// 创建必要的目录
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	// 释放 img 目录中的文件
	if err := extractDir(templateFS, "img", "img"); err != nil {
		return fmt.Errorf("释放 img 目录失败: %w", err)
	}

	// 释放 music 目录中的文件
	if err := extractDir(templateFS, "music", "music"); err != nil {
		return fmt.Errorf("释放 music 目录失败: %w", err)
	}

	return nil
}

// extractDir 从嵌入的文件系统中提取目录
func extractDir(embedFS embed.FS, srcDir, dstDir string) error {
	// 检查目标目录是否已存在
	if _, err := os.Stat(dstDir); err == nil {
		// 目录已存在,跳过提取
		return nil
	}

	// 创建目标目录
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// 读取源目录内容
	entries, err := fs.ReadDir(embedFS, srcDir)
	if err != nil {
		return err
	}

	// 提取所有文件
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// 递归提取子目录
			if err := extractDir(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 提取文件
			if err := extractFile(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractFile 从嵌入的文件系统中提取单个文件
func extractFile(embedFS embed.FS, srcPath, dstPath string) error {
	// 读取源文件
	srcFile, err := embedFS.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
