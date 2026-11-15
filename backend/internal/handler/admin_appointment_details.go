package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminAppointmentDetails: GET /admin/appointments/:id/details
// Запись с клиентом, полной услугой и действующей ссылкой отмены (если есть).
func (h *Handler) AdminAppointmentDetails(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id записи"})
		return
	}
	ctx := c.Request.Context()

	appt, err := h.apptRepo.GetByIDWithRelations(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "запись не найдена"})
		return
	}
	if err != nil {
		log.Printf("admin appointment details %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить запись"})
		return
	}

	if svc, err := h.serviceRepo.GetByID(ctx, appt.ServiceID); err == nil {
		appt.Service = svc
	} else {
		log.Printf("admin appointment details %s: load service: %v", id, err)
	}

	resp := gin.H{
		"appointment":       appt,
		"cancel_url":        "",
		"cancel_expires_at": nil,
	}
	if h.tokenRepo != nil {
		tok, err := h.tokenRepo.GetActiveForAppointment(ctx, id)
		if err != nil {
			log.Printf("admin appointment details %s: load cancel token: %v", id, err)
		} else if tok != nil {
			resp["cancel_url"] = fmt.Sprintf("%s/?cancel_token=%s",
				strings.TrimRight(h.baseURL, "/"), url.QueryEscape(tok.Token))
			resp["cancel_expires_at"] = tok.ExpiresAt
		}
	}

	c.JSON(http.StatusOK, resp)
}
