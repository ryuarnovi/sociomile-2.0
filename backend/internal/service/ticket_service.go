package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"backend/internal/model"

	"backend/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TicketService struct {
	ticketRepo *repository.TicketRepository
	convRepo   *repository.ConversationRepository
	eventRepo  *repository.EventRepository
	rabbitCh   *amqp.Channel
}

func NewTicketService(
	ticketRepo *repository.TicketRepository,
	convRepo *repository.ConversationRepository,
	eventRepo *repository.EventRepository,
	rabbitCh *amqp.Channel,
) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		convRepo:   convRepo,
		eventRepo:  eventRepo,
		rabbitCh:   rabbitCh,
	}
}

func (s *TicketService) Escalate(ctx context.Context, conversationID, tenantID, userID string, req model.EscalateRequest) (*model.Ticket, error) {
	// Verify conversation exists
	conv, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	// If TicketCode provided, attach existing ticket to conversation
	if req.TicketCode != "" {
		ticket, err := s.ticketRepo.GetByCode(ctx, req.TicketCode, tenantID)
		if err != nil {
			return nil, errors.New("ticket not found")
		}

		// add mapping (many-to-many)
		if err := s.ticketRepo.AddConversationMapping(ctx, ticket.ID, conversationID); err != nil {
			return nil, err
		}

		// Log event and publish
		s.logEvent(ctx, tenantID, "ticket.linked", "ticket", ticket.ID, userID, map[string]string{"conversation_id": conversationID})
		s.logEvent(ctx, tenantID, "conversation.escalated", "conversation", conversationID, userID, map[string]string{"ticket_id": ticket.ID})

		s.publishEvent(ctx, "conversation.events", "conversation.escalated", map[string]interface{}{
			"conversation_id": conversationID,
			"ticket_id":       ticket.ID,
		})

		return ticket, nil
	}

	// Create a new ticket (not bound to conversation directly) and map it
	ticket := &model.Ticket{
		TenantID:        tenantID,
		ConversationID:  sql.NullString{Valid: false},
		Title:           req.Title,
		Description:     req.Description,
		Priority:        req.Priority,
		AssignedAgentID: conv.AssignedAgentID,
		CreatedByID:     userID,
	}

	err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	// create mapping
	if err := s.ticketRepo.AddConversationMapping(ctx, ticket.ID, conversationID); err != nil {
		return nil, err
	}

	// Log event and publish to queue
	s.logEvent(ctx, tenantID, "ticket.created", "ticket", ticket.ID, userID, ticket)
	s.logEvent(ctx, tenantID, "conversation.escalated", "conversation", conversationID, userID, map[string]string{"ticket_id": ticket.ID})

	s.publishEvent(ctx, "ticket.events", "ticket.created", map[string]interface{}{
		"ticket_id":       ticket.ID,
		"conversation_id": conversationID,
		"tenant_id":       tenantID,
	})

	s.publishEvent(ctx, "conversation.events", "conversation.escalated", map[string]interface{}{
		"conversation_id": conversationID,
		"ticket_id":       ticket.ID,
	})

	return ticket, nil
}

func (s *TicketService) Create(ctx context.Context, tenantID string, userID string, payload model.Ticket) (*model.Ticket, error) {
	// If conversation id is provided, verify it exists; otherwise allow ticket without conversation
	if payload.ConversationID.Valid && payload.ConversationID.String != "" {
		_, err := s.convRepo.GetByID(ctx, payload.ConversationID.String, tenantID)
		if err != nil {
			return nil, errors.New("conversation not found")
		}
	}

	payload.TenantID = tenantID
	payload.CreatedByID = userID

	err := s.ticketRepo.Create(ctx, &payload)
	if err != nil {
		return nil, err
	}

	s.logEvent(ctx, tenantID, "ticket.created", "ticket", payload.ID, userID, payload)
	s.publishEvent(ctx, "ticket.events", "ticket.created", map[string]interface{}{"ticket_id": payload.ID, "conversation_id": (func() string { if payload.ConversationID.Valid { return payload.ConversationID.String } ; return ""})(), "tenant_id": tenantID})

	return &payload, nil
}

func (s *TicketService) Update(ctx context.Context, id, tenantID, userID string, req model.Ticket) (*model.Ticket, error) {
	t, err := s.ticketRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// apply updates
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Priority != "" {
		t.Priority = req.Priority
	}
	if req.AssignedAgentID.Valid {
		t.AssignedAgentID = req.AssignedAgentID
	}
	// Code is a *string now; update if provided and non-empty
	if req.Code != nil && *req.Code != "" {
		t.Code = req.Code
	}

	err = s.ticketRepo.Update(ctx, t)
	if err != nil {
		return nil, err
	}

	s.logEvent(ctx, tenantID, "ticket.updated", "ticket", t.ID, userID, t)
	return t, nil
}

func (s *TicketService) Delete(ctx context.Context, id, tenantID, userID string) error {
	_, err := s.ticketRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return errors.New("ticket not found")
	}
	err = s.ticketRepo.Delete(ctx, id, tenantID)
	if err != nil {
		return err
	}
	s.logEvent(ctx, tenantID, "ticket.deleted", "ticket", id, userID, nil)
	return nil
}

func (s *TicketService) List(ctx context.Context, tenantID string, filter model.TicketFilter) ([]model.Ticket, *model.PaginationMeta, error) {
	tickets, total, err := s.ticketRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, nil, err
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	meta := &model.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}

	return tickets, meta, nil
}

func (s *TicketService) GetByID(ctx context.Context, id, tenantID string) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}
	return ticket, nil
}

func (s *TicketService) UpdateStatus(ctx context.Context, id, tenantID, userID, status string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return errors.New("ticket not found")
	}

	// Validate status transition
	validStatuses := map[string]bool{"open": true, "in_progress": true, "resolved": true, "closed": true}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	err = s.ticketRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	// Log event
	s.logEvent(ctx, tenantID, "ticket.status_updated", "ticket", id, userID, map[string]string{
		"old_status": ticket.Status,
		"new_status": status,
	})

	return nil
}

func (s *TicketService) logEvent(ctx context.Context, tenantID, eventType, entityType, entityID, userID string, data interface{}) {
	err := s.eventRepo.LogEvent(ctx, tenantID, eventType, entityType, entityID, userID, data)
	if err != nil {
		log.Printf("Failed to log event: %v", err)
	}
}

func (s *TicketService) publishEvent(ctx context.Context, exchange, routingKey string, data interface{}) {
	if s.rabbitCh == nil {
		return
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return
	}

	err = s.rabbitCh.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	}
}
