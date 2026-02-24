package repository

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()

	query := `INSERT INTO messages (id, conversation_id, sender_type, sender_id, sender_name, message, created_at)
			  VALUES (:id, :conversation_id, :sender_type, :sender_id, :sender_name, :message, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, msg)
	return err
}

func (r *MessageRepository) GetByConversationID(ctx context.Context, conversationID string) ([]model.Message, error) {
	var messages []model.Message
	query := `SELECT * FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`
	query = r.db.Rebind(query)
	err := r.db.SelectContext(ctx, &messages, query, conversationID)
	return messages, err
}

func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
