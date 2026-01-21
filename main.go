package main

import (
	"log"
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
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting server with config: %+v", cfg)

	// 初始化数据库
	if err := db.InitDB(cfg.DBDriver, cfg.DBConnStr); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	log.Println("Database initialized successfully")

	// 同步附件文件到数据库
	attachmentService := attachment.NewService()
	if err := attachmentService.SyncToDB(); err != nil {
		log.Printf("Warning: Failed to sync attachments to database: %v", err)
	} else {
		log.Println("Attachments synced to database successfully")
	}

	// 同步音乐文件到数据库
	musicSyncService := service.NewMusicSyncService()
	if err := musicSyncService.SyncMusicFilesToDB(); err != nil {
		log.Printf("Warning: Failed to sync music files to database: %v", err)
	} else {
		log.Println("Music files synced to database successfully")
	}

	// 清理音乐标题，移除时间戳
	if err := musicSyncService.CleanAllTitles(); err != nil {
		log.Printf("Warning: Failed to clean music titles: %v", err)
	}

	// 初始化 GeoIP 服务
	if err := service.InitGeoIP(); err != nil {
		log.Printf("Warning: Failed to initialize GeoIP service: %v", err)
		log.Println("GeoIP features will be disabled. To enable, download GeoLite2-City.mmdb from MaxMind and place it in the data/ directory.")
	} else {
		defer service.CloseGeoIP()
		log.Println("GeoIP service initialized successfully")
	}

	// 初始化 Kafka 服务（需要手动启用）
	if cfg.KafkaBrokers != "" {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		for i, broker := range brokers {
			brokers[i] = strings.TrimSpace(broker)
		}

		if err := kafka.InitKafka(brokers, cfg.KafkaGroupID); err != nil {
			log.Printf("Warning: Failed to initialize Kafka service: %v", err)
			log.Println("Kafka features will be disabled. To enable, ensure Kafka is running and accessible.")
		} else {
			defer kafka.CloseKafka()
			log.Printf("Kafka service initialized successfully with brokers: %v", brokers)
		}

		// 初始化异步生产者（队列大小 1000）
		if err := kafka.InitAsyncProducer(brokers, 1000); err != nil {
			log.Printf("Warning: Failed to initialize async Kafka producer: %v", err)
			log.Println("Async Kafka producer features will be disabled.")
		} else {
			defer kafka.CloseAsyncProducer()
			log.Println("Async Kafka producer initialized successfully with queue size: 1000")
		}
	} else {
		log.Println("Kafka service is disabled. To enable, use --kafka-brokers flag")
	}

	// 初始化工作池（并发优化）
	workerCount := 2
	queueSize := 1000
	service.InitWorkerPool(workerCount, queueSize)
	defer service.CloseWorkerPool()
	log.Printf("Worker pool initialized with %d workers, queue size: %d", workerCount, queueSize)

	// 初始化关于页面仓库
	database := db.GetDB()
	controller.InitAboutRepositories(database)

	// 启动定期清理过期会话的goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			controller.CleanupExpiredSessions()
			log.Printf("Cleaned up expired ECC sessions. Active sessions: %d", controller.GetSessionCount())
		}
	}()

	// 设置路由
	mux := router.SetupRoutes()

	// 应用中间件链
	handler := middleware.Recovery(mux)
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)
	handler = middleware.AuthMiddleware(handler)
	handler = middleware.CheckPassageAccess(handler)
	handler = middleware.RateLimitMiddleware(handler)
	handler = middleware.VisitorTracking(handler)

	// 创建服务器
	srv := server.New(handler, cfg)

	// 启动服务器
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 优雅关闭
	if err := server.GracefulShutdown(srv); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
