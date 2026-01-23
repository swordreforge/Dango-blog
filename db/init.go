package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"myblog-gogogo/auth"
	"myblog-gogogo/db/drivers"
	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
)

var (
	dbInstance          *sql.DB
	passageRepo         repositories.PassageRepository
	userRepo            repositories.UserRepository
	statsRepo           repositories.StatsRepository
	visitorRepo         repositories.VisitorRepository
	commentRepo         repositories.CommentRepository
	settingRepo         repositories.SettingRepository
	aboutMainCardRepo   repositories.AboutMainCardRepository
	aboutSubCardRepo    repositories.AboutSubCardRepository
	attachmentRepo      repositories.AttachmentRepository
)

// InitDB 初始化数据库
func InitDB(driver, dsn string) error {
	var err error
	
	// 根据驱动类型获取DSN
	config := drivers.Config{
		FilePath: dsn,
	}
	
	driverImpl, err := drivers.GetDriver(driver)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}
	
	dbInstance, err = driverImpl.Connect(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 创建表结构
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	// 初始化仓库
	passageRepo = repositories.NewSQLitePassageRepository(dbInstance)
	userRepo = repositories.NewSQLiteUserRepository(dbInstance)
	statsRepo = repositories.NewSQLiteStatsRepository(dbInstance)
	visitorRepo = repositories.NewSQLiteVisitorRepository(dbInstance)
	commentRepo = repositories.NewSQLiteCommentRepository(dbInstance)
	settingRepo = repositories.NewSQLiteSettingRepository(dbInstance)
	aboutMainCardRepo = repositories.NewSQLiteAboutMainCardRepository(dbInstance)
	aboutSubCardRepo = repositories.NewSQLiteAboutSubCardRepository(dbInstance)
	attachmentRepo = repositories.NewSQLiteAttachmentRepository(dbInstance)

	// 插入默认数据
	if err := seedData(); err != nil {
		log.Printf("Warning: failed to seed data: %v", err)
	}
	
	log.Println("Database initialized successfully")
	return nil
}

// createTables 创建数据库表
func createTables() error {
	// 创建文章表
	passageTable := `
	CREATE TABLE IF NOT EXISTS passages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		original_content TEXT,
		summary TEXT,
		author TEXT DEFAULT '管理员',
		tags TEXT DEFAULT '[]',
		category TEXT DEFAULT '未分类',
		status TEXT DEFAULT 'published',
		file_path TEXT,
		visibility TEXT DEFAULT 'public',
		is_scheduled INTEGER DEFAULT 0,
		published_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// passages 表索引
	passageIndexes := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_passages_file_path ON passages(file_path);
	CREATE INDEX IF NOT EXISTS idx_passages_status ON passages(status);
	CREATE INDEX IF NOT EXISTS idx_passages_category ON passages(category);
	CREATE INDEX IF NOT EXISTS idx_passages_created_at ON passages(created_at);
	CREATE INDEX IF NOT EXISTS idx_passages_status_created ON passages(status, created_at DESC);  -- 复合索引：状态+创建时间
	CREATE INDEX IF NOT EXISTS idx_passages_category_status ON passages(category, status);  -- 复合索引：分类+状态
	CREATE INDEX IF NOT EXISTS idx_passages_visibility ON passages(visibility);  -- 可见性索引
	CREATE INDEX IF NOT EXISTS idx_passages_published_at ON passages(published_at);  -- 发布时间索引
	CREATE INDEX IF NOT EXISTS idx_passages_scheduled ON passages(is_scheduled, published_at);  -- 定时发布复合索引
	`

	// 添加original_content字段（如果表已存在）
	alterTable := `
	ALTER TABLE passages ADD COLUMN original_content TEXT;
	`

	// 添加file_path字段（如果表已存在）
	alterFilePathTable := `
	ALTER TABLE passages ADD COLUMN file_path TEXT;
	`

	// 添加category字段（如果表已存在）
	alterCategoryTable := `
	ALTER TABLE passages ADD COLUMN category TEXT DEFAULT '未分类';
	`

	// 添加visibility字段（如果表已存在）
	alterVisibilityTable := `
	ALTER TABLE passages ADD COLUMN visibility TEXT DEFAULT 'public';
	`

	// 添加is_scheduled字段（如果表已存在）
	alterIsScheduledTable := `
	ALTER TABLE passages ADD COLUMN is_scheduled INTEGER DEFAULT 0;
	`

	// 添加published_at字段（如果表已存在）
	alterPublishedAtTable := `
	ALTER TABLE passages ADD COLUMN published_at DATETIME;
	`

	// 添加show_title字段（如果表已存在）
	alterShowTitleTable := `
	ALTER TABLE passages ADD COLUMN show_title INTEGER DEFAULT 1;
	`
	
	// 创建用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		role TEXT DEFAULT 'user',
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	
	// 创建访客表
	visitorTable := `
	CREATE TABLE IF NOT EXISTS visitors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT NOT NULL,
		user_agent TEXT,
		visit_date TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_visitors_ip_date ON visitors(ip, visit_date);
	CREATE INDEX IF NOT EXISTS idx_visitors_date ON visitors(visit_date);
	`

	// 创建文章阅读记录表
	articleViewTable := `
	CREATE TABLE IF NOT EXISTS article_views (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		passage_id INTEGER NOT NULL,
		ip TEXT NOT NULL,
		user_agent TEXT,
		country TEXT DEFAULT '',
		city TEXT DEFAULT '',
		region TEXT DEFAULT '',
		view_date TEXT NOT NULL,
		view_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		duration INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (passage_id) REFERENCES passages(id) ON DELETE CASCADE
	);
	-- 优化索引：添加复合索引支持常见查询
	CREATE INDEX IF NOT EXISTS idx_article_views_passage_id ON article_views(passage_id);
	CREATE INDEX IF NOT EXISTS idx_article_views_passage_date ON article_views(passage_id, view_date);  -- 复合索引：文章+日期统计
	CREATE INDEX IF NOT EXISTS idx_article_views_ip_date ON article_views(ip, view_date);
	CREATE INDEX IF NOT EXISTS idx_article_views_date ON article_views(view_date);
	CREATE INDEX IF NOT EXISTS idx_article_views_country ON article_views(country);
	CREATE INDEX IF NOT EXISTS idx_article_views_city_region ON article_views(city, region);  -- 复合索引：城市+地区
	`

	// 创建评论表
	commentTable := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		content TEXT NOT NULL,
		passage_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (passage_id) REFERENCES passages(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_comments_passage_id ON comments(passage_id);
	CREATE INDEX IF NOT EXISTS idx_comments_passage_created ON comments(passage_id, created_at DESC);  -- 复合索引：文章+创建时间
	CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
	`

	// 创建设置表
	settingTable := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		type TEXT DEFAULT 'string',
		description TEXT,
		category TEXT DEFAULT 'system',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_settings_key ON settings(key);
	CREATE INDEX IF NOT EXISTS idx_settings_category ON settings(category);
	`

	// 创建关于页面主卡片表
	aboutMainCardTable := `
	CREATE TABLE IF NOT EXISTS about_main_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		icon TEXT DEFAULT '',
		layout_type TEXT DEFAULT 'default',
		custom_css TEXT DEFAULT '',
		sort_order INTEGER DEFAULT 0,
		is_enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_main_cards_sort ON about_main_cards(sort_order);
	`

	// 创建关于页面次卡片表
	aboutSubCardTable := `
	CREATE TABLE IF NOT EXISTS about_sub_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		main_card_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		icon TEXT DEFAULT '',
		link_url TEXT DEFAULT '',
		layout_type TEXT DEFAULT 'default',
		custom_css TEXT DEFAULT '',
		sort_order INTEGER DEFAULT 0,
		is_enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (main_card_id) REFERENCES about_main_cards(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_sub_cards_main_id ON about_sub_cards(main_card_id);
	CREATE INDEX IF NOT EXISTS idx_sub_cards_sort ON about_sub_cards(sort_order);
	`

	// 创建分类表
	categoryTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT DEFAULT '',
		icon TEXT DEFAULT '',
		sort_order INTEGER DEFAULT 0,
		is_enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
	CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories(sort_order);
	`

	// 创建标签表
	tagTable := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT DEFAULT '',
		color TEXT DEFAULT '#007bff',
		category_id INTEGER DEFAULT 0,
		sort_order INTEGER DEFAULT 0,
		is_enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
	CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category_id);
	CREATE INDEX IF NOT EXISTS idx_tags_sort ON tags(sort_order);
	`

	// 创建附件表
	attachmentTable := `
	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		stored_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_type TEXT NOT NULL,
		content_type TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		passage_id INTEGER,
		visibility TEXT DEFAULT 'public',
		show_in_passage INTEGER DEFAULT 1,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (passage_id) REFERENCES passages(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_attachments_passage_id ON attachments(passage_id);
	CREATE INDEX IF NOT EXISTS idx_attachments_type ON attachments(file_type);
	CREATE INDEX IF NOT EXISTS idx_attachments_visibility ON attachments(visibility);
	CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_at ON attachments(uploaded_at);
	CREATE INDEX IF NOT EXISTS idx_attachments_passage_visibility ON attachments(passage_id, visibility);  -- 复合索引：文章+可见性
	CREATE INDEX IF NOT EXISTS idx_attachments_show_in_passage ON attachments(show_in_passage);  -- 索引：是否在文章中显示
	`

	// 创建音乐表
	musicTable := `
	CREATE TABLE IF NOT EXISTS music_tracks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		duration TEXT DEFAULT '',
		cover_image TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_music_tracks_created_at ON music_tracks(created_at);
	`

	// 执行创建表语句
	if _, err := dbInstance.Exec(passageTable); err != nil {
		return fmt.Errorf("failed to create passages table: %w", err)
	}

	// 尝试添加original_content字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterTable)

	// 尝试添加file_path字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterFilePathTable)

	// 尝试添加category字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterCategoryTable)

	// 尝试添加visibility字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterVisibilityTable)

	// 尝试添加is_scheduled字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterIsScheduledTable)

	// 尝试添加published_at字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterPublishedAtTable)

	// 尝试添加show_title字段（如果已存在会忽略错误）
	_, _ = dbInstance.Exec(alterShowTitleTable)

	// 创建 passages 表索引（在添加列之后）
	if _, err := dbInstance.Exec(passageIndexes); err != nil {
		return fmt.Errorf("failed to create passages indexes: %w", err)
	}

	if _, err := dbInstance.Exec(userTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if _, err := dbInstance.Exec(visitorTable); err != nil {
		return fmt.Errorf("failed to create visitors table: %w", err)
	}

	if _, err := dbInstance.Exec(articleViewTable); err != nil {
		return fmt.Errorf("failed to create article_views table: %w", err)
	}

	if _, err := dbInstance.Exec(commentTable); err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	if _, err := dbInstance.Exec(settingTable); err != nil {
		return fmt.Errorf("failed to create settings table: %w", err)
	}

	if _, err := dbInstance.Exec(aboutMainCardTable); err != nil {
		return fmt.Errorf("failed to create about_main_cards table: %w", err)
	}

	if _, err := dbInstance.Exec(aboutSubCardTable); err != nil {
		return fmt.Errorf("failed to create about_sub_cards table: %w", err)
	}

	if _, err := dbInstance.Exec(categoryTable); err != nil {
		return fmt.Errorf("failed to create categories table: %w", err)
	}

	if _, err := dbInstance.Exec(tagTable); err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	if _, err := dbInstance.Exec(attachmentTable); err != nil {
		return fmt.Errorf("failed to create attachments table: %w", err)
	}

	if _, err := dbInstance.Exec(musicTable); err != nil {
		return fmt.Errorf("failed to create music tracks table: %w", err)
	}

	// 迁移附件表：添加新字段（如果不存在）
	migrations := []string{
		"ALTER TABLE attachments ADD COLUMN visibility TEXT DEFAULT 'public'",
		"ALTER TABLE attachments ADD COLUMN show_in_passage INTEGER DEFAULT 1",
		"ALTER TABLE music_tracks ADD COLUMN cover_image TEXT DEFAULT ''",
	}

	for _, migration := range migrations {
		if _, err := dbInstance.Exec(migration); err != nil {
			// 如果字段已存在，忽略错误
			if !strings.Contains(err.Error(), "duplicate column name") {
				log.Printf("Warning: migration failed: %v", err)
			}
		}
	}

	// 创建性能优化索引（如果不存在）
	indexMigrations := []string{
		// passages 表复合索引
		"CREATE INDEX IF NOT EXISTS idx_passages_status_created ON passages(status, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_passages_category_status ON passages(category, status)",

		// article_views 表复合索引
		"CREATE INDEX IF NOT EXISTS idx_article_views_passage_date ON article_views(passage_id, view_date)",
		"CREATE INDEX IF NOT EXISTS idx_article_views_city_region ON article_views(city, region)",

		// comments 表复合索引
		"CREATE INDEX IF NOT EXISTS idx_comments_passage_created ON comments(passage_id, created_at DESC)",

		// attachments 表复合索引
		"CREATE INDEX IF NOT EXISTS idx_attachments_passage_visibility ON attachments(passage_id, visibility)",
		"CREATE INDEX IF NOT EXISTS idx_attachments_show_in_passage ON attachments(show_in_passage)",
	}

	for _, indexMigration := range indexMigrations {
		if _, err := dbInstance.Exec(indexMigration); err != nil {
			// 索引创建失败通常不是致命错误，记录日志即可
			log.Printf("Warning: index creation failed: %v", err)
		}
	}

	return nil
}

// 插入默认数据
func seedData() error {
	// 插入默认设置
	if err := seedDefaultSettings(); err != nil {
		log.Printf("Warning: failed to seed default settings: %v", err)
	}

	// 插入卡片示例数据
	if err := seedAboutCards(); err != nil {
		log.Printf("Warning: failed to seed about cards: %v", err)
	}

	// 检查是否已有文章数据
	var count int
	err := dbInstance.QueryRow("SELECT COUNT(*) FROM passages").Scan(&count)
	if err != nil {
		return err
	}
	
	if count > 0 {
		log.Println("Articles already exist, skipping markdown import")
	} else {
		// 从 markdown 目录导入文章
		if err := importMarkdownFiles(); err != nil {
			log.Printf("Warning: failed to import markdown files: %v", err)
		} else {
			log.Println("Markdown files imported successfully")
		}
	}
	
	// 检查是否已有用户数据
	var userCount int
	err = dbInstance.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return err
	}
	
	if userCount > 0 {
		return nil // 已有用户，无需插入
	}
	
	// 插入示例用户
	// 使用 Argon2 哈希默认密码
	hashedPassword, err := auth.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash default password: %w", err)
	}

	sampleUsers := []models.User{
		{
			Username: "admin",
			Password: hashedPassword,
			Email:    "admin@example.com",
			Role:     "admin",
			Status:   "active",
		},
	}
	
	for _, user := range sampleUsers {
		if err := userRepo.Create(&user); err != nil {
			return fmt.Errorf("failed to insert sample user: %w", err)
		}
	}

	log.Println("Sample user inserted successfully")
	return nil
}

// seedDefaultSettings 插入默认设置
func seedDefaultSettings() error {
	// 获取所有现有设置的键名
	existingKeys := make(map[string]bool)
	rows, err := dbInstance.Query("SELECT key FROM settings")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return err
		}
		existingKeys[key] = true
	}

	// 默认外观设置
	defaultSettings := []models.Setting{
		{
			Key:         "background_image",
			Value:       "/img/test.webp",
			Type:        "string",
			Description: "页面背景图片路径",
			Category:    "appearance",
		},
		{
			Key:         "global_opacity",
			Value:       "0.15",
			Type:        "number",
			Description: "全局透明度 (0-1)",
			Category:    "appearance",
		},
		{
			Key:         "background_size",
			Value:       "cover",
			Type:        "string",
			Description: "背景图片尺寸 (cover, contain, auto)",
			Category:    "appearance",
		},
		{
			Key:         "background_position",
			Value:       "center",
			Type:        "string",
			Description: "背景图片位置",
			Category:    "appearance",
		},
		{
			Key:         "background_repeat",
			Value:       "no-repeat",
			Type:        "string",
			Description: "背景图片重复方式",
			Category:    "appearance",
		},
		{
			Key:         "background_attachment",
			Value:       "fixed",
			Type:        "string",
			Description: "背景图片滚动方式",
			Category:    "appearance",
		},
		{
			Key:         "blur_amount",
			Value:       "20px",
			Type:        "string",
			Description: "背景模糊程度",
			Category:    "appearance",
		},
		{
			Key:         "saturate_amount",
			Value:       "180%",
			Type:        "string",
			Description: "背景饱和度",
			Category:    "appearance",
		},
		{
			Key:         "dark_mode_enabled",
			Value:       "false",
			Type:        "boolean",
			Description: "是否启用暗色模式",
			Category:    "appearance",
		},
		{
			Key:         "navbar_glass_color",
			Value:       "rgba(220, 138, 221, 0.15)",
			Type:        "string",
			Description: "导航栏毛玻璃颜色",
			Category:    "appearance",
		},
		{
			Key:         "navbar_text_color",
			Value:       "#333333",
			Type:        "string",
			Description: "导航栏文字颜色",
			Category:    "appearance",
		},
		{
			Key:         "card_glass_color",
			Value:       "rgba(220, 138, 221, 0.2)",
			Type:        "string",
			Description: "页面卡片毛玻璃颜色",
			Category:    "appearance",
		},
		{
			Key:         "footer_glass_color",
			Value:       "rgba(220, 138, 221, 0.25)",
			Type:        "string",
			Description: "底栏毛玻璃颜色",
			Category:    "appearance",
		},
		// 默认模板设置
		{
			Key:         "template_name",
			Value:       "欢迎来到我的博客",
			Type:        "string",
			Description: "个人主页标题",
			Category:    "template",
		},
		{
			Key:         "template_greting",
			Value:       "这是一个使用 Go 语言构建的个人博客系统，支持文章管理、数据分析等功能。",
			Type:        "string",
			Description: "首页欢迎语",
			Category:    "template",
		},
		{
			Key:         "template_year",
			Value:       "2026",
			Type:        "string",
			Description: "版权年份",
			Category:    "template",
		},
		{
			Key:         "template_foods",
			Value:       "我的博客",
			Type:        "string",
			Description: "页脚信息",
			Category:    "template",
		},
		{
			Key:         "template_article_title",
			Value:       "true",
			Type:        "boolean",
			Description: "是否显示文章标题",
			Category:    "template",
		},
		{
			Key:         "template_article_title_prefix",
			Value:       "文章",
			Type:        "string",
			Description: "文章标题前缀",
			Category:    "template",
		},
		{
			Key:         "template_switch_notice",
			Value:       "true",
			Type:        "boolean",
			Description: "是否显示切换界面提示",
			Category:    "template",
		},
		{
			Key:         "template_switch_notice_text",
			Value:       "回来继续阅读",
			Type:        "string",
			Description: "切换标签页时显示的提示文字",
			Category:    "template",
		},
		{
			Key:         "external_link_warning",
			Value:       "true",
			Type:        "boolean",
			Description: "是否启用外部链接跳转警告",
			Category:    "template",
		},
		{
			Key:         "external_link_whitelist",
			Value:       "github.com,gitee.com,stackoverflow.com",
			Type:        "string",
			Description: "外部链接白名单（逗号分隔的域名）",
			Category:    "template",
		},
		{
			Key:         "external_link_warning_text",
			Value:       "您即将离开本站，前往外部链接",
			Type:        "string",
			Description: "外部链接警告提示文字",
			Category:    "template",
		},
		{
			Key:         "live2d_enabled",
			Value:       "false",
			Type:        "boolean",
			Description: "是否启用 Live2D 看板娘",
			Category:    "template",
		},
		{
			Key:         "live2d_show_on_index",
			Value:       "true",
			Type:        "boolean",
			Description: "是否在首页显示 Live2D",
			Category:    "template",
		},
		{
			Key:         "live2d_show_on_passage",
			Value:       "true",
			Type:        "boolean",
			Description: "是否在文章页显示 Live2D",
			Category:    "template",
		},
		{
			Key:         "live2d_show_on_collect",
			Value:       "true",
			Type:        "boolean",
			Description: "是否在归档页显示 Live2D",
			Category:    "template",
		},
		{
			Key:         "live2d_show_on_about",
			Value:       "true",
			Type:        "boolean",
			Description: "是否在关于页显示 Live2D",
			Category:    "template",
		},
		{
			Key:         "live2d_show_on_admin",
			Value:       "false",
			Type:        "boolean",
			Description: "是否在管理页显示 Live2D",
			Category:    "template",
		},
		{
			Key:         "live2d_model_id",
			Value:       "1",
			Type:        "string",
			Description: "Live2D 模型 ID",
			Category:    "template",
		},
		{
			Key:         "live2d_model_path",
			Value:       "",
			Type:        "string",
			Description: "Live2D 自定义模型路径（留空使用 CDN）",
			Category:    "template",
		},
		{
			Key:         "live2d_cdn_path",
			Value:       "https://unpkg.com/live2d-widget-model@1.0.5/",
			Type:        "string",
			Description: "Live2D CDN 路径",
			Category:    "template",
		},
		{
			Key:         "live2d_position",
			Value:       "right",
			Type:        "string",
			Description: "Live2D 显示位置（left/right）",
			Category:    "template",
		},
		{
			Key:         "live2d_width",
			Value:       "280px",
			Type:        "string",
			Description: "Live2D 宽度",
			Category:    "template",
		},
		{
			Key:         "live2d_height",
			Value:       "250px",
			Type:        "string",
			Description: "Live2D 高度",
			Category:    "template",
		},
		{
			Key:         "sponsor_enabled",
			Value:       "false",
			Type:        "boolean",
			Description: "是否启用赞助功能",
			Category:    "template",
		},
		{
			Key:         "sponsor_title",
			Value:       "感谢您的支持",
			Type:        "string",
			Description: "赞助模态框标题",
			Category:    "template",
		},
		{
			Key:         "sponsor_image",
			Value:       "/img/avatar.png",
			Type:        "string",
			Description: "赞助图片路径",
			Category:    "template",
		},
		{
			Key:         "sponsor_description",
			Value:       "如果您觉得这个博客对您有帮助，欢迎赞助支持！",
			Type:        "string",
			Description: "赞助描述文字",
			Category:    "template",
		},
		{
			Key:         "sponsor_button_text",
			Value:       "❤️ 赞助支持",
			Type:        "string",
			Description: "赞助按钮文字",
			Category:    "template",
		},
		{
			Key:         "global_avatar",
			Value:       "/img/avatar.webp",
			Type:        "string",
			Description: "全局头像路径",
			Category:    "template",
		},
		// 默认音乐设置
		{
			Key:         "music_enabled",
			Value:       "false",
			Type:        "boolean",
			Description: "是否启用音乐播放器",
			Category:    "appearance",
		},
		{
			Key:         "music_auto_play",
			Value:       "false",
			Type:        "boolean",
			Description: "音乐是否自动播放",
			Category:    "appearance",
		},
		{
			Key:         "music_control_size",
			Value:       "medium",
			Type:        "string",
			Description: "音乐控件大小 (small, medium, large)",
			Category:    "appearance",
		},
		{
			Key:         "music_custom_css",
			Value:       "",
			Type:        "string",
			Description: "音乐播放器自定义CSS样式",
			Category:    "appearance",
		},
		{
			Key:         "music_player_color",
			Value:       "rgba(66, 133, 244, 0.9)",
			Type:        "string",
			Description: "音乐播放器颜色 (RGBA格式)",
			Category:    "appearance",
		},
		{
			Key:         "music_position",
			Value:       "bottom-right",
			Type:        "string",
			Description: "音乐播放器显示位置 (top-left, top-right, bottom-left, bottom-right)",
			Category:    "template",
		},
	}

	insertedCount := 0
	for _, setting := range defaultSettings {
		// 只插入不存在的设置项
		if !existingKeys[setting.Key] {
			if err := settingRepo.Create(&setting); err != nil {
				return fmt.Errorf("failed to insert default setting %s: %w", setting.Key, err)
			}
			insertedCount++
		}
	}

	if insertedCount > 0 {
		log.Printf("Inserted %d new default settings", insertedCount)
	} else {
		log.Println("All default settings already exist")
	}
	return nil
}

// seedAboutCards 插入关于页面卡片示例数据
func seedAboutCards() error {
	// 检查是否已有主卡片数据
	var mainCount int
	err := dbInstance.QueryRow("SELECT COUNT(*) FROM about_main_cards").Scan(&mainCount)
	if err != nil {
		return err
	}

	if mainCount > 0 {
		log.Println("About cards already exist, skipping default cards")
		return nil
	}

	// 插入主卡片示例
	mainCards := []models.AboutMainCard{
		{
			Title:      "项目简介",
			Icon:       "📖",
			LayoutType: "default",
			SortOrder:  1,
			IsEnabled:  true,
		},
		{
			Title:      "核心特性",
			Icon:       "⚡",
			LayoutType: "grid",
			SortOrder:  2,
			IsEnabled:  true,
		},
		{
			Title:      "开发团队",
			Icon:       "👥",
			LayoutType: "grid",
			SortOrder:  3,
			IsEnabled:  true,
		},
		{
			Title:      "联系我们",
			Icon:       "📞",
			LayoutType: "flex",
			SortOrder:  4,
			IsEnabled:  true,
		},
	}

	mainCardIDs := make(map[string]int)

	for i := range mainCards {
		if err := aboutMainCardRepo.Create(&mainCards[i]); err != nil {
			return fmt.Errorf("failed to insert main card %s: %w", mainCards[i].Title, err)
		}
		mainCardIDs[mainCards[i].Title] = mainCards[i].ID
	}

	// 插入次卡片示例
	subCards := []struct {
		mainCardTitle string
		card          models.AboutSubCard
	}{
		{
			"项目简介",
			models.AboutSubCard{
				Title:       "欢迎",
				Description: "欢迎来到我们的网站！这是一个专注于技术分享与知识管理的平台。",
				SortOrder:   1,
				IsEnabled:   true,
			},
		},
		{
			"项目简介",
			models.AboutSubCard{
				Title:       "目标",
				Description: "我们的目标是构建一个开放、友好、专业的技术社区。",
				SortOrder:   2,
				IsEnabled:   true,
			},
		},
		{
			"核心特性",
			models.AboutSubCard{
				Title:       "高性能",
				Description: "采用现代化技术栈，确保网站快速响应。",
				Icon:        "🚀",
				SortOrder:   1,
				IsEnabled:   true,
			},
		},
		{
			"核心特性",
			models.AboutSubCard{
				Title:       "安全可靠",
				Description: "多层安全防护机制，保护用户数据隐私。",
				Icon:        "🔒",
				SortOrder:   2,
				IsEnabled:   true,
			},
		},
		{
			"核心特性",
			models.AboutSubCard{
				Title:       "全平台",
				Description: "响应式设计，各类设备完美呈现。",
				Icon:        "📱",
				SortOrder:   3,
				IsEnabled:   true,
			},
		},
		{
			"核心特性",
			models.AboutSubCard{
				Title:       "开放API",
				Description: "提供完善的API接口，方便集成扩展。",
				Icon:        "🌐",
				SortOrder:   4,
				IsEnabled:   true,
			},
		},
		{
			"开发团队",
			models.AboutSubCard{
				Title:       "技术总监",
				Description: "负责平台架构设计与技术选型。",
				Icon:        "JD",
				SortOrder:   1,
				IsEnabled:   true,
			},
		},
		{
			"开发团队",
			models.AboutSubCard{
				Title:       "前端负责人",
				Description: "专注于用户体验与交互设计。",
				Icon:        "LW",
				SortOrder:   2,
				IsEnabled:   true,
			},
		},
		{
			"开发团队",
			models.AboutSubCard{
				Title:       "后端工程师",
				Description: "负责服务器端逻辑与数据库设计。",
				Icon:        "ZY",
				SortOrder:   3,
				IsEnabled:   true,
			},
		},
		{
			"联系我们",
			models.AboutSubCard{
				Title:       "电子邮件",
				Description: "contact@example.com",
				Icon:        "📧",
				LinkURL:     "mailto:contact@example.com",
				SortOrder:   1,
				IsEnabled:   true,
			},
		},
		{
			"联系我们",
			models.AboutSubCard{
				Title:       "GitHub",
				Description: "github.com/ourproject",
				Icon:        "🐙",
				LinkURL:     "https://github.com/ourproject",
				SortOrder:   2,
				IsEnabled:   true,
			},
		},
		{
			"联系我们",
			models.AboutSubCard{
				Title:       "社交媒体",
				Description: "@ourproject",
				Icon:        "🐦",
				LinkURL:     "https://twitter.com/ourproject",
				SortOrder:   3,
				IsEnabled:   true,
			},
		},
	}

	for _, item := range subCards {
		mainCardID, ok := mainCardIDs[item.mainCardTitle]
		if !ok {
			continue
		}
		item.card.MainCardID = mainCardID
		if err := aboutSubCardRepo.Create(&item.card); err != nil {
			return fmt.Errorf("failed to insert sub card %s: %w", item.card.Title, err)
		}
	}

	log.Println("About cards inserted successfully")
	return nil
}

// importMarkdownFiles 从 markdown 目录导入所有 markdown 文件
func importMarkdownFiles() error {
	markdownDir := "markdown"
	
	// 遍历 markdown 目录
	entries, err := os.ReadDir(markdownDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Markdown directory not found: %s", markdownDir)
			return nil
		}
		return fmt.Errorf("failed to read markdown directory: %w", err)
	}
	
	for _, yearEntry := range entries {
		if !yearEntry.IsDir() {
			continue
		}
		
		yearPath := filepath.Join(markdownDir, yearEntry.Name())
		monthEntries, err := os.ReadDir(yearPath)
		if err != nil {
			log.Printf("Failed to read year directory %s: %v", yearPath, err)
			continue
		}
		
		for _, monthEntry := range monthEntries {
			if !monthEntry.IsDir() {
				continue
			}
			
			monthPath := filepath.Join(yearPath, monthEntry.Name())
			dayEntries, err := os.ReadDir(monthPath)
			if err != nil {
				log.Printf("Failed to read month directory %s: %v", monthPath, err)
				continue
			}
			
			for _, dayEntry := range dayEntries {
				if !dayEntry.IsDir() {
					continue
				}
				
				dayPath := filepath.Join(monthPath, dayEntry.Name())
				fileEntries, err := os.ReadDir(dayPath)
				if err != nil {
					log.Printf("Failed to read day directory %s: %v", dayPath, err)
					continue
				}
				
				for _, fileEntry := range fileEntries {
					if fileEntry.IsDir() {
						continue
					}
					
					filename := fileEntry.Name()
					if filepath.Ext(filename) != ".md" {
						continue
					}
					
					filePath := filepath.Join(dayPath, filename)
					if err := importSingleMarkdownFile(filePath); err != nil {
						log.Printf("Failed to import markdown file %s: %v", filePath, err)
					}
				}
			}
		}
	}
	
	return nil
}

// importSingleMarkdownFile 导入单个 markdown 文件
func importSingleMarkdownFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 提取标题
	title := extractTitle(string(content))

	// 保存原始Markdown内容
	originalContent := string(content)

	// 转换 markdown 为 HTML
	htmlContent, err := convertMarkdownToHTML(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	// 提取摘要（前100个字符）
	summary := extractSummary(htmlContent)

	// 从文件路径中提取日期和文件名
	// 路径格式: markdown/年/月/日/文件名.md
	relativePath := strings.TrimPrefix(filePath, "markdown/")
	relativePath = strings.TrimSuffix(relativePath, ".md")

	parts := strings.Split(relativePath, "/")
	var year, month, day, filename string

	// 提取年、月、日和文件名
	if len(parts) >= 4 {
		year = parts[0]
		month = parts[1]
		day = parts[2]
		filename = parts[3]
	}

	// 构建日期时间
	var createdAt time.Time
	if year != "" && month != "" && day != "" {
		// 使用路径中的日期
		dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
		createdAt, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// 如果解析失败，使用当前时间
			createdAt = time.Now()
		}
	} else {
		// 如果路径中没有日期信息，使用当前时间
		createdAt = time.Now()
	}

	// 比对标题与文件名，如果不一致则重命名文件
	// 文件名可能包含特殊字符，需要进行清理
	cleanedTitle := sanitizeFilename(title)
	
	// 获取目录路径
	dirPath := filepath.Dir(filePath)
	
	// 如果标题与文件名不一致，重命名文件
	if cleanedTitle != filename {
		newFilePath := filepath.Join(dirPath, cleanedTitle+".md")
		
		// 检查新文件名是否已存在
		if _, err := os.Stat(newFilePath); err == nil {
			// 文件已存在，添加时间戳避免冲突
			timestamp := time.Now().Format("20060102-150405")
			cleanedTitle = fmt.Sprintf("%s-%s", cleanedTitle, timestamp)
			newFilePath = filepath.Join(dirPath, cleanedTitle+".md")
		}
		
		// 重命名文件
		if err := os.Rename(filePath, newFilePath); err != nil {
			log.Printf("Warning: failed to rename file %s to %s: %v", filePath, newFilePath, err)
			// 重命名失败，继续使用原文件名
		} else {
			log.Printf("Renamed file: %s -> %s", filepath.Base(filePath), filepath.Base(newFilePath))
			// 更新文件路径
			filePath = newFilePath
			relativePath = strings.TrimPrefix(filePath, "markdown/")
			relativePath = strings.TrimSuffix(relativePath, ".md")
		}
	}

	// 创建文章记录
	passage := &models.Passage{
		Title:           title,
		Content:         htmlContent,
		OriginalContent: originalContent,
		Summary:         summary,
		Author:          "管理员",
		Tags:            `[]`,
		Status:          "published",
		FilePath:        relativePath,
		Visibility:      "public", // 默认为公开
		IsScheduled:     false,   // 默认不定时发布
		CreatedAt:       createdAt,
		UpdatedAt:       time.Now(),
	}

	if err := passageRepo.Create(passage); err != nil {
		return fmt.Errorf("failed to create passage: %w", err)
	}

	log.Printf("Imported: %s (date: %s)", filePath, createdAt.Format("2006-01-02"))
	return nil
}

// sanitizeFilename 清理文件名，移除或替换不安全的字符
func sanitizeFilename(name string) string {
	// 定义不允许的字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	
	// 替换不允许的字符
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// 移除首尾空格
	result = strings.TrimSpace(result)
	
	// 如果结果为空，使用默认名称
	if result == "" {
		result = "未命名文档"
	}
	
	return result
}

// extractTitle 从 markdown 内容中提取标题
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
	}
	return "未命名文档"
}

// extractSummary 从 HTML 内容中提取摘要
func extractSummary(htmlContent string) string {
	// 简单提取纯文本摘要
	summary := htmlContent
	// 移除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	summary = re.ReplaceAllString(summary, "")
	// 去除空白字符
	summary = strings.TrimSpace(summary)
	if len(summary) > 100 {
		summary = summary[:100] + "..."
	}
	return summary
}

// convertMarkdownToHTML 将 markdown 转换为 HTML
func convertMarkdownToHTML(markdownContent []byte) (string, error) {
	var buf bytes.Buffer
	
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	
	if err := md.Convert(markdownContent, &buf); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return dbInstance
}

// GetPassageRepository 获取文章仓库
func GetPassageRepository() repositories.PassageRepository {
	return passageRepo
}

// GetUserRepository 获取用户仓库
func GetUserRepository() repositories.UserRepository {
	return userRepo
}

// GetStatsRepository 获取统计仓库
func GetStatsRepository() repositories.StatsRepository {
	return statsRepo
}

// GetVisitorRepository 获取访客仓库
func GetVisitorRepository() repositories.VisitorRepository {
	return visitorRepo
}

// GetCommentRepository 获取评论仓库
func GetCommentRepository() repositories.CommentRepository {
	return commentRepo
}

// GetSettingRepository 获取设置仓库
func GetSettingRepository() repositories.SettingRepository {
	return settingRepo
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if dbInstance != nil {
		return dbInstance.Close()
	}
	return nil
}

// GetAboutMainCardRepository 获取关于页面主卡片仓库
func GetAboutMainCardRepository() repositories.AboutMainCardRepository {
	return aboutMainCardRepo
}

// GetAboutSubCardRepository 获取关于页面次卡片仓库
func GetAboutSubCardRepository() repositories.AboutSubCardRepository {
	return aboutSubCardRepo
}

// GetAttachmentRepository 获取附件仓库
func GetAttachmentRepository() repositories.AttachmentRepository {
	return attachmentRepo
}

// GetCategoryRepository 获取分类仓库
func GetCategoryRepository() repositories.CategoryRepository {
	return repositories.NewSQLiteCategoryRepository(dbInstance)
}

// GetTagRepository 获取标签仓库
func GetTagRepository() repositories.TagRepository {
	return repositories.NewSQLiteTagRepository(dbInstance)
}

// GetArticleViewRepository 获取文章阅读仓库
func GetArticleViewRepository() repositories.ArticleViewRepository {
	return repositories.NewSQLiteArticleViewRepository(dbInstance)
}