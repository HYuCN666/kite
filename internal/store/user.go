package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HYuCN666/volans/internal/model"
)

const userColumns = `id, inbound_id, email, uuid, remark, quota_bytes, used_uplink, used_downlink,
	speed_limit_uplink, speed_limit_downlink, expire_at, enabled, created_at, updated_at`

// ListUsers 返回（可选按入站过滤的）用户列表。
func (s *Store) ListUsers(inboundID int64) ([]model.User, error) {
	query := "SELECT " + userColumns + " FROM users"
	args := []any{}
	if inboundID > 0 {
		query += " WHERE inbound_id = ?"
		args = append(args, inboundID)
	}
	query += " ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

// GetUser 按 ID 查询用户。
func (s *Store) GetUser(id int64) (*model.User, error) {
	row := s.db.QueryRow("SELECT "+userColumns+" FROM users WHERE id = ?", id)
	return scanUser(row)
}

// CreateUser 创建用户。
func (s *Store) CreateUser(u *model.User) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO users
		 (inbound_id, email, uuid, remark, quota_bytes, used_uplink, used_downlink,
		  speed_limit_uplink, speed_limit_downlink, expire_at, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.InboundID, u.Email, u.UUID, u.Remark, u.QuotaBytes, u.UsedUplink, u.UsedDownlink,
		u.SpeedLimitUplink, u.SpeedLimitDownlink, u.ExpireAt, boolToInt(u.Enabled),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateUser 更新用户。
func (s *Store) UpdateUser(u *model.User) error {
	_, err := s.db.Exec(
		`UPDATE users SET remark = ?, quota_bytes = ?, speed_limit_uplink = ?,
		 speed_limit_downlink = ?, expire_at = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		u.Remark, u.QuotaBytes, u.SpeedLimitUplink, u.SpeedLimitDownlink,
		u.ExpireAt, boolToInt(u.Enabled), u.ID,
	)
	return err
}

// DeleteUser 删除用户。
func (s *Store) DeleteUser(id int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// AddUserTraffic 累加用户已用流量。
func (s *Store) AddUserTraffic(id int64, uplink, downlink int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET used_uplink = used_uplink + ?, used_downlink = used_downlink + ? WHERE id = ?",
		uplink, downlink, id,
	)
	return err
}

// SetUserEnabled 设置用户启用状态。
func (s *Store) SetUserEnabled(id int64, enabled bool) error {
	_, err := s.db.Exec("UPDATE users SET enabled = ? WHERE id = ?", boolToInt(enabled), id)
	return err
}

// ResetUserTraffic 清零用户已用流量。
func (s *Store) ResetUserTraffic(id int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET used_uplink = 0, used_downlink = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	return err
}

// DisableExpiredUsers 停用到期的用户，返回受影响数量。
func (s *Store) DisableExpiredUsers(now time.Time) (int64, error) {
	res, err := s.db.Exec(
		"UPDATE users SET enabled = 0 WHERE enabled = 1 AND expire_at IS NOT NULL AND expire_at < ?",
		now.UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// TrafficOverview 返回总流量与活跃用户数。
func (s *Store) TrafficOverview() (uplink, downlink int64, activeUsers int, err error) {
	err = s.db.QueryRow(
		"SELECT COALESCE(SUM(used_uplink),0), COALESCE(SUM(used_downlink),0) FROM users",
	).Scan(&uplink, &downlink)
	if err != nil {
		return
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE enabled = 1").Scan(&activeUsers)
	return
}

func scanUser(scanner interface{ Scan(...any) error }) (*model.User, error) {
	var u model.User
	var remark, email, uuid sql.NullString
	var expire sql.NullTime
	var enabled int
	err := scanner.Scan(
		&u.ID, &u.InboundID, &email, &uuid, &remark, &u.QuotaBytes,
		&u.UsedUplink, &u.UsedDownlink, &u.SpeedLimitUplink, &u.SpeedLimitDownlink,
		&expire, &enabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Email = email.String
	u.UUID = uuid.String
	u.Remark = remark.String
	if expire.Valid {
		u.ExpireAt = &expire.Time
	}
	u.Enabled = enabled == 1
	return &u, nil
}
