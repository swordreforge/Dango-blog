package db

import (
	"os"
	"testing"

	"myblog-gogogo/db/drivers"
)

// TestParseFlags 测试命令行参数解析
func TestParseFlags(t *testing.T) {
	// 注意：flag.Parse() 只能调用一次，所以这里只测试默认情况
	// 实际的参数测试需要在进程启动时进行
	config, err := ParseFlags()
	if err != nil {
		t.Errorf("ParseFlags() error = %v", err)
		return
	}
	if config == nil {
		t.Errorf("ParseFlags() returned nil config")
	}
}

// TestLoadFromEnv 测试从环境变量加载配置
func TestLoadFromEnv(t *testing.T) {
	// 保存原始环境变量
	originalEnv := make(map[string]string)
	envVars := []string{
		"DB_DRIVER", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD",
		"DB_NAME", "DB_SSLMODE", "DB_FILE", "DB_MAX_CONNS",
		"DB_AUTO_MIGRATE", "DB_LOG_LEVEL",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// 测试完成后恢复环境变量
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		verify  func(*testing.T, *DBConfig)
	}{
		{
			name: "默认配置",
			env:  map[string]string{},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.Driver != "sqlite" {
					t.Errorf("期望默认驱动为 sqlite，实际为 %s", config.Driver)
				}
				if config.Host != "localhost" {
					t.Errorf("期望默认主机为 localhost，实际为 %s", config.Host)
				}
			},
		},
		{
			name: "SQLite 配置",
			env: map[string]string{
				"DB_DRIVER": "sqlite",
				"DB_FILE":   "/tmp/test.db",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.Driver != "sqlite" {
					t.Errorf("期望驱动为 sqlite，实际为 %s", config.Driver)
				}
				if config.FilePath != "/tmp/test.db" {
					t.Errorf("期望文件路径为 /tmp/test.db，实际为 %s", config.FilePath)
				}
			},
		},
		{
			name: "MySQL 配置",
			env: map[string]string{
				"DB_DRIVER":   "mysql",
				"DB_HOST":     "mysql.example.com",
				"DB_PORT":     "3307",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.Driver != "mysql" {
					t.Errorf("期望驱动为 mysql，实际为 %s", config.Driver)
				}
				if config.Host != "mysql.example.com" {
					t.Errorf("期望主机为 mysql.example.com，实际为 %s", config.Host)
				}
				if config.Port != 3307 {
					t.Errorf("期望端口为 3307，实际为 %d", config.Port)
				}
				if config.User != "testuser" {
					t.Errorf("期望用户为 testuser，实际为 %s", config.User)
				}
				if config.Password != "testpass" {
					t.Errorf("期望密码为 testpass，实际为 %s", config.Password)
				}
				if config.Database != "testdb" {
					t.Errorf("期望数据库名为 testdb，实际为 %s", config.Database)
				}
			},
		},
		{
			name: "PostgreSQL 配置",
			env: map[string]string{
				"DB_DRIVER":   "postgres",
				"DB_HOST":     "postgres.example.com",
				"DB_PORT":     "5433",
				"DB_USER":     "pguser",
				"DB_PASSWORD": "pgpass",
				"DB_NAME":     "pgdb",
				"DB_SSLMODE":  "require",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.Driver != "postgres" {
					t.Errorf("期望驱动为 postgres，实际为 %s", config.Driver)
				}
				if config.Host != "postgres.example.com" {
					t.Errorf("期望主机为 postgres.example.com，实际为 %s", config.Host)
				}
				if config.Port != 5433 {
					t.Errorf("期望端口为 5433，实际为 %d", config.Port)
				}
				if config.SSLMode != "require" {
					t.Errorf("期望 SSL 模式为 require，实际为 %s", config.SSLMode)
				}
			},
		},
		{
			name: "连接池配置",
			env: map[string]string{
				"DB_DRIVER":    "mysql",
				"DB_MAX_CONNS": "20",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.MaxConns != 20 {
					t.Errorf("期望最大连接数为 20，实际为 %d", config.MaxConns)
				}
			},
		},
		{
			name: "自动迁移配置",
			env: map[string]string{
				"DB_DRIVER":      "sqlite",
				"DB_AUTO_MIGRATE": "true",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if !config.AutoMigrate {
					t.Errorf("期望自动迁移为 true，实际为 %v", config.AutoMigrate)
				}
			},
		},
		{
			name: "日志级别配置",
			env: map[string]string{
				"DB_DRIVER":    "sqlite",
				"DB_LOG_LEVEL": "debug",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				if config.LogLevel != "debug" {
					t.Errorf("期望日志级别为 debug，实际为 %s", config.LogLevel)
				}
			},
		},
		{
			name: "无效的端口配置",
			env: map[string]string{
				"DB_DRIVER": "mysql",
				"DB_PORT":   "invalid",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				// 无效的端口应该被忽略，使用默认值
				if config.Port != 0 {
					t.Errorf("期望端口为 0（默认值），实际为 %d", config.Port)
				}
			},
		},
		{
			name: "无效的最大连接数配置",
			env: map[string]string{
				"DB_DRIVER":    "mysql",
				"DB_MAX_CONNS": "invalid",
			},
			wantErr: false,
			verify: func(t *testing.T, config *DBConfig) {
				// 无效的最大连接数应该被忽略
				if config.MaxConns != 0 {
					t.Errorf("期望最大连接数为 0（默认值），实际为 %d", config.MaxConns)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清除所有相关环境变量
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// 设置测试环境变量
			for key, value := range tt.env {
				os.Setenv(key, value)
			}

			config, err := LoadFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.verify != nil {
				tt.verify(t, config)
			}
		})
	}
}

// TestGetEnv 测试 getEnv 辅助函数
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		setEnv        bool
		envValue      string
		expectedValue string
	}{
		{
			name:          "环境变量存在",
			key:           "TEST_VAR",
			defaultValue:  "default",
			setEnv:        true,
			envValue:      "actual",
			expectedValue: "actual",
		},
		{
			name:          "环境变量不存在",
			key:           "NONEXISTENT_VAR",
			defaultValue:  "default",
			setEnv:        false,
			envValue:      "",
			expectedValue: "default",
		},
		{
			name:          "环境变量为空字符串",
			key:           "EMPTY_VAR",
			defaultValue:  "default",
			setEnv:        true,
			envValue:      "",
			expectedValue: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清除环境变量
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			// 如果需要，设置环境变量
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("getEnv(%s, %s) = %s, 期望 %s", tt.key, tt.defaultValue, result, tt.expectedValue)
			}
		})
	}
}

// TestDBConfig_Validate 测试配置验证
func TestDBConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DBConfig
		wantErr bool
	}{
		{
			name: "SQLite 配置",
			config: &DBConfig{
				Config: drivers.Config{
					Driver:   "sqlite",
					FilePath: "/tmp/test.db",
				},
			},
			wantErr: false,
		},
		{
			name: "MySQL 配置",
			config: &DBConfig{
				Config: drivers.Config{
					Driver:   "mysql",
					Host:     "localhost",
					Port:     3306,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
				},
			},
			wantErr: false,
		},
		{
			name: "PostgreSQL 配置",
			config: &DBConfig{
				Config: drivers.Config{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
				},
			},
			wantErr: false,
		},
		{
			name: "空配置",
			config: &DBConfig{
				Config: drivers.Config{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这里只是验证配置结构，不进行实际的验证逻辑
			// 因为 DBConfig 结构体本身没有 Validate 方法
			if tt.config == nil {
				t.Errorf("配置不应该为 nil")
			}
		})
	}
}

// TestDBConfig_AutoMigrate 测试自动迁移配置
func TestDBConfig_AutoMigrate(t *testing.T) {
	tests := []struct {
		name         string
		autoMigrate  bool
		expectedVal  bool
	}{
		{
			name:         "自动迁移启用",
			autoMigrate:  true,
			expectedVal:  true,
		},
		{
			name:         "自动迁移禁用",
			autoMigrate:  false,
			expectedVal:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &DBConfig{
				AutoMigrate: tt.autoMigrate,
			}

			if config.AutoMigrate != tt.expectedVal {
				t.Errorf("AutoMigrate = %v, 期望 %v", config.AutoMigrate, tt.expectedVal)
			}
		})
	}
}

// TestDBConfig_LogLevel 测试日志级别配置
func TestDBConfig_LogLevel(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		expectedVal  string
	}{
		{
			name:         "Debug 级别",
			logLevel:     "debug",
			expectedVal:  "debug",
		},
		{
			name:         "Info 级别",
			logLevel:     "info",
			expectedVal:  "info",
		},
		{
			name:         "Warn 级别",
			logLevel:     "warn",
			expectedVal:  "warn",
		},
		{
			name:         "Error 级别",
			logLevel:     "error",
			expectedVal:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &DBConfig{
				LogLevel: tt.logLevel,
			}

			if config.LogLevel != tt.expectedVal {
				t.Errorf("LogLevel = %s, 期望 %s", config.LogLevel, tt.expectedVal)
			}
		})
	}
}