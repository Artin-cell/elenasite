package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"elena-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	svcTitleMaxLen    = 200
	svcDescMaxLen     = 5000
	svcDurationMinMin = 15
	svcDurationMinMax = 480
	svcPriceMaxKopeks = 100_000_000 // 1 000 000 ₽, защита от опечатки
)

// AdminListServices: GET /admin/services — все услуги, включая скрытые и демо.
func (h *Handler) AdminListServices(c *gin.Context) {
	list, err := h.serviceRepo.ListAll(c.Request.Context())
	if err != nil {
		log.Printf("admin list services: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить список услуг"})
		return
	}
	if list == nil {
		list = []models.Service{}
	}
	c.JSON(http.StatusOK, list)
}

// serviceInput — общее тело для создания и частичного обновления.
// nil означает «поле не передано».
type serviceInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Format      *string `json:"format"`
	PriceKopeks *int64  `json:"price_kopeks"`
	DurationMin *int    `json:"duration_min"`
	IsDemo      *bool   `json:"is_demo"`
	IsActive    *bool   `json:"is_active"`
}

func (in *serviceInput) validate() string {
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return "название услуги не может быть пустым"
		}
		if utf8.RuneCountInString(t) > svcTitleMaxLen {
			return "название услуги слишком длинное (не больше 200 символов)"
		}
	}
	if in.Description != nil && utf8.RuneCountInString(*in.Description) > svcDescMaxLen {
		return "описание слишком длинное (не больше 5000 символов)"
	}
	if in.Format != nil {
		switch models.ServiceFormat(*in.Format) {
		case models.FormatOnline, models.FormatOffline, models.FormatBoth:
		default:
			return "формат услуги: онлайн, очно или оба варианта"
		}
	}
	if in.PriceKopeks != nil {
		if *in.PriceKopeks < 0 {
			return "цена не может быть отрицательной"
		}
		if *in.PriceKopeks > svcPriceMaxKopeks {
			return "цена слишком большая, проверьте сумму"
		}
	}
	if in.DurationMin != nil {
		if *in.DurationMin < svcDurationMinMin || *in.DurationMin > svcDurationMinMax {
			return "длительность должна быть от 15 минут до 8 часов"
		}
	}
	return ""
}

// apply переносит переданные поля в услугу.
func (in *serviceInput) apply(s *models.Service) {
	if in.Title != nil {
		s.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		s.Description = strings.TrimSpace(*in.Description)
	}
	if in.Format != nil {
		s.Format = models.ServiceFormat(*in.Format)
	}
	if in.PriceKopeks != nil {
		s.PriceKopeks = *in.PriceKopeks
	}
	if in.DurationMin != nil {
		s.DurationMin = *in.DurationMin
	}
	if in.IsDemo != nil {
		s.IsDemo = *in.IsDemo
	}
	if in.IsActive != nil {
		s.IsActive = *in.IsActive
	}
}

// AdminCreateService: POST /admin/services
func (h *Handler) AdminCreateService(c *gin.Context) {
	var in serviceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if in.Title == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите название услуги"})
		return
	}
	if in.Format == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите формат услуги"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Значения по умолчанию совпадают с дефолтами таблицы services.
	s := &models.Service{DurationMin: 60, IsActive: true}
	in.apply(s)

	if err := h.serviceRepo.Create(c.Request.Context(), s); err != nil {
		log.Printf("admin create service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать услугу"})
		return
	}
	h.logAction(c, "service.create", "service", uidPtr(s.ID), gin.H{"title": s.Title, "price_kopeks": s.PriceKopeks})
	c.JSON(http.StatusCreated, s)
}

// AdminUpdateService: PATCH и PUT /admin/services/:id — частичное обновление,
// не переданные поля не меняются.
func (h *Handler) AdminUpdateService(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in serviceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	current, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "услуга не найдена"})
			return
		}
		log.Printf("admin update service %s: get: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить услугу"})
		return
	}

	in.apply(current)

	if err := h.serviceRepo.Update(c.Request.Context(), current); err != nil {
		log.Printf("admin update service %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить услугу"})
		return
	}
	h.logAction(c, "service.update", "service", uidPtr(id), gin.H{"changes": in})
	c.JSON(http.StatusOK, current)
}

// AdminDeleteService: DELETE /admin/services/:id
// Услугу с записями удалить нельзя (FK on delete restrict): её нужно скрыть.
func (h *Handler) AdminDeleteService(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	has, err := h.serviceRepo.HasAppointments(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin delete service %s: check appointments: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить услугу"})
		return
	}
	if has {
		c.JSON(http.StatusConflict, gin.H{"error": "на эту услугу уже есть записи, удалить её нельзя; скройте услугу, чтобы она пропала с сайта"})
		return
	}

	deleted, err := h.serviceRepo.DeleteIfUnused(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin delete service %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить услугу"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "услуга не найдена"})
		return
	}
	h.logAction(c, "service.delete", "service", uidPtr(id), nil)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
