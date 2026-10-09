package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"elena-backend/internal/models"
	"elena-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	scheduleMaxDays = 366
	scheduleMaxRows = 2000
)

var ruMonthsGenitive = []string{"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря"}

func scheduleLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Yekaterinburg")
	if err != nil {
		return time.FixedZone("YEKT", 5*3600)
	}
	return loc
}

func isScheduleSlot(t string) bool {
	for _, s := range allSlotTimes {
		if s == t {
			return true
		}
	}
	return false
}

// uniqueSorted убирает дубли часов (занято записью и закрыто одновременно) и сортирует.
func uniqueSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func scheduleClientName(c *models.Client) string {
	if c == nil {
		return "клиент"
	}
	name := strings.TrimSpace(c.FirstName + " " + c.LastName)
	if name == "" {
		return "клиент"
	}
	return name
}

type scheduleBlockDTO struct {
	ID       uuid.UUID `json:"id"`
	Date     string    `json:"date"`
	SlotTime *string   `json:"slot_time"`
	Reason   string    `json:"reason"`
}

type scheduleAppointmentDTO struct {
	ID            uuid.UUID `json:"id"`
	Date          string    `json:"date"`
	Slots         []string  `json:"slots"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	ClientName    string    `json:"client_name"`
	ServiceTitle  string    `json:"service_title"`
}

// AdminSchedule: GET /admin/schedule?month=YYYY-MM — закрытия и записи месяца для календаря.
func (h *Handler) AdminSchedule(c *gin.Context) {
	loc := scheduleLoc()
	var from time.Time
	if monthParam := c.Query("month"); monthParam != "" {
		parsed, err := time.ParseInLocation("2006-01", monthParam, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month, expected YYYY-MM"})
			return
		}
		from = parsed
	} else {
		now := time.Now().In(loc)
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	to := from.AddDate(0, 1, 0)

	blocks, err := h.blockedSlotRepo.ListRange(c.Request.Context(), from, to)
	if err != nil {
		log.Printf("admin schedule blocks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить расписание"})
		return
	}
	appts, err := h.apptRepo.List(c.Request.Context(), repository.ListAppointmentsOpts{From: from, To: to})
	if err != nil {
		log.Printf("admin schedule appointments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить расписание"})
		return
	}

	blockList := make([]scheduleBlockDTO, 0, len(blocks))
	for _, b := range blocks {
		blockList = append(blockList, scheduleBlockDTO{
			ID:       b.ID,
			Date:     b.BlockedDate.Format("2006-01-02"),
			SlotTime: b.SlotTime,
			Reason:   b.Reason,
		})
	}

	apptList := make([]scheduleAppointmentDTO, 0, len(appts))
	for _, a := range appts {
		if a.Status == models.StatusCancelled {
			continue
		}
		start := a.StartsAt.In(loc)
		end := a.EndsAt.In(loc)
		dateKey := start.Format("2006-01-02")
		slots := []string{}
		for hour := start.Truncate(time.Hour); hour.Before(end); hour = hour.Add(time.Hour) {
			if hour.Format("2006-01-02") == dateKey && isScheduleSlot(hour.Format("15:04")) {
				slots = append(slots, hour.Format("15:04"))
			}
		}
		title := ""
		if a.Service != nil {
			title = a.Service.Title
		}
		apptList = append(apptList, scheduleAppointmentDTO{
			ID:            a.ID,
			Date:          dateKey,
			Slots:         slots,
			Status:        string(a.Status),
			PaymentStatus: a.PaymentStatus,
			ClientName:    scheduleClientName(a.Client),
			ServiceTitle:  title,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"month":        from.Format("2006-01"),
		"slots":        allSlotTimes,
		"blocks":       blockList,
		"appointments": apptList,
	})
}

type blocksBatchRequest struct {
	Date      string   `json:"date"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	Weekdays  []int    `json:"weekdays"` // 0 = воскресенье … 6 = суббота; пусто = все дни
	SlotTime  *string  `json:"slot_time"`
	SlotTimes []string `json:"slot_times"` // пусто = закрыть день целиком
	Reason    string   `json:"reason"`
	Force     bool     `json:"force"`
}

func scheduleDays(date, from, to string, weekdays []int, loc *time.Location) ([]time.Time, error) {
	if date != "" {
		from, to = date, date
	}
	if from == "" {
		return nil, errors.New("укажите дату")
	}
	if to == "" {
		to = from
	}
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, errors.New("неверная дата, ожидается формат ГГГГ-ММ-ДД")
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, errors.New("неверная дата, ожидается формат ГГГГ-ММ-ДД")
	}
	if end.Before(start) {
		return nil, errors.New("дата «по» раньше даты «с»")
	}
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if start.Before(today) {
		return nil, errors.New("нельзя закрывать прошедшие даты")
	}
	if end.Sub(start) > time.Duration(scheduleMaxDays)*24*time.Hour {
		return nil, fmt.Errorf("период слишком длинный, максимум %d дней", scheduleMaxDays)
	}

	allowed := map[time.Weekday]bool{}
	for _, w := range weekdays {
		if w < 0 || w > 6 {
			return nil, errors.New("неверный день недели")
		}
		allowed[time.Weekday(w)] = true
	}

	var days []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if len(allowed) > 0 && !allowed[d.Weekday()] {
			continue
		}
		days = append(days, d)
	}
	return days, nil
}

// scheduleSlotList возвращает nil, если нужно закрыть день целиком.
func scheduleSlotList(single *string, list []string) ([]string, error) {
	var raw []string
	if single != nil && *single != "" {
		raw = append(raw, *single)
	}
	raw = append(raw, list...)
	if len(raw) == 0 {
		return nil, nil
	}
	for _, t := range raw {
		if !isScheduleSlot(t) {
			return nil, fmt.Errorf("недопустимое время %q: приём идёт с 10:00 до 20:00, начало каждый час", t)
		}
	}
	return uniqueSorted(raw), nil
}

type scheduleConflict struct {
	When   string `json:"when"`
	Client string `json:"client"`
}

// scheduleConflicts находит действующие записи, которые попадут под новые закрытия.
func (h *Handler) scheduleConflicts(ctx context.Context, days []time.Time, slots []string, loc *time.Location) ([]scheduleConflict, error) {
	if len(days) == 0 {
		return nil, nil
	}
	first, last := days[0], days[len(days)-1]
	from := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, loc)
	to := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	appts, err := h.apptRepo.List(ctx, repository.ListAppointmentsOpts{From: from, To: to})
	if err != nil {
		return nil, err
	}

	target := make(map[string]struct{}, len(days))
	for _, d := range days {
		target[d.Format("2006-01-02")] = struct{}{}
	}

	var out []scheduleConflict
	for _, a := range appts {
		if a.Status != models.StatusPending && a.Status != models.StatusConfirmed {
			continue
		}
		start := a.StartsAt.In(loc)
		end := a.EndsAt.In(loc)
		if _, ok := target[start.Format("2006-01-02")]; !ok {
			continue
		}
		hit := slots == nil
		for _, t := range slots {
			hhmm, err := time.ParseInLocation("15:04", t, loc)
			if err != nil {
				continue
			}
			slotAt := time.Date(start.Year(), start.Month(), start.Day(), hhmm.Hour(), hhmm.Minute(), 0, 0, loc)
			if !slotAt.Before(start) && slotAt.Before(end) {
				hit = true
				break
			}
		}
		if hit {
			out = append(out, scheduleConflict{
				When:   fmt.Sprintf("%d %s, %s", start.Day(), ruMonthsGenitive[int(start.Month())-1], start.Format("15:04")),
				Client: scheduleClientName(a.Client),
			})
		}
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out, nil
}

// AdminCreateBlocks: POST /admin/blocks — один день, один час или период с фильтром по дням недели и часам.
func (h *Handler) AdminCreateBlocks(c *gin.Context) {
	var req blocksBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}
	loc := scheduleLoc()

	days, err := scheduleDays(req.Date, req.From, req.To, req.Weekdays, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slots, err := scheduleSlotList(req.SlotTime, req.SlotTimes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if len([]rune(reason)) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "причина слишком длинная, не больше 200 символов"})
		return
	}

	var items []repository.BlockInput
	for _, day := range days {
		if slots == nil {
			items = append(items, repository.BlockInput{Date: day})
			continue
		}
		for _, t := range slots {
			slot := t
			items = append(items, repository.BlockInput{Date: day, SlotTime: &slot})
		}
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "в выбранном периоде нет подходящих дней"})
		return
	}
	if len(items) > scheduleMaxRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "слишком много закрытий за один раз, выберите период короче"})
		return
	}

	if !req.Force {
		conflicts, err := h.scheduleConflicts(c.Request.Context(), days, slots, loc)
		if err != nil {
			log.Printf("admin create blocks conflicts: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось проверить существующие записи"})
			return
		}
		if len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":      "В выбранное время уже есть записи:",
				"need_force": true,
				"conflicts":  conflicts,
			})
			return
		}
	}

	created, err := h.blockedSlotRepo.BulkCreate(c.Request.Context(), items, reason)
	if err != nil {
		log.Printf("admin create blocks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить расписание"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"created": created, "skipped": len(items) - created})
}

// AdminDeleteBlocksRange: DELETE /admin/blocks?from=YYYY-MM-DD&to=YYYY-MM-DD — открыть день или период.
func (h *Handler) AdminDeleteBlocksRange(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if to == "" {
		to = from
	}
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите дату в формате ГГГГ-ММ-ДД"})
		return
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите дату в формате ГГГГ-ММ-ДД"})
		return
	}
	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "дата «по» раньше даты «с»"})
		return
	}
	if end.Sub(start) > time.Duration(scheduleMaxDays)*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "период слишком длинный"})
		return
	}

	n, err := h.blockedSlotRepo.DeleteRange(c.Request.Context(), start, end)
	if err != nil {
		log.Printf("admin delete blocks range: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть период"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": n})
}
