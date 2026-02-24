package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	// Bind query params for pagination and filtering
	var filter model.TicketFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid query parameters: " + err.Error()})
		return
	}

	// set defaults if not provided
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PerPage == 0 {
		filter.PerPage = 20
	}

	tickets, meta, err := h.ticketService.List(c.Request.Context(), tenantID, filter)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "does not exist") || strings.Contains(strings.ToLower(err.Error()), "no such table") {
			log.Printf("DB schema missing when listing tickets: %v", err)
			empty := []model.Ticket{}
			meta := &model.PaginationMeta{Page: filter.Page, PerPage: filter.PerPage, Total: 0, TotalPages: 0}
			c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: empty, Meta: meta, Message: "database schema not ready or no data"})
			return
		}

		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    tickets,
		Meta:    meta,
	})
}

func (h *TicketHandler) GetByID(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id := c.Param("id")

	ticket, err := h.ticketService.GetByID(c.Request.Context(), id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    ticket,
	})
}

func (h *TicketHandler) Escalate(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var req model.EscalateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	ticket, err := h.ticketService.Escalate(c.Request.Context(), conversationID, tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Data:    ticket,
		Message: "Ticket created successfully",
	})
}

func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req model.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	err := h.ticketService.UpdateStatus(c.Request.Context(), id, tenantID, userID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Ticket status updated successfully",
	})
}

func (h *TicketHandler) Create(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	var req struct {
		Title           string  `json:"title"`
		Description     string  `json:"description"`
		Priority        string  `json:"priority"`
		ConversationID  *string `json:"conversation_id"`
		Code            *string `json:"code"`
		AssignedAgentID *string `json:"assigned_agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	payload := model.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}
	if req.ConversationID != nil && *req.ConversationID != "" {
		payload.ConversationID = sql.NullString{String: *req.ConversationID, Valid: true}
	}
	if req.Code != nil && *req.Code != "" {
		// model.Ticket.Code is *string
		payload.Code = req.Code
	}
	if req.AssignedAgentID != nil && *req.AssignedAgentID != "" {
		payload.AssignedAgentID = sql.NullString{String: *req.AssignedAgentID, Valid: true}
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), tenantID, userID, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{Success: true, Data: ticket})
}

func (h *TicketHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req model.Ticket
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	// Ensure Code is nil or string; no sql.NullString here since model.Ticket.Code is *string
	ticket, err := h.ticketService.Update(c.Request.Context(), id, tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: ticket})
}

func (h *TicketHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	id := c.Param("id")

	err := h.ticketService.Delete(c.Request.Context(), id, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "Ticket deleted"})
}
