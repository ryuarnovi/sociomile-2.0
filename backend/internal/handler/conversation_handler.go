package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	convService *service.ConversationService
}

func NewConversationHandler(convService *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{convService: convService}
}

func (h *ConversationHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	// Manually parse query params to avoid binding errors when SPA sends empty strings
	var filter model.ConversationFilter
	// simple string filters
	filter.Status = c.Query("status")
	filter.AssignedAgentID = c.Query("assigned_agent_id")

	// pagination with defaults
	pageStr := c.Query("page")
	perPageStr := c.Query("per_page")
	page := 1
	perPage := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}
	filter.Page = page
	filter.PerPage = perPage

	conversations, meta, err := h.convService.List(c.Request.Context(), tenantID, filter)
	if err != nil {
		// If DB schema isn't ready (missing tables) return an empty result for easier debugging in dev
		if strings.Contains(strings.ToLower(err.Error()), "does not exist") || strings.Contains(strings.ToLower(err.Error()), "no such table") {
			log.Printf("DB schema missing when listing conversations: %v", err)
			empty := []model.Conversation{}
			meta := &model.PaginationMeta{Page: filter.Page, PerPage: filter.PerPage, Total: 0, TotalPages: 0}
			c.JSON(http.StatusOK, model.APIResponse{
				Success: true,
				Data:    empty,
				Meta:    meta,
				Message: "database schema not ready or no data",
			})
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
		Data:    conversations,
		Meta:    meta,
	})
}

func (h *ConversationHandler) GetByID(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id := c.Param("id")

	conv, messages, err := h.convService.GetByID(c.Request.Context(), id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"conversation": conv,
			"messages":     messages,
		},
	})
}

func (h *ConversationHandler) SendMessage(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	userName := c.GetString("email") // Could be improved to get actual name
	conversationID := c.Param("id")

	var req model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	msg, err := h.convService.SendMessage(c.Request.Context(), conversationID, tenantID, userID, userName, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Data:    msg,
	})
}

func (h *ConversationHandler) Assign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	err := h.convService.Assign(c.Request.Context(), conversationID, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Conversation assigned successfully",
	})
}

func (h *ConversationHandler) Close(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	err := h.convService.Close(c.Request.Context(), conversationID, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Conversation closed successfully",
	})
}

func (h *ConversationHandler) Create(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var req model.Conversation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	conv, err := h.convService.Create(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{Success: true, Data: conv})
}

func (h *ConversationHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id := c.Param("id")

	var payload struct {
		Status        string `json:"status"`
		AssignedAgent string `json:"assigned_agent_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	// allow status update or assign
	if payload.AssignedAgent != "" {
		err := h.convService.Assign(c.Request.Context(), id, tenantID, payload.AssignedAgent)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
			return
		}
	}
	if payload.Status != "" {
		userID := c.GetString("user_id")
		if payload.Status == "closed" {
			err := h.convService.Close(c.Request.Context(), id, tenantID, userID)
			if err != nil {
				c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
				return
			}
		} else {
			err := h.convService.UpdateStatus(c.Request.Context(), id, payload.Status)
			if err != nil {
				c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "Conversation updated"})
}

func (h *ConversationHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id := c.Param("id")

	err := h.convService.Delete(c.Request.Context(), id, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "Conversation deleted"})
}

func (h *ConversationHandler) ListTickets(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	id := c.Param("id")

	tickets, err := h.convService.ListTicketsForConversation(c.Request.Context(), id, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: tickets})
}

func (h *ConversationHandler) SetSelectedTicket(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	id := c.Param("id")

	var body struct{
		TicketID string `json:"ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	err := h.convService.SetSelectedTicket(c.Request.Context(), id, tenantID, userID, body.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "Selected ticket updated"})
}
