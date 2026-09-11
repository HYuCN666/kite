package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HYuCN666/volans/internal/model"
)

const adminColumns = `id, username, password_hash, role, totp_secret, last_login_at, failed_attempts, locked_until, created_at, updated_at`

// CreateAdmin 创建管理员，返回新 ID。
func (s *Store) CreateAdmin(username, passwordHash, role string) (int64, error) {
	if role == "" {
		role = "admin"
	}
	res, err := s.db.Exec(
		"INSERT INTO admins (username, password_hash, role) VALUES (?, ?, ?)",
		username, passwordHash, role,
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

// ListAdmins 返回全部管理员。
func (s *Store) ListAdmins() ([]model.Admin, error) {
	rows, err := s.db.Query("SELECT " + adminColumns + " FROM admins ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Admin
	for rows.Next() {
		a, err := scanAdminRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// GetAdminByID 按 ID 查询管理员。
func (s *Store) GetAdminByID(id int64) (*model.Admin, error) {
	row := s.db.QueryRow("SELECT "+adminColumns+" FROM admins WHERE id = ?", id)
	return scanAdmin(row)
}

// GetAdminByUsername 按用户名查询管理员。
func (s *Store) GetAdminByUsername(username string) (*model.Admin, error) {
	row := s.db.QueryRow(
		"SELECT "+adminColumns+" FROM admins WHERE username = ?",
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

// UpdateAdminRole 更新管理员角色。
func (s *Store) UpdateAdminRole(id int64, role string) error {
	_, err := s.db.Exec(
		"UPDATE admins SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		role, id,
	)
	return err
}

// UpdateAdminTOTPSecret 更新管理员 TOTP 密钥。
func (s *Store) UpdateAdminTOTPSecret(id int64, secret string) error {
	_, err := s.db.Exec(
		"UPDATE admins SET totp_secret = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		secret, id,
	)
	return err
}

// DeleteAdmin 删除管理员。
func (s *Store) DeleteAdmin(id int64) error {
	_, err := s.db.Exec("DELETE FROM admins WHERE id = ?", id)
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
	var totpSecret sql.NullString
	var lastLogin, locked sql.NullTime
	err := row.Scan(
		&a.ID, &a.Username, &a.PasswordHash, &a.Role, &totpSecret, &lastLogin, &a.FailedAttempt, &locked,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.TOTPSecret = totpSecret.String
	if lastLogin.Valid {
		a.LastLoginAt = &lastLogin.Time
	}
	if locked.Valid {
		a.LockedUntil = &locked.Time
	}
	return &a, nil
}

func scanAdminRow(scanner interface{ Scan(...any) error }) (*model.Admin, error) {
	var a model.Admin
	var totpSecret sql.NullString
	var lastLogin, locked sql.NullTime
	err := scanner.Scan(
		&a.ID, &a.Username, &a.PasswordHash, &a.Role, &totpSecret, &lastLogin, &a.FailedAttempt, &locked,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.TOTPSecret = totpSecret.String
	if lastLogin.Valid {
		a.LastLoginAt = &lastLogin.Time
	}
	if locked.Valid {
		a.LockedUntil = &locked.Time
	}
	return &a, nil
}
