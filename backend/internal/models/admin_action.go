package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AdminAction — строка журнала действий администратора (таблица admin_actions).
type AdminAction struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	AdminID       *uuid.UUID      `db:"admin_id" json:"admin_id"`
	AdminUsername string          `db:"admin_username" json:"admin_username"`
	Action        string          `db:"action" json:"action"`
	Entity        string          `db:"entity" json:"entity"`
	EntityID      *uuid.UUID      `db:"entity_id" json:"entity_id"`
	DetailsText   string          `db:"details" json:"-"`
	Details       json.RawMessage `db:"-" json:"details"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}
