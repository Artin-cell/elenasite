package repository

import (
	"context"

	"github.com/google/uuid"
)

// HasAppointments сообщает, есть ли записи (в любом статусе) на услугу.
func (r *ServiceRepo) HasAppointments(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`select exists(select 1 from appointments where service_id = $1)`, id)
	return exists, err
}

// DeleteIfUnused удаляет услугу, только если на неё нет записей.
// Возвращает false, если удалять было нечего (услуги нет или появились записи).
func (r *ServiceRepo) DeleteIfUnused(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		delete from services s
		where s.id = $1
		  and not exists (select 1 from appointments a where a.service_id = s.id)
	`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
