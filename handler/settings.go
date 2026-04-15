package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingtw/bikeparts/config"
)

type settingsResponse struct {
	Port         string `json:"port"`
	EmailUser    string `json:"email_user"`
	EmailPass    string `json:"email_pass"` // always masked on GET
	EmailTo      string `json:"email_to"`
	SMTPPort     string `json:"smtp_port"`
	IsConfigured bool   `json:"is_configured"`
}

func (h *Handler) GetSettings(c *gin.Context) {
	s := h.cfg.Get()
	pass := ""
	if s.EmailPass != "" {
		pass = "****"
	}
	c.JSON(http.StatusOK, settingsResponse{
		Port:         s.Port,
		EmailUser:    s.EmailUser,
		EmailPass:    pass,
		EmailTo:      s.EmailTo,
		SMTPPort:     s.SMTPPort,
		IsConfigured: h.cfg.IsConfigured(),
	})
}

func (h *Handler) SaveSettings(c *gin.Context) {
	var body settingsResponse
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	current := h.cfg.Get()

	// Keep existing password if client sends the masked placeholder
	pass := body.EmailPass
	if pass == "****" || pass == "" {
		pass = current.EmailPass
	}

	port := body.Port
	if port == "" {
		port = "8080"
	}
	smtpPort := body.SMTPPort
	if smtpPort == "" {
		smtpPort = "587"
	}

	next := config.Settings{
		Port:      port,
		EmailUser: body.EmailUser,
		EmailPass: pass,
		EmailTo:   body.EmailTo,
		SMTPPort:  smtpPort,
	}
	if err := h.cfg.Set(next); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to save settings", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	restartRequired := current.Port != next.Port
	c.JSON(http.StatusOK, gin.H{"ok": true, "restart_required": restartRequired})
}

func (h *Handler) TestEmail(c *gin.Context) {
	if err := h.mailer.Send(
		"[測試] 零件庫存系統 Email 設定",
		"這是一封測試信，確認 Email 設定正確。",
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
