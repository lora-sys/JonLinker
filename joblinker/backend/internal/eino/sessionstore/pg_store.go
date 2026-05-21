package sessionstore

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"joblinker/internal/model"
)

// PostgresStore implements Store using GORM + PostgreSQL.
type PostgresStore struct {
	db *gorm.DB
}

// NewPostgresStore creates a new PostgresStore.
func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(ctx context.Context, sessionID string, matchID uuid.UUID, version int) error {
	meta := model.SessionMeta{
		SessionID: sessionID,
		MatchID:   matchID,
		Version:   version,
		Status:    string(SessionStatusActive),
		State:     "idle",
	}
	return s.db.WithContext(ctx).Create(&meta).Error
}

func (s *PostgresStore) Load(ctx context.Context, sessionID string) (*Session, error) {
	var meta model.SessionMeta
	if err := s.db.WithContext(ctx).First(&meta, "session_id = ?", sessionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session %s: %w", sessionID, errNotFound)
		}
		return nil, fmt.Errorf("load session meta %s: %w", sessionID, err)
	}

	var dbMsgs []model.SessionMessage
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("seq ASC").
		Find(&dbMsgs).Error; err != nil {
		return nil, fmt.Errorf("load session messages %s: %w", sessionID, err)
	}

	msgs := make([]Message, len(dbMsgs))
	for i, m := range dbMsgs {
		msgs[i] = Message{
			Seq:     m.Seq,
			Role:    m.Role,
			Content: m.Content,
		}
	}

	return &Session{
		SessionID: meta.SessionID,
		MatchID:   meta.MatchID,
		Version:   meta.Version,
		Status:    SessionStatus(meta.Status),
		State:     meta.State,
		Messages:  msgs,
	}, nil
}

func (s *PostgresStore) AppendMessages(ctx context.Context, sessionID string, msgs []Message) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify session is active
		var meta model.SessionMeta
		if err := tx.First(&meta, "session_id = ?", sessionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("session %s: %w", sessionID, errNotFound)
			}
			return fmt.Errorf("check session %s: %w", sessionID, err)
		}
		if meta.Status != string(SessionStatusActive) {
			return fmt.Errorf("session %s is %s, cannot append messages", sessionID, meta.Status)
		}

		// Get the next seq number
		var maxSeq int
		tx.Model(&model.SessionMessage{}).
			Where("session_id = ?", sessionID).
			Select("COALESCE(MAX(seq), 0)").
			Scan(&maxSeq)

		for i, m := range msgs {
			msg := model.SessionMessage{
				SessionID: sessionID,
				Seq:       maxSeq + i + 1,
				Role:      m.Role,
				Content:   m.Content,
			}
			if err := tx.Create(&msg).Error; err != nil {
				return fmt.Errorf("insert message seq %d: %w", msg.Seq, err)
			}
		}
		return nil
	})
}

func (s *PostgresStore) ListByMatchID(ctx context.Context, matchID uuid.UUID) ([]Session, error) {
	var metas []model.SessionMeta
	if err := s.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		Order("version ASC").
		Find(&metas).Error; err != nil {
		return nil, fmt.Errorf("list sessions for match %s: %w", matchID, err)
	}

	sessions := make([]Session, len(metas))
	for i, meta := range metas {
		sessions[i] = Session{
			SessionID: meta.SessionID,
			MatchID:   meta.MatchID,
			Version:   meta.Version,
			Status:    SessionStatus(meta.Status),
			State:     meta.State,
		}
	}
	return sessions, nil
}

func (s *PostgresStore) Conclude(ctx context.Context, sessionID string, reason string) error {
	result := s.db.WithContext(ctx).
		Model(&model.SessionMeta{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":            string(SessionStatusConcluded),
			"conclusion_reason": reason,
		})
	if result.Error != nil {
		return fmt.Errorf("conclude session %s: %w", sessionID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session %s: %w", sessionID, errNotFound)
	}
	return nil
}

func (s *PostgresStore) LatestVersion(ctx context.Context, matchID uuid.UUID) (int, error) {
	var meta model.SessionMeta
	err := s.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		Order("version DESC").
		First(&meta).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("latest version for match %s: %w", matchID, err)
	}
	return meta.Version, nil
}

func (s *PostgresStore) Delete(ctx context.Context, sessionID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.SessionMessage{}).Error; err != nil {
			return fmt.Errorf("delete messages for %s: %w", sessionID, err)
		}
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.SessionMeta{}).Error; err != nil {
			return fmt.Errorf("delete meta for %s: %w", sessionID, err)
		}
		return nil
	})
}
