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

// PreferenceHandler handles notification preference REST endpoints.
type PreferenceHandler struct {
	service *application.NotificationService
	logger  *zap.Logger
}

// NewPreferenceHandler creates a new preference handler.
func NewPreferenceHandler(service *application.NotificationService, logger *zap.Logger) *PreferenceHandler {
	return &PreferenceHandler{service: service, logger: logger}
}

// RegisterRoutes registers preference API routes under /notifications/preferences.
func (h *PreferenceHandler) RegisterRoutes(rg *gin.RouterGroup, jwtManager *auth.JWTManager) {
	prefs := rg.Group("/notifications/preferences")
	prefs.Use(middleware.AuthMiddleware(jwtManager))
	{
		prefs.GET("", h.GetPreferences)
		prefs.PUT("", h.UpdatePreferences)
	}

	// FCM token registration.
	fcm := rg.Group("/notifications/fcm-token")
	fcm.Use(middleware.AuthMiddleware(jwtManager))
	{
		fcm.POST("", h.RegisterFCMToken)
	}
	devices := rg.Group("/notifications/device-tokens")
	devices.Use(middleware.AuthMiddleware(jwtManager))
	{
		devices.POST("", h.RegisterDeviceToken)
	}
}

// GetPreferences returns the notification preferences for the authenticated user.
func (h *PreferenceHandler) GetPreferences(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}

	prefs, err := h.service.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, prefs)
}

type updatePreferencesRequest struct {
	EnablePush  bool `json:"enable_push"`
	EnableSMS   bool `json:"enable_sms"`
	EnableEmail bool `json:"enable_email"`
}

// UpdatePreferences updates channel preferences.
func (h *PreferenceHandler) UpdatePreferences(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}

	var req updatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.service.UpdatePreferences(c.Request.Context(), userID, req.EnablePush, req.EnableSMS, req.EnableEmail); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "preferences updated"})
}

type registerFCMTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type registerDeviceTokenRequest struct {
	Token     string `json:"token" binding:"required"`
	Platform  string `json:"platform"`
	ScopeType string `json:"scope_type"`
	ScopeID   string `json:"scope_id"`
}

// RegisterFCMToken stores the Firebase device token.
func (h *PreferenceHandler) RegisterFCMToken(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}

	var req registerFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}

	if err := h.service.RegisterFCMToken(c.Request.Context(), userID, req.Token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "FCM token registered"})
}

// RegisterDeviceToken stores a device token, optionally scoped to a shop.
func (h *PreferenceHandler) RegisterDeviceToken(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user context")
		return
	}
	var req registerDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}
	var scopeID *uuid.UUID
	if req.ScopeID != "" {
		parsed, err := uuid.Parse(req.ScopeID)
		if err != nil {
			response.BadRequest(c, "invalid scope_id")
			return
		}
		scopeID = &parsed
	}
	if err := h.service.RegisterScopedFCMToken(c.Request.Context(), userID, req.Token, req.ScopeType, scopeID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "device token registered"})
}
