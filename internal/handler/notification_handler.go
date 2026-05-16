package handler

import (
	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationHandler handles notification REST endpoints. Listing has moved
// to ListNotificationsHandler (cursor-paginated); this handler covers state
// mutation only (mark-as-read for now).
type NotificationHandler struct {
	service *application.NotificationService
	lister  NotificationLister
	logger  *zap.Logger
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(service *application.NotificationService, lister NotificationLister, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, lister: lister, logger: logger}
}

// RegisterRoutes registers notification API routes. The list endpoint is
// served by an inline ListNotificationsHandler that shares the same lister
// dependency so all /notifications routes live under a single auth-protected
// group.
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup, jwtManager *auth.JWTManager) {
	listHandler := NewListNotificationsHandler(h.lister, h.logger)

	notifications := rg.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware(jwtManager))
	{
		notifications.GET("", listHandler.List)
		notifications.PUT("/:id/read", h.MarkAsRead)
	}
}

// MarkAsRead marks a notification as read.
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification ID")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), notifID, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "notification marked as read"})
}
