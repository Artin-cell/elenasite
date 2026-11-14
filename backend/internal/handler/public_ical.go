package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"elena-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const officeAddress = "г. Екатеринбург, ул. Уральских рабочих 67"

func icsEscape(s string) string {
	r := strings.NewReplacer(`\\`, `\\\\`, ";", `\\;`, ",", `\\,`, "\n", `\\n`, "\r", "")
	return r.Replace(s)
}

// AppointmentICal: GET /api/v1/appointments/:id/ical — событие для календаря клиента.
// В файле нет персональных данных: только время, «Психолог», адрес и название услуги.
func (h *Handler) AppointmentICal(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusNotFound, "запись не найдена")
		return
	}
	appt, err := h.apptRepo.GetByIDWithRelations(c.Request.Context(), id)
	if err != nil || appt.Status == models.StatusCancelled {
		c.String(http.StatusNotFound, "запись не найдена или отменена")
		return
	}

	end := appt.EndsAt
	if !end.After(appt.StartsAt) {
		end = appt.StartsAt.Add(time.Hour)
	}
	location := officeAddress
	if appt.Format == models.AppointmentOnline {
		location = "Онлайн"
	}
	description := ""
	if appt.Service != nil {
		description = appt.Service.Title
	}
	utc := func(t time.Time) string { return t.UTC().Format("20060102T150405Z") }

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//ekosta.ru//RU",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:" + appt.ID.String() + "@ekosta.ru",
		"DTSTAMP:" + utc(time.Now()),
		"DTSTART:" + utc(appt.StartsAt),
		"DTEND:" + utc(end),
		"SUMMARY:Психолог",
		"LOCATION:" + icsEscape(location),
		"DESCRIPTION:" + icsEscape(description),
		"END:VEVENT",
		"END:VCALENDAR",
	}
	c.Header("Content-Disposition", `inline; filename="priem.ics"`)
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/calendar; charset=utf-8", []byte(strings.Join(lines, "\r\n")+"\r\n"))
	_ = fmt.Sprint
}
