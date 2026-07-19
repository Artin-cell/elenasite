package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID `db:"id" json:"id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	Patronym  string    `db:"patronym" json:"patronym"`
	Phone     string    `db:"phone" json:"phone"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ServiceFormat string

const (
	FormatOnline  ServiceFormat = "online"
	FormatOffline ServiceFormat = "offline"
	FormatBoth    ServiceFormat = "both"
)

type Service struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	Title       string        `db:"title" json:"title"`
	Description string        `db:"description" json:"description"`
	Format      ServiceFormat `db:"format" json:"format"`
	PriceKopeks int64         `db:"price_kopeks" json:"price_kopeks"`
	DurationMin int           `db:"duration_min" json:"duration_min"`
	IsDemo      bool          `db:"is_demo" json:"is_demo"`
	IsActive    bool          `db:"is_active" json:"is_active"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusCompleted AppointmentStatus = "completed"
	StatusNoShow    AppointmentStatus = "no_show"
)

type AppointmentFormat string

const (
	AppointmentOnline  AppointmentFormat = "online"
	AppointmentOffline AppointmentFormat = "offline"
)

type PaymentMode string

const (
	PaymentModeFull     PaymentMode = "full"
	PaymentModePrepay50 PaymentMode = "prepay_50"
)

const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

type Appointment struct {
	ID        uuid.UUID         `db:"id" json:"id"`
	ClientID  uuid.UUID         `db:"client_id" json:"client_id"`
	ServiceID uuid.UUID         `db:"service_id" json:"service_id"`
	Format    AppointmentFormat `db:"format" json:"format"`
	StartsAt  time.Time         `db:"starts_at" json:"starts_at"`
	EndsAt    time.Time         `db:"ends_at" json:"ends_at"`
	Status    AppointmentStatus `db:"status" json:"status"`
	Notes     string            `db:"notes" json:"notes"`

	PaymentMode   PaymentMode `db:"payment_mode" json:"payment_mode"`
	PaymentStatus string      `db:"payment_status" json:"payment_status"`
	PaymentID     string      `db:"payment_id" json:"payment_id"`
	AmountKopeks  int64       `db:"amount_kopeks" json:"amount_kopeks"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Client  *Client  `db:"-" json:"client,omitempty"`
	Service *Service `db:"-" json:"service,omitempty"`
}

type CancellationToken struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	AppointmentID uuid.UUID  `db:"appointment_id" json:"appointment_id"`
	Token         string     `db:"token" json:"token"`
	ExpiresAt     time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt        *time.Time `db:"used_at" json:"used_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

type Admin struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type News struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Body        string     `db:"body" json:"body"`
	PublishedAt *time.Time `db:"published_at" json:"published_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type Review struct {
	ID         uuid.UUID `db:"id" json:"id"`
	AuthorName string    `db:"author_name" json:"author_name"`
	Body       string    `db:"body" json:"body"`
	IsVisible  bool      `db:"is_visible" json:"is_visible"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}