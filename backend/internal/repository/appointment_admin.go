package repository

import (
	"context"

	"github.com/google/uuid"
)

// MarkRefunded переводит оплату записи в статус «возвращена».
// Возвращает true, только если именно этот вызов сменил статус с «оплачена»:
// так письмо клиенту уходит ровно один раз, а вебхук ЮKassa увидит, что статус уже сменён.
func (r *AppointmentRepo) MarkRefunded(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`update appointments set payment_status = 'refunded', updated_at = now()
		 where id = $1 and payment_status = 'paid'`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// MarkPaidManually отмечает оплаченной подтверждённую запись без онлайн-платежа
// (наличные, перевод). Записи с платежом в ЮKassa так отметить нельзя.
func (r *AppointmentRepo) MarkPaidManually(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`update appointments set payment_status = 'paid', updated_at = now()
		 where id = $1 and status = 'confirmed' and payment_status = 'pending' and payment_id = ''`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
