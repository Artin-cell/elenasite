package repository

import (
	"context"
	"database/sql"
	"errors"

	"elena-backend/internal/models"

	"github.com/google/uuid"
)

// GetActiveForAppointment возвращает самую свежую действующую ссылку отмены записи.
// Если действующей ссылки нет, возвращает nil, nil.
func (r *CancellationTokenRepo) GetActiveForAppointment(ctx context.Context, appointmentID uuid.UUID) (*models.CancellationToken, error) {
	var t models.CancellationToken
	err := r.db.GetContext(ctx, &t, `
		select * from cancellation_tokens
		where appointment_id = $1 and used_at is null and expires_at > now()
		order by created_at desc
		limit 1
	`, appointmentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
