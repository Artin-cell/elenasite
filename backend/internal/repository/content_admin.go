package repository

import (
	"context"

	"elena-backend/internal/models"

	"github.com/google/uuid"
)

// GetByID возвращает новость; если её нет — sql.ErrNoRows.
func (r *NewsRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.News, error) {
	var n models.News
	if err := r.db.GetContext(ctx, &n, `select * from news where id = $1`, id); err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteByID удаляет новость и сообщает, была ли она.
func (r *NewsRepo) DeleteByID(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `delete from news where id = $1`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

// GetByID возвращает отзыв; если его нет — sql.ErrNoRows.
func (r *ReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	var rev models.Review
	if err := r.db.GetContext(ctx, &rev, `select * from reviews where id = $1`, id); err != nil {
		return nil, err
	}
	return &rev, nil
}

// DeleteByID удаляет отзыв и сообщает, был ли он.
func (r *ReviewRepo) DeleteByID(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, `delete from reviews where id = $1`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}
