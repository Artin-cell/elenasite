package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"elena-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AdminActionRepo struct {
	db *sqlx.DB
}

func NewAdminActionRepo(db *sqlx.DB) *AdminActionRepo {
	return &AdminActionRepo{db: db}
}

// Create добавляет строку журнала. Имя администратора подставляется из admins по admin_id.
// detailsJSON — готовый JSON-объект.
func (r *AdminActionRepo) Create(ctx context.Context, adminID *uuid.UUID, action, entity string, entityID *uuid.UUID, detailsJSON string) error {
	_, err := r.db.ExecContext(ctx, `
		insert into admin_actions (admin_id, admin_username, action, entity, entity_id, details)
		values ($1, coalesce((select username from admins where id = $1), ''), $2, $3, $4, $5::jsonb)
	`, adminID, action, entity, entityID, detailsJSON)
	return err
}

type ListAdminActionsOpts struct {
	Entity   string
	EntityID *uuid.UUID
	Limit    int
	Offset   int
}

// List возвращает действия, новые сначала, и общее число подходящих строк.
func (r *AdminActionRepo) List(ctx context.Context, opts ListAdminActionsOpts) ([]models.AdminAction, int, error) {
	var where []string
	var args []interface{}
	if opts.Entity != "" {
		args = append(args, opts.Entity)
		where = append(where, fmt.Sprintf("entity = $%d", len(args)))
	}
	if opts.EntityID != nil {
		args = append(args, *opts.EntityID)
		where = append(where, fmt.Sprintf("entity_id = $%d", len(args)))
	}
	cond := ""
	if len(where) > 0 {
		cond = "where " + strings.Join(where, " and ")
	}

	var total int
	if err := r.db.GetContext(ctx, &total, "select count(*) from admin_actions "+cond, args...); err != nil {
		return nil, 0, err
	}

	args = append(args, opts.Limit, opts.Offset)
	q := fmt.Sprintf(`
		select id, admin_id, admin_username, action, entity, entity_id, details::text as details, created_at
		from admin_actions %s
		order by created_at desc
		limit $%d offset $%d
	`, cond, len(args)-1, len(args))
	items := []models.AdminAction{}
	if err := r.db.SelectContext(ctx, &items, q, args...); err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].Details = json.RawMessage(items[i].DetailsText)
	}
	return items, total, nil
}
