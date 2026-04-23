// Package mysqllogic 的单元测试：验证 Init 行为及连接是否可用。
package mysqllogic

import (
	"os"
	"testing"

	"github.com/wangzhione/sbp/chain"
)

var ctx = chain.Context()

// testMySQLCommand 优先读取环境变量，避免占位命令在新版本解析器下被当成非法参数。
func testMySQLCommand() string {
	return os.Getenv("SBP_TEST_MYSQL")
}

// TestInit_ValidCommand_CanConnect 使用固定 DSN 初始化并 Ping，连接成功则输出成功。
func TestInit_ValidCommand_CanConnect(t *testing.T) {
	command := testMySQLCommand()
	if command == "" {
		t.Skip("skip mysql integration test: env SBP_TEST_MYSQL is empty")
	}

	err := Init(ctx, command)
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	if Main == nil {
		t.Fatal("Init 成功时 Main 不应为 nil")
	}

	db := Main.DB()
	if db == nil {
		t.Fatal("Main.DB() 不应为 nil")
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("数据库 Ping 失败: %v", err)
	}
	t.Log("MySQL 连接成功")
}
