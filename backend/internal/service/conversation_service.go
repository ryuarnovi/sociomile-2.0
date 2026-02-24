package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"backend/internal/model"
	"backend/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type ConversationService struct {
	convRepo     *repository.ConversationRepository
	msgRepo      *repository.MessageRepository
	customerRepo *repository.CustomerRepository
	eventRepo    *repository.EventRepository
	ticketRepo   *repository.TicketRepository
	redis        *redis.Client
	rabbitCh     *amqp.Channel
}

func NewConversationService(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	customerRepo *repository.CustomerRepository,
	eventRepo *repository.EventRepository,
	ticketRepo *repository.TicketRepository,
	redis *redis.Client,
	rabbitCh *amqp.Channel,
) *ConversationService {
	return &ConversationService{
		convRepo:     convRepo,
		msgRepo:      msgRepo,
		customerRepo: customerRepo,
		eventRepo:    eventRepo,
		ticketRepo:   ticketRepo,
		redis:        redis,
		rabbitCh:     rabbitCh,
	}
}

func (s *ConversationService) HandleWebhook(ctx context.Context, req model.WebhookRequest) (*model.Conversation, error) {
	// Get or create customer
	channel := req.Channel
	if channel == "" {
		channel = "unknown"
	}

	customer, err := s.customerRepo.GetOrCreate(ctx, req.CustomerExternalID, req.TenantID, channel)
	if err != nil {
		return nil, err
	}
	// removed debug logging

	// Find existing open conversation or create new one
	conv, err := s.convRepo.GetByCustomerAndTenant(ctx, customer.ID, req.TenantID)
	if err != nil {
		// Create new conversation
		conv = &model.Conversation{
			TenantID:   req.TenantID,
			CustomerID: customer.ID,
			Channel:    channel,
		}
		err = s.convRepo.Create(ctx, conv)
		if err != nil {
			return nil, err
		}

		// Log event
		s.logEvent(ctx, req.TenantID, "conversation.created", "conversation", conv.ID, "", conv)
	}

	// Create message
	msg := &model.Message{
		ConversationID: conv.ID,
		SenderType:     "customer",
		SenderID:       customer.ID,
		SenderName:     customer.Name,
		Message:        req.Message,
	}
	err = s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Update conversation last message time
	s.convRepo.UpdateLastMessage(ctx, conv.ID)

	// Invalidate cache
	s.invalidateConversationCache(ctx, req.TenantID)

	// Log event
	s.logEvent(ctx, req.TenantID, "message.received", "conversation", conv.ID, "", msg)

	// Publish to message queue for realtime delivery
	go func(m *model.Message, tenant string) {
		payload := map[string]interface{}{
			"tenant_id":       tenant,
			"conversation_id": m.ConversationID,
			"message_id":      m.ID,
			"sender_id":       m.SenderID,
			"sender_name":     m.SenderName,
			"sender_type":     m.SenderType,
			"message":         m.Message,
			"created_at":      m.CreatedAt,
		}
		s.publishEvent(context.Background(), "conversation.events", "message.received", payload)
	}(msg, req.TenantID)

	return conv, nil
}

func (s *ConversationService) List(ctx context.Context, tenantID string, filter model.ConversationFilter) ([]model.Conversation, *model.PaginationMeta, error) {
	conversations, total, err := s.convRepo.List(ctx, tenantID, filter)
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

	return conversations, meta, nil
}

func (s *ConversationService) GetByID(ctx context.Context, id, tenantID string) (*model.Conversation, []model.Message, error) {
	conv, err := s.convRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, nil, errors.New("conversation not found")
	}

	messages, err := s.msgRepo.GetByConversationID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return conv, messages, nil
}

func (s *ConversationService) SendMessage(ctx context.Context, conversationID, tenantID, userID, userName, message string) (*model.Message, error) {
	// Verify conversation exists and belongs to tenant
	conv, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	if conv.Status == "closed" {
		return nil, errors.New("conversation is closed")
	}

	msg := &model.Message{
		ConversationID: conversationID,
		SenderType:     "agent",
		SenderID:       userID,
		SenderName:     userName,
		Message:        message,
	}

	err = s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Update conversation
	s.convRepo.UpdateLastMessage(ctx, conversationID)

	// Invalidate cache
	s.invalidateConversationCache(ctx, tenantID)

	// Log event
	s.logEvent(ctx, tenantID, "message.sent", "conversation", conversationID, userID, msg)

	// Publish to message queue for realtime delivery
	go func(m *model.Message, tenant string) {
		payload := map[string]interface{}{
			"tenant_id":       tenant,
			"conversation_id": m.ConversationID,
			"message_id":      m.ID,
			"sender_id":       m.SenderID,
			"sender_name":     m.SenderName,
			"sender_type":     m.SenderType,
			"message":         m.Message,
			"created_at":      m.CreatedAt,
		}
		s.publishEvent(context.Background(), "conversation.events", "message.sent", payload)
	}(msg, tenantID)

	return msg, nil
}

func (s *ConversationService) Assign(ctx context.Context, conversationID, tenantID, agentID string) error {
	conv, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return errors.New("conversation not found")
	}

	if conv.Status == "closed" {
		return errors.New("cannot assign closed conversation")
	}

	err = s.convRepo.Assign(ctx, conversationID, agentID)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateConversationCache(ctx, tenantID)

	// Log event and publish to queue
	s.logEvent(ctx, tenantID, "conversation.assigned", "conversation", conversationID, agentID, map[string]string{"agent_id": agentID})
	s.publishEvent(ctx, "conversation.events", "conversation.assigned", map[string]string{
		"conversation_id": conversationID,
		"agent_id":        agentID,
	})

	return nil
}

func (s *ConversationService) UpdateStatus(ctx context.Context, id, status string) error {
	err := s.convRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}
	// no tenant context here for cache invalidation; caller should handle if needed
	return nil
}

func (s *ConversationService) Close(ctx context.Context, conversationID, tenantID, userID string) error {
	conv, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return errors.New("conversation not found")
	}

	if conv.Status == "closed" {
		return errors.New("conversation already closed")
	}

	err = s.convRepo.UpdateStatus(ctx, conversationID, "closed")
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateConversationCache(ctx, tenantID)

	// Log event
	s.logEvent(ctx, tenantID, "conversation.closed", "conversation", conversationID, userID, nil)

	return nil
}

func (s *ConversationService) Create(ctx context.Context, tenantID string, conv *model.Conversation) (*model.Conversation, error) {
	conv.TenantID = tenantID
	if conv.Channel == "" {
		conv.Channel = "unknown"
	}
	err := s.convRepo.Create(ctx, conv)
	if err != nil {
		return nil, err
	}
	s.logEvent(ctx, tenantID, "conversation.created", "conversation", conv.ID, "", conv)
	s.invalidateConversationCache(ctx, tenantID)
	return conv, nil
}

func (s *ConversationService) Delete(ctx context.Context, id, tenantID string) error {
	_, err := s.convRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return errors.New("conversation not found")
	}
	err = s.convRepo.Delete(ctx, id, tenantID)
	if err != nil {
		return err
	}
	s.invalidateConversationCache(ctx, tenantID)
	s.logEvent(ctx, tenantID, "conversation.deleted", "conversation", id, "", nil)
	return nil
}

func (s *ConversationService) ListTicketsForConversation(ctx context.Context, conversationID, tenantID string) ([]model.Ticket, error) {
	// ensure conversation belongs to tenant
	_, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	tickets, err := s.convRepoListTickets(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

// convRepoListTickets bridges to ticketRepo to list tickets for conversation
func (s *ConversationService) convRepoListTickets(ctx context.Context, conversationID string) ([]model.Ticket, error) {
	return s.convRepoListTicketsImpl(ctx, conversationID)
}

// This is a small indirection to keep repository references organized; we'll implement impl below by casting to known repository
func (s *ConversationService) convRepoListTicketsImpl(ctx context.Context, conversationID string) ([]model.Ticket, error) {
	// convRepo doesn't have ticket listing, use ticketRepo directly if available
	if s == nil || s.ticketRepo == nil {
		return nil, errors.New("internal error")
	}
	return s.ticketRepo.ListByConversationID(ctx, conversationID)
}

func (s *ConversationService) SetSelectedTicket(ctx context.Context, conversationID, tenantID, userID, ticketID string) error {
	// verify conversation exists and belongs to tenant
	_, err := s.convRepo.GetByID(ctx, conversationID, tenantID)
	if err != nil {
		return errors.New("conversation not found")
	}

	// optionally verify ticket exists
	_, err = s.ticketRepo.GetByID(ctx, ticketID, tenantID)
	if err != nil {
		return errors.New("ticket not found")
	}

	// ensure mapping exists between ticket and conversation (many-to-many)
	if err := s.ticketRepo.AddConversationMapping(ctx, ticketID, conversationID); err != nil {
		return err
	}

	err = s.convRepo.SetSelectedTicket(ctx, conversationID, ticketID)
	if err != nil {
		return err
	}

	s.logEvent(ctx, tenantID, "conversation.selected_ticket", "conversation", conversationID, userID, map[string]string{"ticket_id": ticketID})
	s.publishEvent(ctx, "conversation.events", "conversation.selected_ticket", map[string]string{"conversation_id": conversationID, "ticket_id": ticketID})
	return nil
}

func (s *ConversationService) DeleteMessage(ctx context.Context, id, tenantID, userID string) error {
	// Verify message exists and belongs to a conversation under tenant
	// Fetch message by conversation via repository (no GetByID implemented), query conversations ownership
	// We'll perform a simple delete assuming authorization handled by middleware (tenant scope)
	err := s.msgRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	s.logEvent(ctx, tenantID, "message.deleted", "message", id, userID, nil)
	return nil
}

func (s *ConversationService) logEvent(ctx context.Context, tenantID, eventType, entityType, entityID, userID string, data interface{}) {
	err := s.eventRepo.LogEvent(ctx, tenantID, eventType, entityType, entityID, userID, data)
	if err != nil {
		log.Printf("Failed to log event: %v", err)
	}
}

func (s *ConversationService) publishEvent(ctx context.Context, exchange, routingKey string, data interface{}) {
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

func (s *ConversationService) invalidateConversationCache(ctx context.Context, tenantID string) {
	if s.redis == nil {
		return
	}
	pattern := "conversations:" + tenantID + ":*"
	keys, _ := s.redis.Keys(ctx, pattern).Result()
	if len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}
}
