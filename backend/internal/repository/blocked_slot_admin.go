package repository

import (
	"context"
	"fmt"
	"time"

	"elena-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type BlockInput struct {
	Date     time.Time
	SlotTime *string // nil = закрыть весь день
}

// BulkCreate добавляет закрытия одной транзакцией; уже существующие пропускаются.
func (r *BlockedSlotRepo) BulkCreate(ctx context.Context, items []BlockInput, reason string) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

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
		n, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		created += int(n)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return created, nil
}

// DeleteRange снимает все закрытия с даты from по дату to включительно.
func (r *BlockedSlotRepo) DeleteRange(ctx context.Context, from, to time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`delete from blocked_slots where blocked_date between $1::date and $2::date`, from, to)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ErrScheduleConflict возвращается из BulkCreateLocked, если без force=true нашлись записи.
type ErrScheduleConflict struct {
	Conflicts []models.Appointment
}

func (e *ErrScheduleConflict) Error() string {
	return fmt.Sprintf("конфликт с %d записями", len(e.Conflicts))
}

// BulkCreateLocked: проверка конфликтов и вставка блоков внутри одной транзакции с замком.
// При force=false и наличии конфликтов возвращает *ErrScheduleConflict (без вставки).
// При force=true вставляет блоки невзирая на конфликты.
func (r *BlockedSlotRepo) BulkCreateLocked(
	ctx context.Context, db *sqlx.DB,
	apptRepo *AppointmentRepo,
	items []BlockInput, reason string, force bool,
	loc *time.Location,
) (created int, err error) {
	lockErr := WithScheduleLock(ctx, db, func(tx *sqlx.Tx) error {
		if !force {
			conflicts, err := conflictsForItems(ctx, tx, apptRepo, items, loc)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				return &ErrScheduleConflict{Conflicts: conflicts}
			}
		}
		n, err := r.BulkCreateTx(ctx, tx, items, reason)
		if err != nil {
			return err
		}
		created = n
		return nil
	})
	return created, lockErr
}

// conflictsForItems находит pending/confirmed записи, пересекающиеся с закрываемыми слотами.
func conflictsForItems(ctx context.Context, tx *sqlx.Tx, apptRepo *AppointmentRepo, items []BlockInput, loc *time.Location) ([]models.Appointment, error) {
	if len(items) == 0 {
		return nil, nil
	}
	first, last := items[0].Date, items[len(items)-1].Date
	from := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, loc)
	to := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	appts, err := apptRepo.ListPendingConfirmedTx(ctx, tx, from, to)
	if err != nil {
		return nil, err
	}
	// построить индекс закрытий: дата -> часы (nil = весь день)
	type slotSet struct {
		fullDay bool
		slots   []time.Time // время закрытия (часы и минуты)
	}
	index := map[string]*slotSet{}
	for _, it := range items {
		key := it.Date.In(loc).Format("2006-01-02")
		ss := index[key]
		if ss == nil {
			ss = &slotSet{}
			index[key] = ss
		}
		if it.SlotTime == nil {
			ss.fullDay = true
		} else {
			t, err := time.ParseInLocation("15:04", *it.SlotTime, loc)
			if err == nil {
				ss.slots = append(ss.slots, t)
			}
		}
	}
	var out []models.Appointment
	for _, a := range appts {
		start := a.StartsAt.In(loc)
		key := start.Format("2006-01-02")
		ss := index[key]
		if ss == nil {
			continue
		}
		if ss.fullDay {
			out = append(out, a)
			continue
		}
		end := a.EndsAt.In(loc)
		hit := false
		for _, t := range ss.slots {
			slotAt := time.Date(start.Year(), start.Month(), start.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			if !slotAt.Before(start) && slotAt.Before(end) {
				hit = true
				break
			}
		}
		if hit {
			out = append(out, a)
		}
	}
	return out, nil
}
