package store

import (
	"database/sql"
	"errors"

	"github.com/HYuCN666/volans/internal/model"
)

// CreateSubscription 创建订阅令牌。
func (s *Store) CreateSubscription(userID int64, token string) error {
	_, err := s.db.Exec(
		"INSERT INTO subscriptions (user_id, token) VALUES (?, ?)",
		userID, token,
	)
	return err
}

// GetSubscriptionByToken 按令牌查询订阅。
func (s *Store) GetSubscriptionByToken(token string) (*model.Subscription, error) {
	row := s.db.QueryRow(
		"SELECT id, user_id, token, created_at FROM subscriptions WHERE token = ?",
		token,
	)
	var sub model.Subscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.Token, &sub.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubscriptionByUserID 按用户查询订阅。
func (s *Store) GetSubscriptionByUserID(userID int64) (*model.Subscription, error) {
	row := s.db.QueryRow(
		"SELECT id, user_id, token, created_at FROM subscriptions WHERE user_id = ?",
		userID,
	)
	var sub model.Subscription
	err := row.Scan(&sub.ID, &sub.UserID, &sub.Token, &sub.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
