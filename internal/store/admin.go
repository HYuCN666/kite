package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HYuCN666/volans/internal/model"
)

// CreateAdmin 创建管理员，返回新 ID。
func (s *Store) CreateAdmin(username, passwordHash string) (int64, error) {
	res, err := s.db.Exec(
		"INSERT INTO admins (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CountAdmins 返回管理员数量。
func (s *Store) CountAdmins() (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&n)
	return n, err
}

// GetAdminByUsername 按用户名查询管理员。
func (s *Store) GetAdminByUsername(username string) (*model.Admin, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, last_login_at, failed_attempts, locked_until
		 FROM admins WHERE username = ?`,
		username,
	)
	return scanAdmin(row)
}

// UpdateAdminPassword 更新管理员密码。
func (s *Store) UpdateAdminPassword(id int64, hash string) error {
	_, err := s.db.Exec(
		"UPDATE admins SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		hash, id,
	)
	return err
}

// RegisterLoginSuccess 记录登录成功。
func (s *Store) RegisterLoginSuccess(id int64) error {
	_, err := s.db.Exec(
		"UPDATE admins SET last_login_at = ?, failed_attempts = 0, locked_until = NULL WHERE id = ?",
		time.Now().UTC(), id,
	)
	return err
}

// RegisterLoginFailure 记录登录失败，返回失败次数与锁定截止时间。
func (s *Store) RegisterLoginFailure(id int64) (int, time.Time, error) {
	var attempts int
	err := s.db.QueryRow("SELECT failed_attempts FROM admins WHERE id = ?", id).Scan(&attempts)
	if err != nil {
		return 0, time.Time{}, err
	}

	attempts++
	var lockedUntil any
	if attempts >= 5 {
		lockedUntil = time.Now().UTC().Add(15 * time.Minute)
	}
	if _, err := s.db.Exec(
		"UPDATE admins SET failed_attempts = ?, locked_until = ? WHERE id = ?",
		attempts, lockedUntil, id,
	); err != nil {
		return 0, time.Time{}, err
	}

	lu, _ := lockedUntil.(time.Time)
	return attempts, lu, nil
}

func scanAdmin(row *sql.Row) (*model.Admin, error) {
	var a model.Admin
	var lastLogin, locked sql.NullTime
	err := row.Scan(
		&a.ID, &a.Username, &a.PasswordHash, &lastLogin, &a.FailedAttempt, &locked,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		a.LastLoginAt = &lastLogin.Time
	}
	if locked.Valid {
		a.LockedUntil = &locked.Time
	}
	return &a, nil
}
