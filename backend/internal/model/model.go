package model

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Name      string    `json:"name" db:"name"`
	Role      string    `json:"role" db:"role"` // admin, agent
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Customer represents a customer from external channels
type Customer struct {
	ID         string    `json:"id" db:"id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	ExternalID string    `json:"external_id" db:"external_id"`
	Name       string    `json:"name" db:"name"`
	Channel    string    `json:"channel" db:"channel"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Conversation represents a conversation with a customer
type Conversation struct {
	ID              string         `json:"id" db:"id"`
	TenantID        string         `json:"tenant_id" db:"tenant_id"`
	CustomerID      string         `json:"customer_id" db:"customer_id"`
	Status          string         `json:"status" db:"status"` // open, assigned, closed
	AssignedAgentID sql.NullString `json:"assigned_agent_id" db:"assigned_agent_id"`
	Channel         string         `json:"channel" db:"channel"`
	LastMessageAt   sql.NullTime   `json:"last_message_at" db:"last_message_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`

	// Joined fields
	CustomerName       string `json:"customer_name,omitempty" db:"customer_name"`
	CustomerExternalID string `json:"customer_external_id,omitempty" db:"customer_external_id"`
	AssignedAgentName  string `json:"assigned_agent_name,omitempty" db:"assigned_agent_name"`
	LastMessage        string `json:"last_message,omitempty" db:"last_message"`
	HasTicket          bool   `json:"has_ticket" db:"has_ticket"`
	SelectedTicketID   sql.NullString `json:"ticket_id,omitempty" db:"selected_ticket_id"`
}

// Message represents a message in a conversation
type Message struct {
	ID             string    `json:"id" db:"id"`
	ConversationID string    `json:"conversation_id" db:"conversation_id"`
	SenderType     string    `json:"sender_type" db:"sender_type"` // customer, agent
	SenderID       string    `json:"sender_id" db:"sender_id"`
	SenderName     string    `json:"sender_name" db:"sender_name"`
	Message        string    `json:"message" db:"message"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Ticket represents an escalated ticket
type Ticket struct {
	ID              string         `json:"id" db:"id"`
	TenantID        string         `json:"tenant_id" db:"tenant_id"`
	ConversationID  sql.NullString `json:"conversation_id" db:"conversation_id"`
	Code            *string        `json:"code,omitempty" db:"code"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	Status          string         `json:"status" db:"status"`     // open, in_progress, resolved, closed
	Priority        string         `json:"priority" db:"priority"` // low, medium, high, urgent
	AssignedAgentID sql.NullString `json:"assigned_agent_id" db:"assigned_agent_id"`
	CreatedByID     string         `json:"created_by_id" db:"created_by_id"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`

	// Joined fields
	AssignedAgentName string `json:"assigned_agent_name,omitempty" db:"assigned_agent_name"`
	CreatedByName     string `json:"created_by_name,omitempty" db:"created_by_name"`
}

// Event represents an activity log event
type Event struct {
	ID         string    `json:"id" db:"id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	EventType  string    `json:"event_type" db:"event_type"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   string    `json:"entity_id" db:"entity_id"`
	Data       string    `json:"data" db:"data"`
	UserID     string    `json:"user_id" db:"user_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Channel represents an inbound/outbound channel configuration
type Channel struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Request/Response DTOs
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin agent"`
}

type WebhookRequest struct {
	TenantID           string `json:"tenant_id" binding:"required"`
	CustomerExternalID string `json:"customer_external_id" binding:"required"`
	Channel            string `json:"channel"`
	Message            string `json:"message" binding:"required"`
}

type SendMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type EscalateRequest struct {
	Title       string `json:"title" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
	Priority    string `json:"priority" binding:"required_without=TicketCode"`
	TicketCode  string `json:"ticket_code" binding:"omitempty"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress resolved closed"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin agent"`
}

type UpdateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role" binding:"omitempty,oneof=admin agent"`
}

type PaginationParams struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=100"`
}

type ConversationFilter struct {
	Status          string `form:"status"`
	AssignedAgentID string `form:"assigned_agent_id"`
	PaginationParams
}

type TicketFilter struct {
	Status   string `form:"status"`
	Priority string `form:"priority"`
	PaginationParams
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
