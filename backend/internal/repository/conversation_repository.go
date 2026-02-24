package repository

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ConversationRepository struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, conv *model.Conversation) error {
	conv.ID = uuid.New().String()
	conv.Status = "open"
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()

	query := `INSERT INTO conversations (id, tenant_id, customer_id, status, channel, created_at, updated_at)
			  VALUES (:id, :tenant_id, :customer_id, :status, :channel, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, conv)
	return err
}

func (r *ConversationRepository) GetByID(ctx context.Context, id, tenantID string) (*model.Conversation, error) {
	var conv model.Conversation
	query := `
		SELECT c.*, 
			   cu.name as customer_name, 
			   cu.external_id as customer_external_id,
			   COALESCE(u.name, '') as assigned_agent_name,
               COALESCE((SELECT message FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1), '') as last_message,
               EXISTS(SELECT 1 FROM conversation_tickets ct WHERE ct.conversation_id = c.id) as has_ticket,
			   c.selected_ticket_id as selected_ticket_id
		FROM conversations c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		LEFT JOIN users u ON c.assigned_agent_id = u.id
		WHERE c.id = ? AND c.tenant_id = ?`

	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &conv, query, id, tenantID)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) SetSelectedTicket(ctx context.Context, id, ticketID string) error {
	query := `UPDATE conversations SET selected_ticket_id = ?, updated_at = ? WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, ticketID, time.Now(), id)
	return err
}

func (r *ConversationRepository) GetByCustomerAndTenant(ctx context.Context, customerID, tenantID string) (*model.Conversation, error) {
	var conv model.Conversation
	query := `SELECT * FROM conversations WHERE customer_id = ? AND tenant_id = ? AND status != 'closed' ORDER BY created_at DESC LIMIT 1`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &conv, query, customerID, tenantID)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) List(ctx context.Context, tenantID string, filter model.ConversationFilter) ([]model.Conversation, int, error) {
	var conversations []model.Conversation
	var total int

	// Set defaults
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PerPage == 0 {
		filter.PerPage = 20
	}
	offset := (filter.Page - 1) * filter.PerPage

	// Build query
	baseQuery := `
		FROM conversations c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		LEFT JOIN users u ON c.assigned_agent_id = u.id
		WHERE c.tenant_id = ?`

	args := []interface{}{tenantID}

	if filter.Status != "" {
		baseQuery += ` AND c.status = ?`
		args = append(args, filter.Status)
	}

	if filter.AssignedAgentID != "" {
		baseQuery += ` AND c.assigned_agent_id = ?`
		args = append(args, filter.AssignedAgentID)
	}

	// Count total
	countQuery := `SELECT COUNT(*) ` + baseQuery
	countQuery = r.db.Rebind(countQuery)
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get data
	selectQuery := `
		SELECT c.*, 
			   cu.name as customer_name, 
			   cu.external_id as customer_external_id,
			   COALESCE(u.name, '') as assigned_agent_name,
               COALESCE((SELECT message FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1), '') as last_message,
               EXISTS(SELECT 1 FROM conversation_tickets ct WHERE ct.conversation_id = c.id) as has_ticket
		` + baseQuery + ` ORDER BY c.updated_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.PerPage, offset)
	selectQuery = r.db.Rebind(selectQuery)
	err = r.db.SelectContext(ctx, &conversations, selectQuery, args...)

	return conversations, total, err
}

func (r *ConversationRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE conversations SET status = ?, updated_at = ? WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *ConversationRepository) Assign(ctx context.Context, id, agentID string) error {
	query := `UPDATE conversations SET assigned_agent_id = ?, status = 'assigned', updated_at = ? WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, agentID, time.Now(), id)
	return err
}

func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, id string) error {
	query := `UPDATE conversations SET last_message_at = ?, updated_at = ? WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, time.Now(), time.Now(), id)
	return err
}

func (r *ConversationRepository) HasTicket(ctx context.Context, id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM conversation_tickets WHERE conversation_id = ?`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &count, query, id)
	return count > 0, err
}

func (r *ConversationRepository) Delete(ctx context.Context, id, tenantID string) error {
	query := `DELETE FROM conversations WHERE id = ? AND tenant_id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id, tenantID)
	return err
}
