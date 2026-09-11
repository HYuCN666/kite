package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/HYuCN666/volans/internal/model"
)

const inboundColumns = `id, tag, remark, protocol, port, listen, transport, stream_settings,
	tls_enabled, tls_cert, tls_key, tls_server_name, enable_sniffing, enabled, created_at, updated_at`

// ListInbounds 返回全部入站。
func (s *Store) ListInbounds() ([]model.Inbound, error) {
	rows, err := s.db.Query("SELECT " + inboundColumns + " FROM inbounds ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Inbound
	for rows.Next() {
		in, err := scanInbound(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *in)
	}
	return out, rows.Err()
}

// GetInbound 按 ID 查询入站。
func (s *Store) GetInbound(id int64) (*model.Inbound, error) {
	row := s.db.QueryRow("SELECT "+inboundColumns+" FROM inbounds WHERE id = ?", id)
	return scanInbound(row)
}

// CreateInbound 创建入站。
func (s *Store) CreateInbound(in *model.Inbound) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO inbounds
		 (tag, remark, protocol, port, listen, transport, stream_settings,
		  tls_enabled, tls_cert, tls_key, tls_server_name, enable_sniffing, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Tag, in.Remark, in.Protocol, in.Port, in.Listen, in.Transport, in.StreamSettings,
		boolToInt(in.TLSEnabled), in.TLSCert, in.TLSKey, in.TLSServerName,
		boolToInt(in.EnableSniffing), boolToInt(in.Enabled),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateInbound 更新入站。
func (s *Store) UpdateInbound(in *model.Inbound) error {
	_, err := s.db.Exec(
		`UPDATE inbounds SET remark = ?, protocol = ?, port = ?, listen = ?, transport = ?,
		 stream_settings = ?, tls_enabled = ?, tls_cert = ?, tls_key = ?, tls_server_name = ?,
		 enable_sniffing = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		in.Remark, in.Protocol, in.Port, in.Listen, in.Transport, in.StreamSettings,
		boolToInt(in.TLSEnabled), in.TLSCert, in.TLSKey, in.TLSServerName,
		boolToInt(in.EnableSniffing), boolToInt(in.Enabled), in.ID,
	)
	return err
}

// DeleteInbound 删除入站。
func (s *Store) DeleteInbound(id int64) error {
	_, err := s.db.Exec("DELETE FROM inbounds WHERE id = ?", id)
	return err
}

// NextInboundTag 生成下一个入站 tag。
func (s *Store) NextInboundTag() (string, error) {
	var maxID int64
	if err := s.db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM inbounds").Scan(&maxID); err != nil {
		return "", err
	}
	return fmt.Sprintf("inbound-%d", maxID+1), nil
}

func scanInbound(scanner interface{ Scan(...any) error }) (*model.Inbound, error) {
	var in model.Inbound
	var tlsEnabled, sniffing, enabled int
	var stream, cert, key, sni sql.NullString
	err := scanner.Scan(
		&in.ID, &in.Tag, &in.Remark, &in.Protocol, &in.Port, &in.Listen, &in.Transport,
		&stream, &tlsEnabled, &cert, &key, &sni, &sniffing, &enabled,
		&in.CreatedAt, &in.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	in.StreamSettings = stream.String
	in.TLSCert = cert.String
	in.TLSKey = key.String
	in.TLSServerName = sni.String
	in.TLSEnabled = tlsEnabled == 1
	in.EnableSniffing = sniffing == 1
	in.Enabled = enabled == 1
	return &in, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
