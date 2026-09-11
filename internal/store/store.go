package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 数据访问。
type Store struct {
	db *sql.DB
}

// Open 打开（必要时创建）数据库并执行迁移。
func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "kite.db"))
	if err != nil {
		return nil, err
	}

	// SQLite 单写者模型，限制为单连接避免 database is locked。
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// DB 返回底层连接。
func (s *Store) DB() *sql.DB { return s.db }

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS settings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key TEXT NOT NULL UNIQUE,
	value TEXT,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admins (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	last_login_at DATETIME,
	failed_attempts INTEGER NOT NULL DEFAULT 0,
	locked_until DATETIME,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS inbounds (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	tag TEXT NOT NULL,
	remark TEXT,
	protocol TEXT NOT NULL,
	port INTEGER NOT NULL,
	listen TEXT NOT NULL DEFAULT '0.0.0.0',
	transport TEXT NOT NULL,
	stream_settings TEXT,
	tls_enabled INTEGER NOT NULL DEFAULT 0,
	tls_cert TEXT,
	tls_key TEXT,
	tls_server_name TEXT,
	enable_sniffing INTEGER NOT NULL DEFAULT 1,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	inbound_id INTEGER NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
	email TEXT NOT NULL UNIQUE,
	uuid TEXT NOT NULL UNIQUE,
	remark TEXT,
	quota_bytes INTEGER NOT NULL DEFAULT 0,
	used_uplink INTEGER NOT NULL DEFAULT 0,
	used_downlink INTEGER NOT NULL DEFAULT 0,
	speed_limit_uplink INTEGER NOT NULL DEFAULT 0,
	speed_limit_downlink INTEGER NOT NULL DEFAULT 0,
	expire_at DATETIME,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_inbound ON users(inbound_id);

CREATE TABLE IF NOT EXISTS subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	token TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS traffic_snapshots (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	inbound_id INTEGER NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
	uplink INTEGER NOT NULL DEFAULT 0,
	downlink INTEGER NOT NULL DEFAULT 0,
	recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_snapshots_user_time ON traffic_snapshots(user_id, recorded_at);
`
