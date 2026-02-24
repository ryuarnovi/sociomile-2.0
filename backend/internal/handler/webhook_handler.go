package handler

import (
	"net/http"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	convService *service.ConversationService
}

func NewWebhookHandler(convService *service.ConversationService) *WebhookHandler {
	return &WebhookHandler{convService: convService}
}

// HandleWebhook simulates incoming messages from external channels (WhatsApp, Instagram, etc.)
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var req model.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Set default channel if not provided
	if req.Channel == "" {
		req.Channel = "unknown"
	}

	conv, err := h.convService.HandleWebhook(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"conversation_id": conv.ID,
			"status":          conv.Status,
		},
		Message: "Message received and processed",
	})
}
