package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"elena-backend/internal/middleware"
	"elena-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func uidPtr(id uuid.UUID) *uuid.UUID { return &id }

// logAction пишет действие администратора в журнал admin_actions.
// Вызывать после успешного изменения. Ошибка только логируется, ответ клиенту не меняется.
// В details не класть персональные данные клиентов (почта, телефон, ФИО).
func (h *Handler) logAction(c *gin.Context, action, entity string, entityID *uuid.UUID, details gin.H) {
	if h.actionRepo == nil {
		return
	}
	var adminID *uuid.UUID
	if id, ok := middleware.AdminIDFromContext(c); ok {
		adminID = &id
	}
	if details == nil {
		details = gin.H{}
	}
	raw, err := json.Marshal(details)
	if err != nil {
		log.Printf("admin action %s: marshal details: %v", action, err)
		raw = []byte("{}")
	}
	if err := h.actionRepo.Create(c.Request.Context(), adminID, action, entity, entityID, string(raw)); err != nil {
		log.Printf("admin action %s %s: запись в журнал не удалась: %v", action, entity, err)
	}
}

// AdminListActions: GET /admin/actions?entity=&entity_id=&limit=&offset=
func (h *Handler) AdminListActions(c *gin.Context) {
	limit, offset, ok := parsePagination(c, 50)
	if !ok {
		return
	}
	opts := repository.ListAdminActionsOpts{Entity: c.Query("entity"), Limit: limit, Offset: offset}
	if raw := c.Query("entity_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный entity_id"})
			return
		}
		opts.EntityID = &id
	}
	items, total, err := h.actionRepo.List(c.Request.Context(), opts)
	if err != nil {
		log.Printf("admin list actions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить историю"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}
