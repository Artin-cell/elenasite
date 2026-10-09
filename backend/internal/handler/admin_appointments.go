package handler

import (
	"errors"
	"log"
	"net/http"

	"elena-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var adminCancelConflicts = []error{
	service.ErrAlreadyCancelled,
	service.ErrCannotCancel,
	service.ErrNothingToRefund,
	service.ErrManualPayment,
	service.ErrStatusChanged,
}

// AdminCancelAppointment: POST /admin/appointments/:id/cancel  {"refund": true|false}
func (h *Handler) AdminCancelAppointment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Refund bool `json:"refund"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}

	res, err := h.apptSvc.AdminCancel(c.Request.Context(), id, req.Refund)
	if err != nil {
		log.Printf("admin cancel %s (refund=%v): %v", id, req.Refund, err)
		if errors.Is(err, service.ErrRefundFailed) {
			c.JSON(http.StatusBadGateway, gin.H{"error": service.ErrRefundFailed.Error()})
			return
		}
		for _, known := range adminCancelConflicts {
			if errors.Is(err, known) {
				c.JSON(http.StatusConflict, gin.H{"error": known.Error()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось отменить запись"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"refunded":      res.Refunded,
		"refund_status": res.RefundStatus,
	})
}

// AdminMarkPaid: POST /admin/appointments/:id/mark-paid (оплата наличными или переводом)
func (h *Handler) AdminMarkPaid(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	changed, err := h.apptRepo.MarkPaidManually(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin mark paid %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось отметить оплату"})
		return
	}
	if !changed {
		c.JSON(http.StatusConflict, gin.H{"error": "отметить оплаченной можно только подтверждённую запись без онлайн-оплаты"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
