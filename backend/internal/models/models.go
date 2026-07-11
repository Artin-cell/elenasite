package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"` 
	Patronym  string    `db:"patronym"`   
	Phone     string    `db:"phone"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ServiceFormat string

const (
	FormatOnline  ServiceFormat = "online"
	FormatOffline ServiceFormat = "offline"
	FormatBoth    ServiceFormat = "both"
)

type Service struct {
	ID          uuid.UUID     `db:"id"`
	Title       string        `db:"title"`
	Description string        `db:"description"`
	Format      ServiceFormat `db:"format"`
	PriceKopeks int64         `db:"price_kopeks"` 
	DurationMin int           `db:"duration_min"` 
	IsDemo      bool          `db:"is_demo"`
	IsActive    bool          `db:"is_active"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
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
	ID        uuid.UUID         `db:"id"`
	ClientID  uuid.UUID         `db:"client_id"`
	ServiceID uuid.UUID         `db:"service_id"`
	Format    AppointmentFormat `db:"format"`
	StartsAt  time.Time         `db:"starts_at"`
	EndsAt    time.Time         `db:"ends_at"`
	Status    AppointmentStatus `db:"status"`
	Notes     string            `db:"notes"` 

	PaymentMode   PaymentMode `db:"payment_mode"`   
	PaymentStatus string      `db:"payment_status"` 
	PaymentID     string      `db:"payment_id"`     
	AmountKopeks  int64       `db:"amount_kopeks"`  

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Client  *Client  `db:"-"`
	Service *Service `db:"-"`
}

type CancellationToken struct {
	ID            uuid.UUID  `db:"id"`
	AppointmentID uuid.UUID  `db:"appointment_id"`
	Token         string     `db:"token"`      
	ExpiresAt     time.Time  `db:"expires_at"` 
	UsedAt        *time.Time `db:"used_at"`
	CreatedAt     time.Time  `db:"created_at"`
}

type Admin struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"` 
	CreatedAt    time.Time `db:"created_at"`
}

type News struct {
	ID          uuid.UUID  `db:"id"`
	Title       string     `db:"title"`
	Body        string     `db:"body"`
	PublishedAt *time.Time `db:"published_at"` 
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Review struct {
	ID         uuid.UUID `db:"id"`
	AuthorName string    `db:"author_name"`
	Body       string    `db:"body"`
	IsVisible  bool      `db:"is_visible"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
