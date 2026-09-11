package store

import (
	"database/sql"
	"errors"

	"github.com/HYuCN666/volans/internal/model"
)

const serverColumns = `id, name, host, port, username, auth_type, password, private_key, enabled, created_at, updated_at`

// ListServers 返回全部节点服务器。
func (s *Store) ListServers() ([]model.Server, error) {
	rows, err := s.db.Query("SELECT " + serverColumns + " FROM servers ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Server
	for rows.Next() {
		sv, err := scanServer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *sv)
	}
	return out, rows.Err()
}

// GetServer 按 ID 查询服务器。
func (s *Store) GetServer(id int64) (*model.Server, error) {
	row := s.db.QueryRow("SELECT "+serverColumns+" FROM servers WHERE id = ?", id)
	return scanServer(row)
}

// CreateServer 创建服务器。
func (s *Store) CreateServer(sv *model.Server) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO servers (name, host, port, username, auth_type, password, private_key, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sv.Name, sv.Host, sv.Port, sv.Username, sv.AuthType, sv.Password, sv.PrivateKey, boolToInt(sv.Enabled),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateServer 更新服务器（密钥字段仅在非空时更新）。
func (s *Store) UpdateServer(sv *model.Server) error {
	_, err := s.db.Exec(
		`UPDATE servers SET name = ?, host = ?, port = ?, username = ?, auth_type = ?,
		 password = CASE WHEN ? != '' THEN ? ELSE password END,
		 private_key = CASE WHEN ? != '' THEN ? ELSE private_key END,
		 enabled = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		sv.Name, sv.Host, sv.Port, sv.Username, sv.AuthType,
		sv.Password, sv.Password, sv.PrivateKey, sv.PrivateKey,
		boolToInt(sv.Enabled), sv.ID,
	)
	return err
}

// DeleteServer 删除服务器。
func (s *Store) DeleteServer(id int64) error {
	_, err := s.db.Exec("DELETE FROM servers WHERE id = ?", id)
	return err
}

func scanServer(scanner interface{ Scan(...any) error }) (*model.Server, error) {
	var sv model.Server
	var password, privateKey sql.NullString
	var enabled int
	err := scanner.Scan(
		&sv.ID, &sv.Name, &sv.Host, &sv.Port, &sv.Username, &sv.AuthType,
		&password, &privateKey, &enabled, &sv.CreatedAt, &sv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sv.Password = password.String
	sv.PrivateKey = privateKey.String
	sv.Enabled = enabled == 1
	return &sv, nil
}
