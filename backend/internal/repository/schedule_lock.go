package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"elena-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// scheduleAdvisoryKey — общий ключ для pg_advisory_xact_lock.
// Все операции, изменяющие расписание (запись, перенос, закрытие слотов),
// берут этот замок внутри транзакции; замок снимается при commit/rollback.
const scheduleAdvisoryKey = int64(0x656c656e61736368) // "elenasch"

// ErrScheduleLockTimeout возвращается, если замок не удалось взять за 10 с.
var ErrScheduleLockTimeout = errors.New("расписание сейчас обновляется, попробуйте снова")

// WithScheduleLock открывает транзакцию READ COMMITTED, берёт pg_advisory_xact_lock
// и выполняет fn. Коммитит транзакцию при nil-ошибке, откатывает иначе.
func WithScheduleLock(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// lock_timeout чтобы запрос не висел вечно.
	if _, err := tx.ExecContext(ctx, `set local lock_timeout = '10s'`); err != nil {
		return fmt.Errorf("set lock_timeout: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock($1)`, scheduleAdvisoryKey); err != nil {
		return ErrScheduleLockTimeout
	}

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// IsBlockedTx — IsBlocked внутри уже открытой транзакции.
func (r *BlockedSlotRepo) IsBlockedTx(ctx context.Context, tx *sqlx.Tx, startsAt, endsAt time.Time, loc *time.Location) (bool, error) {
	startLocal := startsAt.In(loc)
	endLocal := endsAt.In(loc)
	var rows []struct {
		BlockedDate time.Time `db:"blocked_date"`
		SlotTime    *string   `db:"slot_time"`
	}
	if err := tx.SelectContext(ctx, &rows, `
		select blocked_date, slot_time
		from blocked_slots
		where blocked_date between $1::date and $2::date
	`, startLocal, endLocal); err != nil {
		return false, fmt.Errorf("check blocked slot: %w", err)
	}
	for _, row := range rows {
		date := row.BlockedDate.In(loc)
		if row.SlotTime == nil {
			return true, nil
		}
		slot, err := time.ParseInLocation("15:04", *row.SlotTime, loc)
		if err != nil {
			return false, fmt.Errorf("invalid stored blocked slot time: %w", err)
		}
		blockedAt := time.Date(date.Year(), date.Month(), date.Day(), slot.Hour(), slot.Minute(), 0, 0, loc)
		if !blockedAt.Before(startLocal) && blockedAt.Before(endLocal) {
			return true, nil
		}
	}
	return false, nil
}

// CreateTx — Create внутри уже открытой транзакции.
func (r *AppointmentRepo) CreateTx(ctx context.Context, tx *sqlx.Tx, a *models.Appointment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	_, err := tx.NamedExecContext(ctx, `
		insert into appointments
			(id, client_id, service_id, format, starts_at, ends_at, status,
			 notes, payment_mode, payment_status, payment_id, amount_kopeks, created_at, updated_at)
		values
			(:id, :client_id, :service_id, :format, :starts_at, :ends_at, :status,
			 :notes, :payment_mode, :payment_status, :payment_id, :amount_kopeks, :created_at, :updated_at)
	`, a)
	if err != nil {
		return fmt.Errorf("create appointment: %w", err)
	}
	return nil
}

// RescheduleTx — Reschedule внутри уже открытой транзакции.
func (r *AppointmentRepo) RescheduleTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, startsAt, endsAt time.Time) error {
	_, err := tx.ExecContext(ctx, `
		update appointments
		set starts_at = $1, ends_at = $2, updated_at = now()
		where id = $3
	`, startsAt, endsAt, id)
	return err
}

// ListPendingConfirmedTx — список pending и confirmed записей в диапазоне дат внутри tx.
// Используется для проверки конфликтов перед закрытием слотов.
func (r *AppointmentRepo) ListPendingConfirmedTx(ctx context.Context, tx *sqlx.Tx, from, to time.Time) ([]models.Appointment, error) {
	type row struct {
		models.Appointment
		CFirstName string `db:"c_first_name"`
		CLastName  string `db:"c_last_name"`
		CPatronym  string `db:"c_patronym"`
		CEmail     string `db:"c_email"`
		CPhone     string `db:"c_phone"`
		STitle     string `db:"s_title"`
	}
	var rows []row
	err := tx.SelectContext(ctx, &rows, `
		select a.*,
		       c.first_name as c_first_name, c.last_name as c_last_name, c.patronym as c_patronym,
		       c.email as c_email, c.phone as c_phone,
		       s.title as s_title
		from appointments a
		join clients c on c.id = a.client_id
		join services s on s.id = a.service_id
		where a.starts_at >= $1
		  and a.starts_at < $2
		  and a.status in ('pending','confirmed')
		order by a.starts_at asc
	`, from, to)
	if err != nil {
		return nil, err
	}
	list := make([]models.Appointment, len(rows))
	for i, rr := range rows {
		appt := rr.Appointment
		appt.Client = &models.Client{
			ID:        appt.ClientID,
			FirstName: rr.CFirstName,
			LastName:  rr.CLastName,
			Patronym:  rr.CPatronym,
			Email:     rr.CEmail,
			Phone:     rr.CPhone,
		}
		appt.Service = &models.Service{ID: appt.ServiceID, Title: rr.STitle}
		list[i] = appt
	}
	return list, nil
}

// BulkCreateTx — BulkCreate внутри уже открытой транзакции.
func (r *BlockedSlotRepo) BulkCreateTx(ctx context.Context, tx *sqlx.Tx, items []BlockInput, reason string) (int, error) {
	created := 0
	for _, it := range items {
		res, err := tx.ExecContext(ctx, `
			insert into blocked_slots (id, blocked_date, slot_time, reason)
			values (uuid_generate_v4(), $1, $2, $3)
			on conflict do nothing
		`, it.Date, it.SlotTime, reason)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		created += int(n)
	}
	return created, nil
}
