package repository

import (
	"context"
	"time"
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
