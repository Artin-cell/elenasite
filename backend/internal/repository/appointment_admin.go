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

// SetRefund сохраняет ID и статус возврата в ЮKassa (pending или succeeded).
func (r *AppointmentRepo) SetRefund(ctx context.Context, id uuid.UUID, refundID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`update appointments set refund_id = $1, refund_status = $2, updated_at = now() where id = $3`,
		refundID, status, id)
	return err
}

// MarkRefundSucceeded отмечает, что возврат дошёл (вебхук refund.succeeded).
func (r *AppointmentRepo) MarkRefundSucceeded(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`update appointments set refund_status = 'succeeded', updated_at = now()
		 where id = $1 and refund_status <> 'succeeded'`, id)
	return err
}
