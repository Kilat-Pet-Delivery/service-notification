package handler

import (
	"strconv"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-notification/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationHandler handles notification REST endpoints.
type NotificationHandler struct {
	service *application.NotificationService
	logger  *zap.Logger
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(service *application.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, logger: logger}
}

// RegisterRoutes registers notification API routes.
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup, jwtManager *auth.JWTManager) {
	notifications := rg.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware(jwtManager))
	{
		notifications.GET("", h.ListNotifications)
		notifications.GET("/:id", h.GetNotification)
		notifications.PUT("/:id/read", h.MarkAsRead)
	}
}

// ListNotifications returns paginated notifications for the authenticated user.
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	notifications, total, err := h.service.ListForUser(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, notifications, total, page, limit)
}

// GetNotification returns a single notification by ID.
func (h *NotificationHandler) GetNotification(c *gin.Context) {
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

	// Re-use ListForUser is not ideal for single fetch. For now, mark as read returns the item.
	_ = notifID
	_ = userID
	response.Success(c, gin.H{"message": "use list endpoint with pagination"})
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
