package handler

import (
	"net/http"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	convService *service.ConversationService
	// we access message repo via conversation service's msgRepo is unexported; instead add a small delete wrapper in conversation service
}

func NewMessageHandler(convService *service.ConversationService) *MessageHandler {
	return &MessageHandler{convService: convService}
}

func (h *MessageHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	id := c.Param("id")

	err := h.convService.DeleteMessage(c.Request.Context(), id, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "Message deleted"})
}
