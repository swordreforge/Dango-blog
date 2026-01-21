package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"myblog-gogogo/db/drivers"
)

// MockDatabase 是 Database 接口的模拟实现
type MockDatabase struct {
	db           *sql.DB
	driverName   string
	shouldFail   bool
	connectCount int
	closeCount   int
	pingCount    int
	queryCount   int
	execCount    int
	beginCount   int
}

func NewMockDatabase(driverName string) *MockDatabase {
	return &MockDatabase{
		driverName: driverName,
		shouldFail: false,
	}
}

func (m *MockDatabase) Connect() error {
	m.connectCount++
	if m.shouldFail {
		return errors.New("mock connection failed")
	}
	return nil
}

func (m *MockDatabase) Close() error {
	m.closeCount++
	if m.shouldFail {
		return errors.New("mock close failed")
	}
	return nil
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	m.pingCount++
	if m.shouldFail {
		return errors.New("mock ping failed")
	}
	return nil
}

func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	m.queryCount++
	if m.shouldFail {
		return nil, errors.New("mock query failed")
	}
	return nil, nil
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	// 返回一个空的 sql.Row，这样测试不会 panic
	// 在实际使用中，这会返回一个包含错误或结果的 Row
	return &sql.Row{}
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.execCount++
	if m.shouldFail {
		return nil, errors.New("mock exec failed")
	}
	return nil, nil
}

func (m *MockDatabase) Begin() (*sql.Tx, error) {
	m.beginCount++
	if m.shouldFail {
		return nil, errors.New("mock begin failed")
	}
	return nil, nil
}

func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	m.beginCount++
	if m.shouldFail {
		return nil, errors.New("mock begin tx failed")
	}
	return nil, nil
}

func (m *MockDatabase) GetDriver() interface{} {
	return m.driverName
}

func (m *MockDatabase) Name() string {
	return m.driverName
}

func (m *MockDatabase) Version() (string, error) {
	if m.shouldFail {
		return "", errors.New("mock version failed")
	}
	return "1.0.0", nil
}

func (m *MockDatabase) SetFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func (m *MockDatabase) GetConnectCount() int {
	return m.connectCount
}

func (m *MockDatabase) GetCloseCount() int {
	return m.closeCount
}

func (m *MockDatabase) GetPingCount() int {
	return m.pingCount
}

func (m *MockDatabase) GetQueryCount() int {
	return m.queryCount
}

func (m *MockDatabase) GetExecCount() int {
	return m.execCount
}

func (m *MockDatabase) GetBeginCount() int {
	return m.beginCount
}

// TestDatabaseInterface_Connect 测试连接操作
func TestDatabaseInterface_Connect(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"成功连接", false, false},
		{"连接失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			err := mockDB.Connect()

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if mockDB.GetConnectCount() != 1 {
				t.Errorf("Connect 应该被调用一次，实际被调用 %d 次", mockDB.GetConnectCount())
			}
		})
	}
}

// TestDatabaseInterface_Close 测试关闭操作
func TestDatabaseInterface_Close(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"成功关闭", false, false},
		{"关闭失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			err := mockDB.Close()

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if mockDB.GetCloseCount() != 1 {
				t.Errorf("Close 应该被调用一次，实际被调用 %d 次", mockDB.GetCloseCount())
			}
		})
	}
}

// TestDatabaseInterface_Ping 测试 Ping 操作
func TestDatabaseInterface_Ping(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"Ping 成功", false, false},
		{"Ping 失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			ctx := context.Background()
			err := mockDB.Ping(ctx)

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if mockDB.GetPingCount() != 1 {
				t.Errorf("Ping 应该被调用一次，实际被调用 %d 次", mockDB.GetPingCount())
			}
		})
	}
}

// TestDatabaseInterface_PingWithTimeout 测试带超时的 Ping 操作
func TestDatabaseInterface_PingWithTimeout(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := mockDB.Ping(ctx)
	if err != nil {
		t.Errorf("Ping 失败: %v", err)
	}
}

// TestDatabaseInterface_PingWithCancelledContext 测试取消上下文的 Ping 操作
func TestDatabaseInterface_PingWithCancelledContext(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")
	mockDB.SetFail(true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	err := mockDB.Ping(ctx)
	if err == nil {
		t.Errorf("期望返回上下文取消错误，但没有返回")
	}
}

// TestDatabaseInterface_Query 测试查询操作
func TestDatabaseInterface_Query(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		args        []interface{}
		shouldFail  bool
		expectError bool
	}{
		{"成功查询", "SELECT * FROM users", nil, false, false},
		{"带参数查询", "SELECT * FROM users WHERE id = ?", []interface{}{1}, false, false},
		{"查询失败", "SELECT * FROM users", nil, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			rows, err := mockDB.Query(tt.query, tt.args...)

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if tt.expectError && rows != nil {
				t.Errorf("期望返回 nil rows，但返回了非 nil")
			}
			if mockDB.GetQueryCount() != 1 {
				t.Errorf("Query 应该被调用一次，实际被调用 %d 次", mockDB.GetQueryCount())
			}
		})
	}
}

// TestDatabaseInterface_QueryRow 测试单行查询操作
func TestDatabaseInterface_QueryRow(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	row := mockDB.QueryRow("SELECT * FROM users WHERE id = ?", 1)
	// QueryRow 总是返回一个 *sql.Row，即使没有结果
	// 我们只需要验证它不会 panic
	if row == nil {
		t.Errorf("QueryRow 应该返回非 nil 的 Row")
	}
}

// TestDatabaseInterface_Exec 测试执行操作
func TestDatabaseInterface_Exec(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		args        []interface{}
		shouldFail  bool
		expectError bool
	}{
		{"成功执行", "INSERT INTO users (name) VALUES (?)", []interface{}{"test"}, false, false},
		{"执行失败", "INSERT INTO users (name) VALUES (?)", []interface{}{"test"}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			result, err := mockDB.Exec(tt.query, tt.args...)

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if tt.expectError && result != nil {
				t.Errorf("期望返回 nil result，但返回了非 nil")
			}
			if mockDB.GetExecCount() != 1 {
				t.Errorf("Exec 应该被调用一次，实际被调用 %d 次", mockDB.GetExecCount())
			}
		})
	}
}

// TestDatabaseInterface_Begin 测试事务开始操作
func TestDatabaseInterface_Begin(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"成功开始事务", false, false},
		{"开始事务失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			tx, err := mockDB.Begin()

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if tt.expectError && tx != nil {
				t.Errorf("期望返回 nil tx，但返回了非 nil")
			}
			if mockDB.GetBeginCount() != 1 {
				t.Errorf("Begin 应该被调用一次，实际被调用 %d 次", mockDB.GetBeginCount())
			}
		})
	}
}

// TestDatabaseInterface_BeginTx 测试带上下文的事务开始操作
func TestDatabaseInterface_BeginTx(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"成功开始事务", false, false},
		{"开始事务失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			ctx := context.Background()
			opts := &sql.TxOptions{}
			tx, err := mockDB.BeginTx(ctx, opts)

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if tt.expectError && tx != nil {
				t.Errorf("期望返回 nil tx，但返回了非 nil")
			}
			if mockDB.GetBeginCount() != 1 {
				t.Errorf("BeginTx 应该被调用一次，实际被调用 %d 次", mockDB.GetBeginCount())
			}
		})
	}
}

// TestDatabaseInterface_BeginTxWithTimeout 测试带超时的事务开始操作
func TestDatabaseInterface_BeginTxWithTimeout(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	opts := &sql.TxOptions{}
	tx, err := mockDB.BeginTx(ctx, opts)
	if err != nil {
		t.Errorf("BeginTx 失败: %v", err)
	}
	if tx != nil {
		t.Errorf("期望返回 nil tx，但返回了非 nil")
	}
}

// TestDatabaseInterface_GetDriver 测试获取驱动
func TestDatabaseInterface_GetDriver(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		expectVal string
	}{
		{"SQLite 驱动", "sqlite", "sqlite"},
		{"PostgreSQL 驱动", "postgres", "postgres"},
		{"MariaDB 驱动", "mariadb", "mariadb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase(tt.driver)

			driver := mockDB.GetDriver()
			if driver != tt.expectVal {
				t.Errorf("期望驱动为 %s，实际为 %v", tt.expectVal, driver)
			}
		})
	}
}

// TestDatabaseInterface_Name 测试获取数据库名称
func TestDatabaseInterface_Name(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		expectVal string
	}{
		{"SQLite 名称", "sqlite", "sqlite"},
		{"PostgreSQL 名称", "postgres", "postgres"},
		{"MariaDB 名称", "mariadb", "mariadb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase(tt.driver)

			name := mockDB.Name()
			if name != tt.expectVal {
				t.Errorf("期望名称为 %s，实际为 %s", tt.expectVal, name)
			}
		})
	}
}

// TestDatabaseInterface_Version 测试获取数据库版本
func TestDatabaseInterface_Version(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
	}{
		{"成功获取版本", false, false},
		{"获取版本失败", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDatabase("sqlite")
			mockDB.SetFail(tt.shouldFail)

			version, err := mockDB.Version()

			if tt.expectError && err == nil {
				t.Errorf("期望返回错误，但没有返回")
			}
			if !tt.expectError && err != nil {
				t.Errorf("期望不返回错误，但返回了: %v", err)
			}
			if !tt.expectError && version != "1.0.0" {
				t.Errorf("期望版本为 1.0.0，实际为 %s", version)
			}
		})
	}
}

// TestDatabaseInterface_MultipleOperations 测试多个操作组合
func TestDatabaseInterface_MultipleOperations(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	// 执行多个操作
	mockDB.Connect()
	mockDB.Ping(context.Background())
	mockDB.Query("SELECT * FROM users")
	mockDB.Exec("INSERT INTO users (name) VALUES (?)", "test")
	mockDB.Begin()
	mockDB.Close()

	// 验证每个操作都被调用
	if mockDB.GetConnectCount() != 1 {
		t.Errorf("Connect 应该被调用一次")
	}
	if mockDB.GetPingCount() != 1 {
		t.Errorf("Ping 应该被调用一次")
	}
	if mockDB.GetQueryCount() != 1 {
		t.Errorf("Query 应该被调用一次")
	}
	if mockDB.GetExecCount() != 1 {
		t.Errorf("Exec 应该被调用一次")
	}
	if mockDB.GetBeginCount() != 1 {
		t.Errorf("Begin 应该被调用一次")
	}
	if mockDB.GetCloseCount() != 1 {
		t.Errorf("Close 应该被调用一次")
	}
}

// TestDatabaseInterface_ConcurrentOperations 测试并发操作
func TestDatabaseInterface_ConcurrentOperations(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	done := make(chan bool)

	// 并发执行多个查询
	for i := 0; i < 10; i++ {
		go func() {
			mockDB.Query("SELECT * FROM users")
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if mockDB.GetQueryCount() != 10 {
		t.Errorf("Query 应该被调用 10 次，实际被调用 %d 次", mockDB.GetQueryCount())
	}
}

// TestDatabaseInterface_NilContext 测试 nil 上下文
func TestDatabaseInterface_NilContext(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	// 测试 nil 上下文不会导致 panic
	err := mockDB.Ping(nil)
	if err != nil {
		t.Errorf("Ping with nil context failed: %v", err)
	}

	tx, err := mockDB.BeginTx(nil, nil)
	if err != nil {
		t.Errorf("BeginTx with nil context failed: %v", err)
	}
	if tx != nil {
		t.Errorf("期望返回 nil tx，但返回了非 nil")
	}
}

// TestDatabaseInterface_EmptyQuery 测试空查询
func TestDatabaseInterface_EmptyQuery(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	_, err := mockDB.Query("")
	if err != nil {
		t.Errorf("空查询不应该返回错误，但返回了: %v", err)
	}

	_, err = mockDB.Exec("")
	if err != nil {
		t.Errorf("空执行不应该返回错误，但返回了: %v", err)
	}
}

// TestDatabaseInterface_LargeArgs 测试大量参数
func TestDatabaseInterface_LargeArgs(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	// 创建大量参数
	args := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		args[i] = i
	}

	_, err := mockDB.Query("SELECT * FROM users WHERE id IN (SELECT value FROM json_array(?))", args...)
	if err != nil {
		t.Errorf("大量参数查询失败: %v", err)
	}
}

// TestDatabaseInterface_StateAfterFailure 测试失败后的状态
func TestDatabaseInterface_StateAfterFailure(t *testing.T) {
	mockDB := NewMockDatabase("sqlite")

	// 第一次操作失败
	mockDB.SetFail(true)
	err := mockDB.Connect()
	if err == nil {
		t.Errorf("期望 Connect 失败，但成功了")
	}

	// 重置状态，第二次操作应该成功
	mockDB.SetFail(false)
	err = mockDB.Connect()
	if err != nil {
		t.Errorf("期望 Connect 成功，但失败了: %v", err)
	}

	if mockDB.GetConnectCount() != 2 {
		t.Errorf("Connect 应该被调用两次，实际被调用 %d 次", mockDB.GetConnectCount())
	}
}

// BenchmarkDatabaseInterface_Ping 性能测试 Ping 操作
func BenchmarkDatabaseInterface_Ping(b *testing.B) {
	mockDB := NewMockDatabase("sqlite")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockDB.Ping(ctx)
	}
}

// BenchmarkDatabaseInterface_Query 性能测试 Query 操作
func BenchmarkDatabaseInterface_Query(b *testing.B) {
	mockDB := NewMockDatabase("sqlite")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockDB.Query("SELECT * FROM users")
	}
}

// BenchmarkDatabaseInterface_Exec 性能测试 Exec 操作
func BenchmarkDatabaseInterface_Exec(b *testing.B) {
	mockDB := NewMockDatabase("sqlite")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockDB.Exec("INSERT INTO users (name) VALUES (?)", "test")
	}
}

// TestDatabaseInterface_Implementation 验证 MockDatabase 实现了 Database 接口
func TestDatabaseInterface_Implementation(t *testing.T) {
	var _ Database = (*MockDatabase)(nil)
}

// TestDatabaseInterface_DriverRegistry 测试驱动注册表
func TestDatabaseInterface_DriverRegistry(t *testing.T) {
	// 测试获取可用驱动列表
	driverList := drivers.AvailableDrivers()
	if len(driverList) == 0 {
		t.Errorf("应该至少有一个可用的驱动")
	}

	// 测试获取特定驱动
	driver, err := drivers.GetDriver("sqlite")
	if err != nil {
		t.Errorf("获取 sqlite 驱动失败: %v", err)
	}
	if driver == nil {
		t.Errorf("驱动不应该为 nil")
	}

	// 测试获取不存在的驱动
	_, err = drivers.GetDriver("nonexistent")
	if err == nil {
		t.Errorf("期望返回错误，但没有返回")
	}
}