package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
	service *service.ChannelService
}

func NewChannelHandler(s *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{service: s}
}

func (h *ChannelHandler) List(c *gin.Context) {
	chs, err := h.service.List(c.Request.Context())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "does not exist") || strings.Contains(strings.ToLower(err.Error()), "no such table") {
			log.Printf("DB schema missing when listing channels: %v", err)
			c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: []model.Channel{}, Message: "database schema not ready or no data"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: chs})
}

func (h *ChannelHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	ch, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: ch})
}

func (h *ChannelHandler) Create(c *gin.Context) {
	var req model.Channel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	// set tenant id from context or default
	if req.TenantID == "" {
		req.TenantID = "tenant_001"
	}
	if err := h.service.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model.APIResponse{Success: true, Data: req})
}

func (h *ChannelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.Channel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	req.ID = id
	if err := h.service.Update(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: req})
}

func (h *ChannelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true})
}
