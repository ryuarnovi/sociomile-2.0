package repository

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TicketRepository struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	ticket.ID = uuid.New().String()
	ticket.Status = "open"
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	query := `INSERT INTO tickets (id, tenant_id, conversation_id, code, title, description, status, priority, assigned_agent_id, created_by_id, created_at, updated_at)
			  VALUES (:id, :tenant_id, :conversation_id, :code, :title, :description, :status, :priority, :assigned_agent_id, :created_by_id, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, ticket)
	return err
}

func (r *TicketRepository) GetByID(ctx context.Context, id, tenantID string) (*model.Ticket, error) {
	var ticket model.Ticket
	query := `
		SELECT t.*,
			   COALESCE(u1.name, '') as assigned_agent_name,
			   COALESCE(u2.name, '') as created_by_name
		FROM tickets t
		LEFT JOIN users u1 ON t.assigned_agent_id = u1.id
		LEFT JOIN users u2 ON t.created_by_id = u2.id
		WHERE t.id = ? AND t.tenant_id = ?`

	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &ticket, query, id, tenantID)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) GetByCode(ctx context.Context, code, tenantID string) (*model.Ticket, error) {
	var ticket model.Ticket
	query := `
		SELECT t.*,
			   COALESCE(u1.name, '') as assigned_agent_name,
			   COALESCE(u2.name, '') as created_by_name
		FROM tickets t
		LEFT JOIN users u1 ON t.assigned_agent_id = u1.id
		LEFT JOIN users u2 ON t.created_by_id = u2.id
		WHERE t.code = ? AND t.tenant_id = ? LIMIT 1`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &ticket, query, code, tenantID)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) GetByConversationID(ctx context.Context, conversationID string) (*model.Ticket, error) {
	// Deprecated: single-result; return first ticket linked to conversation via join table
	var ticket model.Ticket
	query := `SELECT t.* FROM tickets t JOIN conversation_tickets ct ON ct.ticket_id = t.id WHERE ct.conversation_id = ? LIMIT 1`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &ticket, query, conversationID)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) ListByConversationID(ctx context.Context, conversationID string) ([]model.Ticket, error) {
	var tickets []model.Ticket
	query := `SELECT t.* FROM tickets t JOIN conversation_tickets ct ON ct.ticket_id = t.id WHERE ct.conversation_id = ? ORDER BY ct.created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.SelectContext(ctx, &tickets, query, conversationID)
	return tickets, err
}

func (r *TicketRepository) List(ctx context.Context, tenantID string, filter model.TicketFilter) ([]model.Ticket, int, error) {
	var tickets []model.Ticket
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
		FROM tickets t
		LEFT JOIN users u1 ON t.assigned_agent_id = u1.id
		LEFT JOIN users u2 ON t.created_by_id = u2.id
		WHERE t.tenant_id = ?`

	args := []interface{}{tenantID}

	if filter.Status != "" {
		baseQuery += ` AND t.status = ?`
		args = append(args, filter.Status)
	}

	if filter.Priority != "" {
		baseQuery += ` AND t.priority = ?`
		args = append(args, filter.Priority)
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
		SELECT t.*,
			   COALESCE(u1.name, '') as assigned_agent_name,
			   COALESCE(u2.name, '') as created_by_name
		` + baseQuery + ` ORDER BY t.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.PerPage, offset)
	selectQuery = r.db.Rebind(selectQuery)
	err = r.db.SelectContext(ctx, &tickets, selectQuery, args...)

	return tickets, total, err
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	ticket.UpdatedAt = time.Now()
	query := `UPDATE tickets SET title = :title, description = :description, priority = :priority, assigned_agent_id = :assigned_agent_id, code = :code, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, ticket)
	return err
}

func (r *TicketRepository) Delete(ctx context.Context, id, tenantID string) error {
	query := `DELETE FROM tickets WHERE id = ? AND tenant_id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id, tenantID)
	return err
}

// UpdateConversationID sets the conversation_id for a ticket (used when attaching ticket to a conversation)
func (r *TicketRepository) UpdateConversationID(ctx context.Context, ticketID, conversationID string) error {
	query := `UPDATE tickets SET conversation_id = :conversation_id, updated_at = :updated_at WHERE id = :id`
	params := map[string]interface{}{
		"conversation_id": conversationID,
		"updated_at":      time.Now(),
		"id":              ticketID,
	}
	_, err := r.db.NamedExecContext(ctx, query, params)
	return err
}

// AddConversationMapping creates a link between a ticket and a conversation
func (r *TicketRepository) AddConversationMapping(ctx context.Context, ticketID, conversationID string) error {
	query := `INSERT INTO conversation_tickets (conversation_id, ticket_id, created_at) VALUES (:conversation_id, :ticket_id, :created_at) ON CONFLICT DO NOTHING`
	params := map[string]interface{}{
		"conversation_id": conversationID,
		"ticket_id":       ticketID,
		"created_at":      time.Now(),
	}
	_, err := r.db.NamedExecContext(ctx, query, params)
	return err
}
