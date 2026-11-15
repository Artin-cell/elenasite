package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"elena-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	newsTitleMaxLen    = 200
	newsBodyMaxLen     = 20000
	reviewAuthorMaxLen = 100
	reviewBodyMaxLen   = 5000
)

// ===== Новости =====

// newsInput — тело создания и частичного обновления новости. nil / пусто = поле не передано.
// published_at: не передано — не меняется; null или "" — снять с публикации (черновик);
// дата RFC3339 — опубликовать с этой даты.
type newsInput struct {
	Title       *string         `json:"title"`
	Body        *string         `json:"body"`
	PublishedAt json.RawMessage `json:"published_at"`
}

// parsePublishedAt возвращает: передано ли поле, значение и текст ошибки.
func parsePublishedAt(raw json.RawMessage) (bool, *time.Time, string) {
	if len(raw) == 0 {
		return false, nil, ""
	}
	if string(raw) == "null" {
		return true, nil, ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return true, nil, "дата публикации указана некорректно"
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil, ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return true, nil, "дата публикации указана некорректно (нужен формат 2026-11-15T10:00:00+05:00)"
	}
	return true, &t, ""
}

func (in *newsInput) validate() string {
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return "заголовок новости не может быть пустым"
		}
		if utf8.RuneCountInString(t) > newsTitleMaxLen {
			return "заголовок слишком длинный (не больше 200 символов)"
		}
	}
	if in.Body != nil && utf8.RuneCountInString(*in.Body) > newsBodyMaxLen {
		return "текст новости слишком длинный (не больше 20000 символов)"
	}
	if _, _, msg := parsePublishedAt(in.PublishedAt); msg != "" {
		return msg
	}
	return ""
}

func (in *newsInput) apply(n *models.News) {
	if in.Title != nil {
		n.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		n.Body = strings.TrimSpace(*in.Body)
	}
	if set, t, _ := parsePublishedAt(in.PublishedAt); set {
		n.PublishedAt = t
	}
}

// AdminListNews: GET /admin/news — все новости, включая черновики.
func (h *Handler) AdminListNews(c *gin.Context) {
	list, err := h.newsRepo.ListAll(c.Request.Context())
	if err != nil {
		log.Printf("admin list news: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить новости"})
		return
	}
	if list == nil {
		list = []models.News{}
	}
	c.JSON(http.StatusOK, list)
}

// AdminCreateNews: POST /admin/news
func (h *Handler) AdminCreateNews(c *gin.Context) {
	var in newsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if in.Title == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите заголовок новости"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	n := &models.News{}
	in.apply(n)
	if err := h.newsRepo.Create(c.Request.Context(), n); err != nil {
		log.Printf("admin create news: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать новость"})
		return
	}
	h.logAction(c, "news.create", "news", uidPtr(n.ID), gin.H{"title": n.Title, "published": n.PublishedAt != nil})
	c.JSON(http.StatusCreated, n)
}

// AdminUpdateNews: PATCH и PUT /admin/news/:id — частичное обновление.
func (h *Handler) AdminUpdateNews(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id новости"})
		return
	}
	var in newsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	ctx := c.Request.Context()
	current, err := h.newsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "новость не найдена"})
			return
		}
		log.Printf("admin update news %s: get: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить новость"})
		return
	}
	in.apply(current)
	if err := h.newsRepo.Update(ctx, current); err != nil {
		log.Printf("admin update news %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить новость"})
		return
	}
	h.logAction(c, "news.update", "news", uidPtr(id), gin.H{"title": current.Title, "published": current.PublishedAt != nil})
	if fresh, err := h.newsRepo.GetByID(ctx, id); err == nil {
		current = fresh
	}
	c.JSON(http.StatusOK, current)
}

// AdminDeleteNews: DELETE /admin/news/:id
func (h *Handler) AdminDeleteNews(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id новости"})
		return
	}
	deleted, err := h.newsRepo.DeleteByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin delete news %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить новость"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "новость не найдена"})
		return
	}
	h.logAction(c, "news.delete", "news", uidPtr(id), nil)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ===== Отзывы =====

// reviewInput — тело создания и частичного обновления отзыва. nil = поле не передано.
type reviewInput struct {
	AuthorName *string `json:"author_name"`
	Body       *string `json:"body"`
	IsVisible  *bool   `json:"is_visible"`
}

func (in *reviewInput) validate() string {
	if in.AuthorName != nil {
		a := strings.TrimSpace(*in.AuthorName)
		if a == "" {
			return "имя автора отзыва не может быть пустым"
		}
		if utf8.RuneCountInString(a) > reviewAuthorMaxLen {
			return "имя автора слишком длинное (не больше 100 символов)"
		}
	}
	if in.Body != nil {
		b := strings.TrimSpace(*in.Body)
		if b == "" {
			return "текст отзыва не может быть пустым"
		}
		if utf8.RuneCountInString(b) > reviewBodyMaxLen {
			return "текст отзыва слишком длинный (не больше 5000 символов)"
		}
	}
	return ""
}

func (in *reviewInput) apply(r *models.Review) {
	if in.AuthorName != nil {
		r.AuthorName = strings.TrimSpace(*in.AuthorName)
	}
	if in.Body != nil {
		r.Body = strings.TrimSpace(*in.Body)
	}
	if in.IsVisible != nil {
		r.IsVisible = *in.IsVisible
	}
}

// AdminListReviews: GET /admin/reviews — все отзывы, включая скрытые.
func (h *Handler) AdminListReviews(c *gin.Context) {
	list, err := h.reviewRepo.ListAll(c.Request.Context())
	if err != nil {
		log.Printf("admin list reviews: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить отзывы"})
		return
	}
	if list == nil {
		list = []models.Review{}
	}
	c.JSON(http.StatusOK, list)
}

// AdminCreateReview: POST /admin/reviews. По умолчанию отзыв скрыт.
func (h *Handler) AdminCreateReview(c *gin.Context) {
	var in reviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if in.AuthorName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите имя автора отзыва"})
		return
	}
	if in.Body == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите текст отзыва"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	r := &models.Review{}
	in.apply(r)
	if err := h.reviewRepo.Create(c.Request.Context(), r); err != nil {
		log.Printf("admin create review: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать отзыв"})
		return
	}
	h.logAction(c, "review.create", "review", uidPtr(r.ID), gin.H{"is_visible": r.IsVisible})
	c.JSON(http.StatusCreated, r)
}

// AdminUpdateReview: PATCH и PUT /admin/reviews/:id — частичное обновление.
func (h *Handler) AdminUpdateReview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id отзыва"})
		return
	}
	var in reviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	ctx := c.Request.Context()
	current, err := h.reviewRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "отзыв не найден"})
			return
		}
		log.Printf("admin update review %s: get: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить отзыв"})
		return
	}
	in.apply(current)
	if err := h.reviewRepo.Update(ctx, current); err != nil {
		log.Printf("admin update review %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить отзыв"})
		return
	}
	h.logAction(c, "review.update", "review", uidPtr(id), gin.H{"changes": in})
	if fresh, err := h.reviewRepo.GetByID(ctx, id); err == nil {
		current = fresh
	}
	c.JSON(http.StatusOK, current)
}

// AdminDeleteReview: DELETE /admin/reviews/:id
func (h *Handler) AdminDeleteReview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id отзыва"})
		return
	}
	deleted, err := h.reviewRepo.DeleteByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("admin delete review %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить отзыв"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "отзыв не найден"})
		return
	}
	h.logAction(c, "review.delete", "review", uidPtr(id), nil)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
