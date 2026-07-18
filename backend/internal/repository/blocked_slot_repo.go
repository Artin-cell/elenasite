package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BlockedSlot struct {
	ID          uuid.UUID `db:"id" json:"id"`
	BlockedDate time.Time `db:"blocked_date" json:"blocked_date"`
	SlotTime    *string   `db:"slot_time" json:"slot_time"`
	Reason      string    `db:"reason" json:"reason"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type BlockedSlotRepo struct {
	db *sqlx.DB
}

func NewBlockedSlotRepo(db *sqlx.DB) *BlockedSlotRepo {
	return &BlockedSlotRepo{db: db}
}

func (r *BlockedSlotRepo) Create(ctx context.Context, date time.Time, slotTime *string, reason string) (*BlockedSlot, error) {
	var bs BlockedSlot
	err := r.db.GetContext(ctx, &bs, `
		insert into blocked_slots (id, blocked_date, slot_time, reason)
		values (uuid_generate_v4(), $1, $2, $3)
		returning id, blocked_date, slot_time, reason, created_at
	`, date, slotTime, reason)
	if err != nil {
		return nil, fmt.Errorf("create blocked slot: %w", err)
	}
	return &bs, nil
}

func (r *BlockedSlotRepo) ListRange(ctx context.Context, from, to time.Time) ([]BlockedSlot, error) {
	var list []BlockedSlot
	err := r.db.SelectContext(ctx, &list, `
		select id, blocked_date, slot_time, reason, created_at
		from blocked_slots
		where blocked_date >= $1::date and blocked_date < $2::date
		order by blocked_date, slot_time nulls first
	`, from, to)
	if err != nil {
		return nil, fmt.Errorf("list blocked slots: %w", err)
	}
	return list, nil
}

func (r *BlockedSlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `delete from blocked_slots where id = $1`, id)
	return err
}